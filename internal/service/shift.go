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

type ShiftService interface {
	GetMetadata(ctx context.Context) (dto.ShiftMetadata, error)
	// Template CRUD
	GetAllTemplates(ctx context.Context) ([]dto.ShiftTemplateResponse, error)
	GetTemplateByID(ctx context.Context, id uint) (dto.ShiftTemplateResponse, error)
	CreateTemplate(ctx context.Context, req dto.CreateShiftRequest) (dto.ShiftTemplateResponse, error)
	UpdateTemplate(ctx context.Context, id uint, req dto.UpdateShiftRequest) (dto.ShiftTemplateResponse, error)
	DeleteTemplate(ctx context.Context, id uint) error
	// Detail
	GetDetailsByTemplateID(ctx context.Context, shiftID uint) ([]dto.ShiftTemplateDetailResp, error)
	// Schedule CRUD
	GetAllSchedules(ctx context.Context, params *dto.ScheduleListParams) ([]dto.ScheduleResponse, error)
	GetScheduleByID(ctx context.Context, id uint) (dto.ScheduleResponse, error)
	CreateSchedule(ctx context.Context, req dto.CreateScheduleRequest) (dto.ScheduleResponse, error)
	UpdateSchedule(ctx context.Context, id uint, req dto.UpdateScheduleRequest) (dto.ScheduleResponse, error)
	DeleteSchedule(ctx context.Context, id uint) error
}

type shiftService struct {
	repo      repository.ShiftRepository
	txManager repository.TxManager
}

func NewShiftService(repo repository.ShiftRepository, txManager repository.TxManager) ShiftService {
	return &shiftService{repo: repo, txManager: txManager}
}

func (s *shiftService) GetMetadata(ctx context.Context) (dto.ShiftMetadata, error) {
	return dto.ShiftMetadata{
		DayOfWeekMeta: data.DayOfWeekMeta,
	}, nil
}

// ── Template ──────────────────────────────────────────

func (s *shiftService) GetAllTemplates(ctx context.Context) ([]dto.ShiftTemplateResponse, error) {
	templates, err := s.repo.GetAllShiftTemplates(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get all shift templates: %w", err)
	}
	return templates, nil
}

func (s *shiftService) GetTemplateByID(ctx context.Context, id uint) (dto.ShiftTemplateResponse, error) {
	tmpl, err := s.repo.GetShiftTemplateByID(ctx, nil, id)
	if err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("get shift template by ID: %w", err)
	}
	return tmpl, nil
}

func (s *shiftService) CreateTemplate(ctx context.Context, req dto.CreateShiftRequest) (dto.ShiftTemplateResponse, error) {
	if req.Name == "" {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("shift name is required")
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("create shift template: begin transaction: %w", err)
	}
	defer tx.Rollback()

	tmpl, err := s.repo.CreateShiftTemplate(ctx, tx, model.ShiftTemplate{
		Name:       req.Name,
		IsFlexible: req.IsFlexible,
	})
	if err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("create shift template: %w", err)
	}

	if len(req.Details) > 0 {
		details, err := s.buildDetails(tmpl.ID, req.Details)
		if err != nil {
			return dto.ShiftTemplateResponse{}, err
		}
		if err := s.repo.CreateDetails(ctx, tx, details); err != nil {
			return dto.ShiftTemplateResponse{}, fmt.Errorf("create shift details: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("create shift template: commit: %w", err)
	}

	return s.repo.GetShiftTemplateByID(ctx, nil, tmpl.ID)
}

func (s *shiftService) UpdateTemplate(ctx context.Context, id uint, req dto.UpdateShiftRequest) (dto.ShiftTemplateResponse, error) {
	existing, err := s.repo.GetShiftTemplateByID(ctx, nil, id)
	if err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("update shift template: get existing: %w", err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("update shift template: begin transaction: %w", err)
	}
	defer tx.Rollback()

	updateModel := model.ShiftTemplate{
		Name:       existing.Name,
		IsFlexible: existing.IsFlexible,
	}
	if req.Name != nil {
		updateModel.Name = *req.Name
	}
	if req.IsFlexible != nil {
		updateModel.IsFlexible = *req.IsFlexible
	}

	if _, err := s.repo.UpdateShiftTemplate(ctx, tx, id, updateModel); err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("update shift template: %w", err)
	}

	if len(req.Details) > 0 {
		if err := s.repo.DeleteDetailsByTemplateID(ctx, tx, id); err != nil {
			return dto.ShiftTemplateResponse{}, fmt.Errorf("update shift template: delete old details: %w", err)
		}
		details, err := s.buildDetails(id, req.Details)
		if err != nil {
			return dto.ShiftTemplateResponse{}, err
		}
		if err := s.repo.CreateDetails(ctx, tx, details); err != nil {
			return dto.ShiftTemplateResponse{}, fmt.Errorf("update shift template: create new details: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("update shift template: commit: %w", err)
	}

	return s.repo.GetShiftTemplateByID(ctx, nil, id)
}

func (s *shiftService) DeleteTemplate(ctx context.Context, id uint) error {
	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("delete shift template: begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.repo.DeleteDetailsByTemplateID(ctx, tx, id); err != nil {
		return fmt.Errorf("delete shift template: delete details: %w", err)
	}
	if err := s.repo.DeleteShiftTemplate(ctx, tx, id); err != nil {
		return fmt.Errorf("delete shift template: %w", err)
	}

	return tx.Commit()
}

// ── Detail ────────────────────────────────────────────

func (s *shiftService) GetDetailsByTemplateID(ctx context.Context, shiftID uint) ([]dto.ShiftTemplateDetailResp, error) {
	details, err := s.repo.GetDetailsByTemplateID(ctx, nil, shiftID)
	if err != nil {
		return nil, fmt.Errorf("get details by template ID: %w", err)
	}
	return details, nil
}

// ── Schedule ──────────────────────────────────────────

func (s *shiftService) GetAllSchedules(ctx context.Context, params *dto.ScheduleListParams) ([]dto.ScheduleResponse, error) {
	schedules, err := s.repo.GetAllSchedules(ctx, nil, params)
	if err != nil {
		return nil, fmt.Errorf("get all schedules: %w", err)
	}
	return schedules, nil
}

func (s *shiftService) GetScheduleByID(ctx context.Context, id uint) (dto.ScheduleResponse, error) {
	schedule, err := s.repo.GetScheduleByID(ctx, nil, id)
	if err != nil {
		return dto.ScheduleResponse{}, fmt.Errorf("get schedule by ID: %w", err)
	}
	return schedule, nil
}

func (s *shiftService) CreateSchedule(ctx context.Context, req dto.CreateScheduleRequest) (dto.ScheduleResponse, error) {
	if req.EmployeeID == 0 {
		return dto.ScheduleResponse{}, fmt.Errorf("employee_id is required")
	}
	if req.ShiftTemplateID == 0 {
		return dto.ScheduleResponse{}, fmt.Errorf("shift_template_id is required")
	}
	if req.EffectiveDate == "" {
		return dto.ScheduleResponse{}, fmt.Errorf("effective_date is required")
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return dto.ScheduleResponse{}, fmt.Errorf("invalid effective_date format: %w", err)
	}

	scheduleModel := model.EmployeeSchedule{
		EmployeeID:      req.EmployeeID,
		ShiftTemplateID: req.ShiftTemplateID,
		EffectiveDate:   effectiveDate,
		IsActive:        req.IsActive,
	}

	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return dto.ScheduleResponse{}, fmt.Errorf("invalid end_date format: %w", err)
		}
		scheduleModel.EndDate = &endDate
	}

	created, err := s.repo.CreateSchedule(ctx, nil, scheduleModel)
	if err != nil {
		return dto.ScheduleResponse{}, fmt.Errorf("create schedule: %w", err)
	}

	return s.repo.GetScheduleByID(ctx, nil, created.ID)
}

