package dto

import "time"

type LeaveBalanceResponse struct {
	ID                   uint       `json:"id"`
	EmployeeID           uint       `json:"employee_id"`
	EmployeeName         *string    `json:"employee_name"`
	LeaveTypeID          uint       `json:"leave_type_id"`
	LeaveTypeName        *string    `json:"leave_type_name"`
	Year                 int        `json:"year"`
	UsedOccurrences      int        `json:"used_occurrences"`
	UsedDuration         int        `json:"used_duration"`
	MaxOccurrences       *int       `json:"max_occurrences"`
	MaxDuration          *int       `json:"max_duration"`
	RemainingOccurrences *int       `json:"remaining_occurrences"`
	RemainingDuration    *int       `json:"remaining_duration"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at"`
}

type LeaveRequestResponse struct {
	ID            uint                    `json:"id"`
	EmployeeID    uint                    `json:"employee_id"`
	EmployeeName  *string                 `json:"employee_name"`
	LeaveTypeID   uint                    `json:"leave_type_id"`
	LeaveTypeName *string                 `json:"leave_type_name"`
	LeaveCategory *string                 `json:"leave_category"`
	StartDate     string                  `json:"start_date"`
	EndDate       string                  `json:"end_date"`
	TotalDays     int                     `json:"total_days"`
	TotalHours    *int                    `json:"total_hours"`
	Reason        *string                 `json:"reason"`
	DocumentURL   *string                 `json:"document_url"`
	Status        string                  `json:"status"`
	Approvals     []LeaveApprovalResponse `json:"approvals,omitempty" gorm:"-"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     *time.Time              `json:"updated_at"`
	DeletedAt     *time.Time              `json:"deleted_at"`
}

type LeaveApprovalResponse struct {
	ID             uint       `json:"id"`
	LeaveRequestID uint       `json:"leave_request_id"`
	ApproverID     *uint      `json:"approver_id"`
	ApproverName   *string    `json:"approver_name"`
	Level          int        `json:"level"`
	Status         string     `json:"status"`
	Notes          *string    `json:"notes"`
	DecidedAt      *time.Time `json:"decided_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateLeaveRequest struct {
	EmployeeID  *uint   `json:"employee_id"`
	LeaveTypeID uint    `json:"leave_type_id"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	TotalDays   int     `json:"total_days"`
	TotalHours  *int    `json:"total_hours"`
	Reason      *string `json:"reason"`
	DocumentURL *string `json:"document_url"`
}

type ApproveLeaveRequest struct {
	Notes *string `json:"notes"`
}

type RejectLeaveRequest struct {
	Notes string `json:"notes"`
}

type LeaveBalanceListParams struct {
	EmployeeID *uint `query:"employee_id"`
	Year       *int  `query:"year"`
}

type LeaveRequestListParams struct {
	EmployeeID  *uint   `query:"employee_id"`
	Status      *string `query:"status"`
	LeaveTypeID *uint   `query:"leave_type_id"`
	Year        *int    `query:"year"`
}

type LeaveMetadata struct {
	LeaveTypeMeta []Meta `json:"leave_type_meta"`
	StatusMeta    []Meta `json:"status_meta"`
	EmployeeMeta  []Meta `json:"employee_meta"`
}
