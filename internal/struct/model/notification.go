package model

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID                uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	EmployeeID        uint           `gorm:"not null;index"           json:"employee_id"`
	Type              string         `gorm:"type:varchar(50);not null" json:"type"`
	Title             string         `gorm:"type:varchar(255);not null" json:"title"`
	Body              string         `gorm:"type:text;not null"       json:"body"`
	ActionURL         *string        `gorm:"type:text"                json:"action_url"`
	ActionTab         *string        `gorm:"type:varchar(50)"         json:"action_tab"`
	IsRead            bool           `gorm:"not null;default:false"   json:"is_read"`
	ReadAt            *time.Time     `                                 json:"read_at"`
	PushStatus        string         `gorm:"type:varchar(20);not null;default:pending" json:"push_status"`
	PushAttempts      int            `gorm:"not null;default:0"       json:"push_attempts"`
	LastAttemptAt     *time.Time     `                                 json:"last_attempt_at"`
	RelatedEntityType *string        `gorm:"type:varchar(50)"         json:"related_entity_type"`
	RelatedEntityID   *uint          `                                 json:"related_entity_id"`
	SendAt            time.Time      `gorm:"not null;default:now()"  json:"send_at"`
	CreatedAt         time.Time      `gorm:"not null;default:now()"  json:"created_at"`
	UpdatedAt         *time.Time     `                                 json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index"                    json:"deleted_at"`

	// Relations
	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}
