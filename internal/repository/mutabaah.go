package repository

import (
	"context"
	"errors"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type MutabaahRepository interface {
	GetTodayLog(ctx context.Context, tx Transaction, employeeID uint, date string) (*dto.MutabaahLogResponse, error)
	GetAllLogs(ctx context.Context, tx Transaction, params dto.MutabaahListParams) ([]dto.MutabaahLogResponse, error)
	CreateLog(ctx context.Context, tx Transaction, m model.MutabaahLog) (model.MutabaahLog, error)
	UpdateLog(ctx context.Context, tx Transaction, id uint, updates map[string]interface{}) error
	GetEmployeesWithAttendanceWithoutMutabaah(ctx context.Context, tx Transaction, date string) ([]struct {
		EmployeeID      uint `db:"employee_id"`
		AttendanceLogID uint `db:"attendance_log_id"`
	}, error)
	GetByID(ctx context.Context, tx Transaction, id uint) (*dto.MutabaahLogResponse, error)
	BulkCreateMissingLogs(ctx context.Context, tx Transaction, logs []model.MutabaahLog) error
	GetDailyReport(ctx context.Context, tx Transaction, date string) ([]dto.MutabaahDailyReport, error)
	GetMonthlyReport(ctx context.Context, tx Transaction, month, year int) ([]dto.MutabaahMonthlySummary, error)
	GetCategoryReport(ctx context.Context, tx Transaction, date string) ([]dto.MutabaahCategorySummary, error)
}

type mutabaahRepository struct {
	db *gorm.DB
}

func NewMutabaahRepository(db *gorm.DB) MutabaahRepository {
	return &mutabaahRepository{db: db}
}

func (r *mutabaahRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *mutabaahRepository) GetTodayLog(ctx context.Context, tx Transaction, employeeID uint, date string) (*dto.MutabaahLogResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var log dto.MutabaahLogResponse
	err = db.Raw(`
		SELECT
			ml.id,
			ml.attendance_log_id,
			ml.employee_id,
			e.full_name    AS employee_name,
			ml.log_date::TEXT AS log_date,
			CASE 
				WHEN e.is_trainer THEN 10
				ELSE 5
			END AS target_pages,
			ml.is_submitted,
			ml.submitted_at,
			ml.is_auto_generated,
			ml.created_at,
			ml.updated_at
		FROM mutabaah_logs ml
		JOIN employees e ON e.id = ml.employee_id
		WHERE ml.employee_id = ? AND ml.log_date = ?::DATE AND ml.deleted_at IS NULL
	`, employeeID, date).Scan(&log).Error
	if err != nil {
		return nil, err
	}
	if log.ID == 0 {
		return nil, nil
	}
	return &log, nil
}

func (r *mutabaahRepository) GetAllLogs(ctx context.Context, tx Transaction, params dto.MutabaahListParams) ([]dto.MutabaahLogResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			ml.id,
			ml.employee_id,
			e.full_name    AS employee_name,
			ml.log_date::TEXT AS log_date,
			CASE 
				WHEN e.is_trainer THEN 10
				ELSE 5
			END AS target_pages,
			ml.is_submitted,
			ml.submitted_at,
			ml.is_auto_generated,
			ml.created_at,
			ml.updated_at
		FROM mutabaah_logs ml
		JOIN employees e ON e.id = ml.employee_id
		WHERE ml.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.EmployeeID != nil {
		query += " AND ml.employee_id = ?"
		args = append(args, *params.EmployeeID)
	}
	if params.StartDate != nil {
		query += " AND ml.log_date >= ?"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		query += " AND ml.log_date <= ?"
		args = append(args, *params.EndDate)
	}
	if params.IsSubmitted != nil {
		query += " AND ml.is_submitted = ?"
		args = append(args, *params.IsSubmitted)
	}
	query += " ORDER BY ml.log_date DESC"

	var logs []dto.MutabaahLogResponse
	if err := db.Raw(query, args...).Scan(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *mutabaahRepository) CreateLog(ctx context.Context, tx Transaction, m model.MutabaahLog) (model.MutabaahLog, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.MutabaahLog{}, err
	}
	if err := db.Create(&m).Error; err != nil {
		return model.MutabaahLog{}, err
	}
	return m, nil
}

func (r *mutabaahRepository) UpdateLog(ctx context.Context, tx Transaction, id uint, updates map[string]interface{}) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.MutabaahLog{}).Where("id = ?", id).Updates(updates).Error
}

