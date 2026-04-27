package model

import (
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// ENUM TYPES
// =============================================================================

type (
	GenderEnum        string
	MaritalStatusEnum string
)

const (
	GenderMale   GenderEnum = "male"
	GenderFemale GenderEnum = "female"
)

const (
	MaritalSingle   MaritalStatusEnum = "single"
	MaritalMarried  MaritalStatusEnum = "married"
	MaritalDivorced MaritalStatusEnum = "divorced"
	MaritalWidowed  MaritalStatusEnum = "widowed"
)

type Employee struct {
	ID             uint               `gorm:"primaryKey;autoIncrement"                  json:"id"`
	EmployeeNumber string             `gorm:"type:varchar(20);not null;uniqueIndex"     json:"employee_number"`
	FullName       string             `gorm:"type:varchar(150);not null"                json:"full_name"`
	NIK            *string            `gorm:"type:varchar(16)"                          json:"nik"`
	NPWP           *string            `gorm:"type:varchar(20)"                          json:"npwp"`
	KKNumber       *string            `gorm:"type:varchar(16)"                          json:"kk_number"`
	BirthDate      time.Time          `gorm:"type:date;not null"                        json:"birth_date"`
	BirthPlace     *string            `gorm:"type:varchar(100)"                         json:"birth_place"`
	Gender         *GenderEnum        `gorm:"type:gender_enum"                          json:"gender"`
	Religion       *string            `gorm:"type:varchar(50)"                          json:"religion"`
	MaritalStatus  *MaritalStatusEnum `gorm:"type:marital_status_enum"                  json:"marital_status"`
	BloodType      *string            `gorm:"type:varchar(5)"                           json:"blood_type"`
	Nationality    *string            `gorm:"type:varchar(50)"                          json:"nationality"`
	Height         *float64           `gorm:"type:numeric(5,2)"                         json:"height"`
	Weight         *float64           `gorm:"type:numeric(5,2)"                         json:"weight"`
	PhotoURL       *string            `gorm:"type:text"                                 json:"photo_url"`
	IsTrainer      *bool              `gorm:"not null;default:false"                    json:"is_trainer"`
	BranchID       *uint              `gorm:"index"                                     json:"branch_id"`
	DepartmentID   *uint              `gorm:"index"                                     json:"department_id"`
	JobPositionsID *uint              `gorm:"index"                                     json:"job_positions_id"`
	CreatedAt      time.Time          `gorm:"not null;default:now()"                   json:"created_at"`
	UpdatedAt      *time.Time         `                                                 json:"updated_at"`
	DeletedAt      gorm.DeletedAt     `gorm:"index"                                     json:"deleted_at"`

	// Relations
	Branch      *Branch      `gorm:"foreignKey:BranchID"      json:"branch,omitempty"`
	Department  *Department  `gorm:"foreignKey:DepartmentID"  json:"department,omitempty"`
	JobPosition *JobPosition `gorm:"foreignKey:JobPositionsID" json:"job_position,omitempty"`
}
