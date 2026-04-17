package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"hris-backend/config/storage"
	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"
)

type ProfileService interface {
	GetProfile(ctx context.Context, accountID uint) (dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, accountID uint, req dto.UpdateProfileRequest) (dto.ProfileResponse, error)
	UploadPhoto(ctx context.Context, accountID uint, req dto.UploadPhotoRequest) (dto.UploadPhotoResponse, error)
	DeletePhoto(ctx context.Context, accountID uint) error
	GetEmployeeProfile(ctx context.Context, accountID uint) (dto.EmployeeProfileResponse, error)
	GetEmployeeContacts(ctx context.Context, accountID uint) ([]dto.EmployeeProfileContactResponse, error)
	ChangePassword(ctx context.Context, accountID uint, req dto.ChangePasswordRequest) error
}

type profileService struct {
	repo  repository.ProfileRepository
	minio storage.MinioClient
}

func NewProfileService(repo repository.ProfileRepository, minio storage.MinioClient) ProfileService {
	return &profileService{repo: repo, minio: minio}
}

// ─── GetProfile ───────────────────────────────────────────────────────────────

func (s *profileService) GetProfile(ctx context.Context, accountID uint) (dto.ProfileResponse, error) {
	employee, err := s.repo.GetEmployeeByAccountID(ctx, nil, accountID)
	if err != nil {
		return dto.ProfileResponse{}, fmt.Errorf("get employee: %w", err)
	}

	photoURL := ""
	if employee.PhotoURL != nil {
		// Generate presigned download URL for the photo
		url, err := s.minio.PresignedGetObject(ctx, storage.BucketProfilePhotos, *employee.PhotoURL, storage.PresignedDownloadExpiry)
		if err == nil {
			photoURL = url
		}
	}

	updatedAt := employee.CreatedAt
	if employee.UpdatedAt != nil {
		updatedAt = *employee.UpdatedAt
	}

	return dto.ProfileResponse{
		ID:        fmt.Sprintf("%d", accountID),
		UserID:    fmt.Sprintf("%d", employee.ID),
		FullName:  employee.FullName,
		PhotoURL:  photoURL,
		CreatedAt: employee.CreatedAt,
		UpdatedAt: updatedAt,
	}, nil
}

// ─── UpdateProfile ────────────────────────────────────────────────────────────

func (s *profileService) UpdateProfile(ctx context.Context, accountID uint, req dto.UpdateProfileRequest) (dto.ProfileResponse, error) {
	if strings.TrimSpace(req.FullName) == "" {
		return dto.ProfileResponse{}, fmt.Errorf("full_name tidak boleh kosong")
	}

	employee, err := s.repo.GetEmployeeByAccountID(ctx, nil, accountID)
	if err != nil {
		return dto.ProfileResponse{}, fmt.Errorf("get employee: %w", err)
	}

	if err := s.repo.UpdateFullName(ctx, nil, employee.ID, req.FullName); err != nil {
		return dto.ProfileResponse{}, fmt.Errorf("update full name: %w", err)
	}

	return s.GetProfile(ctx, accountID)
}

// ─── UploadPhoto ──────────────────────────────────────────────────────────────

func (s *profileService) UploadPhoto(ctx context.Context, accountID uint, req dto.UploadPhotoRequest) (dto.UploadPhotoResponse, error) {
	if req.Base64Image == "" {
		return dto.UploadPhotoResponse{}, fmt.Errorf("base64_image tidak boleh kosong")
	}

	employee, err := s.repo.GetEmployeeByAccountID(ctx, nil, accountID)
	if err != nil {
		return dto.UploadPhotoResponse{}, fmt.Errorf("get employee: %w", err)
	}

	// Decode base64 — handle "data:image/...;base64,..." prefix
	base64Data := req.Base64Image
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	imageBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return dto.UploadPhotoResponse{}, fmt.Errorf("gagal decode base64: %w", err)
	}

	// Validate size (max 2MB)
	if len(imageBytes) > 2*1024*1024 {
		return dto.UploadPhotoResponse{}, fmt.Errorf("ukuran foto maksimal 2MB")
	}

	// Detect image format from magic bytes
	ext := "jpg"
	if len(imageBytes) > 4 {
		if imageBytes[0] == 0x89 && imageBytes[1] == 0x50 {
			ext = "png"
		}
	}

	// Upload to MinIO
	objectKey := fmt.Sprintf("profile/%d/avatar_%d.%s", employee.ID, time.Now().UnixNano(), ext)

	uploadURL, err := s.minio.PresignedPutObject(ctx, storage.BucketProfilePhotos, objectKey, storage.PresignedUploadExpiry)
	if err != nil {
		return dto.UploadPhotoResponse{}, fmt.Errorf("gagal membuat upload URL: %w", err)
	}

	// Upload directly using HTTP PUT to presigned URL
	if err := utils.UploadToPresignedURL(uploadURL, imageBytes, "image/"+ext); err != nil {
		return dto.UploadPhotoResponse{}, fmt.Errorf("gagal upload foto: %w", err)
	}

	// Save object key to DB
	if err := s.repo.UpdatePhotoURL(ctx, nil, employee.ID, &objectKey); err != nil {
		return dto.UploadPhotoResponse{}, fmt.Errorf("gagal update photo_url: %w", err)
	}

	return dto.UploadPhotoResponse{
		Success:  true,
		PhotoURL: objectKey,
		Message:  "Foto profil berhasil diupload",
	}, nil
}

