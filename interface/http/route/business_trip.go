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

func BusinessTripRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewBusinessTripRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)
	txManager := repository.NewTxManager(db)
	svc := service.NewBusinessTripService(repo, attendRepo, txManager)
	h := handler.NewBusinessTripHandler(svc)

	trips := app.Group("/business-trips")
	{
		trips.Get("/", middleware.RBACMiddleware(data.PERM_RequestRead), h.List)
		trips.Get("/:id", middleware.RBACMiddleware(data.PERM_RequestRead), h.Detail)
		trips.Post("/", middleware.RBACMiddleware(data.PERM_RequestCreate), h.Create)
		trips.Put("/:id", middleware.RBACMiddleware(data.PERM_RequestUpdate), h.UpdateStatus)
		trips.Delete("/:id", middleware.RBACMiddleware(data.PERM_RequestDelete), h.Delete)
	}
}
