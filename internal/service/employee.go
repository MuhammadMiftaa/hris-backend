package service

import (
	"context"
	"fmt"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"
	"hris-backend/internal/utils/data"
)

type EmployeeService interface {
	GetMetadata(ctx context.Context) (dto.EmployeeMetadata, error)
	GetAllEmployees(ctx context.Context) ([]dto.Employee, error)
	GetEmployeeByID(ctx context.Context, employeeID string) (dto.Employee, error)
	CreateEmployee(ctx context.Context, req dto.CreateEmployeeRequest) (dto.Employee, error)
	UpdateEmployee(ctx context.Context, id string, req dto.UpdateEmployeeRequest) (dto.Employee, error)
	DeleteEmployee(ctx context.Context, employeeID string) error
}

type employeeService struct {
	repo      repository.EmployeeRepository
	txManager repository.TxManager
}

func NewEmployeeService(repo repository.EmployeeRepository, txManager repository.TxManager) EmployeeService {
	return &employeeService{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *employeeService) GetMetadata(ctx context.Context) (dto.EmployeeMetadata, error) {
	branchMeta, err := s.repo.GetBranchMetadata(ctx, nil)
	if err != nil {
		return dto.EmployeeMetadata{}, fmt.Errorf("get branch metadata: %w", err)
	}

	departmentMeta, err := s.repo.GetDepartmentMetadata(ctx, nil)
	if err != nil {
		return dto.EmployeeMetadata{}, fmt.Errorf("get department metadata: %w", err)
	}

	roleMeta, err := s.repo.GetRoleMetadata(ctx, nil)
	if err != nil {
		return dto.EmployeeMetadata{}, fmt.Errorf("get role metadata: %w", err)
	}

	jobPositionMeta, err := s.repo.GetJobPositionMetadata(ctx, nil)
	if err != nil {
		return dto.EmployeeMetadata{}, fmt.Errorf("get job position metadata: %w", err)
	}

	return dto.EmployeeMetadata{
		BranchMeta:        branchMeta,
		DepartmentMeta:    departmentMeta,
		RoleMeta:          roleMeta,
		JobPositionMeta:   jobPositionMeta,
		ReligionMeta:      data.ReligionMeta,
		MaritalStatusMeta: data.MaritalStatusMeta,
		BloodTypeMeta:     data.BloodTypeMeta,
		StatusMeta:        data.StatusMeta,
	}, nil
}

func (s *employeeService) GetAllEmployees(ctx context.Context) ([]dto.Employee, error) {
	employees, err := s.repo.GetAllEmployees(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get all employees: %w", err)
	}
	return employees, nil
}

func (s *employeeService) GetEmployeeByID(ctx context.Context, employeeID string) (dto.Employee, error) {
	employee, err := s.repo.GetEmployeeByID(ctx, nil, employeeID)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("get employee by ID: %w", err)
	}
	return employee, nil
}

func (s *employeeService) CreateEmployee(ctx context.Context, req dto.CreateEmployeeRequest) (dto.Employee, error) {
	employee, err := s.validateEmployeePayload(req.EmployeeRequest)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("validate employee payload: %w", err)
	}
	account, err := s.validateAccountPayload(employee, req.EmployeeRequest)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("validate account payload: %w", err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("create employee: begin transaction: %w", err)
	}
	defer tx.Rollback()

	createdEmployee, err := s.repo.CreateEmployee(ctx, tx, employee)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("create employee: create employee: %w", err)
	}

	createdAccount, err := s.repo.CreateAccount(ctx, tx, account)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("create employee: create account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.Employee{}, fmt.Errorf("create employee: commit transaction: %w", err)
	}

	return dto.Employee{
		ID:             createdEmployee.ID,
		EmployeeNumber: createdEmployee.EmployeeNumber,
		FullName:       createdEmployee.FullName,
		NIK:            createdEmployee.NIK,
		NPWP:           createdEmployee.NPWP,
		KKNumber:       createdEmployee.KKNumber,
		BirthDate:      createdEmployee.BirthDate.Format("2006-01-02"),
		BirthPlace:     createdEmployee.BirthPlace,
		Gender:         createdEmployee.Gender,
		Religion:       createdEmployee.Religion,
		MaritalStatus:  createdEmployee.MaritalStatus,
		BloodType:      createdEmployee.BloodType,
		Nationality:    createdEmployee.Nationality,
		PhotoURL:       createdEmployee.PhotoURL,
		IsActive:       createdAccount.IsActive,
		IsTrainer:      createdEmployee.IsTrainer,
		BranchID:       createdEmployee.BranchID,
		DepartmentID:   createdEmployee.DepartmentID,
		RoleID:         &createdAccount.RoleID,
		JobPositionsID: createdEmployee.JobPositionsID,
	}, nil
}

