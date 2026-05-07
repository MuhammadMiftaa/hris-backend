package dto

import "time"

type LeaveBalanceResponse struct {
	ID                   uint       `json:"id"`
	EmployeeID           uint       `json:"employee_id"`
	EmployeeName         *string    `json:"employee_name"`
	DepartmentName       *string    `json:"department_name"`
	LeaveTypeID          uint       `json:"leave_type_id"`
	LeaveTypeName        *string    `json:"leave_type_name"`
	Year                 int        `json:"year"`
	UsedOccurrences      int        `json:"used_occurrences"`
	UsedDuration         float64    `json:"used_duration"`
	AllocatedDuration    float64    `json:"allocated_duration"`
	MaxOccurrences       *int       `json:"max_occurrences"`
	MaxDuration          *float64   `json:"max_duration"`
	RemainingOccurrences *int       `json:"remaining_occurrences"`
	RemainingDuration    *float64   `json:"remaining_duration"`
	EffectiveDate        string     `json:"effective_date"`
	Notes                *string    `json:"notes"`
	TotalAdjustment      float64    `json:"total_adjustment"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at"`
}

type EmployeeBalanceSummaryResponse struct {
	EmployeeID       uint    `json:"employee_id"`
	EmployeeName     string  `json:"employee_name"`
	DepartmentName   *string `json:"department_name"`
	JobPositionTitle *string `json:"job_position_title"`
	Year             int     `json:"year"`
	TotalAllocated   float64 `json:"total_allocated"`
	TotalUsed        float64 `json:"total_used"`
	TotalRemaining   float64 `json:"total_remaining"`
}

type EmployeeBalanceSummaryParams struct {
	PaginationParams
	Year             *int    `query:"year"`
	EmployeeName     *string `query:"employee_name"`
	DepartmentID     *uint   `query:"department_id"`
	JobPositionTitle *string `query:"job_position_title"`
}

type UpsertLeaveBalanceRequest struct {
	EmployeeID        uint    `json:"employee_id"`
	LeaveTypeID       uint    `json:"leave_type_id"`
	Year              int     `json:"year"`
	AllocatedDuration float64 `json:"allocated_duration"`
	EffectiveDate     string  `json:"effective_date"`
	Notes             *string `json:"notes"`
}

type AdjustLeaveBalanceRequest struct {
	Delta  float64 `json:"delta"`
	Reason *string `json:"reason"`
}

type EmployeeBalanceDetailResponse struct {
	EmployeeID   uint                   `json:"employee_id"`
	EmployeeName string                 `json:"employee_name"`
	Year         int                    `json:"year"`
	Balances     []LeaveBalanceResponse `json:"balances"`
}

type LeaveBalanceAdjustmentResponse struct {
	ID             uint    `json:"id"`
	LeaveBalanceID uint    `json:"leave_balance_id"`
	AdjusterName   string  `json:"adjuster_name"`
	Delta          float64 `json:"delta"`
	Reason         *string `json:"reason"`
	CreatedAt      string  `json:"created_at"`
}

type LeaveRequestResponse struct {
	ID             uint                    `json:"id"`
	EmployeeID     uint                    `json:"employee_id"`
	EmployeeName   *string                 `json:"employee_name"`
	LeaveTypeID    uint                    `json:"leave_type_id"`
	LeaveTypeName  *string                 `json:"leave_type_name"`
	DepartmentName *string                 `json:"department_name"`
	LeaveCategory  *string                 `json:"leave_category"`
	StartDate      string                  `json:"start_date"`
	EndDate        string                  `json:"end_date"`
	TotalDays      float64                 `json:"total_days"`
	TotalHours     *int                    `json:"total_hours"`
	Reason         *string                 `json:"reason"`
	DocumentURL    *string                 `json:"document_url"`
	Status         string                  `json:"status"`
	Approvals      []LeaveApprovalResponse `json:"approvals,omitempty" gorm:"-"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      *time.Time              `json:"updated_at"`
	DeletedAt      *time.Time              `json:"deleted_at"`
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
	TotalDays   float64 `json:"total_days"`
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
	PaginationParams
	EmployeeID   *uint   `query:"employee_id"`
	Year         *string `query:"year"`
	UsedDuration *string `query:"used_duration"`
	MaxDuration  *string `query:"max_duration"`
	EmployeeName *string `query:"employee_name"`
	DepartmentID *uint   `query:"department_id"`
	LeaveTypeID  *uint   `query:"leave_type_id"`
}

type LeaveRequestListParams struct {
	PaginationParams
	EmployeeID   *uint   `query:"employee_id"`
	EmployeeName *string `query:"employee_name"`
	DepartmentID *uint   `query:"department_id"`
	Status       *string `query:"status"`
	LeaveTypeID  *uint   `query:"leave_type_id"`
	Year         *int    `query:"year"`
	StartDate    *string `query:"start_date"`
	EndDate      *string `query:"end_date"`
	TotalDays    *string `query:"total_days"`
	Reason       *string `query:"reason"`
}

type LeaveMetadata struct {
	LeaveTypeMeta  []Meta `json:"leave_type_meta"`
	StatusMeta     []Meta `json:"status_meta"`
	EmployeeMeta   []Meta `json:"employee_meta"`
	DepartmentMeta []Meta `json:"department_meta"`
}
