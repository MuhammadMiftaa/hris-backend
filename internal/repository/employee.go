package repository

import (
	"context"
	"errors"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type EmployeeRepository interface {
	GetBranchMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetDepartmentMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetRoleMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetJobPositionMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error)

	GetAllEmployees(ctx context.Context, tx Transaction) ([]dto.Employee, error)
	GetEmployeeByID(ctx context.Context, tx Transaction, employeeID string) (dto.Employee, error)
	CreateEmployee(ctx context.Context, tx Transaction, req model.Employee) (model.Employee, error)
	UpdateEmployee(ctx context.Context, tx Transaction, id string, req model.Employee) (model.Employee, error)
	DeleteEmployee(ctx context.Context, tx Transaction, employeeID string) error

	GetAccountByEmployeeID(ctx context.Context, tx Transaction, employeeID string) (model.Account, error)
	CreateAccount(ctx context.Context, tx Transaction, req model.Account) (model.Account, error)
	UpdateAccount(ctx context.Context, tx Transaction, req model.Account) (model.Account, error)
	DeleteAccount(ctx context.Context, tx Transaction, accountID string) error
}

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{
		db: db,
	}
}

func (r *employeeRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

// ~ Metadata
func (r *employeeRepository) GetBranchMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var branchMeta []dto.Meta
	if err := db.Raw(`
		SELECT
			id,
			name
		FROM branches
		WHERE deleted_at IS NULL
	`).Scan(&branchMeta).Error; err != nil {
		return nil, err
	}

	return branchMeta, nil
}

func (r *employeeRepository) GetDepartmentMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var departmentMeta []dto.Meta
	if err := db.Raw(`
		SELECT
			id,
			name
		FROM departments
		WHERE deleted_at IS NULL
	`).Scan(&departmentMeta).Error; err != nil {
		return nil, err
	}

	return departmentMeta, nil
}

func (r *employeeRepository) GetRoleMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var roleMeta []dto.Meta
	if err := db.Raw(`
		SELECT
			id,
			name
		FROM roles
		WHERE deleted_at IS NULL
	`).Scan(&roleMeta).Error; err != nil {
		return nil, err
	}

	return roleMeta, nil
}

func (r *employeeRepository) GetJobPositionMetadata(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var jobPositionMeta []dto.Meta
	if err := db.Raw(`
		SELECT
			id,
			title as name
		FROM job_positions
		WHERE deleted_at IS NULL
	`).Scan(&jobPositionMeta).Error; err != nil {
		return nil, err
	}

	return jobPositionMeta, nil
}

