package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	logger "hris-backend/config/log"
	"hris-backend/config/storage"
	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils/data"
)

type LeaveService interface {
	GetMetadata(ctx context.Context) (dto.LeaveMetadata, error)
	GetAllBalances(ctx context.Context, params dto.LeaveBalanceListParams) ([]dto.LeaveBalanceResponse, error)
	GetAllRequests(ctx context.Context, params dto.LeaveRequestListParams) ([]dto.LeaveRequestResponse, error)
	GetRequestByID(ctx context.Context, id uint) (dto.LeaveRequestResponse, error)
	CreateRequest(ctx context.Context, employeeID uint, roleLevel string, req dto.CreateLeaveRequest) (dto.LeaveRequestResponse, error)
	ApproveRequest(ctx context.Context, approverID uint, requestID uint, req dto.ApproveLeaveRequest) (dto.LeaveRequestResponse, error)
	RejectRequest(ctx context.Context, approverID uint, requestID uint, req dto.RejectLeaveRequest) (dto.LeaveRequestResponse, error)
}

type leaveService struct {
	repo       repository.LeaveRepository
	attendRepo repository.AttendanceRepository
	txManager  repository.TxManager
	minio      storage.MinioClient
}

func NewLeaveService(
	repo repository.LeaveRepository,
	attendRepo repository.AttendanceRepository,
	txManager repository.TxManager,
	minio storage.MinioClient,
) LeaveService {
	return &leaveService{
		repo:       repo,
		attendRepo: attendRepo,
		txManager:  txManager,
		minio:      minio,
	}
}

func (s *leaveService) GetMetadata(ctx context.Context) (dto.LeaveMetadata, error) {
	leaveTypeMeta, err := s.repo.GetLeaveTypeMeta(ctx, nil)
	if err != nil {
		return dto.LeaveMetadata{}, err
	}
	empMeta, err := s.repo.GetEmployeeMetaList(ctx, nil)
	if err != nil {
		return dto.LeaveMetadata{}, err
	}

	return dto.LeaveMetadata{
		LeaveTypeMeta: leaveTypeMeta,
		StatusMeta:    data.LeaveRequestStatusMeta,
		EmployeeMeta:  empMeta,
	}, nil
}

func (s *leaveService) GetAllBalances(ctx context.Context, params dto.LeaveBalanceListParams) ([]dto.LeaveBalanceResponse, error) {
	return s.repo.GetAllBalances(ctx, nil, params)
}

func (s *leaveService) GetAllRequests(ctx context.Context, params dto.LeaveRequestListParams) ([]dto.LeaveRequestResponse, error) {
	reqs, err := s.repo.GetAllRequests(ctx, nil, params)
	if err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].DocumentURL != nil && *reqs[i].DocumentURL != "" && !strings.HasPrefix(*reqs[i].DocumentURL, "http") {
			url, err := s.minio.PresignedGetObject(ctx, storage.BucketLeaveDocuments, *reqs[i].DocumentURL, storage.PresignedDownloadExpiry)
			if err == nil {
				reqs[i].DocumentURL = &url
			} else {
				logger.Warn("presign leave document failed", map[string]any{
					"request_id": reqs[i].ID,
					"key":        *reqs[i].DocumentURL,
					"error":      err.Error(),
				})
			}
		}
	}
	return reqs, nil
}

func (s *leaveService) GetRequestByID(ctx context.Context, id uint) (dto.LeaveRequestResponse, error) {
	res, err := s.repo.GetRequestByID(ctx, nil, id)
	if err != nil {
		return dto.LeaveRequestResponse{}, err
	}
	if res.DocumentURL != nil && *res.DocumentURL != "" && !strings.HasPrefix(*res.DocumentURL, "http") {
		url, err := s.minio.PresignedGetObject(ctx, storage.BucketLeaveDocuments, *res.DocumentURL, storage.PresignedDownloadExpiry)
		if err == nil {
			res.DocumentURL = &url
		} else {
			logger.Warn("presign leave document failed", map[string]any{
				"request_id": res.ID,
				"key":        *res.DocumentURL,
				"error":      err.Error(),
			})
		}
	}
	return *res, nil
}

