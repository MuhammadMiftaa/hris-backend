package dto

import "time"

type ContractResponse struct {
	ID             uint       `json:"id"`
	EmployeeID     uint       `json:"employee_id"`
	ContractNumber string     `json:"contract_number"`
	ContractType   string     `json:"contract_type"`
	StartDate      string     `json:"start_date"`
	EndDate        *string    `json:"end_date"`
	Salary         float64    `json:"salary"`
	Notes          *string    `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
}

type CreateContractRequest struct {
	ContractNumber string  `json:"contract_number"`
	ContractType   string  `json:"contract_type"`
	StartDate      string  `json:"start_date"`
	EndDate        *string `json:"end_date"`
	Salary         float64 `json:"salary"`
	Notes          *string `json:"notes"`
}

type UpdateContractRequest struct {
	ContractNumber *string  `json:"contract_number"`
	ContractType   *string  `json:"contract_type"`
	StartDate      *string  `json:"start_date"`
	EndDate        *string  `json:"end_date"`
	Salary         *float64 `json:"salary"`
	Notes          *string  `json:"notes"`
}

type ContractListParams struct {
	PaginationParams
	Search *string `query:"search"`
}