// ~ Employee
func (r *employeeRepository) GetAllEmployees(ctx context.Context, tx Transaction) ([]dto.Employee, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var employees []dto.Employee
	if err := db.Raw(`
		SELECT
			e.id,
			e.employee_number,
			e.full_name,
			e.nik,
			e.npwp,
			e.kk_number,
			e.birth_date::TEXT         AS birth_date,
			e.birth_place,
			e.gender::TEXT             AS gender,
			e.religion,
			e.marital_status::TEXT     AS marital_status,
			e.blood_type,
			e.nationality,
			e.photo_url,
			a.is_active,
			e.is_trainer,
			e.branch_id,
			e.department_id,
			a.role_id,
			e.job_positions_id,
			e.created_at               AS created_at,
			e.updated_at               AS updated_at,
			e.deleted_at               AS deleted_at,
			b.name                     AS branch_name,
			d.name                     AS department_name,
			r.name                     AS role_name,
			jp.title                   AS job_position_title
		FROM employees e
		LEFT JOIN accounts      a  ON a.employee_id = e.id  AND a.deleted_at IS NULL
		LEFT JOIN branches      b  ON b.id = e.branch_id    AND b.deleted_at IS NULL
		LEFT JOIN departments   d  ON d.id = e.department_id AND d.deleted_at IS NULL
		LEFT JOIN roles         r  ON r.id = a.role_id       AND r.deleted_at IS NULL
		LEFT JOIN job_positions jp ON jp.id = e.job_positions_id AND jp.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
	`).Scan(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *employeeRepository) GetEmployeeByID(ctx context.Context, tx Transaction, employeeID string) (dto.Employee, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.Employee{}, err
	}

	var employee dto.Employee
	if err := db.Raw(`
		SELECT
			e.id,
			e.employee_number,
			e.full_name,
			e.nik,
			e.npwp,
			e.kk_number,
			e.birth_date::TEXT         AS birth_date,
			e.birth_place,
			e.gender::TEXT             AS gender,
			e.religion,
			e.marital_status::TEXT     AS marital_status,
			e.blood_type,
			e.nationality,
			e.photo_url,
			a.is_active,
			e.is_trainer,
			e.branch_id,
			e.department_id,
			a.role_id,
			e.job_positions_id,
			e.created_at               AS created_at,
			e.updated_at               AS updated_at,
			e.deleted_at               AS deleted_at,
			b.name                     AS branch_name,
			d.name                     AS department_name,
			r.name                     AS role_name,
			jp.title                   AS job_position_title
		FROM employees e
		LEFT JOIN accounts      a  ON a.employee_id = e.id  AND a.deleted_at IS NULL
		LEFT JOIN branches      b  ON b.id = e.branch_id    AND b.deleted_at IS NULL
		LEFT JOIN departments   d  ON d.id = e.department_id AND d.deleted_at IS NULL
		LEFT JOIN roles         r  ON r.id = a.role_id       AND r.deleted_at IS NULL
		LEFT JOIN job_positions jp ON jp.id = e.job_positions_id AND jp.deleted_at IS NULL
		WHERE e.deleted_at IS NULL AND e.id = ?
	`, employeeID).Scan(&employee).Error; err != nil {
		return dto.Employee{}, err
	}
	return employee, nil
}

func (r *employeeRepository) CreateEmployee(ctx context.Context, tx Transaction, req model.Employee) (model.Employee, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Employee{}, err
	}

	if err := db.Create(&req).Error; err != nil {
		return model.Employee{}, err
	}

	return req, nil
}

func (r *employeeRepository) UpdateEmployee(ctx context.Context, tx Transaction, id string, req model.Employee) (model.Employee, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Employee{}, err
	}

	if err := db.Model(&req).Where("id = ?", id).Updates(req).Error; err != nil {
		return model.Employee{}, err
	}

	return req, nil
}

func (r *employeeRepository) DeleteEmployee(ctx context.Context, tx Transaction, id string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	if err := db.Where("id = ?", id).Delete(&model.Employee{}).Error; err != nil {
		return err
	}

	return nil
}

// ~ Account
func (r *employeeRepository) GetAccountByEmployeeID(ctx context.Context, tx Transaction, employeeID string) (model.Account, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Account{}, err
	}

	var account model.Account
	if err := db.Where("employee_id = ?", employeeID).First(&account).Error; err != nil {
		return model.Account{}, err
	}

	return account, nil
}

func (r *employeeRepository) CreateAccount(ctx context.Context, tx Transaction, req model.Account) (model.Account, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Account{}, err
	}

	if err := db.Create(&req).Error; err != nil {
		return model.Account{}, err
	}

	return req, nil
}

func (r *employeeRepository) UpdateAccount(ctx context.Context, tx Transaction, req model.Account) (model.Account, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.Account{}, err
	}

	if err := db.Model(&req).Where("employee_id = ?", req.EmployeeID).Updates(req).Error; err != nil {
		return model.Account{}, err
	}

	return req, nil
}

func (r *employeeRepository) DeleteAccount(ctx context.Context, tx Transaction, accountID string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	if err := db.Where("id = ?", accountID).Delete(&model.Account{}).Error; err != nil {
		return err
	}

	return nil
}
