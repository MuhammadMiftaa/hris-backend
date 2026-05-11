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
	GetAllBalances(ctx context.Context, params dto.LeaveBalanceListParams) (dto.PaginatedResponse[dto.LeaveBalanceResponse], error)
	GetAllRequests(ctx context.Context, roleLevel string, params dto.LeaveRequestListParams) (dto.PaginatedResponse[dto.LeaveRequestResponse], error)
	GetRequestByID(ctx context.Context, id uint) (dto.LeaveRequestResponse, error)
	CreateRequest(ctx context.Context, employeeID uint, roleLevel string, req dto.CreateLeaveRequest) (dto.LeaveRequestResponse, error)
	ApproveRequest(ctx context.Context, approverID uint, requestID uint, req dto.ApproveLeaveRequest) (dto.LeaveRequestResponse, error)
	RejectRequest(ctx context.Context, approverID uint, requestID uint, req dto.RejectLeaveRequest) (dto.LeaveRequestResponse, error)

	// Balance management
	GetEmployeeBalanceSummary(ctx context.Context, params dto.EmployeeBalanceSummaryParams) (dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse], error)
	GetEmployeeBalanceDetail(ctx context.Context, employeeID uint, year int) (dto.EmployeeBalanceDetailResponse, error)
	UpsertBalance(ctx context.Context, hrID uint, req dto.UpsertLeaveBalanceRequest) (dto.LeaveBalanceResponse, error)
	DeleteBalance(ctx context.Context, id uint) error
	AdjustBalance(ctx context.Context, hrID uint, balanceID uint, req dto.AdjustLeaveBalanceRequest) (dto.LeaveBalanceResponse, error)
	GetBalanceAdjustments(ctx context.Context, balanceID uint) ([]dto.LeaveBalanceAdjustmentResponse, error)
	ExportEmployeeBalanceSummary(ctx context.Context, params dto.EmployeeBalanceSummaryParams) (dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse], error)
}

type leaveService struct {
	repo          repository.LeaveRepository
	leaveTypeRepo repository.LeaveTypeRepository
	attendRepo    repository.AttendanceRepository
	txManager     repository.TxManager
	minio         storage.MinioClient
	notifSvc      NotificationService
}

func NewLeaveService(
	repo repository.LeaveRepository,
	leaveTypeRepo repository.LeaveTypeRepository,
	attendRepo repository.AttendanceRepository,
	txManager repository.TxManager,
	minio storage.MinioClient,
	notifSvc NotificationService,
) LeaveService {
	return &leaveService{
		repo:          repo,
		leaveTypeRepo: leaveTypeRepo,
		attendRepo:    attendRepo,
		txManager:     txManager,
		minio:         minio,
		notifSvc:      notifSvc,
	}
}

func (s *leaveService) GetMetadata(ctx context.Context) (dto.LeaveMetadata, error) {
	leaveTypeMeta, err := s.repo.GetLeaveTypeMeta(ctx, nil)
	if err != nil {
		return dto.LeaveMetadata{}, err
	}
	parentLeaveTypeMeta, err := s.repo.GetParentLeaveTypeMeta(ctx, nil)
	if err != nil {
		return dto.LeaveMetadata{}, err
	}
	empMeta, err := s.repo.GetEmployeeMetaList(ctx, nil)
	if err != nil {
		return dto.LeaveMetadata{}, err
	}
	deptMeta, err := s.repo.GetDepartmentMetaList(ctx, nil)
	if err != nil {
		return dto.LeaveMetadata{}, err
	}

	return dto.LeaveMetadata{
		LeaveTypeMeta:       leaveTypeMeta,
		ParentLeaveTypeMeta: parentLeaveTypeMeta,
		StatusMeta:          data.LeaveRequestStatusMeta,
		EmployeeMeta:        empMeta,
		DepartmentMeta:      deptMeta,
	}, nil
}

func (s *leaveService) GetAllBalances(ctx context.Context, params dto.LeaveBalanceListParams) (dto.PaginatedResponse[dto.LeaveBalanceResponse], error) {
	return s.repo.GetAllBalances(ctx, nil, params)
}

