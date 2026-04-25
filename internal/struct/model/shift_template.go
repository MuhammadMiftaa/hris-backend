package model

import (
	"time"

	"gorm.io/gorm"
)

type ShiftTemplate struct {
	ID         uint           `gorm:"primaryKey;autoIncrement"   json:"id"`
	Name       string         `gorm:"type:varchar(100);not null" json:"name"`
	IsFlexible bool           `gorm:"not null;default:false"     json:"is_flexible"`
	CanWFA     bool           `gorm:"not null;default:false"     json:"can_wfa"`
	CreatedAt  time.Time      `gorm:"not null;default:now()"    json:"created_at"`
	UpdatedAt  *time.Time     `                                  json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                      json:"deleted_at"`
}
