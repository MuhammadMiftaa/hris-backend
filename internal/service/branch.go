package service

import (
	"context"
	"fmt"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
)

type BranchService interface {
	GetAllBranches(ctx context.Context, params dto.BranchListParams) (dto.PaginatedResponse[dto.BranchResponse], error)
	GetBranchByID(ctx context.Context, id string) (dto.BranchResponse, error)
	CreateBranch(ctx context.Context, req dto.CreateBranchRequest) (dto.BranchResponse, error)
	UpdateBranch(ctx context.Context, id string, req dto.UpdateBranchRequest) (dto.BranchResponse, error)
	DeleteBranch(ctx context.Context, id string) error
}

type branchService struct {
	repo repository.BranchRepository
}

func NewBranchService(repo repository.BranchRepository) BranchService {
	return &branchService{repo: repo}
}

func (s *branchService) GetAllBranches(ctx context.Context, params dto.BranchListParams) (dto.PaginatedResponse[dto.BranchResponse], error) {
	branches, err := s.repo.GetAllBranches(ctx, params)
	if err != nil {
		return dto.PaginatedResponse[dto.BranchResponse]{}, fmt.Errorf("get all branches: %w", err)
	}
	return branches, nil
}

func (s *branchService) GetBranchByID(ctx context.Context, id string) (dto.BranchResponse, error) {
	branch, err := s.repo.GetBranchByID(ctx, id)
	if err != nil {
		return dto.BranchResponse{}, fmt.Errorf("get branch by ID: %w", err)
	}
	return branch, nil
}

func (s *branchService) CreateBranch(ctx context.Context, req dto.CreateBranchRequest) (dto.BranchResponse, error) {
	branch := model.Branch{
		Code:    req.Code,
		Name:    req.Name,
		Address: req.Address,
	}
	if req.Latitude != nil {
		branch.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		branch.Longitude = req.Longitude
	}
	if req.RadiusMeters != nil {
		branch.RadiusMeters = *req.RadiusMeters
	}
	if req.AllowWFH != nil {
		branch.AllowWFH = *req.AllowWFH
	}

	created, err := s.repo.CreateBranch(ctx, branch)
	if err != nil {
		return dto.BranchResponse{}, fmt.Errorf("create branch: %w", err)
	}

	return dto.BranchResponse{
		ID:           created.ID,
		Code:         created.Code,
		Name:         created.Name,
		Address:      created.Address,
		Latitude:     created.Latitude,
		Longitude:    created.Longitude,
		RadiusMeters: created.RadiusMeters,
		AllowWFH:     created.AllowWFH,
		CreatedAt:    created.CreatedAt,
		UpdatedAt:    created.UpdatedAt,
	}, nil
}

func (s *branchService) UpdateBranch(ctx context.Context, id string, req dto.UpdateBranchRequest) (dto.BranchResponse, error) {
	branch := model.Branch{}
	if req.Code != nil {
		branch.Code = *req.Code
	}
	if req.Name != nil {
		branch.Name = *req.Name
	}
	if req.Address != nil {
		branch.Address = req.Address
	}
	if req.Latitude != nil {
		branch.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		branch.Longitude = req.Longitude
	}
	if req.RadiusMeters != nil {
		branch.RadiusMeters = *req.RadiusMeters
	}
	if req.AllowWFH != nil {
		branch.AllowWFH = *req.AllowWFH
	}

	_, err := s.repo.UpdateBranch(ctx, id, branch)
	if err != nil {
		return dto.BranchResponse{}, fmt.Errorf("update branch: %w", err)
	}

	return s.repo.GetBranchByID(ctx, id)
}

func (s *branchService) DeleteBranch(ctx context.Context, id string) error {
	if err := s.repo.DeleteBranch(ctx, id); err != nil {
		return fmt.Errorf("delete branch: %w", err)
	}
	return nil
}
