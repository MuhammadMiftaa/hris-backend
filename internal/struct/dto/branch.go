package dto

import "time"

type BranchResponse struct {
	ID           uint       `json:"id"`
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	Address      *string    `json:"address"`
	Latitude     *float64   `json:"latitude"`
	Longitude    *float64   `json:"longitude"`
	RadiusMeters int        `json:"radius_meters"`
	AllowWFH     bool       `json:"allow_wfh"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type CreateBranchRequest struct {
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Address      *string  `json:"address"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	RadiusMeters *int     `json:"radius_meters"`
	AllowWFH     *bool    `json:"allow_wfh"`
}

type UpdateBranchRequest struct {
	Code         *string  `json:"code"`
	Name         *string  `json:"name"`
	Address      *string  `json:"address"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	RadiusMeters *int     `json:"radius_meters"`
	AllowWFH     *bool    `json:"allow_wfh"`
}

type BranchListParams struct {
	PaginationParams
	Search *string `query:"search"`
}
