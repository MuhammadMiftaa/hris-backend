package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"

	"gorm.io/gorm"
)

type DailyReportRepository interface {
	GetAll(ctx context.Context, tx Transaction, params dto.DailyReportListParams) (dto.PaginatedResponse[dto.DailyReportResponse], error)
	GetByID(ctx context.Context, tx Transaction, id uint) (*dto.DailyReportResponse, error)
	Create(ctx context.Context, tx Transaction, m model.DailyReport) (model.DailyReport, error)
	Update(ctx context.Context, tx Transaction, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, tx Transaction, id uint) error
	GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetEmployeesWithAttendanceWithoutDailyReport(ctx context.Context, tx Transaction, date string) ([]struct {
		EmployeeID      uint `db:"employee_id"`
		AttendanceLogID uint `db:"attendance_log_id"`
	}, error)
	BulkCreateMissingLogs(ctx context.Context, tx Transaction, logs []model.DailyReport) error
}

type dailyReportRepository struct {
	db *gorm.DB
}

func NewDailyReportRepository(db *gorm.DB) DailyReportRepository {
	return &dailyReportRepository{db: db}
}

func (r *dailyReportRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *dailyReportRepository) GetAll(ctx context.Context, tx Transaction, params dto.DailyReportListParams) (dto.PaginatedResponse[dto.DailyReportResponse], error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.PaginatedResponse[dto.DailyReportResponse]{}, err
	}

	baseQuery := `
		FROM daily_reports d
		JOIN employees e ON e.id = d.employee_id
		WHERE d.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.EmployeeID != nil {
		baseQuery += " AND d.employee_id = ?"
		args = append(args, *params.EmployeeID)
	}
	if params.StartDate != nil {
		baseQuery += " AND d.report_date >= ?::DATE"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		baseQuery += " AND d.report_date <= ?::DATE"
		args = append(args, *params.EndDate)
	}
	if params.IsSubmitted != nil {
		baseQuery += " AND d.is_submitted = ?"
		args = append(args, *params.IsSubmitted)
	}
	if params.Activities != nil && *params.Activities != "" {
		baseQuery += " AND d.activities ILIKE ?"
		like := "%" + *params.Activities + "%"
		args = append(args, like)
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.DailyReportResponse]{}, err
	}

	selectQuery := `
		SELECT
			d.id,
			d.employee_id,
			e.full_name AS employee_name,
			d.attendance_log_id,
			d.report_date::TEXT AS report_date,
			d.activities,
			d.is_submitted,
			d.submitted_at,
			d.is_auto_generated,
			'submitted' AS status,
			d.created_at,
			d.updated_at
	` + baseQuery
	
	selectQuery += utils.BuildSortClause("daily_reports", params.SortBy, params.GetSortDir(), "d.report_date DESC, d.created_at DESC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var res []dto.DailyReportResponse
	if err := db.Raw(selectQuery, args...).Scan(&res).Error; err != nil {
		return dto.PaginatedResponse[dto.DailyReportResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.DailyReportResponse]{
		Data: res,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *dailyReportRepository) GetByID(ctx context.Context, tx Transaction, id uint) (*dto.DailyReportResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			d.id,
			d.employee_id,
			e.full_name AS employee_name,
			d.attendance_log_id,
			d.report_date::TEXT AS report_date,
			d.activities,
			d.is_submitted,
			d.submitted_at,
			d.is_auto_generated,
			'submitted' AS status,
			d.created_at,
			d.updated_at
		FROM daily_reports d
		JOIN employees e ON e.id = d.employee_id
		WHERE d.id = ? AND d.deleted_at IS NULL
	`
	var res dto.DailyReportResponse
	if err := db.Raw(query, id).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, fmt.Errorf("daily report not found")
	}
	return &res, nil
}

func (r *dailyReportRepository) Create(ctx context.Context, tx Transaction, m model.DailyReport) (model.DailyReport, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}

	// Check if already exists for the day
	var count int64
	db.Model(&model.DailyReport{}).Where("employee_id = ? AND report_date = ?", m.EmployeeID, m.ReportDate).Count(&count)
	if count > 0 {
		return m, fmt.Errorf("daily report for this date already exists")
	}

	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *dailyReportRepository) Update(ctx context.Context, tx Transaction, id uint, updates map[string]interface{}) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.DailyReport{}).Where("id = ?", id).Updates(updates).Error
}

func (r *dailyReportRepository) Delete(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Delete(&model.DailyReport{}, id).Error
}

func (r *dailyReportRepository) GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var meta []dto.Meta
	err = db.Raw(`
		SELECT id::TEXT, full_name AS name
		FROM employees
		WHERE deleted_at IS NULL
		ORDER BY full_name ASC
	`).Scan(&meta).Error
	return meta, err
}

func (r *dailyReportRepository) GetEmployeesWithAttendanceWithoutDailyReport(ctx context.Context, tx Transaction, date string) ([]struct {
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

	err = db.Raw(`
		SELECT al.employee_id, al.id AS attendance_log_id
		FROM attendance_logs al
		WHERE al.attendance_date = ?::DATE
		  AND al.status IN ('present', 'late', 'business_trip')
		  AND al.deleted_at IS NULL
		  AND NOT EXISTS (
			  SELECT 1 FROM daily_reports dr
			  WHERE dr.employee_id = al.employee_id
			    AND dr.report_date = ?::DATE
			    AND dr.deleted_at IS NULL
		  )
	`, date, date).Scan(&rows).Error
	return rows, err
}

func (r *dailyReportRepository) BulkCreateMissingLogs(ctx context.Context, tx Transaction, logs []model.DailyReport) error {
	if len(logs) == 0 {
		return nil
	}
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Create(&logs).Error
}
