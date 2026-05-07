package model

import (
	"time"
)

type LeaveBalanceAdjustment struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	LeaveBalanceID uint      `gorm:"not null;index"           json:"leave_balance_id"`
	AdjustedBy     uint      `gorm:"not null"                 json:"adjusted_by"`
	Delta          float64   `gorm:"not null"                 json:"delta"`
	Reason         *string   `                                json:"reason"`
	CreatedAt      time.Time `gorm:"not null;default:now()"  json:"created_at"`

	// Relations
	Adjuster Employee `gorm:"foreignKey:AdjustedBy" json:"adjuster,omitempty"`
}