func (s *shiftService) UpdateSchedule(ctx context.Context, id uint, req dto.UpdateScheduleRequest) (dto.ScheduleResponse, error) {
	existing, err := s.repo.GetScheduleByID(ctx, nil, id)
	if err != nil {
		return dto.ScheduleResponse{}, fmt.Errorf("update schedule: get existing: %w", err)
	}

	updateModel := model.EmployeeSchedule{
		EmployeeID:      existing.EmployeeID,
		ShiftTemplateID: existing.ShiftTemplateID,
		IsActive:        existing.IsActive,
	}

	effDate, err := time.Parse("2006-01-02", existing.EffectiveDate)
	if err == nil {
		updateModel.EffectiveDate = effDate
	}

	if req.EmployeeID != nil {
		updateModel.EmployeeID = *req.EmployeeID
	}
	if req.ShiftTemplateID != nil {
		updateModel.ShiftTemplateID = *req.ShiftTemplateID
	}
	if req.IsActive != nil {
		updateModel.IsActive = *req.IsActive
	}
	if req.EffectiveDate != nil {
		d, err := time.Parse("2006-01-02", *req.EffectiveDate)
		if err != nil {
			return dto.ScheduleResponse{}, fmt.Errorf("invalid effective_date format: %w", err)
		}
		updateModel.EffectiveDate = d
	}
	if req.EndDate != nil {
		d, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return dto.ScheduleResponse{}, fmt.Errorf("invalid end_date format: %w", err)
		}
		updateModel.EndDate = &d
	}

	if _, err := s.repo.UpdateSchedule(ctx, nil, id, updateModel); err != nil {
		return dto.ScheduleResponse{}, fmt.Errorf("update schedule: %w", err)
	}

	return s.repo.GetScheduleByID(ctx, nil, id)
}

func (s *shiftService) DeleteSchedule(ctx context.Context, id uint) error {
	if err := s.repo.DeleteSchedule(ctx, nil, id); err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

// ── Helper ────────────────────────────────────────────

func (s *shiftService) buildDetails(templateID uint, reqs []dto.CreateShiftDetailRequest) ([]model.ShiftTemplateDetail, error) {
	details := make([]model.ShiftTemplateDetail, 0, len(reqs))
	for _, d := range reqs {
		if d.DayOfWeek == "" {
			return nil, fmt.Errorf("day_of_week is required for each detail")
		}
		details = append(details, model.ShiftTemplateDetail{
			ShiftTemplateID: templateID,
			DayOfWeek:       model.DayOfWeekEnum(d.DayOfWeek),
			IsWorkingDay:    d.IsWorkingDay,
			ClockInStart:    d.ClockInStart,
			ClockInEnd:      d.ClockInEnd,
			BreakDhuhrStart: d.BreakDhuhrStart,
			BreakDhuhrEnd:   d.BreakDhuhrEnd,
			BreakAsrStart:   d.BreakAsrStart,
			BreakAsrEnd:     d.BreakAsrEnd,
			ClockOutStart:   d.ClockOutStart,
			ClockOutEnd:     d.ClockOutEnd,
		})
	}
	return details, nil
}
