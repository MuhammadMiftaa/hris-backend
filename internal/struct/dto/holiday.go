package dto

import "time"

// ── Holiday Response ───────────────────────────────────

type HolidayResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Year        int        `json:"year"`
	Date        string     `json:"date"`
	Type        string     `json:"type"`
	BranchID    *uint      `json:"branch_id"`
	BranchName  *string    `json:"branch_name"`
	Description *string    `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

// ── Metadata ───────────────────────────────────────────

type HolidayMetadata struct {
	HolidayTypeMeta []Meta `json:"holiday_type_meta"`
	BranchMeta      []Meta `json:"branch_meta"`
}

// ── Requests ───────────────────────────────────────────

type CreateHolidayRequest struct {
	Name        string  `json:"name"`
	Date        string  `json:"date"`
	Type        string  `json:"type"`
	BranchID    *uint   `json:"branch_id"`
	Description *string `json:"description"`
}

type UpdateHolidayRequest struct {
	Name        *string `json:"name"`
	Date        *string `json:"date"`
	Type        *string `json:"type"`
	BranchID    *uint   `json:"branch_id"`
	Description *string `json:"description"`
}

// ── List Params ────────────────────────────────────────

type HolidayListParams struct {
	Year     *int
	Type     *string
	BranchID *uint
}

// ── External API ───────────────────────────────────────

type ExternalHolidayItem struct {
	ID                int    `json:"id"`
	Date              string `json:"date"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	IsNationalHoliday bool   `json:"is_national_holiday"`
}

type ExternalHolidayPaging struct {
	Page      int `json:"page"`
	Size      int `json:"size"`
	TotalItem int `json:"total_item"`
	TotalPage int `json:"total_page"`
}

type ExternalHolidayAPIResponse struct {
	IsSuccess bool                  `json:"is_success"`
	Message   string                `json:"message"`
	Data      []ExternalHolidayItem `json:"data"`
	Paging    ExternalHolidayPaging `json:"paging"`
}

type SyncHolidayRequest struct {
	Year     int   `json:"year"`
	BranchID *uint `json:"branch_id"`
}

type SyncHolidayResponse struct {
	Synced  int      `json:"synced"`
	Skipped int      `json:"skipped"`
	Year    int      `json:"year"`
	Errors  []string `json:"errors,omitempty"`
}
