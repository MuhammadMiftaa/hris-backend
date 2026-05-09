package model

import (
	"time"

	"gorm.io/gorm"
)

type PushSubscription struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	EmployeeID uint           `gorm:"not null;index"           json:"employee_id"`
	Endpoint   string         `gorm:"type:text;not null"       json:"endpoint"`
	P256dh     string         `gorm:"type:text;not null"       json:"p256dh"`
	Auth       string         `gorm:"type:text;not null"       json:"auth"`
	UserAgent  *string        `gorm:"type:text"                json:"user_agent"`
	IsActive   bool           `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt  time.Time      `gorm:"not null;default:now()"  json:"created_at"`
	UpdatedAt  *time.Time     `                                 json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                    json:"deleted_at"`

	// Relations
	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}
