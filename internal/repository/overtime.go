package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type OvertimeRepository interface {
	GetAll(ctx context.Context, tx Transaction, params dto.OvertimeListParams) ([]dto.OvertimeRequestResponse, error)
	GetByID(ctx context.Context, tx Transaction, id uint) (*dto.OvertimeRequestResponse, error)
	Create(ctx context.Context, tx Transaction, m model.OvertimeRequest) (model.OvertimeRequest, error)
	UpdateStatus(ctx context.Context, tx Transaction, id uint, status string, approverID uint, notes *string) error
	Delete(ctx context.Context, tx Transaction, id uint) error
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

func (r *overtimeRepository) GetAll(ctx context.Context, tx Transaction, params dto.OvertimeListParams) ([]dto.OvertimeRequestResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			o.id,
			o.employee_id,
			e.full_name AS employee_name,
			o.overtime_date::TEXT AS date,
			o.planned_minutes,
			o.reason,
			o.status,
			o.approved_by AS approver_id,
			a.full_name AS approver_name,
			o.approver_notes,
			o.created_at,
			o.updated_at
		FROM overtime_requests o
		JOIN employees e ON e.id = o.employee_id
		LEFT JOIN employees a ON a.id = o.approved_by
		WHERE o.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.EmployeeID != nil {
		query += " AND o.employee_id = ?"
		args = append(args, *params.EmployeeID)
	}
	if params.Status != nil {
		query += " AND o.status = ?"
		args = append(args, *params.Status)
	}
	if params.StartDate != nil {
		query += " AND o.overtime_date >= ?::DATE"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		query += " AND o.overtime_date <= ?::DATE"
		args = append(args, *params.EndDate)
	}
	query += " ORDER BY o.created_at DESC"

	var res []dto.OvertimeRequestResponse
	if err := db.Raw(query, args...).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
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
			o.approved_by AS approver_id,
			a.full_name AS approver_name,
			o.approver_notes,
			o.created_at,
			o.updated_at
		FROM overtime_requests o
		JOIN employees e ON e.id = o.employee_id
		LEFT JOIN employees a ON a.id = o.approved_by
		WHERE o.id = ? AND o.deleted_at IS NULL
	`
	if err := db.Raw(query, id).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, fmt.Errorf("overtime request not found")
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

func (r *overtimeRepository) UpdateStatus(ctx context.Context, tx Transaction, id uint, status string, approverID uint, notes *string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	upd := map[string]interface{}{
		"status":         status,
		"approver_id":    approverID,
		"approver_notes": notes,
	}
	return db.Model(&model.OvertimeRequest{}).Where("id = ?", id).Updates(upd).Error
}

func (r *overtimeRepository) Delete(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Delete(&model.OvertimeRequest{}, id).Error
}
