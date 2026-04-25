package repository

import (
	"context"
	"errors"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error)
	GetRoleByID(ctx context.Context, id string) (dto.RoleDetailResponse, error)
	CreateRole(ctx context.Context, req model.Role) (model.Role, error)
	UpdateRole(ctx context.Context, id string, req model.Role) (model.Role, error)
	DeleteRole(ctx context.Context, id string) error
	GetAllPermissions(ctx context.Context) ([]dto.PermissionResponse, error)
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]dto.PermissionResponse, error)
	DeleteRolePermissions(ctx context.Context, tx Transaction, roleID uint) error
	CreateRolePermissions(ctx context.Context, tx Transaction, perms []model.RolePermission) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *roleRepository) GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	var roles []dto.RoleResponse
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			r.id, r.name, r.level, r.description,
			COUNT(rp.id) AS permission_count,
			r.created_at, r.updated_at
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		WHERE r.deleted_at IS NULL
		GROUP BY r.id
		ORDER BY r.name ASC
	`).Scan(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) GetRoleByID(ctx context.Context, id string) (dto.RoleDetailResponse, error) {
	var role dto.RoleResponse
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			r.id, r.name, r.level, r.description,
			COUNT(rp.id) AS permission_count,
			r.created_at, r.updated_at
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		WHERE r.deleted_at IS NULL AND r.id = ?
		GROUP BY r.id
	`, id).Scan(&role).Error; err != nil {
		return dto.RoleDetailResponse{}, err
	}
	if role.ID == 0 {
		return dto.RoleDetailResponse{}, errors.New("role not found")
	}

	perms, err := r.GetPermissionsByRoleID(ctx, id)
	if err != nil {
		return dto.RoleDetailResponse{}, err
	}

	return dto.RoleDetailResponse{
		RoleResponse: role,
		Permissions:  perms,
	}, nil
}

func (r *roleRepository) CreateRole(ctx context.Context, req model.Role) (model.Role, error) {
	if err := r.db.WithContext(ctx).Create(&req).Error; err != nil {
		return model.Role{}, err
	}
	return req, nil
}

func (r *roleRepository) UpdateRole(ctx context.Context, id string, req model.Role) (model.Role, error) {
	if err := r.db.WithContext(ctx).Model(&req).Where("id = ?", id).Updates(req).Error; err != nil {
		return model.Role{}, err
	}
	return req, nil
}

func (r *roleRepository) DeleteRole(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Role{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *roleRepository) GetAllPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	var perms []dto.PermissionResponse
	if err := r.db.WithContext(ctx).Raw(`
		SELECT code, module, action, description, created_at, updated_at
		FROM permissions
		WHERE deleted_at IS NULL
		ORDER BY module ASC, action ASC
	`).Scan(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *roleRepository) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]dto.PermissionResponse, error) {
	var perms []dto.PermissionResponse
	if err := r.db.WithContext(ctx).Raw(`
		SELECT p.code, p.module, p.action, p.description, p.created_at, p.updated_at
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_code = p.code
		WHERE p.deleted_at IS NULL AND rp.role_id = ?
		ORDER BY p.module ASC, p.action ASC
	`, roleID).Scan(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *roleRepository) DeleteRolePermissions(ctx context.Context, tx Transaction, roleID uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	if err := db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *roleRepository) CreateRolePermissions(ctx context.Context, tx Transaction, perms []model.RolePermission) error {
	if len(perms) == 0 {
		return nil
	}
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	if err := db.Create(&perms).Error; err != nil {
		return err
	}
	return nil
}
