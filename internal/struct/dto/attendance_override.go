package dto

import "time"

type CreateOverrideRequest struct {
	AttendanceLogID   uint    `json:"attendance_log_id"`
	OverrideType      string  `json:"override_type"`
	CorrectedClockIn  *string `json:"corrected_clock_in"`
	CorrectedClockOut *string `json:"corrected_clock_out"`
	Reason            string  `json:"reason"`
}

type UpdateOverrideStatusRequest struct {
	Status        string  `json:"status"`
	ApproverNotes *string `json:"approver_notes"`
}

type OverrideListParams struct {
	EmployeeID *uint   `query:"employee_id"`
	Status     *string `query:"status"`
}

type AttendanceOverrideResponse struct {
	ID                uint       `json:"id"`
	AttendanceLogID   uint       `json:"attendance_log_id"`
	AttendanceDate    *string    `json:"attendance_date"`
	RequestedBy       uint       `json:"requested_by"`
	RequesterName     *string    `json:"requester_name"`
	ApprovedBy        *uint      `json:"approved_by"`
	ApproverName      *string    `json:"approver_name"`
	OverrideType      string     `json:"override_type"`
	OriginalClockIn   *time.Time `json:"original_clock_in"`
	OriginalClockOut  *time.Time `json:"original_clock_out"`
	CorrectedClockIn  *time.Time `json:"corrected_clock_in"`
	CorrectedClockOut *time.Time `json:"corrected_clock_out"`
	Reason            string     `json:"reason"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
}

type CreateManualAttendanceRequest struct {
	EmployeeID     uint    `json:"employee_id"`
	AttendanceDate string  `json:"attendance_date"`
	ClockInAt      string  `json:"clock_in_at"`
	ClockOutAt     *string `json:"clock_out_at"`
	Notes          string  `json:"notes"`
}

type AttendanceMetadata struct {
	StatusMeta       []Meta `json:"status_meta"`
	ClockMethodMeta  []Meta `json:"clock_method_meta"`
	OverrideTypeMeta []Meta `json:"override_type_meta"`
	EmployeeMeta     []Meta `json:"employee_meta"`
	BranchMeta       []Meta `json:"branch_meta"`
}
