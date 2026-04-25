package model

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"             json:"id"`
	Code        string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"`
	Name        string         `gorm:"type:varchar(100);not null"           json:"name"`
	BranchID    *uint          `gorm:"index"                                json:"branch_id"`
	Description *string        `gorm:"type:text"                            json:"description"`
	IsActive    *bool           `gorm:"not null;default:true"                json:"is_active"`
	CreatedAt   time.Time      `gorm:"not null;default:now()"              json:"created_at"`
	UpdatedAt   *time.Time     `                                            json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                json:"deleted_at"`
	
	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}
