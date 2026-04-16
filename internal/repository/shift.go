package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type ShiftRepository interface {
	// Template
	GetAllShiftTemplates(ctx context.Context, tx Transaction) ([]dto.ShiftTemplateResponse, error)
	GetShiftTemplateByID(ctx context.Context, tx Transaction, id uint) (dto.ShiftTemplateResponse, error)
	CreateShiftTemplate(ctx context.Context, tx Transaction, m model.ShiftTemplate) (model.ShiftTemplate, error)
	UpdateShiftTemplate(ctx context.Context, tx Transaction, id uint, m model.ShiftTemplate) (model.ShiftTemplate, error)
	DeleteShiftTemplate(ctx context.Context, tx Transaction, id uint) error

	// Detail
	GetDetailsByTemplateID(ctx context.Context, tx Transaction, shiftID uint) ([]dto.ShiftTemplateDetailResp, error)
	DeleteDetailsByTemplateID(ctx context.Context, tx Transaction, shiftID uint) error
	CreateDetails(ctx context.Context, tx Transaction, details []model.ShiftTemplateDetail) error

	// Schedule
	GetAllSchedules(ctx context.Context, tx Transaction, params *dto.ScheduleListParams) ([]dto.ScheduleResponse, error)
	GetScheduleByID(ctx context.Context, tx Transaction, id uint) (dto.ScheduleResponse, error)
	CreateSchedule(ctx context.Context, tx Transaction, m model.EmployeeSchedule) (model.EmployeeSchedule, error)
	UpdateSchedule(ctx context.Context, tx Transaction, id uint, m model.EmployeeSchedule) (model.EmployeeSchedule, error)
	DeleteSchedule(ctx context.Context, tx Transaction, id uint) error
}

type shiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

// ── Template ──────────────────────────────────────────

