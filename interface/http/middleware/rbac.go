package middleware

import (
	"fmt"
	"slices"

	"hris-backend/config/log"
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
			account, _ := c.Locals("account").(dto.GetEmployeeByIDResponse)
			log.Debug(fmt.Sprintf("Permission denied for user with ID: %d and Name: %s", account.AccountID, account.FullName), map[string]any{
				"permissions":        permList,
				"permissions_needed": permNeed,
			})
			return c.Status(fiber.StatusForbidden).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 403,
				Message:    "Forbidden",
			})
		}

		return c.Next()
	}
}
