package service

import (
	"context"
	"fmt"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils/data"
)

type LeaveTypeService interface {
	GetMetadata(ctx context.Context) dto.LeaveTypeMetadata
	GetAllLeaveTypes(ctx context.Context, params dto.LeaveTypeListParams) (dto.PaginatedResponse[dto.LeaveTypeResponse], error)
	GetLeaveTypeByID(ctx context.Context, id string) (dto.LeaveTypeResponse, error)
	CreateLeaveType(ctx context.Context, req dto.CreateLeaveTypeRequest) (dto.LeaveTypeResponse, error)
	UpdateLeaveType(ctx context.Context, id string, req dto.UpdateLeaveTypeRequest) (dto.LeaveTypeResponse, error)
	DeleteLeaveType(ctx context.Context, id string) error
	ExportLeaveTypes(ctx context.Context, params dto.LeaveTypeListParams) (dto.PaginatedResponse[dto.LeaveTypeResponse], error)
}

type leaveTypeService struct {
	repo repository.LeaveTypeRepository
}

func NewLeaveTypeService(repo repository.LeaveTypeRepository) LeaveTypeService {
	return &leaveTypeService{repo: repo}
}

func (s *leaveTypeService) GetMetadata(ctx context.Context) dto.LeaveTypeMetadata {
	return dto.LeaveTypeMetadata{
		CategoryMeta:     data.LeaveCategoryMeta,
	}
}

func (s *leaveTypeService) GetAllLeaveTypes(ctx context.Context, params dto.LeaveTypeListParams) (dto.PaginatedResponse[dto.LeaveTypeResponse], error) {
	result, err := s.repo.GetAllLeaveTypes(ctx, params)
	if err != nil {
		return dto.PaginatedResponse[dto.LeaveTypeResponse]{}, fmt.Errorf("get all leave types: %w", err)
	}
	return result, nil
}

func (s *leaveTypeService) GetLeaveTypeByID(ctx context.Context, id string) (dto.LeaveTypeResponse, error) {
	lt, err := s.repo.GetLeaveTypeByID(ctx, id)
	if err != nil {
		return dto.LeaveTypeResponse{}, fmt.Errorf("get leave type by ID: %w", err)
	}
	return lt, nil
}

func (s *leaveTypeService) CreateLeaveType(ctx context.Context, req dto.CreateLeaveTypeRequest) (dto.LeaveTypeResponse, error) {
	category := model.LeaveCategoryEnum(req.Category)

	lt := model.LeaveType{
		Name:                    req.Name,
		Category:                category,
		RequiresDocument:        req.RequiresDocument,
		RequiresDocumentType:    req.RequiresDocumentType,
		MaxDurationPerRequest:   req.MaxDurationPerRequest,
		MaxOccurrencesPerYear:   req.MaxOccurrencesPerYear,
		MaxTotalDurationPerYear: req.MaxTotalDurationPerYear,
		MaxPerMonth:             req.MaxPerMonth,
		ParentLeaveTypeID:       req.ParentLeaveTypeID,
		DeductDays:              1.0,
	}

	if req.DeductDays != nil {
		lt.DeductDays = *req.DeductDays
	}

	created, err := s.repo.CreateLeaveType(ctx, lt)
	if err != nil {
		return dto.LeaveTypeResponse{}, fmt.Errorf("create leave type: %w", err)
	}

	return dto.LeaveTypeResponse{
		ID:                      created.ID,
		Name:                    created.Name,
		Category:                created.Category,
		RequiresDocument:        created.RequiresDocument,
		RequiresDocumentType:    created.RequiresDocumentType,
		MaxDurationPerRequest:   created.MaxDurationPerRequest,
		MaxOccurrencesPerYear:   created.MaxOccurrencesPerYear,
		MaxTotalDurationPerYear: created.MaxTotalDurationPerYear,
		MaxPerMonth:             created.MaxPerMonth,
		ParentLeaveTypeID:       created.ParentLeaveTypeID,
		DeductDays:              created.DeductDays,
		CreatedAt:               created.CreatedAt,
		UpdatedAt:               created.UpdatedAt,
	}, nil
}

func (s *leaveTypeService) UpdateLeaveType(ctx context.Context, id string, req dto.UpdateLeaveTypeRequest) (dto.LeaveTypeResponse, error) {
	lt := model.LeaveType{}
	if req.Name != nil {
		lt.Name = *req.Name
	}
	if req.Category != nil {
		lt.Category = model.LeaveCategoryEnum(*req.Category)
	}
	if req.RequiresDocument != nil {
		lt.RequiresDocument = *req.RequiresDocument
	}
	if req.RequiresDocumentType != nil {
		lt.RequiresDocumentType = req.RequiresDocumentType
	}
	if req.MaxDurationPerRequest != nil {
		lt.MaxDurationPerRequest = req.MaxDurationPerRequest
	}
	if req.MaxOccurrencesPerYear != nil {
		lt.MaxOccurrencesPerYear = req.MaxOccurrencesPerYear
	}
	if req.MaxTotalDurationPerYear != nil {
		lt.MaxTotalDurationPerYear = req.MaxTotalDurationPerYear
	}
	if req.MaxPerMonth != nil {
		lt.MaxPerMonth = req.MaxPerMonth
	}
	if req.ParentLeaveTypeID != nil {
		lt.ParentLeaveTypeID = req.ParentLeaveTypeID
	}
	if req.DeductDays != nil {
		lt.DeductDays = *req.DeductDays
	}

	_, err := s.repo.UpdateLeaveType(ctx, id, lt)
	if err != nil {
		return dto.LeaveTypeResponse{}, fmt.Errorf("update leave type: %w", err)
	}

	return s.repo.GetLeaveTypeByID(ctx, id)
}

func (s *leaveTypeService) DeleteLeaveType(ctx context.Context, id string) error {
	if err := s.repo.DeleteLeaveType(ctx, id); err != nil {
		return fmt.Errorf("delete leave type: %w", err)
	}
	return nil
}

func (s *leaveTypeService) ExportLeaveTypes(ctx context.Context, params dto.LeaveTypeListParams) (dto.PaginatedResponse[dto.LeaveTypeResponse], error) {
	allPerPage := -1
	params.PerPage = &allPerPage
	return s.repo.GetAllLeaveTypes(ctx, params)
}
