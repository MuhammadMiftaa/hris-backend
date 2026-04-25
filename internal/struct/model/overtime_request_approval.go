package model

import "time"

type OvertimeRequestApproval struct {
	ID                uint               `gorm:"primaryKey;autoIncrement"             json:"id"`
	OvertimeRequestID uint               `gorm:"not null;index"                       json:"overtime_request_id"`
	ApproverID        *uint              `gorm:"index"                                json:"approver_id"`
	Level             int                `gorm:"not null"                             json:"level"`
	Status            ApprovalStatusEnum `gorm:"type:approval_status_enum;not null;default:pending" json:"status"`
	Notes             *string            `gorm:"type:text"                            json:"notes"`
	DecidedAt         *time.Time         `                                            json:"decided_at"`
	CreatedAt         time.Time          `gorm:"not null;default:now()"              json:"created_at"`

	// Relations
	OvertimeRequest OvertimeRequest `gorm:"foreignKey:OvertimeRequestID" json:"overtime_request,omitempty"`
	Approver        Employee        `gorm:"foreignKey:ApproverID"       json:"approver,omitempty"`
}
