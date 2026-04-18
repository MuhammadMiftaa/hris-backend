package middleware

import (
	"context"
	"fmt"
	"time"

	"hris-backend/internal/redis"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type RateLimitConfig struct {
	Max    int
	Window time.Duration
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:    300,
		Window: 1 * time.Minute,
	}
}

func RateLimiterMiddleware(redisClient redis.Redis, cfg ...RateLimitConfig) fiber.Handler {
	config := DefaultRateLimitConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		ctx := context.Background()

		count, err := redisClient.Incr(ctx, key)
		if err != nil {
			return c.Next()
		}

		if count == 1 {
			redisClient.Expire(ctx, key, config.Window)
		}

		ttl, _ := redisClient.TTL(ctx, key)

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(config.Max)-count)))

		if count > int64(config.Max) {
			retryAfter := int(ttl.Seconds())
			if retryAfter <= 0 {
				retryAfter = int(config.Window.Seconds())
			}
			c.Set("Retry-After", fmt.Sprintf("%d", retryAfter))

			return c.Status(fiber.StatusTooManyRequests).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 429,
				Message:    "Too many requests. Please try again later.",
			})
		}

		return c.Next()
	}
}
