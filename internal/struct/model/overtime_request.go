package model

import (
	"time"

	"gorm.io/gorm"
)

type WorkLocationTypeEnum string

const (
	WorkLocationOnsite  WorkLocationTypeEnum = "onsite"
	WorkLocationWFH     WorkLocationTypeEnum = "wfh"
	WorkLocationOffsite WorkLocationTypeEnum = "offsite"
)

type OvertimeRequest struct {
	ID               uint                  `gorm:"primaryKey;autoIncrement"                          json:"id"`
	EmployeeID       uint                  `gorm:"not null;index"                                    json:"employee_id"`
	AttendanceLogID  *uint                 `gorm:"index"                                             json:"attendance_log_id"`
	OvertimeDate     time.Time             `gorm:"type:date;not null"                                json:"overtime_date"`
	PlannedStart     *time.Time            `                                                         json:"planned_start"`
	PlannedEnd       *time.Time            `                                                         json:"planned_end"`
	ActualStart      *time.Time            `                                                         json:"actual_start"`
	ActualEnd        *time.Time            `                                                         json:"actual_end"`
	PlannedMinutes   int                   `gorm:"not null"                                          json:"planned_minutes"`
	ActualMinutes    *int                  `                                                         json:"actual_minutes"`
	Reason           string                `gorm:"type:text;not null"                                json:"reason"`
	WorkLocationType *WorkLocationTypeEnum `gorm:"type:work_location_type_enum"                      json:"work_location_type"`
	Status           LeaveRequestStatusEnum `gorm:"type:leave_request_status_enum;not null;default:pending" json:"status"`
	CreatedAt        time.Time             `gorm:"not null;default:now()"                           json:"created_at"`
	UpdatedAt        *time.Time            `                                                         json:"updated_at"`
	DeletedAt        gorm.DeletedAt        `gorm:"index"                                             json:"deleted_at"`

	// Relations
	Employee      Employee       `gorm:"foreignKey:EmployeeID"      json:"employee,omitempty"`
	AttendanceLog *AttendanceLog `gorm:"foreignKey:AttendanceLogID" json:"attendance_log,omitempty"`
}
