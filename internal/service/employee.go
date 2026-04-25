package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"hris-backend/config/storage"
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
	CreateEmployee(ctx context.Context, req dto.CreateEmployeeRequest) (dto.Employee, dto.NewEmployeeCred, error)
	UpdateEmployee(ctx context.Context, id string, req dto.UpdateEmployeeRequest) (dto.Employee, error)
	DeleteEmployee(ctx context.Context, employeeID string) error
	ResetPassword(ctx context.Context, employeeID string, req dto.ResetPasswordRequest) error

	// Contact
	GetContactsByEmployeeID(ctx context.Context, employeeID string) ([]dto.EmployeeContactResponse, error)
	CreateContact(ctx context.Context, employeeID string, req dto.CreateContactRequest) (dto.EmployeeContactResponse, error)
	UpdateContact(ctx context.Context, contactID string, req dto.UpdateContactRequest) (dto.EmployeeContactResponse, error)
	DeleteContact(ctx context.Context, contactID string) error

	// Contract
	GetContractsByEmployeeID(ctx context.Context, employeeID string) ([]dto.ContractResponse, error)
	CreateContract(ctx context.Context, employeeID string, req dto.CreateContractRequest) (dto.ContractResponse, error)
	UpdateContract(ctx context.Context, contractID string, req dto.UpdateContractRequest) (dto.ContractResponse, error)
	DeleteContract(ctx context.Context, contractID string) error
}

type employeeService struct {
	repo      repository.EmployeeRepository
	txManager repository.TxManager
	minio     storage.MinioClient
}

func NewEmployeeService(repo repository.EmployeeRepository, txManager repository.TxManager, minio storage.MinioClient) EmployeeService {
	return &employeeService{
		repo:      repo,
		txManager: txManager,
		minio:     minio,
	}
}

// resolvePhotoURL converts raw MinIO object key to presigned download URL
func (s *employeeService) resolvePhotoURL(ctx context.Context, objectKey *string) *string {
	if objectKey == nil || *objectKey == "" {
		return nil
	}
	url, err := s.minio.PresignedGetObject(ctx, storage.BucketProfilePhotos, *objectKey, storage.PresignedDownloadExpiry)
	if err != nil {
		return nil
	}
	return &url
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
		GenderMeta:        data.GenderMeta,
		ReligionMeta:      data.ReligionMeta,
		MaritalStatusMeta: data.MaritalStatusMeta,
		BloodTypeMeta:     data.BloodTypeMeta,
		StatusMeta:        data.StatusMeta,
		ContractTypeMeta:  data.ContractTypeMeta,
	}, nil
}

func (s *employeeService) GetAllEmployees(ctx context.Context) ([]dto.Employee, error) {
	employees, err := s.repo.GetAllEmployees(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get all employees: %w", err)
	}

	// Resolve presigned URLs for photo_url
	for i := range employees {
		employees[i].PhotoURL = s.resolvePhotoURL(ctx, employees[i].PhotoURL)
	}

	return employees, nil
}

func (s *employeeService) GetEmployeeByID(ctx context.Context, employeeID string) (dto.Employee, error) {
	employee, err := s.repo.GetEmployeeByID(ctx, nil, employeeID)
	if err != nil {
		return dto.Employee{}, fmt.Errorf("get employee by ID: %w", err)
	}

	// Resolve presigned URL for photo_url
	employee.PhotoURL = s.resolvePhotoURL(ctx, employee.PhotoURL)

	return employee, nil
}

