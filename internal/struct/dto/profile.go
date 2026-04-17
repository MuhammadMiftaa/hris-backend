package dto

import "time"

// ── Simple Profile (untuk header/sidebar cache) ────────────────────

type ProfileResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FullName  string    `json:"full_name"`
	PhotoURL  string    `json:"photo_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
}

// ── Photo ──────────────────────────────────────────────────────────

type UploadPhotoRequest struct {
	Base64Image string `json:"base64_image"`
}

type UploadPhotoResponse struct {
	Success  bool   `json:"success"`
	PhotoURL string `json:"photo_url"`
	Message  string `json:"message"`
}

// ── Employee Profile (detail untuk ProfilePage) ────────────────────

type EmployeeProfileResponse struct {
	ID             uint    `json:"id"`
	EmployeeNumber string  `json:"employee_number"`
	FullName       string  `json:"full_name"`
	PhotoURL       *string `json:"photo_url"`

	NIK           *string `json:"nik"`
	NPWP          *string `json:"npwp"`
	KKNumber      *string `json:"kk_number"`
	BirthDate     *string `json:"birth_date"`
	BirthPlace    *string `json:"birth_place"`
	Gender        *string `json:"gender"`
	Religion      *string `json:"religion"`
	MaritalStatus *string `json:"marital_status"`
	BloodType     *string `json:"blood_type"`
	Nationality   *string `json:"nationality"`

	BranchID         *uint   `json:"branch_id"`
	DepartmentID     *uint   `json:"department_id"`
	RoleID           *uint   `json:"role_id"`
	JobPositionsID   *uint   `json:"job_positions_id"`
	BranchName       *string `json:"branch_name"`
	DepartmentName   *string `json:"department_name"`
	RoleName         *string `json:"role_name"`
	JobPositionTitle *string `json:"job_position_title"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type EmployeeProfileContactResponse struct {
	ID           uint    `json:"id"`
	ContactType  string  `json:"contact_type"`
	ContactValue string  `json:"contact_value"`
	ContactLabel *string `json:"contact_label"`
	IsPrimary    bool    `json:"is_primary"`
}

// ── Change Password ────────────────────────────────────────────────

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}
