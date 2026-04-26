package service

import (
	"context"
	"fmt"
	"time"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"
)

type OvertimeService interface {
	GetAll(ctx context.Context, params dto.OvertimeListParams) ([]dto.OvertimeRequestResponse, error)
	GetByID(ctx context.Context, id uint) (dto.OvertimeRequestResponse, error)
	Create(ctx context.Context, employeeID uint, roleLevel string, req dto.CreateOvertimeRequest) (dto.OvertimeRequestResponse, error)
	ApproveRequest(ctx context.Context, approverID uint, requestID uint, req dto.ApproveOvertimeRequest) (dto.OvertimeRequestResponse, error)
	RejectRequest(ctx context.Context, approverID uint, requestID uint, req dto.RejectOvertimeRequest) (dto.OvertimeRequestResponse, error)
	Delete(ctx context.Context, id uint) error
}

type overtimeService struct {
	repo       repository.OvertimeRepository
	attendRepo repository.AttendanceRepository
	txManager  repository.TxManager
}

func NewOvertimeService(
	repo repository.OvertimeRepository,
	attendRepo repository.AttendanceRepository,
	txManager repository.TxManager,
) OvertimeService {
	return &overtimeService{
		repo:       repo,
		attendRepo: attendRepo,
		txManager:  txManager,
	}
}

func (s *overtimeService) GetAll(ctx context.Context, params dto.OvertimeListParams) ([]dto.OvertimeRequestResponse, error) {
	return s.repo.GetAll(ctx, nil, params)
}

func (s *overtimeService) GetByID(ctx context.Context, id uint) (dto.OvertimeRequestResponse, error) {
	res, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}
	return *res, nil
}

