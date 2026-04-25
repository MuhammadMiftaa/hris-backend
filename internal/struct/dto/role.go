package dto

import "time"

type RoleResponse struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Level           string     `json:"level"`
	Description     *string    `json:"description"`
	PermissionCount *int       `json:"permission_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type RoleDetailResponse struct {
	RoleResponse
	Permissions []PermissionResponse `json:"permissions"`
}

type PermissionResponse struct {
	Code        string     `json:"code"`
	Module      string     `json:"module"`
	Action      string     `json:"action"`
	Description *string    `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type CreateRoleRequest struct {
	Name        string  `json:"name"`
	Level       string  `json:"level"`
	Description *string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name"`
	Level       *string `json:"level"`
	Description *string `json:"description"`
}

type UpdateRolePermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes"`
}

type RoleMetadata struct {
	ModuleMeta []Meta `json:"module_meta"`
	ActionMeta []Meta `json:"action_meta"`
	LevelMeta  []Meta `json:"level_meta"`
}