func (s *leaveService) CreateRequest(ctx context.Context, employeeID uint, roleLevel string, req dto.CreateLeaveRequest) (dto.LeaveRequestResponse, error) {
	targetEmployeeID := employeeID
	isAdminSubmission := false

	if req.EmployeeID != nil && *req.EmployeeID != employeeID {
		if roleLevel != string(model.RoleLevelSuperAdmin) && roleLevel != string(model.RoleLevelAdmin) {
			return dto.LeaveRequestResponse{}, fmt.Errorf("unauthorized: only admin/superadmin can submit for other employees")
		}
		targetEmployeeID = *req.EmployeeID
		isAdminSubmission = true
	}

	overlap, err := s.repo.CheckOverlap(ctx, nil, targetEmployeeID, req.StartDate, req.EndDate, nil)
	if err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("check overlap: %w", err)
	}
	if overlap {
		return dto.LeaveRequestResponse{}, fmt.Errorf("you have another leave request overlapping with these dates")
	}

	totalHours := 0
	if req.TotalHours != nil {
		totalHours = *req.TotalHours
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Ensure balance exists
	currYear := time.Now().Year()
	bal, err := s.repo.GetBalanceByEmployeeAndType(ctx, tx, targetEmployeeID, req.LeaveTypeID, currYear)
	if err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("check balance: %w", err)
	}
	if bal == nil {
		newBal := model.LeaveBalance{
			EmployeeID:  targetEmployeeID,
			LeaveTypeID: req.LeaveTypeID,
			Year:        currYear,
		}
		if _, e := s.repo.CreateBalance(ctx, tx, newBal); e != nil {
			return dto.LeaveRequestResponse{}, fmt.Errorf("create balance: %w", e)
		}
	} else {
		if bal.MaxDuration != nil && bal.UsedDuration+req.TotalDays > *bal.MaxDuration {
			return dto.LeaveRequestResponse{}, fmt.Errorf("leave duration exceeds max duration")
		}
		if bal.MaxOccurrences != nil && bal.UsedOccurrences+1 > *bal.MaxOccurrences {
			return dto.LeaveRequestResponse{}, fmt.Errorf("leave occurrences exceeds max occurrences")
		}
	}

	sDate, _ := time.Parse("2006-01-02", req.StartDate)
	eDate, _ := time.Parse("2006-01-02", req.EndDate)

	status := "pending"
	apprLevel1Status := "pending"
	apprLevel2Status := "pending"
	var decidedAt *time.Time
	var apprBy *uint
	now := time.Now()

	if isAdminSubmission {
		status = "approved_hr"
		apprLevel1Status = "approved"
		apprLevel2Status = "approved"
		decidedAt = &now
		apprBy = &employeeID
	}

	m := model.LeaveRequest{
		EmployeeID:  targetEmployeeID,
		LeaveTypeID: req.LeaveTypeID,
		StartDate:   sDate,
		EndDate:     eDate,
		TotalDays:   req.TotalDays,
		TotalHours:  &totalHours,
		Reason:      req.Reason,
		DocumentURL: req.DocumentURL,
		Status:      model.LeaveRequestStatusEnum(status),
	}

	created, err := s.repo.CreateRequest(ctx, tx, m)
	if err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("create leave request: %w", err)
	}

	// create approvals
	_, err = s.repo.CreateApproval(ctx, tx, model.LeaveRequestApproval{
		LeaveRequestID: created.ID,
		Level:          1,
		Status:         model.ApprovalStatusEnum(apprLevel1Status),
		ApproverID:     apprBy,
		DecidedAt:      decidedAt,
	})
	if err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("create approval: %w", err)
	}
	_, err = s.repo.CreateApproval(ctx, tx, model.LeaveRequestApproval{
		LeaveRequestID: created.ID,
		Level:          2,
		Status:         model.ApprovalStatusEnum(apprLevel2Status),
		ApproverID:     apprBy,
		DecidedAt:      decidedAt,
	})
	if err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("create approval: %w", err)
	}

	if isAdminSubmission {
		balRefresher, _ := s.repo.GetBalanceByEmployeeAndType(ctx, tx, targetEmployeeID, req.LeaveTypeID, currYear)
		if balRefresher != nil {
			err = s.repo.UpdateBalanceUsage(ctx, tx, balRefresher.ID, balRefresher.UsedOccurrences+1, balRefresher.UsedDuration+req.TotalDays)
			if err != nil {
				return dto.LeaveRequestResponse{}, fmt.Errorf("update balance: %w", err)
			}
		}

		for d := sDate; !d.After(eDate); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			existingLog, _ := s.attendRepo.GetTodayLog(ctx, tx, targetEmployeeID, dateStr)
			if existingLog != nil {
				upd := map[string]interface{}{
					"status":           string(model.AttendanceLeave),
					"leave_request_id": created.ID,
				}
				s.attendRepo.UpdateLog(ctx, tx, existingLog.ID, upd)
			} else {
				cm := model.ClockMethodManual
				newLog := model.AttendanceLog{
					EmployeeID:      targetEmployeeID,
					AttendanceDate:  d,
					Status:          model.AttendanceLeave,
					LeaveRequestID:  &created.ID,
					ClockInMethod:   &cm,
					ClockOutMethod:  &cm,
					IsAutoGenerated: true,
				}
				s.attendRepo.CreateLog(ctx, tx, newLog)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetRequestByID(ctx, created.ID)
}

func (s *leaveService) ApproveRequest(ctx context.Context, approverID uint, requestID uint, req dto.ApproveLeaveRequest) (dto.LeaveRequestResponse, error) {
	request, err := s.repo.GetRequestByID(ctx, nil, requestID)
	if err != nil {
		return dto.LeaveRequestResponse{}, err
	}
	if request.Status == "approved_hr" || request.Status == "rejected" {
		return dto.LeaveRequestResponse{}, fmt.Errorf("request is already %s", request.Status)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.LeaveRequestResponse{}, err
	}
	defer tx.Rollback()

	// determine which level is pending
	level1, _ := s.repo.GetPendingApprovalForLevel(ctx, tx, requestID, 1)
	level2, _ := s.repo.GetPendingApprovalForLevel(ctx, tx, requestID, 2)

	var targetAppr *dto.LeaveApprovalResponse
	targetStatus := ""

	if level1 != nil {
		targetAppr = level1
		targetStatus = "approved_leader"
	} else if level2 != nil {
		targetAppr = level2
		targetStatus = "approved_hr"
	} else {
		return dto.LeaveRequestResponse{}, fmt.Errorf("no pending approval found")
	}

	// Update approval record
	if err := s.repo.UpdateApprovalStatus(ctx, tx, targetAppr.ID, "approved", approverID, req.Notes); err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("update approval: %w", err)
	}

	// Update main request status
	if err := s.repo.UpdateRequestStatus(ctx, tx, requestID, targetStatus); err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("update main status: %w", err)
	}

	// If it's final approval (HR)
	if targetStatus == "approved_hr" {
		currYear := time.Now().Year() // assuming it starts from current request year realistically
		sDate, _ := time.Parse("2006-01-02", request.StartDate)
		currYear = sDate.Year()
		bal, _ := s.repo.GetBalanceByEmployeeAndType(ctx, tx, request.EmployeeID, request.LeaveTypeID, currYear)
		if bal != nil {
			err = s.repo.UpdateBalanceUsage(ctx, tx, bal.ID, bal.UsedOccurrences+1, bal.UsedDuration+request.TotalDays)
			if err != nil {
				return dto.LeaveRequestResponse{}, fmt.Errorf("update balance: %w", err)
			}
		}

		// update attendance logs (mock iteration for daily coverage within dates)
		eDate, _ := time.Parse("2006-01-02", request.EndDate)
		for d := sDate; !d.After(eDate); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			existingLog, _ := s.attendRepo.GetTodayLog(ctx, tx, request.EmployeeID, dateStr)
			if existingLog != nil {
				upd := map[string]interface{}{
					"status":           string(model.AttendanceLeave),
					"leave_request_id": requestID,
				}
				s.attendRepo.UpdateLog(ctx, tx, existingLog.ID, upd)
			} else {
				cm := model.ClockMethodManual
				newLog := model.AttendanceLog{
					EmployeeID:      request.EmployeeID,
					AttendanceDate:  d,
					Status:          model.AttendanceLeave,
					LeaveRequestID:  &requestID,
					ClockInMethod:   &cm,
					ClockOutMethod:  &cm,
					IsAutoGenerated: true,
				}
				s.attendRepo.CreateLog(ctx, tx, newLog)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.LeaveRequestResponse{}, err
	}

	return s.GetRequestByID(ctx, requestID)
}

