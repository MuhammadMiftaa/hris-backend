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

func PermissionRequestRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewPermissionRequestRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)
	txManager := repository.NewTxManager(db)
	svc := service.NewPermissionRequestService(repo, attendRepo, txManager)
	h := handler.NewPermissionRequestHandler(svc)

	perms := app.Group("/permission-requests")
	{
		perms.Get("/metadata", h.Metadata)
		perms.Get("/", middleware.RBACMiddleware(data.PERM_RequestRead), h.List)
		perms.Get("/:id", middleware.RBACMiddleware(data.PERM_RequestRead), h.Detail)
		perms.Post("/", middleware.RBACMiddleware(data.PERM_RequestCreate), h.Create)
		perms.Put("/:id", middleware.RBACMiddleware(data.PERM_RequestUpdate), h.UpdateStatus)
		perms.Delete("/:id", middleware.RBACMiddleware(data.PERM_RequestDelete), h.Delete)
	}
}
