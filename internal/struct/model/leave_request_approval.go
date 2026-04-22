package model

import "time"

type ApprovalStatusEnum string

const (
	ApprovalPending  ApprovalStatusEnum = "pending"
	ApprovalApproved ApprovalStatusEnum = "approved"
	ApprovalRejected ApprovalStatusEnum = "rejected"
)

type LeaveRequestApproval struct {
	ID             uint               `gorm:"primaryKey;autoIncrement"             json:"id"`
	LeaveRequestID uint               `gorm:"not null;index"                       json:"leave_request_id"`
	ApproverID     *uint              `gorm:"index"                                json:"approver_id"`
	Level          int                `gorm:"not null"                             json:"level"`
	Status         ApprovalStatusEnum `gorm:"type:approval_status_enum;not null;default:pending" json:"status"`
	Notes          *string            `gorm:"type:text"                            json:"notes"`
	DecidedAt      *time.Time         `                                            json:"decided_at"`
	CreatedAt      time.Time          `gorm:"not null;default:now()"              json:"created_at"`

	// Relations
	LeaveRequest LeaveRequest `gorm:"foreignKey:LeaveRequestID" json:"leave_request,omitempty"`
	Approver     Employee     `gorm:"foreignKey:ApproverID"     json:"approver,omitempty"`
}
