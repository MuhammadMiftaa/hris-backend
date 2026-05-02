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

type OvertimeRepository interface {
	GetAll(ctx context.Context, tx Transaction, params dto.OvertimeListParams) (dto.PaginatedResponse[dto.OvertimeRequestResponse], error)
	GetByID(ctx context.Context, tx Transaction, id uint) (*dto.OvertimeRequestResponse, error)
	Create(ctx context.Context, tx Transaction, m model.OvertimeRequest) (model.OvertimeRequest, error)
	UpdateRequestStatus(ctx context.Context, tx Transaction, id uint, status string) error
	CreateApproval(ctx context.Context, tx Transaction, m model.OvertimeRequestApproval) (model.OvertimeRequestApproval, error)
	GetApprovalsByRequestID(ctx context.Context, tx Transaction, requestID uint) ([]dto.OvertimeApprovalResponse, error)
	UpdateApprovalStatus(ctx context.Context, tx Transaction, approvalID uint, status string, approverID uint, notes *string) error
	GetPendingApprovalForLevel(ctx context.Context, tx Transaction, requestID uint, level int) (*dto.OvertimeApprovalResponse, error)
	Delete(ctx context.Context, tx Transaction, id uint) error

	// Metadata
	GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetDepartmentMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
}

type overtimeRepository struct {
	db *gorm.DB
}

func NewOvertimeRepository(db *gorm.DB) OvertimeRepository {
	return &overtimeRepository{db: db}
}

func (r *overtimeRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *overtimeRepository) GetAll(ctx context.Context, tx Transaction, params dto.OvertimeListParams) (dto.PaginatedResponse[dto.OvertimeRequestResponse], error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.PaginatedResponse[dto.OvertimeRequestResponse]{}, err
	}

	baseQuery := `
		FROM overtime_requests o
		JOIN employees e ON e.id = o.employee_id
		LEFT JOIN departments d ON d.id = e.department_id AND d.deleted_at IS NULL
		WHERE o.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.EmployeeID != nil {
		baseQuery += " AND o.employee_id = ?"
		args = append(args, *params.EmployeeID)
	}
	if params.DepartmentID != nil {
		baseQuery += " AND e.department_id = ?"
		args = append(args, *params.DepartmentID)
	}
	if params.Status != nil {
		baseQuery += " AND o.status = ?"
		args = append(args, *params.Status)
	}
	if params.StartDate != nil {
		baseQuery += " AND o.overtime_date >= ?::DATE"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		baseQuery += " AND o.overtime_date <= ?::DATE"
		args = append(args, *params.EndDate)
	}
	if params.Search != nil && *params.Search != "" {
		baseQuery += " AND (e.full_name ILIKE ? OR o.reason ILIKE ?)"
		like := "%" + *params.Search + "%"
		args = append(args, like, like)
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.OvertimeRequestResponse]{}, err
	}

	selectQuery := `
		SELECT
			o.id,
			o.employee_id,
			e.full_name AS employee_name,
			d.name AS department_name,
			o.overtime_date::TEXT,
			o.planned_start AS planned_start,
			o.planned_end AS planned_end,
			o.planned_minutes,
			o.actual_start,
			o.actual_end,
			o.actual_minutes,
			o.work_location_type,
			o.reason,
			o.status,
			o.created_at,
			o.updated_at
	` + baseQuery
	
	selectQuery += utils.BuildSortClause("overtime", params.SortBy, params.GetSortDir(), "o.created_at DESC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var res []dto.OvertimeRequestResponse
	if err := db.Raw(selectQuery, args...).Scan(&res).Error; err != nil {
		return dto.PaginatedResponse[dto.OvertimeRequestResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.OvertimeRequestResponse]{
		Data: res,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *overtimeRepository) GetByID(ctx context.Context, tx Transaction, id uint) (*dto.OvertimeRequestResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var res dto.OvertimeRequestResponse
	query := `
		SELECT
			o.id,
			o.employee_id,
			e.full_name AS employee_name,
			o.overtime_date::TEXT AS date,
			o.planned_minutes,
			o.reason,
			o.status,
			o.created_at,
			o.updated_at
		FROM overtime_requests o
		JOIN employees e ON e.id = o.employee_id
		WHERE o.id = ? AND o.deleted_at IS NULL
	`
	if err := db.Raw(query, id).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, fmt.Errorf("overtime request not found")
	}

	apprs, err := r.GetApprovalsByRequestID(ctx, tx, id)
	if err == nil {
		res.Approvals = apprs
	}

	return &res, nil
}

func (r *overtimeRepository) Create(ctx context.Context, tx Transaction, m model.OvertimeRequest) (model.OvertimeRequest, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *overtimeRepository) UpdateRequestStatus(ctx context.Context, tx Transaction, id uint, status string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.OvertimeRequest{}).Where("id = ?", id).Update("status", status).Error
}

func (r *overtimeRepository) CreateApproval(ctx context.Context, tx Transaction, m model.OvertimeRequestApproval) (model.OvertimeRequestApproval, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *overtimeRepository) GetApprovalsByRequestID(ctx context.Context, tx Transaction, requestID uint) ([]dto.OvertimeApprovalResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT
			a.id,
			a.overtime_request_id,
			a.approver_id,
			e.full_name AS approver_name,
			a.level,
			a.status,
			a.notes,
			a.decided_at,
			a.created_at
		FROM overtime_request_approvals a
		LEFT JOIN employees e ON e.id = a.approver_id
		WHERE a.overtime_request_id = ?
		ORDER BY a.level ASC
	`
	var res []dto.OvertimeApprovalResponse
	if err := db.Raw(query, requestID).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *overtimeRepository) UpdateApprovalStatus(ctx context.Context, tx Transaction, approvalID uint, status string, approverID uint, notes *string) error {
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
	return db.Model(&model.OvertimeRequestApproval{}).Where("id = ?", approvalID).Updates(upd).Error
}

func (r *overtimeRepository) GetPendingApprovalForLevel(ctx context.Context, tx Transaction, requestID uint, level int) (*dto.OvertimeApprovalResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, overtime_request_id, approver_id, level, status, notes, decided_at, created_at
		FROM overtime_request_approvals
		WHERE overtime_request_id = ? AND level = ? AND status = 'pending'
		LIMIT 1
	`
	var res dto.OvertimeApprovalResponse
	if err := db.Raw(query, requestID, level).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, nil // not found
	}
	return &res, nil
}

func (r *overtimeRepository) Delete(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Delete(&model.OvertimeRequest{}, id).Error
}

// Metadata
func (r *overtimeRepository) GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
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

func (r *overtimeRepository) GetDepartmentMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var meta []dto.Meta
	err = db.Raw(`
		SELECT id::TEXT, name
		FROM departments
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`).Scan(&meta).Error
	return meta, err
}
