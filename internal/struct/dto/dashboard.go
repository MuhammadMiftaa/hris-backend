package dto

type TodayAttendanceStatus struct {
	HasClockedIn  bool    `json:"has_clocked_in"`
	HasClockedOut bool    `json:"has_clocked_out"`
	ClockInAt     *string `json:"clock_in_at"`
	ClockOutAt    *string `json:"clock_out_at"`
	Status        *string `json:"status"`
	LateMinutes   int     `json:"late_minutes"`
}

type AttendanceSummaryDTO struct {
	TotalPresent      int `json:"total_present"`
	TotalLate         int `json:"total_late"`
	TotalAbsent       int `json:"total_absent"`
	TotalLeave        int `json:"total_leave"`
	TotalBusinessTrip int `json:"total_business_trip"`
	TotalHalfDay      int `json:"total_half_day"`
}

type LeaveBalanceSummaryDTO struct {
	LeaveTypeID   uint   `json:"leave_type_id"`
	LeaveTypeName string `json:"leave_type_name"`
	TotalQuota    *int   `json:"total_quota"`
	Used          int    `json:"used"`
	Remaining     *int   `json:"remaining"`
}

type PendingRequestDTO struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Label     string `json:"label"`
	CreatedAt string `json:"created_at"`
	Status    string `json:"status"`
}

// EmployeeDashboardResponse — sesuai kontrak frontend EmployeeDashboardData (§13.1)
type EmployeeDashboardResponse struct {
	Today           TodayAttendanceStatus    `json:"today"`
	MutabaahToday   interface{}              `json:"mutabaah_today"`
	MonthlySummary  AttendanceSummaryDTO     `json:"monthly_summary"`
	LeaveBalances   []LeaveBalanceSummaryDTO `json:"leave_balances"`
	PendingRequests []PendingRequestDTO      `json:"pending_requests"`
}

type ApprovalQueueItemDTO struct {
	ID           uint   `json:"id"`
	Type         string `json:"type"`
	EmployeeName string `json:"employee_name"`
	Label        string `json:"label"`
	CreatedAt    string `json:"created_at"`
}

type ApprovalCountsDTO struct {
	Leave        int `json:"leave"`
	Permission   int `json:"permission"`
	Overtime     int `json:"overtime"`
	BusinessTrip int `json:"business_trip"`
	Override     int `json:"override"`
	Total        int `json:"total"`
}

type TeamAttendanceSummaryDTO struct {
	TotalEmployees int `json:"total_employees"`
	PresentToday   int `json:"present_today"`
	LateToday      int `json:"late_today"`
	NotClockedIn   int `json:"not_clocked_in"`
	OnLeave        int `json:"on_leave"`
}

type TeamMutabaahSummaryDTO struct {
	TotalEmployees    int `json:"total_employees"`
	SubmittedCount    int `json:"submitted_count"`
	NotSubmittedCount int `json:"not_submitted_count"`
}

type NotClockedInDTO struct {
	EmployeeID     uint    `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	EmployeeNumber string  `json:"employee_number"`
	DepartmentName *string `json:"department_name"`
	ShiftStart     *string `json:"shift_start"`
}

type ExpiringContractDTO struct {
	EmployeeID     uint   `json:"employee_id"`
	EmployeeName   string `json:"employee_name"`
	EmployeeNumber string `json:"employee_number"`
	ContractType   string `json:"contract_type"`
	EndDate        string `json:"end_date"`
	DaysRemaining  int    `json:"days_remaining"`
}

// HRDDashboardResponse — sesuai kontrak frontend HRDDashboardData (§13.2)
type HRDDashboardResponse struct {
	ApprovalQueue     []ApprovalQueueItemDTO   `json:"approval_queue"`
	ApprovalCounts    ApprovalCountsDTO        `json:"approval_counts"`
	TeamAttendance    TeamAttendanceSummaryDTO `json:"team_attendance"`
	TeamMutabaah      TeamMutabaahSummaryDTO   `json:"team_mutabaah"`
	NotClockedIn      []NotClockedInDTO        `json:"not_clocked_in"`
	ExpiringContracts []ExpiringContractDTO    `json:"expiring_contracts"`
}
