package model

import (
	"time"

	"gorm.io/gorm"
)

type EmployeeSchedule struct {
	ID              uint           `gorm:"primaryKey;autoIncrement"   json:"id"`
	EmployeeID      uint           `gorm:"not null;index"             json:"employee_id"`
	ShiftTemplateID uint           `gorm:"not null;index"             json:"shift_template_id"`
	EffectiveDate   time.Time      `gorm:"type:date;not null"         json:"effective_date"`
	EndDate         *time.Time     `gorm:"type:date"                  json:"end_date"`
	IsActive        *bool          `gorm:"not null;default:true"      json:"is_active"`
	CreatedAt       time.Time      `gorm:"not null;default:now()"    json:"created_at"`
	UpdatedAt       *time.Time     `                                  json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index"                      json:"deleted_at"`

	// Relations
	Employee      Employee      `gorm:"foreignKey:EmployeeID"      json:"employee,omitempty"`
	ShiftTemplate ShiftTemplate `gorm:"foreignKey:ShiftTemplateID" json:"shift_template,omitempty"`
}
