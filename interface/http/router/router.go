package router

import (
	"hris-backend/config/db"
	"hris-backend/interface/http/middleware"
	"hris-backend/interface/http/route"
	"hris-backend/internal/redis"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupHTTPServer(dbInstance db.DatabaseClient, redisInstance redis.Redis) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "WAFA HRIS",
		ServerHeader: "WAFA HRIS",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: code,
				Message:    err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.LoggerMiddleware())
	// app.Use(middleware.RateLimiterMiddleware(redisClient))

	// Health check
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(dto.APIResponse{
			Status:     true,
			StatusCode: 200,
			Message:    "BFF is healthy",
		})
	})

	// Auth routes
	route.AuthRoutes(app, dbInstance.GetDB(), redisInstance)

	return app
}
