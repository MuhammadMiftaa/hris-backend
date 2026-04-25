package model

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"                  json:"id"`
	EmployeeID  uint           `gorm:"not null;index"                            json:"employee_id"`
	RoleID      uint           `gorm:"not null;index"                            json:"role_id"`
	Email       string         `gorm:"type:varchar(150);not null;uniqueIndex"    json:"email"`
	Password    string         `gorm:"type:text;not null"                        json:"password"`
	LastLoginAt *time.Time     `                                                 json:"last_login_at"`
	IsActive    *bool           `gorm:"not null;default:true"                     json:"is_active"`
	CreatedAt   time.Time      `gorm:"not null;default:now()"                   json:"created_at"`
	UpdatedAt   *time.Time     `                                                 json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                     json:"deleted_at"`

	// Relations
	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Role     Role     `gorm:"foreignKey:RoleID"     json:"role,omitempty"`
}
