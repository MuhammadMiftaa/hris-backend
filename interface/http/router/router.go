package router

import (
	"hris-backend/config/db"
	"hris-backend/config/storage"
	"hris-backend/interface/http/middleware"
	"hris-backend/interface/http/route"
	"hris-backend/internal/redis"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupHTTPServer(dbInstance db.DatabaseClient, redisInstance redis.Redis, minioClient storage.MinioClient) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "WAFA HRIS",
		ServerHeader: "WAFA HRIS",
		// Maksimum body size 10MB (untuk presigned URL request body kecil, tapi jika ada upload langsung ke server)
		BodyLimit: 10 * 1024 * 1024,
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

	app.Use(recover.New())
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.RateLimiterMiddleware(redisInstance))

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(dto.APIResponse{
			Status:     true,
			StatusCode: 200,
			Message:    "BFF is healthy",
		})
	})

	// ── Auth (tidak butuh AuthMiddleware) ─────────────────────────
	route.AuthRoutes(app, dbInstance.GetDB(), redisInstance)

	// ── Internal / Cron (juga tanpa auth — amankan via network/firewall) ──
	route.InternalRoutes(app, dbInstance.GetDB())

	// ── Protected Routes ──────────────────────────────────────────
	app.Use(middleware.AuthMiddleware(redisInstance))

	route.EmployeeRoutes(app, dbInstance.GetDB(), minioClient)
	route.BranchRoutes(app, dbInstance.GetDB())
	route.DepartmentRoutes(app, dbInstance.GetDB())
	route.PositionRoutes(app, dbInstance.GetDB())
	route.RoleRoutes(app, dbInstance.GetDB())
	route.LeaveTypeRoutes(app, dbInstance.GetDB())
	route.ShiftRoutes(app, dbInstance.GetDB())
	route.HolidayRoutes(app, dbInstance.GetDB())
	route.AttendanceRoutes(app, dbInstance.GetDB(), minioClient)
	route.MutabaahRoutes(app, dbInstance.GetDB())
	route.LeaveRoutes(app, dbInstance.GetDB())
	route.PermissionRequestRoutes(app, dbInstance.GetDB())
	route.BusinessTripRoutes(app, dbInstance.GetDB())
	route.OvertimeRoutes(app, dbInstance.GetDB())
	route.DailyReportRoutes(app, dbInstance.GetDB())
	route.DashboardRoutes(app, dbInstance.GetDB())
	route.ProfileRoutes(app, dbInstance.GetDB(), minioClient)

	return app
}
