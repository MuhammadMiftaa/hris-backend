package dto

type TodayAttendanceStatus struct {
	HasClockedIn  bool    `json:"has_clocked_in"`
	HasClockedOut bool    `json:"has_clocked_out"`
	ClockInAt     *string `json:"clock_in_at"`
	ClockOutAt    *string `json:"clock_out_at"`
	Status        *string `json:"status"`
	LateMinutes   int     `json:"late_minutes"`
}

type MutabaahTodayStatus struct {
	HasRecord       bool    `json:"has_record"`
	IsSubmitted     bool    `json:"is_submitted"`
	SubmittedAt     *string `json:"submitted_at"`
	TargetPages     int     `json:"target_pages"`
	MutabaahLogID   *uint   `json:"mutabaah_log_id"`
	AttendanceLogID *uint   `json:"attendance_log_id"`
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

type EmployeeRequestDTO struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Label     string `json:"label"`
	CreatedAt string `json:"created_at"`
	Date      string `json:"date"`
	Status    string `json:"status"`
}

type EmployeeDashboardResponse struct {
	Today            TodayAttendanceStatus    `json:"today"`
	MutabaahToday    *MutabaahTodayStatus     `json:"mutabaah_today"`
	MonthlySummary   AttendanceSummaryDTO     `json:"monthly_summary"`
	LeaveBalances    []LeaveBalanceSummaryDTO `json:"leave_balances"`
	EmployeeRequests []EmployeeRequestDTO     `json:"employee_requests"`
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

// TeamEmployeeAttendanceDTO — pegawai dengan status kehadiran & mutabaah hari ini
type TeamEmployeeAttendanceDTO struct {
	EmployeeID       uint    `json:"employee_id"`
	EmployeeName     string  `json:"employee_name"`
	DepartmentName   *string `json:"department_name"`
	JobPosition      string  `json:"job_position"`
	AttendanceStatus string  `json:"attendance_status"` // present, late, absent, leave, business_trip, half_day
	MutabaahStatus   string  `json:"mutabaah_status"`   // submitted, not_submitted, not_applicable
}

// TeamEmployeeRequestDTO — pengajuan cuti/izin/tugas pegawai hari ini
type TeamEmployeeRequestDTO struct {
	RequestID     uint   `json:"request_id"`
	EmployeeID    uint   `json:"employee_id"`
	EmployeeName  string `json:"employee_name"`
	JobPosition   string `json:"job_position"`
	RequestType   string `json:"request_type"`   // leave, permission, business_trip
	CreatedAt     string `json:"created_at"`
	RequestedDate string `json:"requested_date"`
	Status        string `json:"status"`
	Label         string `json:"label"`
}

// NotClockedOutDTO — pegawai yang sudah clock in tapi belum clock out setelah window jam pulang
type NotClockedOutDTO struct {
	EmployeeID     uint    `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	DepartmentName *string `json:"department_name"`
}

type ExpiringContractDTO struct {
	EmployeeID     uint   `json:"employee_id"`
	EmployeeName   string `json:"employee_name"`
	EmployeeNumber string `json:"employee_number"`
	ContractType   string `json:"contract_type"`
	EndDate        string `json:"end_date"`
	DaysRemaining  int    `json:"days_remaining"`
}

// TeamDashboardResponse — data untuk tab Tim
type TeamDashboardResponse struct {
	TeamAttendance         TeamAttendanceSummaryDTO    `json:"team_attendance"`
	TeamMutabaah           TeamMutabaahSummaryDTO      `json:"team_mutabaah"`
	NotClockedIn           []NotClockedInDTO           `json:"not_clocked_in"`
	NotClockedOut          []NotClockedOutDTO          `json:"not_clocked_out"`
	EmployeeAttendanceList []TeamEmployeeAttendanceDTO `json:"employee_attendance_list"`
	EmployeeRequestList    []TeamEmployeeRequestDTO    `json:"employee_request_list"`
}

// ReportsDashboardResponse — data untuk tab Laporan
type ReportsDashboardResponse struct {
	ApprovalQueue     []ApprovalQueueItemDTO `json:"approval_queue"`
	ApprovalCounts    ApprovalCountsDTO      `json:"approval_counts"`
	ExpiringContracts []ExpiringContractDTO  `json:"expiring_contracts"`
}

// HRDDashboardResponse — DEPRECATED, gunakan TeamDashboardResponse + ReportsDashboardResponse
// Tetap dipertahankan untuk backward compatibility endpoint /dashboard/hrd
type HRDDashboardResponse struct {
	Team    TeamDashboardResponse    `json:"team"`
	Reports ReportsDashboardResponse `json:"reports"`
}

// RankingEntryDTO — satu entry ranking generik
type RankingEntryDTO struct {
	Rank           int     `json:"rank"`
	EmployeeID     uint    `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	EmployeeNumber string  `json:"employee_number"`
	Value          float64 `json:"value"`
	ValueLabel     string  `json:"value_label"` // "64m", "150 hal", dsb
}

// DepartmentRankingDTO — ranking per departemen
type DepartmentRankingDTO struct {
	Rank           int     `json:"rank"`
	DepartmentID   uint    `json:"department_id"`
	DepartmentName string  `json:"department_name"`
	Value          float64 `json:"value"`
	ValueLabel     string  `json:"value_label"` // "95%", dsb
}

// DashboardRankingsResponse — 3 ranking sekaligus
type DashboardRankingsResponse struct {
	FastestArrival  []RankingEntryDTO      `json:"fastest_arrival"`
	TopTilawah      []DepartmentRankingDTO `json:"top_tilawah"`
	FastestMutabaah []RankingEntryDTO      `json:"fastest_mutabaah"`
}

// DashboardMetadataResponse — return used by Dashboard Quick Requests Modal
type DashboardMetadataResponse struct {
	LeaveTypeMeta        []Meta `json:"leave_type_meta"`
	RecentAttendanceMeta []Meta `json:"recent_attendance_meta"`
	EmployeeMeta         []Meta `json:"employee_meta"`
}