func (s *employeeService) UpdateEmployee(ctx context.Context, id string, req dto.UpdateEmployeeRequest) (dto.Employee, error) {
	employee, err := s.validateEmployeePayload(req.EmployeeRequest)
	if err != nil {
		return dto.Employee{}, err
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("update employee: begin transaction: %w", err)
	}
	defer tx.Rollback()

	existingEmployee, err := s.repo.GetEmployeeByID(ctx, tx, id)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("update employee: get existing employee: %w", err)
	}

	updatedEmployee, err := s.repo.UpdateEmployee(ctx, tx, id, employee)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("update employee: update employee: %w", err)
	}

	if existingEmployee.IsActive != req.IsActive || existingEmployee.RoleID != req.RoleID {
		existingAccount, err := s.repo.GetAccountByEmployeeID(ctx, tx, id)
		if err != nil {
			return dto.Employee{}, fmt.Errorf("update employee: get existing account: %w", err)
		}

		existingAccount.IsActive = req.IsActive
		existingAccount.RoleID = *req.RoleID
		updatedAccount, err := s.repo.UpdateAccount(ctx, tx, existingAccount)
		if err != nil {
			return dto.Employee{}, fmt.Errorf("update employee: update account: %w", err)
		}

		existingEmployee.IsActive = updatedAccount.IsActive
		existingEmployee.RoleID = &updatedAccount.RoleID
	}

	if err := tx.Commit(); err != nil {
		return dto.Employee{}, fmt.Errorf("update employee: commit transaction: %w", err)
	}

	return dto.Employee{
		ID:             updatedEmployee.ID,
		EmployeeNumber: updatedEmployee.EmployeeNumber,
		FullName:       updatedEmployee.FullName,
		NIK:            updatedEmployee.NIK,
		NPWP:           updatedEmployee.NPWP,
		KKNumber:       updatedEmployee.KKNumber,
		BirthDate:      updatedEmployee.BirthDate.Format("2006-01-02"),
		BirthPlace:     updatedEmployee.BirthPlace,
		Gender:         updatedEmployee.Gender,
		Religion:       updatedEmployee.Religion,
		MaritalStatus:  updatedEmployee.MaritalStatus,
		BloodType:      updatedEmployee.BloodType,
		Nationality:    updatedEmployee.Nationality,
		PhotoURL:       updatedEmployee.PhotoURL,
		IsActive:       existingEmployee.IsActive,
		IsTrainer:      updatedEmployee.IsTrainer,
		BranchID:       updatedEmployee.BranchID,
		DepartmentID:   updatedEmployee.DepartmentID,
		RoleID:         existingEmployee.RoleID,
		JobPositionsID: updatedEmployee.JobPositionsID,
	}, nil
}

func (s *employeeService) DeleteEmployee(ctx context.Context, employeeID string) error {
	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("delete employee: begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.repo.DeleteEmployee(ctx, tx, employeeID); err != nil {
		return fmt.Errorf("delete employee: delete employee: %w", err)
	}

	if err := s.repo.DeleteAccount(ctx, tx, employeeID); err != nil {
		return fmt.Errorf("delete employee: delete account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete employee: commit transaction: %w", err)
	}

	return nil
}

// ~ Helper Method
func (s *employeeService) validateEmployeePayload(req dto.EmployeeRequest) (model.Employee, error) {
	birthDate, err := utils.ParseAuto(req.BirthDate)
	if err != nil {
		return model.Employee{}, fmt.Errorf("invalid birth date format: %w", err)
	}

	var gender *model.GenderEnum
	if req.Gender == nil {
		return model.Employee{}, fmt.Errorf("gender is required")
	} else {
		v := model.GenderEnum(*req.Gender)
		switch v {
		case model.GenderMale,
			model.GenderFemale:
			gender = &v
		default:
			return model.Employee{}, fmt.Errorf("invalid gender: %q", *req.Gender)
		}
	}

	var maritalStatus *model.MaritalStatusEnum
	if req.MaritalStatus == nil {
		return model.Employee{}, fmt.Errorf("marital status is required")
	} else {
		v := model.MaritalStatusEnum(*req.MaritalStatus)
		switch v {
		case model.MaritalSingle,
			model.MaritalMarried,
			model.MaritalDivorced:
			maritalStatus = &v
		default:
			return model.Employee{}, fmt.Errorf("invalid marital status: %q", *req.MaritalStatus)
		}
	}

	return model.Employee{
		EmployeeNumber: req.EmployeeNumber,
		FullName:       req.FullName,
		NIK:            req.NIK,
		NPWP:           req.NPWP,
		KKNumber:       req.KKNumber,
		BirthDate:      birthDate,
		BirthPlace:     req.BirthPlace,
		Gender:         gender,
		Religion:       req.Religion,
		MaritalStatus:  maritalStatus,
		BloodType:      req.BloodType,
		Nationality:    req.Nationality,
		PhotoURL:       req.PhotoURL,
		BranchID:       req.BranchID,
		DepartmentID:   req.DepartmentID,
		JobPositionsID: req.JobPositionsID,
	}, nil
}

func (s *employeeService) validateAccountPayload(employee model.Employee, req dto.EmployeeRequest) (model.Account, error) {
	if req.RoleID == nil {
		return model.Account{}, fmt.Errorf("role ID is required")
	}

	randompassword := utils.GenerateRandomString(10)
	hashpassword, err := utils.PasswordHashing(randompassword)
	if err != nil {
		return model.Account{}, fmt.Errorf("failed to generate password: %w", err)
	}

	return model.Account{
		EmployeeID: employee.ID,
		RoleID:     *req.RoleID,
		Email:      utils.GenerateEmail(employee.FullName),
		Password:   hashpassword,
		IsActive:   true,
	}, nil
}
