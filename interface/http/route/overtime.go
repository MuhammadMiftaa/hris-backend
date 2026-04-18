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

func OvertimeRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewOvertimeRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)
	txManager := repository.NewTxManager(db)
	svc := service.NewOvertimeService(repo, attendRepo, txManager)
	h := handler.NewOvertimeHandler(svc)

	ots := app.Group("/overtime-requests")
	{
		ots.Get("/", middleware.RBACMiddleware(data.PERM_RequestRead), h.List)
		ots.Get("/:id", middleware.RBACMiddleware(data.PERM_RequestRead), h.Detail)
		ots.Post("/", middleware.RBACMiddleware(data.PERM_RequestCreate), h.Create)
		ots.Put("/:id", middleware.RBACMiddleware(data.PERM_RequestUpdate), h.UpdateStatus)
		ots.Delete("/:id", middleware.RBACMiddleware(data.PERM_RequestDelete), h.Delete)
	}
}