// ─── DeletePhoto ──────────────────────────────────────────────────────────────

func (s *profileService) DeletePhoto(ctx context.Context, accountID uint) error {
	employee, err := s.repo.GetEmployeeByAccountID(ctx, nil, accountID)
	if err != nil {
		return fmt.Errorf("get employee: %w", err)
	}

	if err := s.repo.UpdatePhotoURL(ctx, nil, employee.ID, nil); err != nil {
		return fmt.Errorf("gagal menghapus photo_url: %w", err)
	}

	return nil
}

// ─── GetEmployeeProfile ───────────────────────────────────────────────────────

func (s *profileService) GetEmployeeProfile(ctx context.Context, accountID uint) (dto.EmployeeProfileResponse, error) {
	profile, err := s.repo.GetEmployeeProfileByAccountID(ctx, nil, accountID)
	if err != nil {
		return dto.EmployeeProfileResponse{}, fmt.Errorf("get employee profile: %w", err)
	}

	// Resolve photo URL via MinIO presigned GET if object key exists
	if profile.PhotoURL != nil && *profile.PhotoURL != "" {
		url, err := s.minio.PresignedGetObject(ctx, storage.BucketProfilePhotos, *profile.PhotoURL, storage.PresignedDownloadExpiry)
		if err == nil {
			profile.PhotoURL = &url
		}
	}

	return profile, nil
}

// ─── GetEmployeeContacts ──────────────────────────────────────────────────────

func (s *profileService) GetEmployeeContacts(ctx context.Context, accountID uint) ([]dto.EmployeeProfileContactResponse, error) {
	employee, err := s.repo.GetEmployeeByAccountID(ctx, nil, accountID)
	if err != nil {
		return nil, fmt.Errorf("get employee: %w", err)
	}

	contacts, err := s.repo.GetContactsByEmployeeID(ctx, nil, employee.ID)
	if err != nil {
		return nil, fmt.Errorf("get contacts: %w", err)
	}

	if contacts == nil {
		contacts = []dto.EmployeeProfileContactResponse{}
	}

	return contacts, nil
}

// ─── ChangePassword ───────────────────────────────────────────────────────────

func (s *profileService) ChangePassword(ctx context.Context, accountID uint, req dto.ChangePasswordRequest) error {
	// Validation
	if strings.TrimSpace(req.OldPassword) == "" {
		return fmt.Errorf("password lama wajib diisi")
	}
	if strings.TrimSpace(req.NewPassword) == "" {
		return fmt.Errorf("password baru wajib diisi")
	}
	if strings.TrimSpace(req.ConfirmPassword) == "" {
		return fmt.Errorf("konfirmasi password wajib diisi")
	}
	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("password baru dan konfirmasi tidak cocok")
	}
	if req.OldPassword == req.NewPassword {
		return fmt.Errorf("password baru tidak boleh sama dengan password lama")
	}
	if len(req.NewPassword) < 8 {
		return fmt.Errorf("password baru minimal 8 karakter")
	}

	// Get account
	account, err := s.repo.GetAccountByID(ctx, nil, accountID)
	if err != nil {
		return fmt.Errorf("akun tidak ditemukan: %w", err)
	}

	// Verify old password
	if !utils.IsPasswordMatch(account.Password, req.OldPassword) {
		return fmt.Errorf("password lama salah")
	}

	// Hash new password
	hashedPassword, err := utils.PasswordHashing(req.NewPassword)
	if err != nil {
		return fmt.Errorf("gagal hash password: %w", err)
	}

	// Update
	if err := s.repo.UpdatePasswordByAccountID(ctx, nil, accountID, hashedPassword); err != nil {
		return fmt.Errorf("gagal update password: %w", err)
	}

	return nil
}
