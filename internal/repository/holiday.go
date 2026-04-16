package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type HolidayRepository interface {
	GetAllHolidays(ctx context.Context, tx Transaction, params *dto.HolidayListParams) ([]dto.HolidayResponse, error)
	GetHolidayByID(ctx context.Context, tx Transaction, id uint) (dto.HolidayResponse, error)
	CreateHoliday(ctx context.Context, tx Transaction, m model.Holiday) (model.Holiday, error)
	UpdateHoliday(ctx context.Context, tx Transaction, id uint, m model.Holiday) (model.Holiday, error)
	DeleteHoliday(ctx context.Context, tx Transaction, id uint) error
	GetBranchMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error)
}

type holidayRepository struct {
	db *gorm.DB
}

func NewHolidayRepository(db *gorm.DB) HolidayRepository {
	return &holidayRepository{db: db}
}

func (r *holidayRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *holidayRepository) GetAllHolidays(ctx context.Context, tx Transaction, params *dto.HolidayListParams) ([]dto.HolidayResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			h.id,
			h.name,
			h.year,
			h.date::TEXT AS date,
			h.type::TEXT AS type,
			h.branch_id,
			b.name       AS branch_name,
			h.description,
			h.created_at, h.updated_at, h.deleted_at
		FROM holidays h
		LEFT JOIN branches b ON b.id = h.branch_id AND b.deleted_at IS NULL
		WHERE h.deleted_at IS NULL
	`
	args := []interface{}{}

	if params != nil {
		if params.Year != nil {
			query += " AND h.year = ?"
			args = append(args, *params.Year)
		}
		if params.Type != nil {
			query += " AND h.type = ?"
			args = append(args, *params.Type)
		}
		if params.BranchID != nil {
			query += " AND h.branch_id = ?"
			args = append(args, *params.BranchID)
		}
	}
	query += " ORDER BY h.date DESC"

	var holidays []dto.HolidayResponse
	if err := db.Raw(query, args...).Scan(&holidays).Error; err != nil {
		return nil, err
	}
	return holidays, nil
}

func (r *holidayRepository) GetHolidayByID(ctx context.Context, tx Transaction, id uint) (dto.HolidayResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.HolidayResponse{}, err
	}

	var holiday dto.HolidayResponse
	if err := db.Raw(`
		SELECT
			h.id,
			h.name,
			h.year,
			h.date::TEXT AS date,
			h.type::TEXT AS type,
			h.branch_id,
			b.name       AS branch_name,
			h.description,
			h.created_at, h.updated_at, h.deleted_at
		FROM holidays h
		LEFT JOIN branches b ON b.id = h.branch_id AND b.deleted_at IS NULL
		WHERE h.deleted_at IS NULL AND h.id = ?
	`, id).Scan(&holiday).Error; err != nil {
		return dto.HolidayResponse{}, err
	}
	if holiday.ID == 0 {
		return dto.HolidayResponse{}, fmt.Errorf("holiday not found")
	}
	return holiday, nil
}

func (r *holidayRepository) CreateHoliday(ctx context.Context, tx Transaction, m model.Holiday) (model.Holiday, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Holiday{}, err
	}

	if err := db.Create(&m).Error; err != nil {
		return model.Holiday{}, err
	}
	return m, nil
}

func (r *holidayRepository) UpdateHoliday(ctx context.Context, tx Transaction, id uint, m model.Holiday) (model.Holiday, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Holiday{}, err
	}

	updates := map[string]interface{}{
		"name":        m.Name,
		"year":        m.Year,
		"date":        m.Date,
		"type":        m.Type,
		"branch_id":   m.BranchID,
		"description": m.Description,
	}
	if err := db.Model(&model.Holiday{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return model.Holiday{}, err
	}
	m.ID = id
	return m, nil
}

func (r *holidayRepository) DeleteHoliday(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	if err := db.Where("id = ?", id).Delete(&model.Holiday{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *holidayRepository) GetBranchMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var branchMeta []dto.Meta
	if err := db.Raw(`
		SELECT id::TEXT AS id, name
		FROM branches
		WHERE deleted_at IS NULL
	`).Scan(&branchMeta).Error; err != nil {
		return nil, err
	}
	return branchMeta, nil
}
