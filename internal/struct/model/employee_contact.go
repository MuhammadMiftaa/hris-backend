package model

import (
	"time"

	"gorm.io/gorm"
)

type EmployeeContact struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	EmployeeID   uint           `gorm:"not null;index"           json:"employee_id"`
	ContactType  string         `gorm:"type:varchar(20);not null" json:"contact_type"`
	ContactValue string         `gorm:"type:text;not null"       json:"contact_value"`
	ContactLabel *string        `gorm:"type:varchar(100)"        json:"contact_label"`
	IsPrimary    bool           `gorm:"not null;default:false"   json:"is_primary"`
	CreatedAt    time.Time      `gorm:"not null;default:now()"   json:"created_at"`
	UpdatedAt    *time.Time     `                                 json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                     json:"deleted_at"`

	// Relations
	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}
