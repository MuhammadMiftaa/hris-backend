package model

import (
	"time"

	"gorm.io/gorm"
)

type ContractTypeEnum string

const (
	ContractPKWT      ContractTypeEnum = "pkwt"
	ContractPKWTT     ContractTypeEnum = "pkwtt"
	ContractProbation ContractTypeEnum = "probation"
	ContractIntern    ContractTypeEnum = "intern"
	ContractPartTime  ContractTypeEnum = "part_time"
	ContractFreelance ContractTypeEnum = "freelance"
)

type EmploymentContract struct {
	ID             uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	EmployeeID     uint             `gorm:"not null;index"           json:"employee_id"`
	ContractNumber string           `gorm:"type:varchar(50);not null" json:"contract_number"`
	ContractType   ContractTypeEnum `gorm:"type:contract_type_enum;not null" json:"contract_type"`
	StartDate      *time.Time       `gorm:"type:date"                json:"start_date"`
	EndDate        *time.Time       `gorm:"type:date"                json:"end_date"`
	Salary         *float64         `gorm:"type:numeric(12,2)"       json:"salary"`
	Notes          *string          `gorm:"type:text"                json:"notes"`
	CreatedAt      time.Time        `gorm:"not null;default:now()"  json:"created_at"`
	UpdatedAt      *time.Time       `                                json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `gorm:"index"                    json:"deleted_at"`

	// Relations
	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}
