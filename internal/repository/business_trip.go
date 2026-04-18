package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type BusinessTripRepository interface {
	GetAll(ctx context.Context, tx Transaction, params dto.BusinessTripListParams) ([]dto.BusinessTripRequestResponse, error)
	GetByID(ctx context.Context, tx Transaction, id uint) (*dto.BusinessTripRequestResponse, error)
	Create(ctx context.Context, tx Transaction, m model.BusinessTripRequest) (model.BusinessTripRequest, error)
	UpdateStatus(ctx context.Context, tx Transaction, id uint, status string, approverID uint, notes *string) error
	Delete(ctx context.Context, tx Transaction, id uint) error
}

type businessTripRepository struct {
	db *gorm.DB
}

func NewBusinessTripRepository(db *gorm.DB) BusinessTripRepository {
	return &businessTripRepository{db: db}
}

func (r *businessTripRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *businessTripRepository) GetAll(ctx context.Context, tx Transaction, params dto.BusinessTripListParams) ([]dto.BusinessTripRequestResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			b.id,
			b.employee_id,
			e.full_name AS employee_name,
			b.start_date::TEXT AS start_date,
			b.end_date::TEXT AS end_date,
			b.destination,
			b.purpose,
			b.document_url,
			b.status,
			b.approved_by AS approver_id,
			a.full_name AS approver_name,
			b.approver_notes,
			b.created_at,
			b.updated_at
		FROM business_trip_requests b
		JOIN employees e ON e.id = b.employee_id
		LEFT JOIN employees a ON a.id = b.approved_by
		WHERE b.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.EmployeeID != nil {
		query += " AND b.employee_id = ?"
		args = append(args, *params.EmployeeID)
	}
	if params.Status != nil {
		query += " AND b.status = ?"
		args = append(args, *params.Status)
	}
	if params.StartDate != nil {
		query += " AND b.start_date >= ?::DATE"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		query += " AND b.end_date <= ?::DATE"
		args = append(args, *params.EndDate)
	}
	query += " ORDER BY b.created_at DESC"

	var res []dto.BusinessTripRequestResponse
	if err := db.Raw(query, args...).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *businessTripRepository) GetByID(ctx context.Context, tx Transaction, id uint) (*dto.BusinessTripRequestResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var res dto.BusinessTripRequestResponse
	query := `
		SELECT
			b.id,
			b.employee_id,
			e.full_name AS employee_name,
			b.start_date::TEXT AS start_date,
			b.end_date::TEXT AS end_date,
			b.destination,
			b.purpose,
			b.document_url,
			b.status,
			b.approved_by AS approver_id,
			a.full_name AS approver_name,
			b.approver_notes,
			b.created_at,
			b.updated_at
		FROM business_trip_requests b
		JOIN employees e ON e.id = b.employee_id
		LEFT JOIN employees a ON a.id = b.approved_by
		WHERE b.id = ? AND b.deleted_at IS NULL
	`
	if err := db.Raw(query, id).Scan(&res).Error; err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, fmt.Errorf("business trip request not found")
	}
	return &res, nil
}

func (r *businessTripRepository) Create(ctx context.Context, tx Transaction, m model.BusinessTripRequest) (model.BusinessTripRequest, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *businessTripRepository) UpdateStatus(ctx context.Context, tx Transaction, id uint, status string, approverID uint, notes *string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	upd := map[string]interface{}{
		"status":         status,
		"approved_by":    approverID,
		"approver_notes": notes,
	}
	return db.Model(&model.BusinessTripRequest{}).Where("id = ?", id).Updates(upd).Error
}

func (r *businessTripRepository) Delete(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Delete(&model.BusinessTripRequest{}, id).Error
}
