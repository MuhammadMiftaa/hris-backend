package middleware

import (
	"fmt"
	"time"

	logger "hris-backend/config/log"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils/data"

	"github.com/gofiber/fiber/v2"
)

func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		latency := time.Since(start)
		statusCode := c.Response().StatusCode()

		requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)
		account, _ := c.Locals("account").(dto.GetEmployeeByIDResponse)

		fields := map[string]any{
			"request_id":    requestID,
			"method":        c.Method(),
			"uri":           c.OriginalURL(),
			"status":        statusCode,
			"latency":       fmt.Sprintf("%.3fms", float64(latency.Nanoseconds())/1e6),
			"client_ip":     c.IP(),
			"response_size": len(c.Response().Body()),
			"account_id":    account.AccountID,
			"full_name":     account.FullName,
			"email":         account.Email,
		}

		switch {
		case statusCode >= 500:
			logger.Error("http_request", fields)
		case statusCode >= 400:
			logger.Warn("http_request", fields)
		default:
			logger.Info("http_request", fields)
		}

		return err
	}
}