func (r *mutabaahRepository) GetEmployeesWithAttendanceWithoutMutabaah(ctx context.Context, tx Transaction, date string) ([]struct {
	EmployeeID      uint `db:"employee_id"`
	AttendanceLogID uint `db:"attendance_log_id"`
}, error,
) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var rows []struct {
		EmployeeID      uint `db:"employee_id"`
		AttendanceLogID uint `db:"attendance_log_id"`
	}

	// Hanya untuk pegawai yang present/late di hari ini tapi belum ada mutabaah log
	err = db.Raw(`
		SELECT al.employee_id, al.id AS attendance_log_id
		FROM attendance_logs al
		WHERE al.attendance_date = ?::DATE
		  AND al.status IN ('present', 'late')
		  AND al.deleted_at IS NULL
		  AND NOT EXISTS (
			  SELECT 1 FROM mutabaah_logs ml
			  WHERE ml.employee_id = al.employee_id
			    AND ml.log_date = ?::DATE
			    AND ml.deleted_at IS NULL
		  )
	`, date, date).Scan(&rows).Error
	return rows, err
}

func (r *mutabaahRepository) GetByID(ctx context.Context, tx Transaction, id uint) (*dto.MutabaahLogResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var log dto.MutabaahLogResponse
	err = db.Raw(`
		SELECT
			ml.id,
			ml.employee_id,
			e.full_name    AS employee_name,
			ml.log_date::TEXT AS log_date,
			CASE 
				WHEN e.is_trainer THEN 10
				ELSE 5
			END AS target_pages,
			ml.is_submitted,
			ml.submitted_at,
			ml.is_auto_generated,
			ml.created_at,
			ml.updated_at
		FROM mutabaah_logs ml
		JOIN employees e ON e.id = ml.employee_id
		WHERE ml.id = ? AND ml.deleted_at IS NULL
	`, id).Scan(&log).Error
	if err != nil {
		return nil, err
	}
	if log.ID == 0 {
		return nil, nil
	}
	return &log, nil
}

func (r *mutabaahRepository) BulkCreateMissingLogs(ctx context.Context, tx Transaction, logs []model.MutabaahLog) error {
	if len(logs) == 0 {
		return nil
	}
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Create(&logs).Error
}

func (r *mutabaahRepository) GetDailyReport(ctx context.Context, tx Transaction, date string) ([]dto.MutabaahDailyReport, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var results []dto.MutabaahDailyReport
	query := `
		SELECT
			e.id AS employee_id,
			e.full_name AS employee_name,
			e.employee_number,
			d.name AS department_name,
			CASE 
				WHEN e.is_trainer THEN 10
				ELSE 5
			END AS target_pages,
			COALESCE(ml.is_submitted, false) AS is_submitted,
			ml.submitted_at::TEXT AS submitted_at
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		INNER JOIN employee_schedules es
			ON es.employee_id = e.id
			AND es.is_active = TRUE
			AND es.effective_date <= ?::DATE
			AND (es.end_date IS NULL OR es.end_date >= ?::DATE)
			AND es.deleted_at IS NULL
		INNER JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		INNER JOIN shift_template_details std
			ON std.shift_template_id = st.id
			AND std.day_of_week = LOWER(TRIM(TO_CHAR(?::DATE, 'Day')))::day_of_week_enum
			AND std.is_working_day = TRUE
			AND std.deleted_at IS NULL
		LEFT JOIN mutabaah_logs ml ON ml.employee_id = e.id AND ml.log_date = ?::DATE AND ml.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
		ORDER BY e.full_name ASC
	`
	err = db.Raw(query, date, date, date, date).Scan(&results).Error
	return results, err
}

