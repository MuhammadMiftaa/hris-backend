package middleware

import (
	"errors"
	"strings"

	"hris-backend/internal/redis"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(rdb redis.Redis) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 401,
				Message:    "Authorization token is required",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 401,
				Message:    "Invalid authorization format",
			})
		}

		accessToken := parts[1]

		isExist, err := rdb.Exists(c.Context(), accessToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 401,
				Message:    "Failed to check token existence",
			})
		}

		if isExist == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 401,
				Message:    "Invalid or expired token",
			})
		}

		token, err := redis.GetToken(c.Context(), rdb, accessToken)
		if err != nil {
			if errors.Is(err, redis.ServerErrInvalidToken) || errors.Is(err, redis.ServerErrTokenExpired) {
				return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
					Status:     false,
					StatusCode: 401,
					Message:    "Invalid or expired token",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 500,
				Message:    "Internal Server Error",
			})
		}

		c.Locals("token", accessToken)
		c.Locals("account", token.Account)
		c.Locals("permissions", token.Permissions)
		c.Locals("refresh_token", token.Refresh)
		return c.Next()
	}
}