func (s *overtimeService) Create(ctx context.Context, employeeID uint, roleLevel string, req dto.CreateOvertimeRequest) (dto.OvertimeRequestResponse, error) {
	d, err := time.Parse("2006-01-02", req.OvertimeDate)
	if err != nil {
		return dto.OvertimeRequestResponse{}, fmt.Errorf("invalid date format")
	}

	targetEmployeeID := employeeID
	isAdminSubmission := false

	if req.EmployeeID != nil && *req.EmployeeID != employeeID {
		if roleLevel != string(model.RoleLevelSuperAdmin) && roleLevel != string(model.RoleLevelAdmin) {
			return dto.OvertimeRequestResponse{}, fmt.Errorf("unauthorized: only admin/superadmin can submit for other employees")
		}
		targetEmployeeID = *req.EmployeeID
		isAdminSubmission = true
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}
	defer tx.Rollback()

	initialStatus := "pending"
	if isAdminSubmission {
		initialStatus = "approved_hr"
	}

	var plannedStart, plannedEnd time.Time
	if req.PlannedStart != nil {
		plannedStart, err = utils.ParseTimeString(*req.PlannedStart, "")
		if err != nil {
			return dto.OvertimeRequestResponse{}, fmt.Errorf("invalid planned start time format")
		}
	}
	if req.PlannedEnd != nil {
		plannedEnd, err = utils.ParseTimeString(*req.PlannedEnd, "")
		if err != nil {
			return dto.OvertimeRequestResponse{}, fmt.Errorf("invalid planned end time format")
		}
	}

	m := model.OvertimeRequest{
		EmployeeID:       targetEmployeeID,
		OvertimeDate:     d,
		PlannedMinutes:   req.PlannedMinutes,
		PlannedStart:     &plannedStart,
		PlannedEnd:       &plannedEnd,
		WorkLocationType: (*model.WorkLocationTypeEnum)(&req.WorkLocationType),

		Reason: req.Reason,
		Status: model.LeaveRequestStatusEnum(initialStatus),
	}

	created, err := s.repo.Create(ctx, tx, m)
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	// Create approvals
	apprLeaderStatus := "pending"
	apprHRStatus := "pending"

	var decidedAt *time.Time
	var apprBy *uint
	now := time.Now()

	if isAdminSubmission {
		apprLeaderStatus = "approved"
		apprHRStatus = "approved"
		decidedAt = &now
		apprBy = &employeeID
	}

	_, err = s.repo.CreateApproval(ctx, tx, model.OvertimeRequestApproval{
		OvertimeRequestID: created.ID,
		Level:             1,
		Status:            model.ApprovalStatusEnum(apprLeaderStatus),
		ApproverID:        apprBy,
		DecidedAt:         decidedAt,
	})
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	_, err = s.repo.CreateApproval(ctx, tx, model.OvertimeRequestApproval{
		OvertimeRequestID: created.ID,
		Level:             2,
		Status:            model.ApprovalStatusEnum(apprHRStatus),
		ApproverID:        apprBy,
		DecidedAt:         decidedAt,
	})
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	if isAdminSubmission {
		// Asosiasikan dengan log attendance (AttendanceRepository LinkOvertimeToLog)
		log, _ := s.attendRepo.GetTodayLog(ctx, tx, targetEmployeeID, created.OvertimeDate.Format("2006-01-02"))
		if log != nil {
			_ = s.attendRepo.LinkOvertimeToLog(ctx, tx, targetEmployeeID, created.OvertimeDate.Format("2006-01-02"), log.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	return s.GetByID(ctx, created.ID)
}

func (s *overtimeService) ApproveRequest(ctx context.Context, approverID uint, requestID uint, req dto.ApproveOvertimeRequest) (dto.OvertimeRequestResponse, error) {
	reqData, err := s.repo.GetByID(ctx, nil, requestID)
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}
	if reqData.Status == "approved_hr" || reqData.Status == "rejected" {
		return dto.OvertimeRequestResponse{}, fmt.Errorf("request is fully approved or rejected")
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}
	defer tx.Rollback()

	// Find the lowest pending approval
	pendingLvl1, _ := s.repo.GetPendingApprovalForLevel(ctx, tx, requestID, 1)
	var targetApprovalID uint
	newMainStatus := ""
	isFullyApproved := false

	if pendingLvl1 != nil {
		targetApprovalID = pendingLvl1.ID
		newMainStatus = "approved_leader"
	} else {
		pendingLvl2, _ := s.repo.GetPendingApprovalForLevel(ctx, tx, requestID, 2)
		if pendingLvl2 != nil {
			targetApprovalID = pendingLvl2.ID
			newMainStatus = "approved_hr"
			isFullyApproved = true
		} else {
			return dto.OvertimeRequestResponse{}, fmt.Errorf("no pending approval found")
		}
	}

	if err := s.repo.UpdateApprovalStatus(ctx, tx, targetApprovalID, "approved", approverID, req.Notes); err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	if err := s.repo.UpdateRequestStatus(ctx, tx, requestID, newMainStatus); err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	if isFullyApproved {
		// Asosiasikan dengan log attendance
		log, _ := s.attendRepo.GetTodayLog(ctx, tx, reqData.EmployeeID, reqData.OvertimeDate)
		if log != nil {
			_ = s.attendRepo.LinkOvertimeToLog(ctx, tx, reqData.EmployeeID, reqData.OvertimeDate, log.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	return s.GetByID(ctx, requestID)
}

func (s *overtimeService) RejectRequest(ctx context.Context, approverID uint, requestID uint, req dto.RejectOvertimeRequest) (dto.OvertimeRequestResponse, error) {
	reqData, err := s.repo.GetByID(ctx, nil, requestID)
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}
	if reqData.Status == "approved_hr" || reqData.Status == "rejected" {
		return dto.OvertimeRequestResponse{}, fmt.Errorf("request is fully approved or rejected")
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.OvertimeRequestResponse{}, err
	}
	defer tx.Rollback()

	// Find the lowest pending approval
	pendingLvl1, _ := s.repo.GetPendingApprovalForLevel(ctx, tx, requestID, 1)
	var targetApprovalID uint

	if pendingLvl1 != nil {
		targetApprovalID = pendingLvl1.ID
	} else {
		pendingLvl2, _ := s.repo.GetPendingApprovalForLevel(ctx, tx, requestID, 2)
		if pendingLvl2 != nil {
			targetApprovalID = pendingLvl2.ID
		} else {
			return dto.OvertimeRequestResponse{}, fmt.Errorf("no pending approval found")
		}
	}

	notes := req.Notes
	if err := s.repo.UpdateApprovalStatus(ctx, tx, targetApprovalID, "rejected", approverID, &notes); err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	if err := s.repo.UpdateRequestStatus(ctx, tx, requestID, "rejected"); err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return dto.OvertimeRequestResponse{}, err
	}

	return s.GetByID(ctx, requestID)
}

func (s *overtimeService) Delete(ctx context.Context, id uint) error {
	reqData, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return err
	}
	if reqData.Status != "pending" && reqData.Status != "approved_leader" {
		return fmt.Errorf("cannot delete processed overtime request")
	}
	return s.repo.Delete(ctx, nil, id)
}
