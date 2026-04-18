package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type PermissionRequestRepository interface {
	GetAll(ctx context.Context, tx Transaction, params dto.PermissionListParams) ([]dto.PermissionRequestResponse, error)
	GetByID(ctx context.Context, tx Transaction, id uint) (*dto.PermissionRequestResponse, error)
	Create(ctx context.Context, tx Transaction, m model.PermissionRequest) (model.PermissionRequest, error)
	UpdateStatus(ctx context.Context, tx Transaction, id uint, status string, approverID uint, notes *string) error
	Delete(ctx context.Context, tx Transaction, id uint) error

	// Metadata
	GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
}

type permissionRequestRepository struct {
	db *gorm.DB
}

func NewPermissionRequestRepository(db *gorm.DB) PermissionRequestRepository {
	return &permissionRequestRepository{db: db}
}

func (r *permissionRequestRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *permissionRequestRepository) GetAll(ctx context.Context, tx Transaction, params dto.PermissionListParams) ([]dto.PermissionRequestResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			pr.id,
			pr.employee_id,
			e.full_name AS employee_name,
			pr.date::TEXT AS date,
			pr.permission_type,
			pr.leave_time::TEXT AS start_time,
			pr.return_time::TEXT AS end_time,
			pr.leave_time - pr.return_time AS duration,
			pr.reason,
			pr.document_url,
			pr.status,
			pr.approved_by AS approver_id,
			a.full_name AS approver_name,
			pr.approver_notes,
			pr.created_at,
			pr.updated_at
		FROM permission_requests pr
		JOIN employees e ON e.id = pr.employee_id
		LEFT JOIN employees a ON a.id = pr.approved_by
		WHERE pr.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.EmployeeID != nil {
		query += " AND pr.employee_id = ?"
		args = append(args, *params.EmployeeID)
	}
	if params.Status != nil {
		query += " AND pr.status = ?"
		args = append(args, *params.Status)
	}
	if params.StartDate != nil {
		query += " AND pr.date >= ?::DATE"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		query += " AND pr.date <= ?::DATE"
		args = append(args, *params.EndDate)
	}
	query += " ORDER BY pr.created_at DESC"

	var res []dto.PermissionRequestResponse
	if err := db.Raw(query, args...).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *permissionRequestRepository) GetByID(ctx context.Context, tx Transaction, id uint) (*dto.PermissionRequestResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var res dto.PermissionRequestResponse
	query := `
		SELECT
			pr.id,
			pr.employee_id,
			e.full_name AS employee_name,
			pr.date::TEXT AS date,
			pr.permission_type,
			pr.leave_time::TEXT AS start_time,
			pr.return_time::TEXT AS end_time,
			pr.leave_time - pr.return_time AS duration,
			pr.reason,
			pr.document_url,
			pr.status,
			pr.approved_by AS approver_id,
			a.full_name AS approver_name,
			pr.approver_notes,
			pr.created_at,
			pr.updated_at
		FROM permission_requests pr
		JOIN employees e ON e.id = pr.employee_id
		LEFT JOIN employees a ON a.id = pr.approved_by
		WHERE pr.id = ? AND pr.deleted_at IS NULL
	`
	if err := db.Raw(query, id).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, fmt.Errorf("permission request not found")
	}
	return &res, nil
}

func (r *permissionRequestRepository) Create(ctx context.Context, tx Transaction, m model.PermissionRequest) (model.PermissionRequest, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *permissionRequestRepository) UpdateStatus(ctx context.Context, tx Transaction, id uint, status string, approverID uint, notes *string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	upd := map[string]interface{}{
		"status":         status,
		"approved_by":    approverID,
		"approver_notes": notes,
	}
	return db.Model(&model.PermissionRequest{}).Where("id = ?", id).Updates(upd).Error
}

func (r *permissionRequestRepository) Delete(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Delete(&model.PermissionRequest{}, id).Error
}

// Metadata
func (r *permissionRequestRepository) GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
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
