package model

import (
	"time"

	"gorm.io/gorm"
)

type (
	LeaveCategoryEnum string
)

const (
	LeaveCategoryAnnual    LeaveCategoryEnum = "annual"
	LeaveCategorySick      LeaveCategoryEnum = "sick"
	LeaveCategoryMaternity LeaveCategoryEnum = "maternity"
	LeaveCategoryPaternity LeaveCategoryEnum = "paternity"
	LeaveCategoryUnpaid    LeaveCategoryEnum = "unpaid"
	LeaveCategoryOther     LeaveCategoryEnum = "other"
)


type LeaveType struct {
	ID                      uint              `gorm:"primaryKey;autoIncrement"        json:"id"`
	Name                    string            `gorm:"type:varchar(100);not null"      json:"name"`
	Category                LeaveCategoryEnum `gorm:"type:leave_category_enum;not null" json:"category"`
	RequiresDocument        bool              `gorm:"not null;default:false"          json:"requires_document"`
	RequiresDocumentType    *string           `gorm:"type:varchar(100)"               json:"requires_document_type"`
	MaxDurationPerRequest   *float64          `                                       json:"max_duration_per_request"`
	MaxOccurrencesPerYear   *int              `                                       json:"max_occurrences_per_year"`
	MaxTotalDurationPerYear *float64          `                                       json:"max_total_duration_per_year"`
	MaxPerMonth             *float64          `                                       json:"max_per_month"`
	ParentLeaveTypeID       *uint             `                                       json:"parent_leave_type_id"`
	DeductDays              float64           `gorm:"not null;default:1.0"            json:"deduct_days"`
	CreatedAt               time.Time         `gorm:"not null;default:now()"         json:"created_at"`
	UpdatedAt               *time.Time        `                                       json:"updated_at"`
	DeletedAt               gorm.DeletedAt    `gorm:"index"                           json:"deleted_at"`
}
