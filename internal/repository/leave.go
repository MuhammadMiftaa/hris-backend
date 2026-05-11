package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"

	"gorm.io/gorm"
)

type LeaveRepository interface {
	// Balance
	GetAllBalances(ctx context.Context, tx Transaction, params dto.LeaveBalanceListParams) (dto.PaginatedResponse[dto.LeaveBalanceResponse], error)
	GetBalanceByEmployeeAndType(ctx context.Context, tx Transaction, employeeID uint, leaveTypeID uint, year int) (*dto.LeaveBalanceResponse, error)
	GetBalanceByID(ctx context.Context, tx Transaction, id uint) (*dto.LeaveBalanceResponse, error)
	CreateBalance(ctx context.Context, tx Transaction, m model.LeaveBalance) (model.LeaveBalance, error)
	UpdateBalanceUsage(ctx context.Context, tx Transaction, id uint, usedOccurrences int, usedDuration float64) error

	// Balance management
	GetEmployeeBalanceSummary(ctx context.Context, tx Transaction, params dto.EmployeeBalanceSummaryParams) (dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse], error)
	GetBalanceDetailByEmployee(ctx context.Context, tx Transaction, employeeID uint, year int) ([]dto.LeaveBalanceResponse, error)
	UpsertBalance(ctx context.Context, tx Transaction, req dto.UpsertLeaveBalanceRequest) (model.LeaveBalance, error)
	DeleteBalance(ctx context.Context, tx Transaction, id uint) error

	// Adjustments
	CreateAdjustment(ctx context.Context, tx Transaction, m model.LeaveBalanceAdjustment) (model.LeaveBalanceAdjustment, error)
	GetAdjustmentsByBalanceID(ctx context.Context, tx Transaction, balanceID uint) ([]dto.LeaveBalanceAdjustmentResponse, error)

	// Request
	GetAllRequests(ctx context.Context, tx Transaction, params dto.LeaveRequestListParams) (dto.PaginatedResponse[dto.LeaveRequestResponse], error)
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
	GetParentLeaveTypeMeta(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetDepartmentMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
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
func (r *leaveRepository) GetAllBalances(ctx context.Context, tx Transaction, params dto.LeaveBalanceListParams) (dto.PaginatedResponse[dto.LeaveBalanceResponse], error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.PaginatedResponse[dto.LeaveBalanceResponse]{}, err
	}

	baseQuery := `
		FROM leave_balances b
		JOIN employees e ON e.id = b.employee_id
		LEFT JOIN departments d ON d.id = e.department_id AND d.deleted_at IS NULL
		JOIN leave_types t ON t.id = b.leave_type_id
		LEFT JOIN (
		  SELECT leave_balance_id, SUM(delta) AS total_delta
		  FROM leave_balance_adjustments
		  GROUP BY leave_balance_id
		) adj ON adj.leave_balance_id = b.id
		WHERE b.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.LeaveTypeID != nil {
		baseQuery += " AND b.leave_type_id = ?"
		args = append(args, *params.LeaveTypeID)
	}
	if params.DepartmentID != nil {
		baseQuery += " AND e.department_id = ?"
		args = append(args, *params.DepartmentID)
	}
	if params.Year != nil && *params.Year != "" {
		baseQuery += " AND b.year::TEXT ILIKE ?"
		like := "%" + *params.Year + "%"
		args = append(args, like)
	}
	if params.UsedDuration != nil && *params.UsedDuration != "" {
		baseQuery += " AND b.used_duration = ?"
		args = append(args, *params.UsedDuration)
	}
	if params.MaxDuration != nil && *params.MaxDuration != "" {
		baseQuery += " AND t.max_total_duration_per_year = ?"
		args = append(args, *params.MaxDuration)
	}
	if params.EmployeeName != nil && *params.EmployeeName != "" {
		baseQuery += " AND (e.full_name ILIKE ?)"
		like := "%" + *params.EmployeeName + "%"
		args = append(args, like)
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.LeaveBalanceResponse]{}, err
	}

	selectQuery := `
		SELECT
			b.id,
			b.employee_id,
			e.full_name AS employee_name,
			d.name AS department_name,
			b.leave_type_id,
			t.name AS leave_type_name,
			b.year,
			b.used_occurrences,
			b.used_duration,
			b.allocated_duration,
			t.max_occurrences_per_year AS max_occurrences,
			t.max_total_duration_per_year AS max_duration,
			(t.max_occurrences_per_year - b.used_occurrences) AS remaining_occurrences,
			(b.allocated_duration + COALESCE(adj.total_delta, 0) - b.used_duration) AS remaining_duration,
			b.effective_date::TEXT AS effective_date,
			b.notes,
			COALESCE(adj.total_delta, 0) AS total_adjustment,
			b.created_at,
			b.updated_at
	` + baseQuery

	selectQuery += utils.BuildSortClause("leave_balances", params.SortBy, params.GetSortDir(), "b.created_at DESC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var res []dto.LeaveBalanceResponse
	if err := db.Raw(selectQuery, args...).Scan(&res).Error; err != nil {
		return dto.PaginatedResponse[dto.LeaveBalanceResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.LeaveBalanceResponse]{
		Data: res,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
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
			b.allocated_duration,
			t.max_occurrences_per_year AS max_occurrences,
			t.max_total_duration_per_year AS max_duration,
			(t.max_occurrences_per_year - b.used_occurrences) AS remaining_occurrences,
			(b.allocated_duration + COALESCE(adj.total_delta, 0) - b.used_duration) AS remaining_duration,
			b.effective_date::TEXT AS effective_date,
			b.notes,
			COALESCE(adj.total_delta, 0) AS total_adjustment,
			b.created_at,
			b.updated_at
		FROM leave_balances b
		JOIN employees e ON e.id = b.employee_id
		JOIN leave_types t ON t.id = b.leave_type_id
		LEFT JOIN (
		  SELECT leave_balance_id, SUM(delta) AS total_delta
		  FROM leave_balance_adjustments
		  GROUP BY leave_balance_id
		) adj ON adj.leave_balance_id = b.id
		WHERE b.employee_id = ? AND b.leave_type_id = ? AND b.year = ? AND b.deleted_at IS NULL AND b.effective_date <= NOW()
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

func (r *leaveRepository) GetBalanceByID(ctx context.Context, tx Transaction, id uint) (*dto.LeaveBalanceResponse, error) {
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
			d.name AS department_name,
			b.leave_type_id,
			t.name AS leave_type_name,
			b.year,
			b.used_occurrences,
			b.used_duration,
			b.allocated_duration,
			t.max_occurrences_per_year AS max_occurrences,
			t.max_total_duration_per_year AS max_duration,
			(t.max_occurrences_per_year - b.used_occurrences) AS remaining_occurrences,
			(b.allocated_duration + COALESCE(adj.total_delta, 0) - b.used_duration) AS remaining_duration,
			b.effective_date::TEXT AS effective_date,
			b.notes,
			COALESCE(adj.total_delta, 0) AS total_adjustment,
			b.created_at,
			b.updated_at
		FROM leave_balances b
		JOIN employees e ON e.id = b.employee_id
		LEFT JOIN departments d ON d.id = e.department_id AND d.deleted_at IS NULL
		JOIN leave_types t ON t.id = b.leave_type_id
		LEFT JOIN (
		  SELECT leave_balance_id, SUM(delta) AS total_delta
		  FROM leave_balance_adjustments
		  GROUP BY leave_balance_id
		) adj ON adj.leave_balance_id = b.id
		WHERE b.id = ? AND b.deleted_at IS NULL
		LIMIT 1
	`
	if err := db.Raw(query, id).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, nil
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

func (r *leaveRepository) UpdateBalanceUsage(ctx context.Context, tx Transaction, id uint, usedOccurrences int, usedDuration float64) error {
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
func (r *leaveRepository) GetAllRequests(ctx context.Context, tx Transaction, params dto.LeaveRequestListParams) (dto.PaginatedResponse[dto.LeaveRequestResponse], error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.PaginatedResponse[dto.LeaveRequestResponse]{}, err
	}

	baseQuery := `
		FROM leave_requests r
		JOIN employees e ON e.id = r.employee_id
		LEFT JOIN departments d ON d.id = e.department_id AND d.deleted_at IS NULL
		JOIN leave_types t ON t.id = r.leave_type_id
		WHERE r.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.DepartmentID != nil {
		baseQuery += " AND e.department_id = ?"
		args = append(args, *params.DepartmentID)
	}
	if params.Status != nil && *params.Status != "" {
		baseQuery += " AND r.status = ?"
		args = append(args, *params.Status)
	}
	if params.LeaveTypeID != nil {
		baseQuery += " AND r.leave_type_id = ?"
		args = append(args, *params.LeaveTypeID)
	}
	if params.StartDate != nil && *params.StartDate != "" {
		baseQuery += " AND r.start_date::DATE >= ?::DATE"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil && *params.EndDate != "" {
		baseQuery += " AND r.end_date::DATE <= ?::DATE"
		args = append(args, *params.EndDate)
	}
	if params.EmployeeName != nil && *params.EmployeeName != "" {
		baseQuery += " AND e.full_name ILIKE ?"
		like := "%" + *params.EmployeeName + "%"
		args = append(args, like)
	}
	if params.TotalDays != nil && *params.TotalDays != "" {
		baseQuery += " AND r.total_days::TEXT ILIKE ?"
		like := "%" + *params.TotalDays + "%"
		args = append(args, like)
	}
	if params.Reason != nil && *params.Reason != "" {
		baseQuery += " AND r.reason ILIKE ?"
		args = append(args, "%"+*params.Reason+"%")
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.LeaveRequestResponse]{}, err
	}

	selectQuery := `
		SELECT
			r.id,
			r.employee_id,
			e.full_name AS employee_name,
			d.name AS department_name,
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
	` + baseQuery

	selectQuery += utils.BuildSortClause("leave_requests", params.SortBy, params.GetSortDir(), "r.created_at DESC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var res []dto.LeaveRequestResponse
	if err := db.Raw(selectQuery, args...).Scan(&res).Error; err != nil {
		return dto.PaginatedResponse[dto.LeaveRequestResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.LeaveRequestResponse]{
		Data: res,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
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

func (r *leaveRepository) GetParentLeaveTypeMeta(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var res []dto.Meta
	err = db.Raw(`SELECT id::TEXT, name FROM leave_types WHERE deleted_at IS NULL AND parent_leave_type_id IS NULL ORDER BY id ASC`).Scan(&res).Error
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

func (r *leaveRepository) GetDepartmentMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var res []dto.Meta
	err = db.Raw(`SELECT id::TEXT, name FROM departments WHERE deleted_at IS NULL ORDER BY name ASC`).Scan(&res).Error
	return res, err
}

func (r *leaveRepository) GetEmployeeBalanceSummary(ctx context.Context, tx Transaction, params dto.EmployeeBalanceSummaryParams) (dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse], error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse]{}, err
	}

	baseQuery := `
		FROM leave_balances b
		JOIN employees e ON e.id = b.employee_id
		LEFT JOIN departments d ON d.id = e.department_id AND d.deleted_at IS NULL
		LEFT JOIN job_positions jp ON jp.id = e.job_positions_id AND jp.deleted_at IS NULL
		LEFT JOIN (
		  SELECT leave_balance_id, SUM(delta) AS total_delta
		  FROM leave_balance_adjustments
		  GROUP BY leave_balance_id
		) adj ON adj.leave_balance_id = b.id
		WHERE b.deleted_at IS NULL AND b.effective_date <= NOW()
	`
	args := []interface{}{}

	if params.Year != nil {
		baseQuery += " AND b.year = ?"
		args = append(args, *params.Year)
	}
	if params.EmployeeName != nil && *params.EmployeeName != "" {
		baseQuery += " AND e.full_name ILIKE ?"
		args = append(args, "%"+*params.EmployeeName+"%")
	}
	if params.DepartmentID != nil {
		baseQuery += " AND e.department_id = ?"
		args = append(args, *params.DepartmentID)
	}
	if params.JobPositionTitle != nil && *params.JobPositionTitle != "" {
		baseQuery += " AND (jp.title ILIKE ?)"
		args = append(args, "%"+*params.JobPositionTitle+"%")
	}

	var total int
	countQuery := `
		SELECT COUNT(*) FROM (
			SELECT e.id, b.year
	` + baseQuery + `
			GROUP BY e.id, b.year
		) sub
	`
	if err := db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse]{}, err
	}

	selectQuery := `
		SELECT
		  e.id AS employee_id,
		  e.full_name AS employee_name,
		  d.name AS department_name,
		  jp.title AS job_position_title,
		  b.year,
		  SUM(b.allocated_duration + COALESCE(adj.total_delta, 0)) AS total_allocated,
		  SUM(b.used_duration) AS total_used,
		  SUM(b.allocated_duration + COALESCE(adj.total_delta, 0) - b.used_duration) AS total_remaining
	` + baseQuery + `
		GROUP BY e.id, e.full_name, d.name, jp.title, b.year
	`

	selectQuery += utils.BuildSortClause("balance_summary", params.SortBy, params.GetSortDir(), "e.full_name ASC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var res []dto.EmployeeBalanceSummaryResponse
	if err := db.Raw(selectQuery, args...).Scan(&res).Error; err != nil {
		return dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.EmployeeBalanceSummaryResponse]{
		Data: res,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *leaveRepository) GetBalanceDetailByEmployee(ctx context.Context, tx Transaction, employeeID uint, year int) ([]dto.LeaveBalanceResponse, error) {
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
			b.allocated_duration,
			t.max_occurrences_per_year AS max_occurrences,
			t.max_total_duration_per_year AS max_duration,
			(t.max_occurrences_per_year - b.used_occurrences) AS remaining_occurrences,
			(b.allocated_duration + COALESCE(adj.total_delta, 0) - b.used_duration) AS remaining_duration,
			b.effective_date::TEXT AS effective_date,
			b.notes,
			COALESCE(adj.total_delta, 0) AS total_adjustment,
			b.created_at,
			b.updated_at
		FROM leave_balances b
		JOIN employees e ON e.id = b.employee_id
		JOIN leave_types t ON t.id = b.leave_type_id
		LEFT JOIN (
		  SELECT leave_balance_id, SUM(delta) AS total_delta
		  FROM leave_balance_adjustments
		  GROUP BY leave_balance_id
		) adj ON adj.leave_balance_id = b.id
		WHERE b.employee_id = ? AND b.year = ? AND b.deleted_at IS NULL
		ORDER BY t.name ASC
	`
	var res []dto.LeaveBalanceResponse
	if err := db.Raw(query, employeeID, year).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *leaveRepository) UpsertBalance(ctx context.Context, tx Transaction, req dto.UpsertLeaveBalanceRequest) (model.LeaveBalance, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.LeaveBalance{}, err
	}

	effDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return model.LeaveBalance{}, fmt.Errorf("invalid effective_date: %w", err)
	}

	var m model.LeaveBalance
	query := `
		INSERT INTO leave_balances
			(employee_id, leave_type_id, year, allocated_duration, effective_date, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (employee_id, leave_type_id, year) WHERE deleted_at IS NULL
		DO UPDATE SET
			allocated_duration = EXCLUDED.allocated_duration,
			effective_date     = EXCLUDED.effective_date,
			notes              = EXCLUDED.notes,
			updated_at         = NOW()
		RETURNING *
	`
	if err := db.Raw(query, req.EmployeeID, req.LeaveTypeID, req.Year, req.AllocatedDuration, effDate, req.Notes).Scan(&m).Error; err != nil {
		return model.LeaveBalance{}, err
	}
	return m, nil
}

func (r *leaveRepository) DeleteBalance(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Where("id = ?", id).Delete(&model.LeaveBalance{}).Error
}

func (r *leaveRepository) CreateAdjustment(ctx context.Context, tx Transaction, m model.LeaveBalanceAdjustment) (model.LeaveBalanceAdjustment, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *leaveRepository) GetAdjustmentsByBalanceID(ctx context.Context, tx Transaction, balanceID uint) ([]dto.LeaveBalanceAdjustmentResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			a.id,
			a.leave_balance_id,
			e.full_name AS adjuster_name,
			a.delta,
			a.reason,
			a.created_at::TEXT AS created_at
		FROM leave_balance_adjustments a
		JOIN employees e ON e.id = a.adjusted_by
		WHERE a.leave_balance_id = ?
		ORDER BY a.created_at DESC
	`
	var res []dto.LeaveBalanceAdjustmentResponse
	if err := db.Raw(query, balanceID).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
