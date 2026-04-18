package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type LeaveRepository interface {
	// Balance
	GetAllBalances(ctx context.Context, tx Transaction, params dto.LeaveBalanceListParams) ([]dto.LeaveBalanceResponse, error)
	GetBalanceByEmployeeAndType(ctx context.Context, tx Transaction, employeeID uint, leaveTypeID uint, year int) (*dto.LeaveBalanceResponse, error)
	CreateBalance(ctx context.Context, tx Transaction, m model.LeaveBalance) (model.LeaveBalance, error)
	UpdateBalanceUsage(ctx context.Context, tx Transaction, id uint, usedOccurrences int, usedDuration int) error

	// Request
	GetAllRequests(ctx context.Context, tx Transaction, params dto.LeaveRequestListParams) ([]dto.LeaveRequestResponse, error)
	GetRequestByID(ctx context.Context, tx Transaction, id uint) (*dto.LeaveRequestResponse, error)
	CreateRequest(ctx context.Context, tx Transaction, m model.LeaveRequest) (model.LeaveRequest, error)
	UpdateRequestStatus(ctx context.Context, tx Transaction, id uint, status string) error

	// Approval
	CreateApproval(ctx context.Context, tx Transaction, m model.LeaveRequestApproval) (model.LeaveRequestApproval, error)
	GetApprovalsByRequestID(ctx context.Context, tx Transaction, requestID uint) ([]dto.LeaveApprovalResponse, error)
	UpdateApprovalStatus(ctx context.Context, tx Transaction, approvalID uint, status string, approverID uint, notes *string) error
	GetPendingApprovalForLevel(ctx context.Context, tx Transaction, requestID uint, level int) (*dto.LeaveApprovalResponse, error)

	// Validation
	CheckOverlap(ctx context.Context, tx Transaction, employeeID uint, startDate string, endDate string, excludeID *uint) (bool, error)

	// Metadata
	GetLeaveTypeMeta(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
}

type leaveRepository struct {
	db *gorm.DB
}

func NewLeaveRepository(db *gorm.DB) LeaveRepository {
	return &leaveRepository{db: db}
}

func (r *leaveRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

// Balance
func (r *leaveRepository) GetAllBalances(ctx context.Context, tx Transaction, params dto.LeaveBalanceListParams) ([]dto.LeaveBalanceResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			b.id,
			b.employee_id,
			e.full_name AS employee_name,
			b.leave_type_id,
			t.name AS leave_type_name,
			b.year,
			b.used_occurrences,
			b.used_duration,
			t.max_occurrences_per_year AS max_occurrences,
			t.max_total_duration_per_year AS max_duration,
			(t.max_occurrences_per_year - b.used_occurrences) AS remaining_occurrences,
			(t.max_total_duration_per_year - b.used_duration) AS remaining_duration,
			b.created_at,
			b.updated_at
		FROM leave_balances b
		JOIN employees e ON e.id = b.employee_id
		JOIN leave_types t ON t.id = b.leave_type_id
		WHERE b.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.EmployeeID != nil {
		query += " AND b.employee_id = ?"
		args = append(args, *params.EmployeeID)
	}
	if params.Year != nil {
		query += " AND b.year = ?"
		args = append(args, *params.Year)
	}

	var res []dto.LeaveBalanceResponse
	if err := db.Raw(query, args...).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *leaveRepository) GetBalanceByEmployeeAndType(ctx context.Context, tx Transaction, employeeID uint, leaveTypeID uint, year int) (*dto.LeaveBalanceResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var res dto.LeaveBalanceResponse
	query := `
		SELECT
			b.id,
			b.employee_id,
			e.full_name AS employee_name,
			b.leave_type_id,
			t.name AS leave_type_name,
			b.year,
			b.used_occurrences,
			b.used_duration,
			t.max_occurrences_per_year AS max_occurrences,
			t.max_total_duration_per_year AS max_duration,
			(t.max_occurrences_per_year - b.used_occurrences) AS remaining_occurrences,
			(t.max_total_duration_per_year - b.used_duration) AS remaining_duration,
			b.created_at,
			b.updated_at
		FROM leave_balances b
		JOIN employees e ON e.id = b.employee_id
		JOIN leave_types t ON t.id = b.leave_type_id
		WHERE b.employee_id = ? AND b.leave_type_id = ? AND b.year = ? AND b.deleted_at IS NULL
		LIMIT 1
	`
	err = db.Raw(query, employeeID, leaveTypeID, year).Scan(&res).Error
	if err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, nil // not found
	}
	return &res, nil
}

func (r *leaveRepository) CreateBalance(ctx context.Context, tx Transaction, m model.LeaveBalance) (model.LeaveBalance, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *leaveRepository) UpdateBalanceUsage(ctx context.Context, tx Transaction, id uint, usedOccurrences int, usedDuration int) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.LeaveBalance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used_occurrences": usedOccurrences,
			"used_duration":    usedDuration,
		}).Error
}

