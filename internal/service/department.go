package service

import (
	"context"
	"fmt"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
)

type DepartmentService interface {
	GetMetadata(ctx context.Context) (dto.DepartmentMetadata, error)
	GetAllDepartments(ctx context.Context, params dto.DepartmentListParams) ([]dto.DepartmentResponse, error)
	GetDepartmentByID(ctx context.Context, id string) (dto.DepartmentResponse, error)
	CreateDepartment(ctx context.Context, req dto.CreateDepartmentRequest) (dto.DepartmentResponse, error)
	UpdateDepartment(ctx context.Context, id string, req dto.UpdateDepartmentRequest) (dto.DepartmentResponse, error)
	DeleteDepartment(ctx context.Context, id string) error
}

type departmentService struct {
	repo repository.DepartmentRepository
}

func NewDepartmentService(repo repository.DepartmentRepository) DepartmentService {
	return &departmentService{repo: repo}
}

func (s *departmentService) GetMetadata(ctx context.Context) (dto.DepartmentMetadata, error) {
	branchMeta, err := s.repo.GetBranchMetadata(ctx)
	if err != nil {
		return dto.DepartmentMetadata{}, fmt.Errorf("get branch metadata: %w", err)
	}
	return dto.DepartmentMetadata{BranchMeta: branchMeta}, nil
}

func (s *departmentService) GetAllDepartments(ctx context.Context, params dto.DepartmentListParams) ([]dto.DepartmentResponse, error) {
	departments, err := s.repo.GetAllDepartments(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get all departments: %w", err)
	}
	return departments, nil
}

func (s *departmentService) GetDepartmentByID(ctx context.Context, id string) (dto.DepartmentResponse, error) {
	dept, err := s.repo.GetDepartmentByID(ctx, id)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("get department by ID: %w", err)
	}
	return dept, nil
}

func (s *departmentService) CreateDepartment(ctx context.Context, req dto.CreateDepartmentRequest) (dto.DepartmentResponse, error) {
	dept := model.Department{
		Code:        req.Code,
		Name:        req.Name,
		BranchID:    req.BranchID,
		Description: req.Description,
	}

	created, err := s.repo.CreateDepartment(ctx, dept)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("create department: %w", err)
	}

	return s.repo.GetDepartmentByID(ctx, fmt.Sprintf("%d", created.ID))
}

func (s *departmentService) UpdateDepartment(ctx context.Context, id string, req dto.UpdateDepartmentRequest) (dto.DepartmentResponse, error) {
	dept := model.Department{}
	if req.Name != nil {
		dept.Name = *req.Name
	}
	if req.BranchID != nil {
		dept.BranchID = req.BranchID
	}
	if req.Description != nil {
		dept.Description = req.Description
	}
	if req.IsActive != nil {
		dept.IsActive = req.IsActive
	}

	_, err := s.repo.UpdateDepartment(ctx, id, dept)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("update department: %w", err)
	}

	return s.repo.GetDepartmentByID(ctx, id)
}

func (s *departmentService) DeleteDepartment(ctx context.Context, id string) error {
	if err := s.repo.DeleteDepartment(ctx, id); err != nil {
		return fmt.Errorf("delete department: %w", err)
	}
	return nil
}
