package model

import (
	"time"

	"gorm.io/gorm"
)

type LeaveRequestStatusEnum string

const (
	LeaveRequestPending  LeaveRequestStatusEnum = "pending"
	LeaveRequestApproved LeaveRequestStatusEnum = "approved"
	LeaveRequestRejected LeaveRequestStatusEnum = "rejected"
	LeaveRequestCanceled LeaveRequestStatusEnum = "canceled"
)

type LeaveRequest struct {
	ID          uint                   `gorm:"primaryKey;autoIncrement"                     json:"id"`
	EmployeeID  uint                   `gorm:"not null;index"                               json:"employee_id"`
	LeaveTypeID uint                   `gorm:"not null;index"                               json:"leave_type_id"`
	StartDate   time.Time              `gorm:"type:date;not null"                           json:"start_date"`
	EndDate     time.Time              `gorm:"type:date;not null"                           json:"end_date"`
	TotalDays   float64                `gorm:"not null"                                     json:"total_days"`
	TotalHours  *int                   `                                                    json:"total_hours"`
	Reason      *string                `gorm:"type:text"                                    json:"reason"`
	DocumentURL *string                `gorm:"type:text"                                    json:"document_url"`
	Status      LeaveRequestStatusEnum `gorm:"type:leave_request_status_enum;not null;default:pending" json:"status"`
	CreatedAt   time.Time              `gorm:"not null;default:now()"                      json:"created_at"`
	UpdatedAt   *time.Time             `                                                    json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index"                                        json:"deleted_at"`

	// Relations
	Employee  Employee  `gorm:"foreignKey:EmployeeID"  json:"employee,omitempty"`
	LeaveType LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
}
