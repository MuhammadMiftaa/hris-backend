package model

import (
	"time"

	"gorm.io/gorm"
)

type HolidayTypeEnum string

const (
	HolidayNational   HolidayTypeEnum = "national"
	HolidayJoint      HolidayTypeEnum = "joint"
	HolidayObservance HolidayTypeEnum = "observance"
	HolidayCompany    HolidayTypeEnum = "company"
)

type Holiday struct {
	ID          uint            `gorm:"primaryKey;autoIncrement"          json:"id"`
	Name        string          `gorm:"type:varchar(100);not null"        json:"name"`
	Year        int             `gorm:"not null"                          json:"year"`
	Date        time.Time       `gorm:"type:date;not null"                json:"date"`
	Type        HolidayTypeEnum `gorm:"type:holiday_type_enum;not null"   json:"type"`
	BranchID    *uint           `gorm:"index"                             json:"branch_id"`
	Description *string         `gorm:"type:text"                         json:"description"`
	CreatedAt   time.Time       `gorm:"not null;default:now()"           json:"created_at"`
	UpdatedAt   *time.Time      `                                         json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index"                             json:"deleted_at"`

	// Relations
	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}
