package repository

import (
	"context"
	"errors"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type PositionRepository interface {
	GetDepartmentMetadata(ctx context.Context) ([]dto.Meta, error)
	GetAllPositions(ctx context.Context, departmentID *uint) ([]dto.PositionResponse, error)
	GetPositionByID(ctx context.Context, id string) (dto.PositionResponse, error)
	CreatePosition(ctx context.Context, req model.JobPosition) (model.JobPosition, error)
	UpdatePosition(ctx context.Context, id string, req model.JobPosition) (model.JobPosition, error)
	DeletePosition(ctx context.Context, id string) error
}

type positionRepository struct {
	db *gorm.DB
}

func NewPositionRepository(db *gorm.DB) PositionRepository {
	return &positionRepository{db: db}
}

func (r *positionRepository) GetDepartmentMetadata(ctx context.Context) ([]dto.Meta, error) {
	var meta []dto.Meta
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			b.id::TEXT AS parent_id,
			d.id::TEXT AS id,
			d.name
		FROM departments d
		JOIN branches b ON b.id = d.branch_id
		WHERE d.deleted_at IS NULL AND b.deleted_at IS NULL
		ORDER BY d.name ASC
	`).Scan(&meta).Error; err != nil {
		return nil, err
	}
	return meta, nil
}

func (r *positionRepository) GetAllPositions(ctx context.Context, departmentID *uint) ([]dto.PositionResponse, error) {
	query := `
		SELECT
			jp.id, jp.title, jp.department_id,
			d.name AS department_name,
			jp.created_at, jp.updated_at
		FROM job_positions jp
		LEFT JOIN departments d ON d.id = jp.department_id AND d.deleted_at IS NULL
		WHERE jp.deleted_at IS NULL
	`
	args := []interface{}{}

	if departmentID != nil {
		query += " AND jp.department_id = ?"
		args = append(args, *departmentID)
	}

	query += " ORDER BY jp.title ASC"

	var positions []dto.PositionResponse
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&positions).Error; err != nil {
		return nil, err
	}
	return positions, nil
}

func (r *positionRepository) GetPositionByID(ctx context.Context, id string) (dto.PositionResponse, error) {
	var pos dto.PositionResponse
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			jp.id, jp.title, jp.department_id,
			d.name AS department_name,
			jp.created_at, jp.updated_at
		FROM job_positions jp
		LEFT JOIN departments d ON d.id = jp.department_id AND d.deleted_at IS NULL
		WHERE jp.deleted_at IS NULL AND jp.id = ?
	`, id).Scan(&pos).Error; err != nil {
		return dto.PositionResponse{}, err
	}
	if pos.ID == 0 {
		return dto.PositionResponse{}, errors.New("position not found")
	}
	return pos, nil
}

func (r *positionRepository) CreatePosition(ctx context.Context, req model.JobPosition) (model.JobPosition, error) {
	if err := r.db.WithContext(ctx).Create(&req).Error; err != nil {
		return model.JobPosition{}, err
	}
	return req, nil
}

func (r *positionRepository) UpdatePosition(ctx context.Context, id string, req model.JobPosition) (model.JobPosition, error) {
	if err := r.db.WithContext(ctx).Model(&req).Where("id = ?", id).Updates(req).Error; err != nil {
		return model.JobPosition{}, err
	}
	return req, nil
}

func (r *positionRepository) DeletePosition(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.JobPosition{}).Error; err != nil {
		return err
	}
	return nil
}