func (s *leaveService) RejectRequest(ctx context.Context, approverID uint, requestID uint, req dto.RejectLeaveRequest) (dto.LeaveRequestResponse, error) {
	_, err := s.repo.GetRequestByID(ctx, nil, requestID)
	if err != nil {
		return dto.LeaveRequestResponse{}, err
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.LeaveRequestResponse{}, err
	}
	defer tx.Rollback()

	// find any pending level
	level1, _ := s.repo.GetPendingApprovalForLevel(ctx, tx, requestID, 1)
	level2, _ := s.repo.GetPendingApprovalForLevel(ctx, tx, requestID, 2)

	var targetAppr *dto.LeaveApprovalResponse
	if level1 != nil {
		targetAppr = level1
	} else if level2 != nil {
		targetAppr = level2
	} else {
		return dto.LeaveRequestResponse{}, fmt.Errorf("no pending approval found")
	}

	// Update approval record to rejected
	if err := s.repo.UpdateApprovalStatus(ctx, tx, targetAppr.ID, "rejected", approverID, &req.Notes); err != nil {
		return dto.LeaveRequestResponse{}, err
	}

	// Reject main request
	if err := s.repo.UpdateRequestStatus(ctx, tx, requestID, "rejected"); err != nil {
		return dto.LeaveRequestResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return dto.LeaveRequestResponse{}, err
	}

	return s.GetRequestByID(ctx, requestID)
}
