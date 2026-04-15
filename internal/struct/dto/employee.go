package dto

import (
	"time"

	"hris-backend/internal/struct/model"
)

type Employee struct {
	ID               uint                     `json:"id"`
	EmployeeNumber   string                   `json:"employee_number"`
	FullName         string                   `json:"full_name"`
	NIK              *string                  `json:"nik"`
	NPWP             *string                  `json:"npwp"`
	KKNumber         *string                  `json:"kk_number"`
	BirthDate        string                   `json:"birth_date"`
	BirthPlace       *string                  `json:"birth_place"`
	Gender           *model.GenderEnum        `json:"gender"`
	Religion         *string                  `json:"religion"`
	MaritalStatus    *model.MaritalStatusEnum `json:"marital_status"`
	BloodType        *string                  `json:"blood_type"`
	Nationality      *string                  `json:"nationality"`
	PhotoURL         *string                  `json:"photo_url"`
	IsActive         bool                     `json:"is_active"`
	IsTrainer        bool                     `json:"is_trainer"`
	BranchID         *uint                    `json:"branch_id"`
	DepartmentID     *uint                    `json:"department_id"`
	RoleID           *uint                    `json:"role_id"`
	JobPositionsID   *uint                    `json:"job_positions_id"`
	BranchName       *string                  `json:"branch_name"`
	DepartmentName   *string                  `json:"department_name"`
	RoleName         *string                  `json:"role_name"`
	JobPositionTitle *string                  `json:"job_position_title"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
	DeletedAt        *time.Time               `json:"deleted_at"`
}

type EmployeeRequest struct {
	EmployeeNumber string  `json:"employee_number"`
	FullName       string  `json:"full_name"`
	NIK            *string `json:"nik"`
	NPWP           *string `json:"npwp"`
	KKNumber       *string `json:"kk_number"`
	BirthDate      string  `json:"birth_date"`
	BirthPlace     *string `json:"birth_place"`
	Gender         *string `json:"gender"`
	Religion       *string `json:"religion"`
	MaritalStatus  *string `json:"marital_status"`
	BloodType      *string `json:"blood_type"`
	Nationality    *string `json:"nationality"`
	PhotoURL       *string `json:"photo_url"`
	IsTrainer      bool    `json:"is_trainer"`
	BranchID       *uint   `json:"branch_id"`
	DepartmentID   *uint   `json:"department_id"`
	RoleID         *uint   `json:"role_id"`
	JobPositionsID *uint   `json:"job_positions_id"`
}

type CreateEmployeeRequest struct {
	EmployeeRequest
}

type UpdateEmployeeRequest struct {
	EmployeeRequest
	IsActive bool `json:"is_active"`
}

type EmployeeMetadata struct {
	BranchMeta        []Meta `json:"branch_meta"`
	DepartmentMeta    []Meta `json:"department_meta"`
	RoleMeta          []Meta `json:"role_meta"`
	JobPositionMeta   []Meta `json:"job_position_meta"`
	ReligionMeta      []Meta `json:"religion_meta"`
	MaritalStatusMeta []Meta `json:"marital_status_meta"`
	BloodTypeMeta     []Meta `json:"blood_type_meta"`
	StatusMeta        []Meta `json:"status_meta"`
}
