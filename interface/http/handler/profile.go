package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type ProfileHandler struct {
	service service.ProfileService
}

func NewProfileHandler(service service.ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

// GetProfile — GET /profile
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)
	result, err := h.service.GetProfile(c.Context(), account.AccountID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "profile data",
		Data:       result,
	})
}

// UpdateProfile — PUT /profile
func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request payload")
	}

	account := getAccountFromCtx(c)
	result, err := h.service.UpdateProfile(c.Context(), account.AccountID, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "profile updated",
		Data:       result,
	})
}

// UploadPhoto — POST /profile/photo
func (h *ProfileHandler) UploadPhoto(c *fiber.Ctx) error {
	var req dto.UploadPhotoRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request payload")
	}

	account := getAccountFromCtx(c)
	result, err := h.service.UploadPhoto(c.Context(), account.AccountID, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "photo uploaded",
		Data:       result,
	})
}

// DeletePhoto — DELETE /profile/photo
func (h *ProfileHandler) DeletePhoto(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)
	if err := h.service.DeletePhoto(c.Context(), account.AccountID); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "photo deleted",
	})
}

// GetEmployeeProfile — GET /profile/employee
func (h *ProfileHandler) GetEmployeeProfile(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)
	result, err := h.service.GetEmployeeProfile(c.Context(), account.AccountID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "employee profile",
		Data:       result,
	})
}

// GetEmployeeContacts — GET /profile/employee/contacts
func (h *ProfileHandler) GetEmployeeContacts(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)
	result, err := h.service.GetEmployeeContacts(c.Context(), account.AccountID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "employee contacts",
		Data:       result,
	})
}

// ChangePassword — POST /profile/change-password
func (h *ProfileHandler) ChangePassword(c *fiber.Ctx) error {
	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request payload")
	}

	account := getAccountFromCtx(c)
	if err := h.service.ChangePassword(c.Context(), account.AccountID, req); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "password changed successfully",
	})
}
