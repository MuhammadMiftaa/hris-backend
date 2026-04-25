package dto

import "time"

type PermissionRequestResponse struct {
	ID             uint       `json:"id"`
	EmployeeID     uint       `json:"employee_id"`
	EmployeeName   *string    `json:"employee_name"`
	Date           string     `json:"date"`
	PermissionType string     `json:"permission_type"`
	LeaveTime      *string    `json:"leave_time"`
	ReturnTime     *string    `json:"return_time"`
	Reason         string     `json:"reason"`
	DocumentURL    *string    `json:"document_url"`
	Status         string     `json:"status"`
	ApprovedBy     *uint      `json:"approved_by"`
	ApproverName   *string    `json:"approver_name"`
	ApproverNotes  *string    `json:"approver_notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
}

type CreatePermissionRequest struct {
	EmployeeID     *uint   `json:"employee_id"`
	Date           string  `json:"date"`
	PermissionType string  `json:"permission_type"`
	LeaveTime      *string `json:"leave_time"`
	ReturnTime     *string `json:"return_time"`
	Reason         string  `json:"reason"`
	DocumentURL    *string `json:"document_url"`
}

type UpdatePermissionStatusRequest struct {
	Status string  `json:"status"`
	Notes  *string `json:"approver_notes"`
}

type PermissionListParams struct {
	EmployeeID     *uint   `query:"employee_id"`
	Status         *string `query:"status"`
	PermissionType *string `query:"permission_type"`
	StartDate      *string `query:"start_date"`
	EndDate        *string `query:"end_date"`
}

type RequestMetadata struct {
	PermissionTypeMeta []Meta `json:"permission_type_meta"`
	WorkLocationMeta   []Meta `json:"work_location_meta"`
	StatusMeta         []Meta `json:"status_meta"`
	EmployeeMeta       []Meta `json:"employee_meta"`
}