// Request
func (r *leaveRepository) GetAllRequests(ctx context.Context, tx Transaction, params dto.LeaveRequestListParams) ([]dto.LeaveRequestResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			r.id,
			r.employee_id,
			e.full_name AS employee_name,
			r.leave_type_id,
			t.name AS leave_type_name,
			t.category AS leave_category,
			r.start_date::TEXT AS start_date,
			r.end_date::TEXT AS end_date,
			r.total_days,
			r.total_hours,
			r.reason,
			r.document_url,
			r.status,
			r.created_at,
			r.updated_at
		FROM leave_requests r
		JOIN employees e ON e.id = r.employee_id
		JOIN leave_types t ON t.id = r.leave_type_id
		WHERE r.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.EmployeeID != nil {
		query += " AND r.employee_id = ?"
		args = append(args, *params.EmployeeID)
	}
	if params.Status != nil {
		query += " AND r.status = ?"
		args = append(args, *params.Status)
	}
	if params.LeaveTypeID != nil {
		query += " AND r.leave_type_id = ?"
		args = append(args, *params.LeaveTypeID)
	}
	if params.Year != nil {
		query += " AND EXTRACT(YEAR FROM r.start_date) = ?"
		args = append(args, *params.Year)
	}
	query += " ORDER BY r.created_at DESC"

	var res []dto.LeaveRequestResponse
	if err := db.Raw(query, args...).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *leaveRepository) GetRequestByID(ctx context.Context, tx Transaction, id uint) (*dto.LeaveRequestResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			r.id,
			r.employee_id,
			e.full_name AS employee_name,
			r.leave_type_id,
			t.name AS leave_type_name,
			t.category AS leave_category,
			r.start_date::TEXT AS start_date,
			r.end_date::TEXT AS end_date,
			r.total_days,
			r.total_hours,
			r.reason,
			r.document_url,
			r.status,
			r.created_at,
			r.updated_at
		FROM leave_requests r
		JOIN employees e ON e.id = r.employee_id
		JOIN leave_types t ON t.id = r.leave_type_id
		WHERE r.id = ? AND r.deleted_at IS NULL
	`
	var res dto.LeaveRequestResponse
	if err := db.Raw(query, id).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, fmt.Errorf("leave request not found")
	}

	apprs, err := r.GetApprovalsByRequestID(ctx, tx, id)
	if err == nil {
		res.Approvals = apprs
	}

	return &res, nil
}

func (r *leaveRepository) CreateRequest(ctx context.Context, tx Transaction, m model.LeaveRequest) (model.LeaveRequest, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *leaveRepository) UpdateRequestStatus(ctx context.Context, tx Transaction, id uint, status string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.LeaveRequest{}).Where("id = ?", id).Update("status", status).Error
}

// Approval
func (r *leaveRepository) CreateApproval(ctx context.Context, tx Transaction, m model.LeaveRequestApproval) (model.LeaveRequestApproval, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *leaveRepository) GetApprovalsByRequestID(ctx context.Context, tx Transaction, requestID uint) ([]dto.LeaveApprovalResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			a.id,
			a.leave_request_id,
			a.approver_id,
			e.full_name AS approver_name,
			a.level,
			a.status,
			a.notes,
			a.decided_at,
			a.created_at
		FROM leave_request_approvals a
		LEFT JOIN employees e ON e.id = a.approver_id
		WHERE a.leave_request_id = ?
		ORDER BY a.level ASC
	`
	var res []dto.LeaveApprovalResponse
	if err := db.Raw(query, requestID).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *leaveRepository) UpdateApprovalStatus(ctx context.Context, tx Transaction, approvalID uint, status string, approverID uint, notes *string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	upd := map[string]interface{}{
		"status":      status,
		"approver_id": approverID,
		"notes":       notes,
		"decided_at":  gorm.Expr("NOW()"),
	}
	return db.Model(&model.LeaveRequestApproval{}).Where("id = ?", approvalID).Updates(upd).Error
}

func (r *leaveRepository) GetPendingApprovalForLevel(ctx context.Context, tx Transaction, requestID uint, level int) (*dto.LeaveApprovalResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, leave_request_id, approver_id, level, status, notes, decided_at, created_at
		FROM leave_request_approvals
		WHERE leave_request_id = ? AND level = ? AND status = 'pending'
		LIMIT 1
	`
	var res dto.LeaveApprovalResponse
	if err := db.Raw(query, requestID, level).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, nil // not found
	}
	return &res, nil
}

// Validation
func (r *leaveRepository) CheckOverlap(ctx context.Context, tx Transaction, employeeID uint, startDate string, endDate string, excludeID *uint) (bool, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return false, err
	}

	query := `
		SELECT COUNT(*) FROM leave_requests
		WHERE employee_id = ?
		  AND status IN ('pending', 'approved_leader', 'approved_hr')
		  AND (start_date <= ?::DATE AND end_date >= ?::DATE)
		  AND deleted_at IS NULL
	`
	args := []interface{}{employeeID, endDate, startDate}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	var count int64
	if err := db.Raw(query, args...).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Metadata
func (r *leaveRepository) GetLeaveTypeMeta(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var res []dto.Meta
	err = db.Raw(`SELECT id::TEXT, name FROM leave_types WHERE deleted_at IS NULL ORDER BY id ASC`).Scan(&res).Error
	return res, err
}

func (r *leaveRepository) GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
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
