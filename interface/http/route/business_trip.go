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

func BusinessTripRoutes(app *fiber.App, db *gorm.DB, minio storage.MinioClient, notifSvc service.NotificationService) {
	repo := repository.NewBusinessTripRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)
	txManager := repository.NewTxManager(db)
	svc := service.NewBusinessTripService(repo, attendRepo, txManager, minio, notifSvc)
	h := handler.NewBusinessTripHandler(svc)

	trips := app.Group("/business-trips")
	{
		trips.Get("/", middleware.RBACMiddleware(data.PERM_BusinessTripRead), h.List)
		trips.Get("/export", middleware.RBACMiddleware(data.PERM_BusinessTripExport), h.Export)
		trips.Get("/:id", middleware.RBACMiddleware(data.PERM_BusinessTripRead), h.Detail)
		trips.Post("/", middleware.RBACMiddleware(data.PERM_BusinessTripCreate), h.Create)
		trips.Put("/:id", middleware.RBACMiddleware(data.PERM_BusinessTripApprove), h.UpdateStatus)
		// trips.Delete("/:id", middleware.RBACMiddleware(data.PERM_BusinessTripDelete), h.Delete)
	}
}