func (r *shiftRepository) GetAllShiftTemplates(ctx context.Context, tx Transaction) ([]dto.ShiftTemplateResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var templates []struct {
		ID         uint       `db:"id"`
		Name       string     `db:"name"`
		IsFlexible bool       `db:"is_flexible"`
		CreatedAt  interface{} `db:"created_at"`
		UpdatedAt  *interface{} `db:"updated_at"`
		DeletedAt  *interface{} `db:"deleted_at"`
	}

	rows, err := db.Raw(`
		SELECT id, name, is_flexible, created_at, updated_at, deleted_at
		FROM shift_templates
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan templates
	var rawTemplates []dto.ShiftTemplateResponse
	type rawTemplate struct {
		ID         uint
		Name       string
		IsFlexible bool
		CreatedAt  interface{}
		UpdatedAt  *interface{}
		DeletedAt  *interface{}
	}
	_ = templates

	if err := db.Raw(`
		SELECT id, name, is_flexible, created_at, updated_at, deleted_at
		FROM shift_templates
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`).Scan(&rawTemplates).Error; err != nil {
		return nil, err
	}

	// For each template, fetch details
	for i, tmpl := range rawTemplates {
		details, err := r.GetDetailsByTemplateID(ctx, tx, tmpl.ID)
		if err != nil {
			return nil, fmt.Errorf("get details for template %d: %w", tmpl.ID, err)
		}
		rawTemplates[i].Details = details
	}

	return rawTemplates, nil
}

func (r *shiftRepository) GetShiftTemplateByID(ctx context.Context, tx Transaction, id uint) (dto.ShiftTemplateResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.ShiftTemplateResponse{}, err
	}

	var tmpl dto.ShiftTemplateResponse
	if err := db.Raw(`
		SELECT id, name, is_flexible, created_at, updated_at, deleted_at
		FROM shift_templates
		WHERE deleted_at IS NULL AND id = ?
	`, id).Scan(&tmpl).Error; err != nil {
		return dto.ShiftTemplateResponse{}, err
	}
	if tmpl.ID == 0 {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("shift template not found")
	}

	details, err := r.GetDetailsByTemplateID(ctx, tx, id)
	if err != nil {
		return dto.ShiftTemplateResponse{}, fmt.Errorf("get details: %w", err)
	}
	tmpl.Details = details

	return tmpl, nil
}

func (r *shiftRepository) CreateShiftTemplate(ctx context.Context, tx Transaction, m model.ShiftTemplate) (model.ShiftTemplate, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.ShiftTemplate{}, err
	}

	if err := db.Create(&m).Error; err != nil {
		return model.ShiftTemplate{}, err
	}
	return m, nil
}

func (r *shiftRepository) UpdateShiftTemplate(ctx context.Context, tx Transaction, id uint, m model.ShiftTemplate) (model.ShiftTemplate, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.ShiftTemplate{}, err
	}

	if err := db.Model(&m).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        m.Name,
		"is_flexible": m.IsFlexible,
	}).Error; err != nil {
		return model.ShiftTemplate{}, err
	}
	m.ID = id
	return m, nil
}

func (r *shiftRepository) DeleteShiftTemplate(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	if err := db.Where("id = ?", id).Delete(&model.ShiftTemplate{}).Error; err != nil {
		return err
	}
	return nil
}

// ── Detail ────────────────────────────────────────────

func (r *shiftRepository) GetDetailsByTemplateID(ctx context.Context, tx Transaction, shiftID uint) ([]dto.ShiftTemplateDetailResp, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var details []dto.ShiftTemplateDetailResp
	if err := db.Raw(`
		SELECT
			id, shift_template_id, day_of_week::TEXT AS day_of_week,
			is_working_day,
			clock_in_start::TEXT    AS clock_in_start,
			clock_in_end::TEXT      AS clock_in_end,
			break_dhuhr_start::TEXT AS break_dhuhr_start,
			break_dhuhr_end::TEXT   AS break_dhuhr_end,
			break_asr_start::TEXT   AS break_asr_start,
			break_asr_end::TEXT     AS break_asr_end,
			clock_out_start::TEXT   AS clock_out_start,
			clock_out_end::TEXT     AS clock_out_end,
			created_at, updated_at, deleted_at
		FROM shift_template_details
		WHERE deleted_at IS NULL AND shift_template_id = ?
		ORDER BY
			CASE day_of_week::TEXT
				WHEN 'monday'    THEN 1
				WHEN 'tuesday'   THEN 2
				WHEN 'wednesday' THEN 3
				WHEN 'thursday'  THEN 4
				WHEN 'friday'    THEN 5
				WHEN 'saturday'  THEN 6
				WHEN 'sunday'    THEN 7
			END
	`, shiftID).Scan(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

func (r *shiftRepository) DeleteDetailsByTemplateID(ctx context.Context, tx Transaction, shiftID uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	if err := db.Where("shift_template_id = ?", shiftID).Delete(&model.ShiftTemplateDetail{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *shiftRepository) CreateDetails(ctx context.Context, tx Transaction, details []model.ShiftTemplateDetail) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	if len(details) == 0 {
		return nil
	}
	if err := db.Create(&details).Error; err != nil {
		return err
	}
	return nil
}

// ── Schedule ──────────────────────────────────────────

func (r *shiftRepository) GetAllSchedules(ctx context.Context, tx Transaction, params *dto.ScheduleListParams) ([]dto.ScheduleResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			es.id,
			es.employee_id,
			e.full_name        AS employee_name,
			e.employee_number  AS employee_number,
			es.shift_template_id,
			st.name            AS shift_name,
			es.effective_date::TEXT AS effective_date,
			es.end_date::TEXT  AS end_date,
			es.is_active,
			es.created_at, es.updated_at, es.deleted_at
		FROM employee_schedules es
		LEFT JOIN employees       e  ON e.id  = es.employee_id       AND e.deleted_at IS NULL
		LEFT JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		WHERE es.deleted_at IS NULL
	`
	args := []interface{}{}

	if params != nil {
		if params.EmployeeID != nil {
			query += " AND es.employee_id = ?"
			args = append(args, *params.EmployeeID)
		}
		if params.ShiftTemplateID != nil {
			query += " AND es.shift_template_id = ?"
			args = append(args, *params.ShiftTemplateID)
		}
		if params.IsActive != nil {
			query += " AND es.is_active = ?"
			args = append(args, *params.IsActive)
		}
	}
	query += " ORDER BY es.effective_date DESC"

	var schedules []dto.ScheduleResponse
	if err := db.Raw(query, args...).Scan(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *shiftRepository) GetScheduleByID(ctx context.Context, tx Transaction, id uint) (dto.ScheduleResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.ScheduleResponse{}, err
	}

	var schedule dto.ScheduleResponse
	if err := db.Raw(`
		SELECT
			es.id,
			es.employee_id,
			e.full_name        AS employee_name,
			e.employee_number  AS employee_number,
			es.shift_template_id,
			st.name            AS shift_name,
			es.effective_date::TEXT AS effective_date,
			es.end_date::TEXT  AS end_date,
			es.is_active,
			es.created_at, es.updated_at, es.deleted_at
		FROM employee_schedules es
		LEFT JOIN employees       e  ON e.id  = es.employee_id       AND e.deleted_at IS NULL
		LEFT JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		WHERE es.deleted_at IS NULL AND es.id = ?
	`, id).Scan(&schedule).Error; err != nil {
		return dto.ScheduleResponse{}, err
	}
	if schedule.ID == 0 {
		return dto.ScheduleResponse{}, fmt.Errorf("schedule not found")
	}
	return schedule, nil
}

func (r *shiftRepository) CreateSchedule(ctx context.Context, tx Transaction, m model.EmployeeSchedule) (model.EmployeeSchedule, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.EmployeeSchedule{}, err
	}

	if err := db.Create(&m).Error; err != nil {
		return model.EmployeeSchedule{}, err
	}
	return m, nil
}

func (r *shiftRepository) UpdateSchedule(ctx context.Context, tx Transaction, id uint, m model.EmployeeSchedule) (model.EmployeeSchedule, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.EmployeeSchedule{}, err
	}

	updates := map[string]interface{}{
		"employee_id":       m.EmployeeID,
		"shift_template_id": m.ShiftTemplateID,
		"effective_date":    m.EffectiveDate,
		"end_date":          m.EndDate,
		"is_active":         m.IsActive,
	}
	if err := db.Model(&model.EmployeeSchedule{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return model.EmployeeSchedule{}, err
	}
	m.ID = id
	return m, nil
}

func (r *shiftRepository) DeleteSchedule(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	if err := db.Where("id = ?", id).Delete(&model.EmployeeSchedule{}).Error; err != nil {
		return err
	}
	return nil
}
