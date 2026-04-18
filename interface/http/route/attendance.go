package route

import (
	"hris-backend/config/storage"
	"hris-backend/interface/http/handler"
	"hris-backend/interface/http/middleware"
	"hris-backend/internal/repository"
	"hris-backend/internal/service"
	"hris-backend/internal/utils/data"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AttendanceRoutes(app *fiber.App, db *gorm.DB, minio storage.MinioClient) {
	repo := repository.NewAttendanceRepository(db)
	txManager := repository.NewTxManager(db)
	svc := service.NewAttendanceService(repo, txManager, minio)
	h := handler.NewAttendanceHandler(svc)

	attendance := app.Group("/attendance")
	{
		// Pegawai: status hari ini
		attendance.Get("/today", middleware.RBACMiddleware(data.PERM_AttendanceRead), h.GetTodayStatus)

		// Pegawai: presign upload foto
		attendance.Post("/presign", middleware.RBACMiddleware(data.PERM_AttendanceCreate), h.PresignClockPhoto)

		// Pegawai: signed download URL untuk foto
		attendance.Get("/photo", middleware.RBACMiddleware(data.PERM_AttendanceRead), h.GetPhotoURL)

		// Pegawai: clock in / clock out
		attendance.Post("/clock-in", middleware.RBACMiddleware(data.PERM_AttendanceCreate), h.ClockIn)
		attendance.Post("/clock-out", middleware.RBACMiddleware(data.PERM_AttendanceCreate), h.ClockOut)

		// Admin: daftar semua presensi
		attendance.Get("/", middleware.RBACMiddleware(data.PERM_AttendanceRead), h.List)

		// Metadata
		attendance.Get("/metadata", middleware.RBACMiddleware(data.PERM_AttendanceRead), h.Metadata)

		// Admin: manual clock in
		attendance.Post("/manual", middleware.RBACMiddleware(data.PERM_AttendanceCreate), h.CreateManual)
	}

	overrides := app.Group("/attendance-overrides")
	{
		overrides.Get("/", middleware.RBACMiddleware(data.PERM_AttendanceRead), h.ListOverrides)
		overrides.Get("/:id", middleware.RBACMiddleware(data.PERM_AttendanceRead), h.DetailOverride)
		overrides.Post("/", middleware.RBACMiddleware(data.PERM_AttendanceCreate), h.CreateOverride)
		overrides.Put("/:id", middleware.RBACMiddleware(data.PERM_AttendanceUpdate), h.UpdateOverride)
	}
}
