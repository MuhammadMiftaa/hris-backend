package repository

import (
	"context"
	"errors"
	"time"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type AuthRepository interface {
	GetAccountByEmail(ctx context.Context, tx Transaction, email string) (dto.GetAccountByEmailResponse, error)
	GetEmployeeByID(ctx context.Context, tx Transaction, id uint) (dto.GetEmployeeByIDResponse, error)
	UpdateAccountLastLogin(ctx context.Context, tx Transaction, time time.Time, id uint) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *authRepository) GetAccountByEmail(ctx context.Context, tx Transaction, email string) (dto.GetAccountByEmailResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.GetAccountByEmailResponse{}, err
	}

	var account dto.GetAccountByEmailResponse

	err = db.Raw(`
		SELECT 
			a.id,
			a.email,
			a.password,
			a.is_active,
			a.last_login_at
		FROM accounts a
		WHERE a.email = $1 AND a.deleted_at IS NULL
	`, email).Scan(&account).Error
	if err != nil {
		return account, err
	}

	return account, nil
}

func (r *authRepository) GetEmployeeByID(ctx context.Context, tx Transaction, id uint) (dto.GetEmployeeByIDResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.GetEmployeeByIDResponse{}, err
	}

	var user dto.GetEmployeeByIDResponse

	err = db.Raw(`
        SELECT 
            a.id AS account_id,
            e.id AS employee_id,
            COALESCE(a.email, '') AS email,
            COALESCE(a.is_active, false) AS is_active,
            COALESCE(a.last_login_at, NULL) AS last_login_at,
            COALESCE(e.employee_number, '') AS employee_number,
            COALESCE(e.full_name, '') AS full_name,
            COALESCE(e.photo_url, '') AS photo_url,
            COALESCE(e.is_trainer, false) AS is_trainer,
            COALESCE(e.branch_id, 0) AS branch_id,
            COALESCE(e.department_id, 0) AS department_id,
            COALESCE(e.job_positions_id, 0) AS job_positions_id,
            COALESCE(r.name, '') AS role_name,
            COALESCE(r.level::TEXT, 'staff') AS role_level,
            COALESCE(
                JSONB_AGG(DISTINCT p.code ORDER BY p.code ASC) FILTER (WHERE p.code IS NOT NULL), 
                '[]'::jsonb
            ) AS permissions
        FROM accounts a
        JOIN roles r ON a.role_id = r.id
        JOIN employees e ON a.employee_id = e.id
        LEFT JOIN role_permissions rp ON a.role_id = rp.role_id
        LEFT JOIN permissions p ON rp.permission_code = p.code
        WHERE a.id = $1 AND a.deleted_at IS NULL AND e.deleted_at IS NULL AND r.deleted_at IS NULL
        GROUP BY 
            a.id, a.email, a.is_active, a.last_login_at,
            e.id, e.employee_number, e.full_name, e.photo_url, 
            e.is_trainer, e.branch_id, e.department_id, e.job_positions_id,
            r.id, r.name, r.level
    `, id).Scan(&user).Error
	if err != nil {
		return user, err
	}

	return user, nil
}

func (r *authRepository) UpdateAccountLastLogin(ctx context.Context, tx Transaction, time time.Time, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	if err := db.Model(&model.Account{}).Where("id = ?", id).Update("last_login_at", time).Error; err != nil {
		return err
	}

	return nil
}
