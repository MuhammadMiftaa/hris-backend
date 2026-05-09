package repository

import (
	"context"
	"errors"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"

	"gorm.io/gorm"
)

type LeaveTypeRepository interface {
	GetAllLeaveTypes(ctx context.Context, params dto.LeaveTypeListParams) (dto.PaginatedResponse[dto.LeaveTypeResponse], error)
	GetLeaveTypeByID(ctx context.Context, id string) (dto.LeaveTypeResponse, error)
	CreateLeaveType(ctx context.Context, req model.LeaveType) (model.LeaveType, error)
	UpdateLeaveType(ctx context.Context, id string, req model.LeaveType) (model.LeaveType, error)
	DeleteLeaveType(ctx context.Context, id string) error
}

type leaveTypeRepository struct {
	db *gorm.DB
}

func NewLeaveTypeRepository(db *gorm.DB) LeaveTypeRepository {
	return &leaveTypeRepository{db: db}
}

func (r *leaveTypeRepository) GetAllLeaveTypes(ctx context.Context, params dto.LeaveTypeListParams) (dto.PaginatedResponse[dto.LeaveTypeResponse], error) {
	baseQuery := `
		FROM leave_types lt
		LEFT JOIN leave_types parent_lt ON parent_lt.id = lt.parent_leave_type_id AND parent_lt.deleted_at IS NULL
		WHERE lt.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.Name != nil && *params.Name != "" {
		baseQuery += " AND lt.name ILIKE ?"
		args = append(args, "%"+*params.Name+"%")
	}
	if params.Category != nil && *params.Category != "" {
		baseQuery += " AND lt.category = ?"
		args = append(args, *params.Category)
	}
	if params.RequiresDocument != nil {
		if *params.RequiresDocument {
			baseQuery += " AND lt.requires_document = TRUE"
		} else {
			baseQuery += " AND lt.requires_document = FALSE"
		}
	}
	if params.MaxTotalDurationPerYear != nil && *params.MaxTotalDurationPerYear != "" {
		baseQuery += " AND lt.max_total_duration_per_year::TEXT LIKE ?"
		args = append(args, "%"+*params.MaxTotalDurationPerYear+"%")
	}

	if params.MaxDurationPerRequest != nil && *params.MaxDurationPerRequest != "" {
		baseQuery += " AND lt.max_duration_per_request::TEXT LIKE ?"
		args = append(args, "%"+*params.MaxDurationPerRequest+"%")
	}

	if params.ParentLeaveTypeID != nil {
		baseQuery += " AND lt.parent_leave_type_id = ?"
		args = append(args, *params.ParentLeaveTypeID)
	}

	if params.ParentLeaveTypeName != nil && *params.ParentLeaveTypeName != "" {
		baseQuery += " AND parent_lt.name ILIKE ?"
		args = append(args, "%"+*params.ParentLeaveTypeName+"%")
	}

	var total int
	if err := r.db.WithContext(ctx).Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.LeaveTypeResponse]{}, err
	}

	selectQuery := `
		SELECT
			lt.id, lt.name, lt.category, lt.requires_document, lt.requires_document_type,
			lt.max_duration_per_request,
			lt.max_occurrences_per_year,
			lt.max_total_duration_per_year,
			lt.max_per_month, lt.parent_leave_type_id, lt.deduct_days,
			lt.created_at, lt.updated_at,
			parent_lt.name AS parent_leave_type_name
	` + baseQuery

	selectQuery += utils.BuildSortClause("leave_type", params.SortBy, params.GetSortDir(), "lt.name ASC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var leaveTypes []dto.LeaveTypeResponse
	if err := r.db.WithContext(ctx).Raw(selectQuery, args...).Scan(&leaveTypes).Error; err != nil {
		return dto.PaginatedResponse[dto.LeaveTypeResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.LeaveTypeResponse]{
		Data: leaveTypes,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *leaveTypeRepository) GetLeaveTypeByID(ctx context.Context, id string) (dto.LeaveTypeResponse, error) {
	var lt dto.LeaveTypeResponse
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			lt.id, lt.name, lt.category, lt.requires_document, lt.requires_document_type,
			lt.max_duration_per_request,
			lt.max_occurrences_per_year,
			lt.max_total_duration_per_year,
			lt.max_per_month, lt.parent_leave_type_id, lt.deduct_days,
			lt.created_at, lt.updated_at,
			parent_lt.name AS parent_leave_type_name
		FROM leave_types lt
		LEFT JOIN leave_types parent_lt ON parent_lt.id = lt.parent_leave_type_id AND parent_lt.deleted_at IS NULL
		WHERE lt.deleted_at IS NULL AND lt.id = ?
	`, id).Scan(&lt).Error; err != nil {
		return dto.LeaveTypeResponse{}, err
	}
	if lt.ID == 0 {
		return dto.LeaveTypeResponse{}, errors.New("leave type not found")
	}
	return lt, nil
}

func (r *leaveTypeRepository) CreateLeaveType(ctx context.Context, req model.LeaveType) (model.LeaveType, error) {
	if err := r.db.WithContext(ctx).Create(&req).Error; err != nil {
		return model.LeaveType{}, err
	}
	return req, nil
}

func (r *leaveTypeRepository) UpdateLeaveType(ctx context.Context, id string, req model.LeaveType) (model.LeaveType, error) {
	if err := r.db.WithContext(ctx).Model(&req).Where("id = ?", id).Updates(req).Error; err != nil {
		return model.LeaveType{}, err
	}
	return req, nil
}

func (r *leaveTypeRepository) DeleteLeaveType(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.LeaveType{}).Error; err != nil {
		return err
	}
	return nil
}
