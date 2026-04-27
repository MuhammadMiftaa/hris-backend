package handler

import (
	"fmt"
	"strings"

	"hris-backend/config/log"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

// getAccountFromCtx — ambil account dari context (di-set oleh AuthMiddleware)
func getAccountFromCtx(c *fiber.Ctx) dto.GetEmployeeByIDResponse {
	account, _ := c.Locals("account").(dto.GetEmployeeByIDResponse)
	return account
}

func respondBadRequest(c *fiber.Ctx, msg string) error {
	account := getAccountFromCtx(c)
	log.Error(fmt.Sprintf("Error: %s", msg), map[string]any{
		"account_id": account.AccountID,
		"full_name":  account.FullName,
		"email":      account.Email,
	})
	return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
		Status:     false,
		StatusCode: fiber.StatusBadRequest,
		Message:    msg,
	})
}

func respondError(c *fiber.Ctx, err error) error {
	statusCode := fiber.StatusInternalServerError
	errMsg := err.Error()
	lowerMsg := strings.ToLower(errMsg)

	if strings.Contains(lowerMsg, "record not found") || strings.Contains(lowerMsg, "not found") {
		statusCode = fiber.StatusNotFound
	} else if strings.Contains(lowerMsg, "bad request") || strings.Contains(lowerMsg, "invalid") {
		errMsg = strings.ReplaceAll(errMsg, "bad request: ", "")
		statusCode = fiber.StatusBadRequest
	} else if strings.Contains(lowerMsg, "unauthorized") {
		statusCode = fiber.StatusUnauthorized
	} else if strings.Contains(lowerMsg, "conflict") || strings.Contains(lowerMsg, "already exists") {
		statusCode = fiber.StatusConflict
	}

	account := getAccountFromCtx(c)
	log.Error(fmt.Sprintf("Error: %s", err.Error()), map[string]any{
		"status_code": statusCode,
		"account_id":  account.AccountID,
		"full_name":   account.FullName,
		"email":       account.Email,
	})
	return c.Status(statusCode).JSON(dto.APIResponse{
		Status:     false,
		StatusCode: statusCode,
		Message:    errMsg,
	})
}
