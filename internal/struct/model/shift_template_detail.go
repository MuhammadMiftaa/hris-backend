package model

import (
	"time"

	"gorm.io/gorm"
)

type DayOfWeekEnum string

const (
	DayMonday    DayOfWeekEnum = "monday"
	DayTuesday   DayOfWeekEnum = "tuesday"
	DayWednesday DayOfWeekEnum = "wednesday"
	DayThursday  DayOfWeekEnum = "thursday"
	DayFriday    DayOfWeekEnum = "friday"
	DaySaturday  DayOfWeekEnum = "saturday"
	DaySunday    DayOfWeekEnum = "sunday"
)

type ShiftTemplateDetail struct {
	ID              uint           `gorm:"primaryKey;autoIncrement"                                    json:"id"`
	ShiftTemplateID uint           `gorm:"not null;index;uniqueIndex:uq_shift_day"                     json:"shift_template_id"`
	DayOfWeek       DayOfWeekEnum  `gorm:"type:day_of_week_enum;not null;uniqueIndex:uq_shift_day"     json:"day_of_week"`
	IsWorkingDay    *bool           `gorm:"not null;default:true"                                       json:"is_working_day"`
	ClockInStart    *string        `gorm:"type:time"                                                   json:"clock_in_start"`
	ClockInEnd      *string        `gorm:"type:time"                                                   json:"clock_in_end"`
	BreakDhuhrStart *string        `gorm:"type:time"                                                   json:"break_dhuhr_start"`
	BreakDhuhrEnd   *string        `gorm:"type:time"                                                   json:"break_dhuhr_end"`
	BreakAsrStart   *string        `gorm:"type:time"                                                   json:"break_asr_start"`
	BreakAsrEnd     *string        `gorm:"type:time"                                                   json:"break_asr_end"`
	ClockOutStart   *string        `gorm:"type:time"                                                   json:"clock_out_start"`
	ClockOutEnd     *string        `gorm:"type:time"                                                   json:"clock_out_end"`
	CreatedAt       time.Time      `gorm:"not null;default:now()"                                      json:"created_at"`
	UpdatedAt       *time.Time     `                                                                   json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index"                                                       json:"deleted_at"`

	// Relations
	ShiftTemplate ShiftTemplate `gorm:"foreignKey:ShiftTemplateID" json:"shift_template,omitempty"`
}
