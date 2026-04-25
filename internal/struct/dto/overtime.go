package dto

import "time"

type OvertimeRequestResponse struct {
	ID               uint       `json:"id"`
	EmployeeID       uint       `json:"employee_id"`
	EmployeeName     *string    `json:"employee_name"`
	AttendanceLogID  *uint      `json:"attendance_log_id"`
	OvertimeDate     string     `json:"overtime_date"`
	PlannedStart     *time.Time `json:"planned_start"`
	PlannedEnd       *time.Time `json:"planned_end"`
	ActualStart      *time.Time `json:"actual_start"`
	ActualEnd        *time.Time `json:"actual_end"`
	PlannedMinutes   int        `json:"planned_minutes"`
	ActualMinutes    *int       `json:"actual_minutes"`
	Reason           string     `json:"reason"`
	WorkLocationType string     `json:"work_location_type"`
	Status           string                     `json:"status"`
	Approvals        []OvertimeApprovalResponse `json:"approvals,omitempty" gorm:"-"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        *time.Time                 `json:"updated_at"`
	DeletedAt        *time.Time                 `json:"deleted_at"`
}

type OvertimeApprovalResponse struct {
	ID                uint       `json:"id"`
	OvertimeRequestID uint       `json:"overtime_request_id"`
	ApproverID        *uint      `json:"approver_id"`
	ApproverName      *string    `json:"approver_name"`
	Level             int        `json:"level"`
	Status            string     `json:"status"`
	Notes             *string    `json:"notes"`
	DecidedAt         *time.Time `json:"decided_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

type CreateOvertimeRequest struct {
	EmployeeID       *uint   `json:"employee_id"`
	AttendanceLogID  *uint   `json:"attendance_log_id"`
	OvertimeDate     string  `json:"overtime_date"`
	PlannedStart     *string `json:"planned_start"`
	PlannedEnd       *string `json:"planned_end"`
	PlannedMinutes   int     `json:"planned_minutes"`
	Reason           string  `json:"reason"`
	WorkLocationType string  `json:"work_location_type"`
}

type ApproveOvertimeRequest struct {
	Notes *string `json:"notes"`
}

type RejectOvertimeRequest struct {
	Notes string `json:"notes"`
}

type OvertimeListParams struct {
	EmployeeID *uint   `query:"employee_id"`
	Status     *string `query:"status"`
	StartDate  *string `query:"start_date"`
	EndDate    *string `query:"end_date"`
}
