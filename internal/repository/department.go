package repository

import (
	"context"
	"errors"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"

	"gorm.io/gorm"
)

type DepartmentRepository interface {
	GetBranchMetadata(ctx context.Context) ([]dto.Meta, error)
	GetAllDepartments(ctx context.Context, params dto.DepartmentListParams) (dto.PaginatedResponse[dto.DepartmentResponse], error)
	GetDepartmentByID(ctx context.Context, id string) (dto.DepartmentResponse, error)
	CreateDepartment(ctx context.Context, req model.Department) (model.Department, error)
	UpdateDepartment(ctx context.Context, id string, req model.Department) (model.Department, error)
	DeleteDepartment(ctx context.Context, id string) error
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) GetBranchMetadata(ctx context.Context) ([]dto.Meta, error) {
	var meta []dto.Meta
	if err := r.db.WithContext(ctx).Raw(`
		SELECT id::TEXT AS id, name
		FROM branches
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`).Scan(&meta).Error; err != nil {
		return nil, err
	}
	return meta, nil
}

func (r *departmentRepository) GetAllDepartments(ctx context.Context, params dto.DepartmentListParams) (dto.PaginatedResponse[dto.DepartmentResponse], error) {
	db := r.db.WithContext(ctx)

	baseQuery := `
		FROM departments d
		LEFT JOIN branches b ON b.id = d.branch_id AND b.deleted_at IS NULL
		WHERE d.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.BranchID != nil {
		baseQuery += " AND d.branch_id = ?"
		args = append(args, *params.BranchID)
	}
	if params.IsActive != nil {
		baseQuery += " AND d.is_active = ?"
		args = append(args, *params.IsActive)
	}
	if params.Search != nil && *params.Search != "" {
		baseQuery += " AND (d.name ILIKE ? OR d.code ILIKE ?)"
		like := "%" + *params.Search + "%"
		args = append(args, like, like)
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.DepartmentResponse]{}, err
	}

	selectQuery := `
		SELECT
			d.id, d.code, d.name, d.branch_id,
			b.name AS branch_name,
			d.description, d.is_active,
			d.created_at, d.updated_at
	` + baseQuery

	selectQuery += utils.BuildSortClause("departments", params.SortBy, params.GetSortDir(), "d.name ASC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var departments []dto.DepartmentResponse
	if err := db.Raw(selectQuery, args...).Scan(&departments).Error; err != nil {
		return dto.PaginatedResponse[dto.DepartmentResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.DepartmentResponse]{
		Data: departments,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *departmentRepository) GetDepartmentByID(ctx context.Context, id string) (dto.DepartmentResponse, error) {
	var dept dto.DepartmentResponse
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			d.id, d.code, d.name, d.branch_id,
			b.name AS branch_name,
			d.description, d.is_active,
			d.created_at, d.updated_at
		FROM departments d
		LEFT JOIN branches b ON b.id = d.branch_id AND b.deleted_at IS NULL
		WHERE d.deleted_at IS NULL AND d.id = ?
	`, id).Scan(&dept).Error; err != nil {
		return dto.DepartmentResponse{}, err
	}
	if dept.ID == 0 {
		return dto.DepartmentResponse{}, errors.New("department not found")
	}
	return dept, nil
}

func (r *departmentRepository) CreateDepartment(ctx context.Context, req model.Department) (model.Department, error) {
	if err := r.db.WithContext(ctx).Create(&req).Error; err != nil {
		return model.Department{}, err
	}
	return req, nil
}

func (r *departmentRepository) UpdateDepartment(ctx context.Context, id string, req model.Department) (model.Department, error) {
	if err := r.db.WithContext(ctx).Model(&req).Where("id = ?", id).Updates(req).Error; err != nil {
		return model.Department{}, err
	}
	return req, nil
}

func (r *departmentRepository) DeleteDepartment(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Department{}).Error; err != nil {
		return err
	}
	return nil
}