func (s *employeeService) CreateEmployee(ctx context.Context, req dto.CreateEmployeeRequest) (dto.Employee, dto.NewEmployeeCred, error) {
	employee, err := s.validateEmployeePayload(req.EmployeeRequest)
	if err != nil {
		return dto.Employee{}, dto.NewEmployeeCred{}, fmt.Errorf("validate employee payload: %w", err)
	}
	account, newCredentials, err := s.validateAccountPayload(employee, req.EmployeeRequest)
	if err != nil {
		return dto.Employee{}, dto.NewEmployeeCred{}, fmt.Errorf("validate account payload: %w", err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.Employee{}, dto.NewEmployeeCred{}, fmt.Errorf("create employee: begin transaction: %w", err)
	}
	defer tx.Rollback()

	createdEmployee, err := s.repo.CreateEmployee(ctx, tx, employee)
	if err != nil {
		return dto.Employee{}, dto.NewEmployeeCred{}, fmt.Errorf("create employee: create employee: %w", err)
	}

	account.EmployeeID = createdEmployee.ID
	createdAccount, err := s.repo.CreateAccount(ctx, tx, account)
	if err != nil {
		return dto.Employee{}, dto.NewEmployeeCred{}, fmt.Errorf("create employee: create account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.Employee{}, dto.NewEmployeeCred{}, fmt.Errorf("create employee: commit transaction: %w", err)
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
		PhotoURL:       s.resolvePhotoURL(ctx, createdEmployee.PhotoURL),
		IsActive:       *createdAccount.IsActive,
		IsTrainer:      createdEmployee.IsTrainer,
		BranchID:       createdEmployee.BranchID,
		DepartmentID:   createdEmployee.DepartmentID,
		RoleID:         &createdAccount.RoleID,
		JobPositionsID: createdEmployee.JobPositionsID,
	}, newCredentials, nil
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

		existingAccount.IsActive = &req.IsActive
		existingAccount.RoleID = *req.RoleID
		updatedAccount, err := s.repo.UpdateAccount(ctx, tx, existingAccount)
		if err != nil {
			return dto.Employee{}, fmt.Errorf("update employee: update account: %w", err)
		}

		existingEmployee.IsActive = *updatedAccount.IsActive
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
		PhotoURL:       s.resolvePhotoURL(ctx, updatedEmployee.PhotoURL),
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
		IsTrainer:      req.IsTrainer,
		BranchID:       req.BranchID,
		DepartmentID:   req.DepartmentID,
		JobPositionsID: req.JobPositionsID,
	}, nil
}

func (s *employeeService) validateAccountPayload(employee model.Employee, req dto.EmployeeRequest) (model.Account, dto.NewEmployeeCred, error) {
	if req.RoleID == nil {
		return model.Account{}, dto.NewEmployeeCred{}, fmt.Errorf("role ID is required")
	}

	email := utils.GenerateEmail(employee.FullName)
	randomPassword := utils.GenerateRandomString(10)
	hashPassword, err := utils.PasswordHashing(randomPassword)
	if err != nil {
		return model.Account{}, dto.NewEmployeeCred{}, fmt.Errorf("failed to generate password: %w", err)
	}

	return model.Account{
			EmployeeID: employee.ID,
			RoleID:     *req.RoleID,
			Email:      email,
			Password:   hashPassword,
		}, dto.NewEmployeeCred{
			Email:    email,
			Password: randomPassword,
		}, nil
}

func (s *employeeService) ResetPassword(ctx context.Context, employeeID string, req dto.ResetPasswordRequest) error {
	if req.NewPassword == "" {
		return fmt.Errorf("new password is required")
	}
	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}
	if len(req.NewPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	hashed, err := utils.PasswordHashing(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdateAccountPassword(ctx, nil, employeeID, hashed); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	return nil
}

// ~ Contact
func (s *employeeService) GetContactsByEmployeeID(ctx context.Context, employeeID string) ([]dto.EmployeeContactResponse, error) {
	return s.repo.GetContactsByEmployeeID(ctx, nil, employeeID)
}

func (s *employeeService) CreateContact(ctx context.Context, employeeID string, req dto.CreateContactRequest) (dto.EmployeeContactResponse, error) {
	if req.ContactType == "" || req.ContactValue == "" {
		return dto.EmployeeContactResponse{}, fmt.Errorf("contact_type and contact_value are required")
	}

	empIDParams, err := strconv.ParseUint(employeeID, 10, 64)
	if err != nil {
		return dto.EmployeeContactResponse{}, err
	}

	modelReq := model.EmployeeContact{
		EmployeeID:   uint(empIDParams),
		ContactType:  req.ContactType,
		ContactValue: req.ContactValue,
		ContactLabel: req.ContactLabel,
	}
	if req.IsPrimary != nil {
		modelReq.IsPrimary = *req.IsPrimary
	}

	created, err := s.repo.CreateContact(ctx, nil, modelReq)
	if err != nil {
		return dto.EmployeeContactResponse{}, err
	}
	return s.repo.GetContactByID(ctx, nil, fmt.Sprintf("%d", created.ID))
}

func (s *employeeService) UpdateContact(ctx context.Context, contactID string, req dto.UpdateContactRequest) (dto.EmployeeContactResponse, error) {
	existing, err := s.repo.GetContactByID(ctx, nil, contactID)
	if err != nil {
		return dto.EmployeeContactResponse{}, err
	}

	modelReq := model.EmployeeContact{
		ContactType:  existing.ContactType,
		ContactValue: existing.ContactValue,
		ContactLabel: existing.ContactLabel,
		IsPrimary:    existing.IsPrimary,
	}

	if req.ContactType != nil {
		modelReq.ContactType = *req.ContactType
	}
	if req.ContactValue != nil {
		modelReq.ContactValue = *req.ContactValue
	}
	if req.ContactLabel != nil {
		modelReq.ContactLabel = req.ContactLabel
	}
	if req.IsPrimary != nil {
		modelReq.IsPrimary = *req.IsPrimary
	}

	if _, err := s.repo.UpdateContact(ctx, nil, contactID, modelReq); err != nil {
		return dto.EmployeeContactResponse{}, err
	}

	return s.repo.GetContactByID(ctx, nil, contactID)
}

func (s *employeeService) DeleteContact(ctx context.Context, contactID string) error {
	return s.repo.DeleteContact(ctx, nil, contactID)
}

// ~ Contract
func (s *employeeService) GetContractsByEmployeeID(ctx context.Context, employeeID string) ([]dto.ContractResponse, error) {
	return s.repo.GetContractsByEmployeeID(ctx, nil, employeeID)
}

func (s *employeeService) CreateContract(ctx context.Context, employeeID string, req dto.CreateContractRequest) (dto.ContractResponse, error) {
	if req.ContractNumber == "" || req.ContractType == "" || req.StartDate == "" {
		return dto.ContractResponse{}, fmt.Errorf("contract_number, contract_type, and start_date are required")
	}

	empIDParams, err := strconv.ParseUint(employeeID, 10, 64)
	if err != nil {
		return dto.ContractResponse{}, err
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return dto.ContractResponse{}, fmt.Errorf("invalid start_date format")
	}

	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return dto.ContractResponse{}, fmt.Errorf("invalid end_date format")
		}
		endDate = &parsed
	}

	modelReq := model.EmploymentContract{
		EmployeeID:     uint(empIDParams),
		ContractNumber: req.ContractNumber,
		ContractType:   model.ContractTypeEnum(req.ContractType),
		StartDate:      &startDate,
		EndDate:        endDate,
		Salary:         &req.Salary,
		Notes:          req.Notes,
	}

	created, err := s.repo.CreateContract(ctx, nil, modelReq)
	if err != nil {
		return dto.ContractResponse{}, err
	}

	return s.repo.GetContractByID(ctx, nil, fmt.Sprintf("%d", created.ID))
}

