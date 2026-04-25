package service

import (
	"context"
	"fmt"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils/data"
)

type RoleService interface {
	GetMetadata(ctx context.Context) dto.RoleMetadata
	GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error)
	GetRoleByID(ctx context.Context, id string) (dto.RoleDetailResponse, error)
	CreateRole(ctx context.Context, req dto.CreateRoleRequest) (dto.RoleResponse, error)
	UpdateRole(ctx context.Context, id string, req dto.UpdateRoleRequest) (dto.RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error
	GetAllPermissions(ctx context.Context) ([]dto.PermissionResponse, error)
	UpdateRolePermissions(ctx context.Context, roleId string, req dto.UpdateRolePermissionsRequest) (dto.RoleDetailResponse, error)
}

type roleService struct {
	repo      repository.RoleRepository
	txManager repository.TxManager
}

func NewRoleService(repo repository.RoleRepository, txManager repository.TxManager) RoleService {
	return &roleService{repo: repo, txManager: txManager}
}

func (s *roleService) GetMetadata(ctx context.Context) dto.RoleMetadata {
	return dto.RoleMetadata{
		ModuleMeta: data.PermissionModuleMeta,
		ActionMeta: data.PermissionActionMeta,
		LevelMeta:  data.RoleLevelMeta,
	}
}

func (s *roleService) GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.repo.GetAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all roles: %w", err)
	}
	return roles, nil
}

func (s *roleService) GetRoleByID(ctx context.Context, id string) (dto.RoleDetailResponse, error) {
	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return dto.RoleDetailResponse{}, fmt.Errorf("get role by ID: %w", err)
	}
	return role, nil
}

func (s *roleService) CreateRole(ctx context.Context, req dto.CreateRoleRequest) (dto.RoleResponse, error) {
	role := model.Role{
		Name:        req.Name,
		Level:       model.RoleLevelEnum(req.Level),
		Description: req.Description,
	}

	created, err := s.repo.CreateRole(ctx, role)
	if err != nil {
		return dto.RoleResponse{}, fmt.Errorf("create role: %w", err)
	}

	return dto.RoleResponse{
		ID:          created.ID,
		Name:        created.Name,
		Level:       string(created.Level),
		Description: created.Description,
		CreatedAt:   created.CreatedAt,
		UpdatedAt:   created.UpdatedAt,
	}, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id string, req dto.UpdateRoleRequest) (dto.RoleResponse, error) {
	role := model.Role{}
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Level != nil {
		role.Level = model.RoleLevelEnum(*req.Level)
	}
	if req.Description != nil {
		role.Description = req.Description
	}

	_, err := s.repo.UpdateRole(ctx, id, role)
	if err != nil {
		return dto.RoleResponse{}, fmt.Errorf("update role: %w", err)
	}

	detail, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return dto.RoleResponse{}, fmt.Errorf("get updated role: %w", err)
	}
	return detail.RoleResponse, nil
}

func (s *roleService) DeleteRole(ctx context.Context, id string) error {
	if err := s.repo.DeleteRole(ctx, id); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (s *roleService) GetAllPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	perms, err := s.repo.GetAllPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all permissions: %w", err)
	}
	return perms, nil
}

func (s *roleService) UpdateRolePermissions(ctx context.Context, roleId string, req dto.UpdateRolePermissionsRequest) (dto.RoleDetailResponse, error) {
	// Validate the role exists
	existingRole, err := s.repo.GetRoleByID(ctx, roleId)
	if err != nil {
		return dto.RoleDetailResponse{}, fmt.Errorf("get role: %w", err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.RoleDetailResponse{}, fmt.Errorf("update role permissions: begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing permissions
	if err := s.repo.DeleteRolePermissions(ctx, tx, existingRole.ID); err != nil {
		return dto.RoleDetailResponse{}, fmt.Errorf("update role permissions: delete existing: %w", err)
	}

	// Create new permissions
	var perms []model.RolePermission
	for _, code := range req.PermissionCodes {
		perms = append(perms, model.RolePermission{
			RoleID:         existingRole.ID,
			PermissionCode: code,
		})
	}

	if err := s.repo.CreateRolePermissions(ctx, tx, perms); err != nil {
		return dto.RoleDetailResponse{}, fmt.Errorf("update role permissions: create new: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.RoleDetailResponse{}, fmt.Errorf("update role permissions: commit transaction: %w", err)
	}

	return s.repo.GetRoleByID(ctx, roleId)
}
