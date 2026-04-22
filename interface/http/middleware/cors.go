package middleware

import (
	"strings"

	"hris-backend/config/env"
	"hris-backend/internal/utils/data"

	"github.com/gofiber/fiber/v2"
)

func CORSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		if isAllowedOrigin(origin) {
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Access-Control-Allow-Credentials", "true")
			c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Request-ID")
			c.Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, X-Request-ID")
			c.Set("Access-Control-Max-Age", "43200")
		}

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

func isAllowedOrigin(origin string) bool {
	if env.Cfg.Server.Mode == data.DEVELOPMENT_MODE {
		if strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:") ||
			strings.Contains(origin, "asse.devtunnels.ms") {
			return true
		}
	}

	allowedSuffixes := []string{".miftech.web.id", ".miv.best", ".wafaindonesia.or.id"}
	for _, suffix := range allowedSuffixes {
		if strings.HasSuffix(origin, suffix) {
			return true
		}
	}

	return false
}