func (s *leaveService) GetAllRequests(ctx context.Context, roleLevel string, params dto.LeaveRequestListParams) (dto.PaginatedResponse[dto.LeaveRequestResponse], error) {
	reqs, err := s.repo.GetAllRequests(ctx, nil, params)
	if err != nil {
		return dto.PaginatedResponse[dto.LeaveRequestResponse]{}, err
	}
	for i := range reqs.Data {
		if roleLevel == string(model.RoleLevelAdmin) || roleLevel == string(model.RoleLevelSuperAdmin) {
			if reqs.Data[i].DocumentURL != nil && *reqs.Data[i].DocumentURL != "" && !strings.HasPrefix(*reqs.Data[i].DocumentURL, "http") {
				url, err := s.minio.PresignedGetObject(ctx, storage.BucketLeaveDocuments, *reqs.Data[i].DocumentURL, storage.PresignedDownloadExpiry)
				if err == nil {
					reqs.Data[i].DocumentURL = &url
				} else {
					logger.Warn("presign leave document failed", map[string]any{
						"request_id": reqs.Data[i].ID,
						"key":        *reqs.Data[i].DocumentURL,
						"error":      err.Error(),
					})
				}
			}
		} else {
			reqs.Data[i].DocumentURL = nil
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

	// Fetch Leave Type
	lt, err := s.leaveTypeRepo.GetLeaveTypeByID(ctx, fmt.Sprint(req.LeaveTypeID))
	if err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("get leave type: %w", err)
	}

	balanceTypeID := req.LeaveTypeID
	if lt.ParentLeaveTypeID != nil {
		balanceTypeID = *lt.ParentLeaveTypeID
	}

	deductDays := lt.DeductDays

	// Ensure balance exists
	currYear := time.Now().Year()
	bal, err := s.repo.GetBalanceByEmployeeAndType(ctx, tx, targetEmployeeID, balanceTypeID, currYear)
	if err != nil {
		return dto.LeaveRequestResponse{}, fmt.Errorf("check balance: %w", err)
	}
	if bal == nil {
		newBal := model.LeaveBalance{
			EmployeeID:  targetEmployeeID,
			LeaveTypeID: balanceTypeID,
			Year:        currYear,
		}
		if _, e := s.repo.CreateBalance(ctx, tx, newBal); e != nil {
			return dto.LeaveRequestResponse{}, fmt.Errorf("create balance: %w", e)
		}
	} else {
		remaining := bal.AllocatedDuration + bal.TotalAdjustment - bal.UsedDuration
		if remaining < deductDays {
			return dto.LeaveRequestResponse{}, fmt.Errorf("saldo cuti tidak mencukupi (sisa: %.1f, dibutuhkan: %.1f)", remaining, deductDays)
		}
		if bal.MaxDuration != nil && bal.UsedDuration+deductDays > *bal.MaxDuration {
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
		balRefresher, _ := s.repo.GetBalanceByEmployeeAndType(ctx, tx, targetEmployeeID, balanceTypeID, currYear)
		if balRefresher != nil {
			err = s.repo.UpdateBalanceUsage(ctx, tx, balRefresher.ID, balRefresher.UsedOccurrences+1, balRefresher.UsedDuration+deductDays)
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

	// Trigger approval notification (fire-and-forget)
	if !isAdminSubmission {
		if err := s.notifSvc.TriggerRequestApprovalNotification(ctx, "leave", created.ID, targetEmployeeID); err != nil {
			logger.Error("failed to trigger leave approval notification", map[string]any{
				"request_id": created.ID,
				"error":      err.Error(),
			})
		}
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
		
		lt, err := s.leaveTypeRepo.GetLeaveTypeByID(ctx, fmt.Sprint(request.LeaveTypeID))
		if err != nil {
			return dto.LeaveRequestResponse{}, fmt.Errorf("get leave type for approval: %w", err)
		}
		balanceTypeID := request.LeaveTypeID
		deductDays := request.TotalDays
		if lt.ID != 0 {
			if lt.ParentLeaveTypeID != nil {
				balanceTypeID = *lt.ParentLeaveTypeID
			}
			deductDays = lt.DeductDays
		}

		bal, _ := s.repo.GetBalanceByEmployeeAndType(ctx, tx, request.EmployeeID, balanceTypeID, currYear)
		if bal != nil {
			err = s.repo.UpdateBalanceUsage(ctx, tx, bal.ID, bal.UsedOccurrences+1, bal.UsedDuration+deductDays)
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

	// Trigger result notification to requester
	if request != nil {
		if err := s.notifSvc.TriggerApprovalResultNotification(ctx, "leave", requestID, request.EmployeeID, targetStatus); err != nil {
			logger.Error("failed to trigger leave approval result notification", map[string]any{
				"request_id": requestID,
				"error":      err.Error(),
			})
		}
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

	// Trigger rejection notification to requester
	request, _ := s.repo.GetRequestByID(ctx, nil, requestID)
	if request != nil {
		if err := s.notifSvc.TriggerApprovalResultNotification(ctx, "leave", requestID, request.EmployeeID, "rejected"); err != nil {
			logger.Error("failed to trigger leave rejection notification", map[string]any{
				"request_id": requestID,
				"error":      err.Error(),
			})
		}
	}

	return s.GetRequestByID(ctx, requestID)
}

func (s *leaveService) GetEmployeeBalanceSummary(ctx context.Context, params dto.EmployeeBalanceSummaryParams) (dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse], error) {
	return s.repo.GetEmployeeBalanceSummary(ctx, nil, params)
}

func (s *leaveService) GetEmployeeBalanceDetail(ctx context.Context, employeeID uint, year int) (dto.EmployeeBalanceDetailResponse, error) {
	balances, err := s.repo.GetBalanceDetailByEmployee(ctx, nil, employeeID, year)
	if err != nil {
		return dto.EmployeeBalanceDetailResponse{}, err
	}

	name := ""
	if len(balances) > 0 && balances[0].EmployeeName != nil {
		name = *balances[0].EmployeeName
	}

	return dto.EmployeeBalanceDetailResponse{
		EmployeeID:   employeeID,
		EmployeeName: name,
		Year:         year,
		Balances:     balances,
	}, nil
}

func (s *leaveService) UpsertBalance(ctx context.Context, hrID uint, req dto.UpsertLeaveBalanceRequest) (dto.LeaveBalanceResponse, error) {
	if _, err := time.Parse("2006-01-02", req.EffectiveDate); err != nil {
		return dto.LeaveBalanceResponse{}, fmt.Errorf("effective_date invalid: %w", err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.LeaveBalanceResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	m, err := s.repo.UpsertBalance(ctx, tx, req)
	if err != nil {
		return dto.LeaveBalanceResponse{}, fmt.Errorf("upsert balance: %w", err)
	}

	updated, err := s.repo.GetBalanceByID(ctx, tx, m.ID)
	if err != nil {
		return dto.LeaveBalanceResponse{}, err
	}
	if updated == nil {
		return dto.LeaveBalanceResponse{}, fmt.Errorf("balance not found after upsert")
	}

	if err := tx.Commit(); err != nil {
		return dto.LeaveBalanceResponse{}, fmt.Errorf("commit tx: %w", err)
	}
	return *updated, nil
}

func (s *leaveService) DeleteBalance(ctx context.Context, id uint) error {
	return s.repo.DeleteBalance(ctx, nil, id)
}

func (s *leaveService) AdjustBalance(ctx context.Context, hrID uint, balanceID uint, req dto.AdjustLeaveBalanceRequest) (dto.LeaveBalanceResponse, error) {
	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.LeaveBalanceResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	adj := model.LeaveBalanceAdjustment{
		LeaveBalanceID: balanceID,
		AdjustedBy:     hrID,
		Delta:          req.Delta,
		Reason:         req.Reason,
	}
	if _, err := s.repo.CreateAdjustment(ctx, tx, adj); err != nil {
		return dto.LeaveBalanceResponse{}, err
	}

	updated, err := s.repo.GetBalanceByID(ctx, tx, balanceID)
	if err != nil {
		return dto.LeaveBalanceResponse{}, err
	}
	if updated == nil {
		return dto.LeaveBalanceResponse{}, fmt.Errorf("balance not found")
	}

	if err := tx.Commit(); err != nil {
		return dto.LeaveBalanceResponse{}, fmt.Errorf("commit tx: %w", err)
	}
	return *updated, nil
}

func (s *leaveService) GetBalanceAdjustments(ctx context.Context, balanceID uint) ([]dto.LeaveBalanceAdjustmentResponse, error) {
	return s.repo.GetAdjustmentsByBalanceID(ctx, nil, balanceID)
}

func (s *leaveService) ExportEmployeeBalanceSummary(ctx context.Context, params dto.EmployeeBalanceSummaryParams) (dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse], error) {
	allPerPage := -1
	params.PerPage = &allPerPage
	return s.repo.GetEmployeeBalanceSummary(ctx, nil, params)
}