func (r *mutabaahRepository) GetMonthlyReport(ctx context.Context, tx Transaction, month, year int) ([]dto.MutabaahMonthlySummary, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var results []dto.MutabaahMonthlySummary
	query := `
		WITH month_dates AS (
			SELECT generate_series(
				DATE_TRUNC('month', MAKE_DATE(?, ?, 1)),
				(DATE_TRUNC('month', MAKE_DATE(?, ?, 1)) + INTERVAL '1 month - 1 day')::DATE,
				'1 day'::INTERVAL
			)::DATE AS dt
		),
		employee_working_days AS (
			SELECT 
				e.id AS employee_id,
				COUNT(md.dt) AS total_working_days
			FROM employees e
			INNER JOIN employee_schedules es 
				ON es.employee_id = e.id AND es.is_active = TRUE AND es.deleted_at IS NULL
			INNER JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
			INNER JOIN shift_template_details std 
				ON std.shift_template_id = st.id AND std.is_working_day = TRUE AND std.deleted_at IS NULL
			CROSS JOIN month_dates md
			WHERE e.deleted_at IS NULL
				AND es.effective_date <= md.dt
				AND (es.end_date IS NULL OR es.end_date >= md.dt)
				AND std.day_of_week = LOWER(TRIM(TO_CHAR(md.dt, 'Day')))::day_of_week_enum
			GROUP BY e.id
		)
		SELECT
			e.id AS employee_id,
			e.full_name AS employee_name,
			e.is_trainer,
			COALESCE(ewd.total_working_days, 0) AS total_working_days,
			COUNT(CASE WHEN ml.is_submitted = true THEN 1 END) AS total_submitted,
			CASE 
				WHEN COALESCE(ewd.total_working_days, 0) > 0 
				THEN (COUNT(CASE WHEN ml.is_submitted = true THEN 1 END)::FLOAT / ewd.total_working_days) * 100
				ELSE 0 
			END AS compliance_percentage
		FROM employees e
		LEFT JOIN employee_working_days ewd ON ewd.employee_id = e.id
		LEFT JOIN attendance_logs al ON al.employee_id = e.id 
			AND EXTRACT(MONTH FROM al.attendance_date) = ? 
			AND EXTRACT(YEAR FROM al.attendance_date) = ?
			AND al.deleted_at IS NULL
		LEFT JOIN mutabaah_logs ml ON ml.attendance_log_id = al.id AND ml.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
		GROUP BY e.id, e.full_name, e.is_trainer, ewd.total_working_days
		ORDER BY compliance_percentage DESC, e.full_name ASC
	`
	err = db.Raw(query, year, month, year, month, month, year).Scan(&results).Error
	return results, err
}

func (r *mutabaahRepository) GetCategoryReport(ctx context.Context, tx Transaction, date string) ([]dto.MutabaahCategorySummary, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var results []dto.MutabaahCategorySummary
	query := `
		SELECT
			CASE WHEN e.is_trainer THEN 'trainer' ELSE 'non_trainer' END AS category,
			COUNT(e.id) AS total_employees,
			COUNT(CASE WHEN ml.is_submitted = true THEN 1 END) AS total_submitted_today,
			COUNT(e.id) - COUNT(CASE WHEN ml.is_submitted = true THEN 1 END) AS total_not_submitted_today,
			CASE 
				WHEN COUNT(e.id) > 0 
				THEN (COUNT(CASE WHEN ml.is_submitted = true THEN 1 END)::FLOAT / COUNT(e.id)) * 100
				ELSE 0
			END AS average_compliance
		FROM employees e
		INNER JOIN employee_schedules es
			ON es.employee_id = e.id
			AND es.is_active = TRUE
			AND es.effective_date <= ?::DATE
			AND (es.end_date IS NULL OR es.end_date >= ?::DATE)
			AND es.deleted_at IS NULL
		INNER JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		INNER JOIN shift_template_details std
			ON std.shift_template_id = st.id
			AND std.day_of_week = LOWER(TRIM(TO_CHAR(?::DATE, 'Day')))::day_of_week_enum
			AND std.is_working_day = TRUE
			AND std.deleted_at IS NULL
		LEFT JOIN mutabaah_logs ml ON ml.employee_id = e.id AND ml.log_date = ?::DATE AND ml.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
		GROUP BY e.is_trainer
	`
	err = db.Raw(query, date, date, date, date).Scan(&results).Error
	return results, err
}
