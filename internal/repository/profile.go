package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type ProfileRepository interface {
	GetEmployeeProfileByAccountID(ctx context.Context, tx Transaction, accountID uint) (dto.EmployeeProfileResponse, error)
	GetContactsByEmployeeID(ctx context.Context, tx Transaction, employeeID uint) ([]dto.EmployeeProfileContactResponse, error)
	GetAccountByID(ctx context.Context, tx Transaction, accountID uint) (model.Account, error)
	GetEmployeeByAccountID(ctx context.Context, tx Transaction, accountID uint) (model.Employee, error)
	UpdatePasswordByAccountID(ctx context.Context, tx Transaction, accountID uint, hashedPassword string) error
	UpdatePhotoURL(ctx context.Context, tx Transaction, employeeID uint, photoURL *string) error
	UpdateFullName(ctx context.Context, tx Transaction, employeeID uint, fullName string) error
}

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *profileRepository) GetEmployeeProfileByAccountID(ctx context.Context, tx Transaction, accountID uint) (dto.EmployeeProfileResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.EmployeeProfileResponse{}, err
	}

	var profile dto.EmployeeProfileResponse
	if err := db.Raw(`
		SELECT
			e.id,
			e.employee_number,
			e.full_name,
			e.photo_url,
			e.nik,
			e.npwp,
			e.kk_number,
			e.birth_date::TEXT AS birth_date,
			e.birth_place,
			e.gender::TEXT AS gender,
			e.religion,
			e.marital_status::TEXT AS marital_status,
			e.blood_type,
			e.nationality,
			e.branch_id,
			e.department_id,
			a.role_id,
			e.job_positions_id,
			b.name AS branch_name,
			d.name AS department_name,
			r.name AS role_name,
			jp.title AS job_position_title,
			e.created_at,
			e.updated_at
		FROM accounts a
		JOIN employees e ON a.employee_id = e.id AND e.deleted_at IS NULL
		LEFT JOIN branches b ON b.id = e.branch_id AND b.deleted_at IS NULL
		LEFT JOIN departments d ON d.id = e.department_id AND d.deleted_at IS NULL
		LEFT JOIN roles r ON r.id = a.role_id AND r.deleted_at IS NULL
		LEFT JOIN job_positions jp ON jp.id = e.job_positions_id AND jp.deleted_at IS NULL
		WHERE a.id = ? AND a.deleted_at IS NULL
	`, accountID).Scan(&profile).Error; err != nil {
		return dto.EmployeeProfileResponse{}, err
	}

	if profile.ID == 0 {
		return dto.EmployeeProfileResponse{}, fmt.Errorf("profile not found for account %d", accountID)
	}

	return profile, nil
}

func (r *profileRepository) GetContactsByEmployeeID(ctx context.Context, tx Transaction, employeeID uint) ([]dto.EmployeeProfileContactResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var contacts []dto.EmployeeProfileContactResponse
	if err := db.Raw(`
		SELECT id, contact_type, contact_value, contact_label, is_primary
		FROM employee_contacts
		WHERE employee_id = ? AND deleted_at IS NULL
		ORDER BY is_primary DESC, created_at DESC
	`, employeeID).Scan(&contacts).Error; err != nil {
		return nil, err
	}

	return contacts, nil
}

func (r *profileRepository) GetAccountByID(ctx context.Context, tx Transaction, accountID uint) (model.Account, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Account{}, err
	}

	var account model.Account
	if err := db.Where("id = ? AND deleted_at IS NULL", accountID).First(&account).Error; err != nil {
		return model.Account{}, err
	}

	return account, nil
}

func (r *profileRepository) GetEmployeeByAccountID(ctx context.Context, tx Transaction, accountID uint) (model.Employee, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Employee{}, err
	}

	var account model.Account
	if err := db.Where("id = ? AND deleted_at IS NULL", accountID).First(&account).Error; err != nil {
		return model.Employee{}, fmt.Errorf("account not found: %w", err)
	}

	var employee model.Employee
	if err := db.Where("id = ? AND deleted_at IS NULL", account.EmployeeID).First(&employee).Error; err != nil {
		return model.Employee{}, fmt.Errorf("employee not found: %w", err)
	}

	return employee, nil
}

func (r *profileRepository) UpdatePasswordByAccountID(ctx context.Context, tx Transaction, accountID uint, hashedPassword string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	result := db.Model(&model.Account{}).Where("id = ? AND deleted_at IS NULL", accountID).Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("account not found for ID %d", accountID)
	}
	return nil
}

func (r *profileRepository) UpdatePhotoURL(ctx context.Context, tx Transaction, employeeID uint, photoURL *string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	result := db.Model(&model.Employee{}).Where("id = ? AND deleted_at IS NULL", employeeID).Update("photo_url", photoURL)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("employee not found for ID %d", employeeID)
	}
	return nil
}

func (r *profileRepository) UpdateFullName(ctx context.Context, tx Transaction, employeeID uint, fullName string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	result := db.Model(&model.Employee{}).Where("id = ? AND deleted_at IS NULL", employeeID).Update("full_name", fullName)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("employee not found for ID %d", employeeID)
	}
	return nil
}
