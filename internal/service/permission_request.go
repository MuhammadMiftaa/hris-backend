package service

import (
	"context"
	"fmt"
	"time"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils/data"
)

type PermissionRequestService interface {
	GetMetadata(ctx context.Context) (dto.RequestMetadata, error)
	GetAll(ctx context.Context, params dto.PermissionListParams) ([]dto.PermissionRequestResponse, error)
	GetByID(ctx context.Context, id uint) (dto.PermissionRequestResponse, error)
	Create(ctx context.Context, employeeID uint, roleLevel string, req dto.CreatePermissionRequest) (dto.PermissionRequestResponse, error)
	UpdateStatus(ctx context.Context, employeeID uint, id uint, req dto.UpdatePermissionStatusRequest) (dto.PermissionRequestResponse, error)
	Delete(ctx context.Context, id uint) error
}

type permissionRequestService struct {
	repo       repository.PermissionRequestRepository
	attendRepo repository.AttendanceRepository
	txManager  repository.TxManager
}

func NewPermissionRequestService(
	repo repository.PermissionRequestRepository,
	attendRepo repository.AttendanceRepository,
	txManager repository.TxManager,
) PermissionRequestService {
	return &permissionRequestService{
		repo:       repo,
		attendRepo: attendRepo,
		txManager:  txManager,
	}
}

func (s *permissionRequestService) GetMetadata(ctx context.Context) (dto.RequestMetadata, error) {
	empMeta, err := s.repo.GetEmployeeMetaList(ctx, nil)
	if err != nil {
		return dto.RequestMetadata{}, err
	}
	return dto.RequestMetadata{
		PermissionTypeMeta: data.PermissionTypeMeta,
		WorkLocationMeta:   data.WorkLocationMeta,
		StatusMeta:         data.LeaveRequestStatusMeta,
		EmployeeMeta:       empMeta,
	}, nil
}

func (s *permissionRequestService) GetAll(ctx context.Context, params dto.PermissionListParams) ([]dto.PermissionRequestResponse, error) {
	return s.repo.GetAll(ctx, nil, params)
}

func (s *permissionRequestService) GetByID(ctx context.Context, id uint) (dto.PermissionRequestResponse, error) {
	res, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return dto.PermissionRequestResponse{}, err
	}
	return *res, nil
}

func (s *permissionRequestService) Create(ctx context.Context, employeeID uint, roleLevel string, req dto.CreatePermissionRequest) (dto.PermissionRequestResponse, error) {
	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return dto.PermissionRequestResponse{}, fmt.Errorf("invalid date format: %w", err)
	}

	targetEmployeeID := employeeID
	isAdminSubmission := false

	if req.EmployeeID != nil && *req.EmployeeID != employeeID {
		if roleLevel != string(model.RoleLevelSuperAdmin) && roleLevel != string(model.RoleLevelAdmin) {
			return dto.PermissionRequestResponse{}, fmt.Errorf("unauthorized: only admin/superadmin can submit for other employees")
		}
		targetEmployeeID = *req.EmployeeID
		isAdminSubmission = true
	}

	status := model.RequestStatusPending
	var approvedBy *uint

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.PermissionRequestResponse{}, err
	}
	defer tx.Rollback()

	if isAdminSubmission {
		status = model.RequestStatusEnum("approved")
		approvedBy = &employeeID
	}

	m := model.PermissionRequest{
		EmployeeID:     targetEmployeeID,
		Date:           parsedDate,
		PermissionType: model.PermissionTypeEnum(req.PermissionType),
		LeaveTime:      req.LeaveTime,
		ReturnTime:     req.ReturnTime,
		Reason:         req.Reason,
		DocumentURL:    req.DocumentURL,
		Status:         status,
		ApprovedBy:     approvedBy,
	}

	created, err := s.repo.Create(ctx, tx, m)
	if err != nil {
		return dto.PermissionRequestResponse{}, err
	}

	if isAdminSubmission {
		log, _ := s.attendRepo.GetTodayLog(ctx, tx, created.EmployeeID, parsedDate.Format("2006-01-02"))
		if log != nil {
			upd := map[string]interface{}{}
			upd["permission_request_id"] = created.ID

			if created.PermissionType == "late_arrival" && log.Status == string(model.AttendanceLate) {
				upd["status"] = model.AttendancePresent
				upd["late_minutes"] = 0
			} else if created.PermissionType == "early_leave" && log.Status == string(model.AttendanceHalfDay) {
				upd["status"] = model.AttendancePresent
				upd["early_leave_minutes"] = 0
			}
			s.attendRepo.UpdateLog(ctx, tx, log.ID, upd)
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.PermissionRequestResponse{}, err
	}

	return s.GetByID(ctx, created.ID)
}

func (s *permissionRequestService) UpdateStatus(ctx context.Context, employeeID uint, id uint, req dto.UpdatePermissionStatusRequest) (dto.PermissionRequestResponse, error) {
	perm, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return dto.PermissionRequestResponse{}, err
	}
	if perm.Status != string(model.RequestStatusPending) {
		return dto.PermissionRequestResponse{}, fmt.Errorf("permission request is no longer pending")
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.PermissionRequestResponse{}, err
	}
	defer tx.Rollback()

	if err := s.repo.UpdateStatus(ctx, tx, id, req.Status, employeeID, req.Notes); err != nil {
		return dto.PermissionRequestResponse{}, err
	}

	if req.Status == "approved" {
		log, _ := s.attendRepo.GetTodayLog(ctx, tx, perm.EmployeeID, perm.Date)
		if log != nil {
			upd := map[string]interface{}{}
			upd["permission_request_id"] = id

			if perm.PermissionType == "late_arrival" && log.Status == string(model.AttendanceLate) {
				upd["status"] = model.AttendancePresent
				upd["late_minutes"] = 0
			} else if perm.PermissionType == "early_leave" && log.Status == string(model.AttendanceHalfDay) {
				upd["status"] = model.AttendancePresent
				upd["early_leave_minutes"] = 0
			}
			s.attendRepo.UpdateLog(ctx, tx, log.ID, upd)
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.PermissionRequestResponse{}, err
	}

	return s.GetByID(ctx, id)
}

func (s *permissionRequestService) Delete(ctx context.Context, id uint) error {
	perm, err := s.repo.GetByID(ctx, nil, id)
	if err != nil {
		return err
	}
	if perm.Status != string(model.RequestStatusPending) {
		return fmt.Errorf("cannot delete processed permission request")
	}
	return s.repo.Delete(ctx, nil, id)
}
