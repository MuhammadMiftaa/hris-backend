package dto

import (
	"time"

	"gorm.io/datatypes"
)

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRes struct {
	Token       string                  `json:"token"`
	Refresh     string                  `json:"refresh"`
	Permissions []string                `json:"permissions"`
	Account     GetEmployeeByIDResponse `json:"account"`
}

type GetAccountByEmailResponse struct {
	ID          uint       `json:"id"`
	Email       string     `json:"email"`
	Password    string     `json:"password"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

type GetEmployeeByIDResponse struct {
	AccountID      uint           `json:"account_id" gorm:"column:account_id"`
	Email          string         `json:"email" gorm:"column:email"`
	IsActive       bool           `json:"is_active" gorm:"column:is_active"`
	LastLoginAt    *time.Time     `json:"last_login_at" gorm:"column:last_login_at"`
	EmployeeNumber string         `json:"employee_number" gorm:"column:employee_number"`
	FullName       string         `json:"full_name" gorm:"column:full_name"`
	PhotoURL       *string        `json:"photo_url" gorm:"column:photo_url"`
	IsTrainer      bool           `json:"is_trainer" gorm:"column:is_trainer"`
	BranchID       *uint          `json:"branch_id" gorm:"column:branch_id"`
	DepartmentID   *uint          `json:"department_id" gorm:"column:department_id"`
	JobPositionsID *uint          `json:"job_positions_id" gorm:"column:job_positions_id"`
	RoleName       string         `json:"role_name" gorm:"column:role_name"`
	Permissions    datatypes.JSON `json:"permissions,omitempty" gorm:"column:permissions"`
}

type Token struct {
	Issuer        string `json:"iss"`
	Audience      string `json:"aud"`
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	EmailVerified string `json:"email_verified"`
	Nonce         string `json:"nonce"`
	IssuedAt      string `json:"iat"`
	Expires       string `json:"exp"`

	Refresh     string                  `json:"refresh"`
	Permissions []string                `json:"permissions"`
	Account     GetEmployeeByIDResponse `json:"account"`
}
