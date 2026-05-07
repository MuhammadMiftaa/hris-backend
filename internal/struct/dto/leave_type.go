package dto

import (
	"time"

	"hris-backend/internal/struct/model"
)

type LeaveTypeResponse struct {
	ID                      uint                    `json:"id"`
	Name                    string                  `json:"name"`
	Category                model.LeaveCategoryEnum `json:"category"`
	RequiresDocument        bool                    `json:"requires_document"`
	RequiresDocumentType    *string                 `json:"requires_document_type"`
	MaxDurationPerRequest   *float64                `json:"max_duration_per_request"`
	MaxOccurrencesPerYear   *int                    `json:"max_occurrences_per_year"`
	MaxTotalDurationPerYear *float64                `json:"max_total_duration_per_year"`
	MaxPerMonth             *float64                `json:"max_per_month"`
	ParentLeaveTypeID       *uint                   `json:"parent_leave_type_id"`
	DeductDays              float64                 `json:"deduct_days"`
	CreatedAt               time.Time               `json:"created_at"`
	UpdatedAt               *time.Time              `json:"updated_at"`
}

type CreateLeaveTypeRequest struct {
	Name                    string   `json:"name"`
	Category                string   `json:"category"`
	RequiresDocument        bool     `json:"requires_document"`
	RequiresDocumentType    *string  `json:"requires_document_type"`
	MaxDurationPerRequest   *float64 `json:"max_duration_per_request"`
	MaxOccurrencesPerYear   *int     `json:"max_occurrences_per_year"`
	MaxTotalDurationPerYear *float64 `json:"max_total_duration_per_year"`
	MaxPerMonth             *float64 `json:"max_per_month"`
	ParentLeaveTypeID       *uint    `json:"parent_leave_type_id"`
	DeductDays              *float64 `json:"deduct_days"`
}

type UpdateLeaveTypeRequest struct {
	Name                    *string  `json:"name"`
	Category                *string  `json:"category"`
	RequiresDocument        *bool    `json:"requires_document"`
	RequiresDocumentType    *string  `json:"requires_document_type"`
	MaxDurationPerRequest   *float64 `json:"max_duration_per_request"`
	MaxOccurrencesPerYear   *int     `json:"max_occurrences_per_year"`
	MaxTotalDurationPerYear *float64 `json:"max_total_duration_per_year"`
	MaxPerMonth             *float64 `json:"max_per_month"`
	ParentLeaveTypeID       *uint    `json:"parent_leave_type_id"`
	DeductDays              *float64 `json:"deduct_days"`
}

type LeaveTypeMetadata struct {
	CategoryMeta []Meta `json:"category_meta"`
}

type LeaveTypeListParams struct {
	PaginationParams
	MaxTotalDurationPerYear *string `query:"max_total_duration_per_year"`
	MaxDurationPerRequest   *string `query:"max_duration_per_request"`
	Name                    *string `query:"name"`
	Category                *string `query:"category"`
	RequiresDocument        *bool   `query:"requires_document"`
}
