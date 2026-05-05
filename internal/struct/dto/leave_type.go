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
	MaxDurationPerRequest   *int                    `json:"max_duration_per_request"`
	MaxDurationUnit         *model.DurationUnitEnum `json:"max_duration_unit"`
	MaxOccurrencesPerYear   *int                    `json:"max_occurrences_per_year"`
	MaxTotalDurationPerYear *int                    `json:"max_total_duration_per_year"`
	MaxTotalDurationUnit    *model.DurationUnitEnum `json:"max_total_duration_unit"`
	CreatedAt               time.Time               `json:"created_at"`
	UpdatedAt               *time.Time              `json:"updated_at"`
}

type CreateLeaveTypeRequest struct {
	Name                    string  `json:"name"`
	Category                string  `json:"category"`
	RequiresDocument        bool    `json:"requires_document"`
	RequiresDocumentType    *string `json:"requires_document_type"`
	MaxDurationPerRequest   *int    `json:"max_duration_per_request"`
	MaxDurationUnit         *string `json:"max_duration_unit"`
	MaxOccurrencesPerYear   *int    `json:"max_occurrences_per_year"`
	MaxTotalDurationPerYear *int    `json:"max_total_duration_per_year"`
	MaxTotalDurationUnit    *string `json:"max_total_duration_unit"`
}

type UpdateLeaveTypeRequest struct {
	Name                    *string `json:"name"`
	Category                *string `json:"category"`
	RequiresDocument        *bool   `json:"requires_document"`
	RequiresDocumentType    *string `json:"requires_document_type"`
	MaxDurationPerRequest   *int    `json:"max_duration_per_request"`
	MaxDurationUnit         *string `json:"max_duration_unit"`
	MaxOccurrencesPerYear   *int    `json:"max_occurrences_per_year"`
	MaxTotalDurationPerYear *int    `json:"max_total_duration_per_year"`
	MaxTotalDurationUnit    *string `json:"max_total_duration_unit"`
}

type LeaveTypeMetadata struct {
	CategoryMeta     []Meta `json:"category_meta"`
	DurationUnitMeta []Meta `json:"duration_unit_meta"`
}

type LeaveTypeListParams struct {
	PaginationParams
	MaxTotalDurationPerYear *string `query:"max_total_duration_per_year"`
	MaxDurationPerRequest   *string `query:"max_duration_per_request"`
	Name                    *string `query:"name"`
	Category                *string `query:"category"`
	RequiresDocument        *bool   `query:"requires_document"`
}
