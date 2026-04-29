package repository

import (
	"context"
	"errors"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"

	"gorm.io/gorm"
)

type BranchRepository interface {
	GetAllBranches(ctx context.Context, params dto.BranchListParams) (dto.PaginatedResponse[dto.BranchResponse], error)
	GetBranchByID(ctx context.Context, id string) (dto.BranchResponse, error)
	CreateBranch(ctx context.Context, req model.Branch) (model.Branch, error)
	UpdateBranch(ctx context.Context, id string, req model.Branch) (model.Branch, error)
	DeleteBranch(ctx context.Context, id string) error
}

type branchRepository struct {
	db *gorm.DB
}

func NewBranchRepository(db *gorm.DB) BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) GetAllBranches(ctx context.Context, params dto.BranchListParams) (dto.PaginatedResponse[dto.BranchResponse], error) {
	db := r.db.WithContext(ctx)

	baseQuery := `
		FROM branches
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}

	if params.Search != nil && *params.Search != "" {
		baseQuery += " AND (name ILIKE ? OR code ILIKE ?)"
		like := "%" + *params.Search + "%"
		args = append(args, like, like)
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.BranchResponse]{}, err
	}

	selectQuery := `
		SELECT
			id, code, name, address, latitude, longitude,
			radius_meters, allow_wfh, created_at, updated_at
	` + baseQuery

	selectQuery += utils.BuildSortClause("branches", params.SortBy, params.GetSortDir(), "name ASC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var branches []dto.BranchResponse
	if err := db.Raw(selectQuery, args...).Scan(&branches).Error; err != nil {
		return dto.PaginatedResponse[dto.BranchResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.BranchResponse]{
		Data: branches,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *branchRepository) GetBranchByID(ctx context.Context, id string) (dto.BranchResponse, error) {
	var branch dto.BranchResponse
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			id, code, name, address, latitude, longitude,
			radius_meters, allow_wfh, created_at, updated_at
		FROM branches
		WHERE deleted_at IS NULL AND id = ?
	`, id).Scan(&branch).Error; err != nil {
		return dto.BranchResponse{}, err
	}
	if branch.ID == 0 {
		return dto.BranchResponse{}, errors.New("branch not found")
	}
	return branch, nil
}

func (r *branchRepository) CreateBranch(ctx context.Context, req model.Branch) (model.Branch, error) {
	if err := r.db.WithContext(ctx).Create(&req).Error; err != nil {
		return model.Branch{}, err
	}
	return req, nil
}

func (r *branchRepository) UpdateBranch(ctx context.Context, id string, req model.Branch) (model.Branch, error) {
	if err := r.db.WithContext(ctx).Model(&req).Where("id = ?", id).Updates(req).Error; err != nil {
		return model.Branch{}, err
	}
	return req, nil
}

func (r *branchRepository) DeleteBranch(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Branch{}).Error; err != nil {
		return err
	}
	return nil
}