func (s *employeeService) UpdateContract(ctx context.Context, contractID string, req dto.UpdateContractRequest) (dto.ContractResponse, error) {
	existing, err := s.repo.GetContractByID(ctx, nil, contractID)
	if err != nil {
		return dto.ContractResponse{}, err
	}

	startDate, _ := time.Parse("2006-01-02", existing.StartDate)
	modelReq := model.EmploymentContract{
		ContractNumber: existing.ContractNumber,
		ContractType:   model.ContractTypeEnum(existing.ContractType),
		StartDate:      &startDate,
		Salary:         &existing.Salary,
		Notes:          existing.Notes,
	}
	if existing.EndDate != nil {
		endDate, _ := time.Parse("2006-01-02", *existing.EndDate)
		modelReq.EndDate = &endDate
	}

	if req.ContractNumber != nil {
		modelReq.ContractNumber = *req.ContractNumber
	}
	if req.ContractType != nil {
		modelReq.ContractType = model.ContractTypeEnum(*req.ContractType)
	}
	if req.StartDate != nil {
		d, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return dto.ContractResponse{}, fmt.Errorf("invalid start_date format")
		}
		modelReq.StartDate = &d
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			modelReq.EndDate = nil
		} else {
			d, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				return dto.ContractResponse{}, fmt.Errorf("invalid end_date format")
			}
			modelReq.EndDate = &d
		}
	}
	if req.Salary != nil {
		modelReq.Salary = req.Salary
	}
	if req.Notes != nil {
		modelReq.Notes = req.Notes
	}

	if _, err := s.repo.UpdateContract(ctx, nil, contractID, modelReq); err != nil {
		return dto.ContractResponse{}, err
	}

	return s.repo.GetContractByID(ctx, nil, contractID)
}

func (s *employeeService) DeleteContract(ctx context.Context, contractID string) error {
	return s.repo.DeleteContract(ctx, nil, contractID)
}
