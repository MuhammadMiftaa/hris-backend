package model

import (
	"time"

	"gorm.io/gorm"
)

type RoleLevelEnum string

const (
	RoleLevelSuperAdmin RoleLevelEnum = "superadmin"
	RoleLevelAdmin      RoleLevelEnum = "admin"
	RoleLevelManager    RoleLevelEnum = "manager"
	RoleLevelStaff      RoleLevelEnum = "staff"
)

type Role struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"             json:"id"`
	Name        string         `gorm:"type:varchar(100);not null"           json:"name"`
	Level       RoleLevelEnum  `gorm:"type:role_level_enum;not null;default:staff" json:"level"`
	Description *string        `gorm:"type:text"                            json:"description"`
	CreatedAt   time.Time      `gorm:"not null;default:now()"              json:"created_at"`
	UpdatedAt   *time.Time     `                                            json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                json:"deleted_at"`
}
