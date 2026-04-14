package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusBadRequest,
			Message:    "invalid request",
		})
	}

	result, err := h.service.Login(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusUnauthorized,
			Message:    "login failed: " + err.Error(),
		})
	}

	c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: fiber.StatusOK,
		Message:    "Login successful",
		Data:       result,
	})
	return nil
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	accessToken := c.Locals("token").(string)
	refreshToken := c.Locals("refresh_token").(string)
	account := c.Locals("account").(dto.GetEmployeeByIDResponse)
	permissions := c.Locals("permissions").([]string)

	result, err := h.service.Refresh(c.Context(), dto.LoginRes{
		Account:     account,
		Permissions: permissions,
		Token:       accessToken,
		Refresh:     refreshToken,
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusUnauthorized,
			Message:    "refresh token failed: " + err.Error(),
		})
	}

	c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: fiber.StatusOK,
		Message:    "Refresh token successful",
		Data:       result,
	})
	return nil
}
