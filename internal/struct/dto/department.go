package dto

import "time"

type DepartmentResponse struct {
	ID          uint       `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	BranchID    *uint      `json:"branch_id"`
	BranchName  *string    `json:"branch_name"`
	Description *string    `json:"description"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type CreateDepartmentRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	BranchID    *uint   `json:"branch_id"`
	Description *string `json:"description"`
}

type UpdateDepartmentRequest struct {
	Name        *string `json:"name"`
	BranchID    *uint   `json:"branch_id"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type DepartmentListParams struct {
	PaginationParams
	BranchID *uint   `query:"branch_id"`
	IsActive *bool   `query:"is_active"`
	Search   *string `query:"search"`
}

type DepartmentMetadata struct {
	BranchMeta   []Meta `json:"branch_meta"`
	PositionMeta []Meta `json:"position_meta"`
	EmployeeMeta []Meta `json:"employee_meta"`
}
