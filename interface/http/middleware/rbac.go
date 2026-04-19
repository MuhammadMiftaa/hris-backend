package middleware

import (
	"slices"

	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

func RBACMiddleware(permNeed ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permList := c.Locals("permissions").([]string)
		if permList == nil {
			return c.Status(fiber.StatusForbidden).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 403,
				Message:    "Forbidden",
			})
		}

		isAllowed := false
		for _, perm := range permNeed {
			if slices.Contains(permList, perm) {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 403,
				Message:    "Forbidden",
			})
		}

		return c.Next()
	}
}
