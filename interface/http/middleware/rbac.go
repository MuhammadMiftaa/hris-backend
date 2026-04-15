package middleware

import (
	"slices"

	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

func RBACMiddleware(permAllowed string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permList := c.Locals("permissions").([]string)
		if permList == nil {
			return c.Status(fiber.StatusForbidden).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 403,
				Message:    "Forbidden",
			})
		}

		if !slices.Contains(permList, permAllowed) {
			return c.Status(fiber.StatusForbidden).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 403,
				Message:    "Forbidden",
			})
		}

		return c.Next()
	}
}
