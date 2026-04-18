package route

import (
	"hris-backend/interface/http/handler"
	"hris-backend/interface/http/middleware"
	"hris-backend/internal/repository"
	"hris-backend/internal/service"
	"hris-backend/internal/utils/data"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func MutabaahRoutes(app *fiber.App, db *gorm.DB) {
	mutaRepo := repository.NewMutabaahRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)
	txManager := repository.NewTxManager(db)
	svc := service.NewMutabaahService(mutaRepo, attendRepo, txManager)
	h := handler.NewMutabaahHandler(svc)

	mutabaah := app.Group("/mutabaah")
	{
		// Pegawai: status hari ini
		mutabaah.Get("/today", middleware.RBACMiddleware(data.PERM_MutabaahRead), h.GetTodayStatus)

		// Pegawai: submit & cancel
		mutabaah.Post("/submit", middleware.RBACMiddleware(data.PERM_MutabaahCreate), h.Submit)
		mutabaah.Post("/cancel", middleware.RBACMiddleware(data.PERM_MutabaahUpdate), h.Cancel)

		// Admin: daftar semua mutabaah
		mutabaah.Get("/", middleware.RBACMiddleware(data.PERM_MutabaahRead), h.List)
		
		// Admin: HRD cancel mutabaah
		mutabaah.Put("/:id/cancel", middleware.RBACMiddleware(data.PERM_MutabaahUpdate), h.HRDCancel)
	}
}
