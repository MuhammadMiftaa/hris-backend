package dto

import "time"

// ── Shift Template ──────────────────────────────────────

type ShiftTemplateResponse struct {
	ID         uint                       `json:"id"`
	Name       string                     `json:"name"`
	IsFlexible bool                       `json:"is_flexible"`
	Details    []ShiftTemplateDetailResp  `json:"details"`
	CreatedAt  time.Time                  `json:"created_at"`
	UpdatedAt  *time.Time                 `json:"updated_at"`
	DeletedAt  *time.Time                 `json:"deleted_at"`
}

type ShiftTemplateDetailResp struct {
	ID              uint       `json:"id"`
	ShiftTemplateID uint       `json:"shift_template_id"`
	DayOfWeek       string     `json:"day_of_week"`
	IsWorkingDay    bool       `json:"is_working_day"`
	ClockInStart    *string    `json:"clock_in_start"`
	ClockInEnd      *string    `json:"clock_in_end"`
	BreakDhuhrStart *string    `json:"break_dhuhr_start"`
	BreakDhuhrEnd   *string    `json:"break_dhuhr_end"`
	BreakAsrStart   *string    `json:"break_asr_start"`
	BreakAsrEnd     *string    `json:"break_asr_end"`
	ClockOutStart   *string    `json:"clock_out_start"`
	ClockOutEnd     *string    `json:"clock_out_end"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}

// ── Employee Schedule ──────────────────────────────────

type ScheduleResponse struct {
	ID              uint       `json:"id"`
	EmployeeID      uint       `json:"employee_id"`
	EmployeeName    *string    `json:"employee_name"`
	EmployeeNumber  *string    `json:"employee_number"`
	ShiftTemplateID uint       `json:"shift_template_id"`
	ShiftName       *string    `json:"shift_name"`
	EffectiveDate   string     `json:"effective_date"`
	EndDate         *string    `json:"end_date"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}

// ── Metadata ───────────────────────────────────────────

type ShiftMetadata struct {
	DayOfWeekMeta []Meta `json:"day_of_week_meta"`
}

// ── Requests ───────────────────────────────────────────

type CreateShiftDetailRequest struct {
	DayOfWeek       string  `json:"day_of_week"`
	IsWorkingDay    bool    `json:"is_working_day"`
	ClockInStart    *string `json:"clock_in_start"`
	ClockInEnd      *string `json:"clock_in_end"`
	BreakDhuhrStart *string `json:"break_dhuhr_start"`
	BreakDhuhrEnd   *string `json:"break_dhuhr_end"`
	BreakAsrStart   *string `json:"break_asr_start"`
	BreakAsrEnd     *string `json:"break_asr_end"`
	ClockOutStart   *string `json:"clock_out_start"`
	ClockOutEnd     *string `json:"clock_out_end"`
}

type CreateShiftRequest struct {
	Name       string                     `json:"name"`
	IsFlexible bool                       `json:"is_flexible"`
	Details    []CreateShiftDetailRequest `json:"details"`
}

type UpdateShiftRequest struct {
	Name       *string                    `json:"name"`
	IsFlexible *bool                      `json:"is_flexible"`
	Details    []CreateShiftDetailRequest `json:"details"`
}

type CreateScheduleRequest struct {
	EmployeeID      uint    `json:"employee_id"`
	ShiftTemplateID uint    `json:"shift_template_id"`
	EffectiveDate   string  `json:"effective_date"`
	EndDate         *string `json:"end_date"`
	IsActive        bool    `json:"is_active"`
}

type UpdateScheduleRequest struct {
	EmployeeID      *uint   `json:"employee_id"`
	ShiftTemplateID *uint   `json:"shift_template_id"`
	EffectiveDate   *string `json:"effective_date"`
	EndDate         *string `json:"end_date"`
	IsActive        *bool   `json:"is_active"`
}

type ScheduleListParams struct {
	EmployeeID      *uint
	ShiftTemplateID *uint
	IsActive        *bool
}
