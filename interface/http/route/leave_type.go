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

func LeaveTypeRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewLeaveTypeRepository(db)
	h := handler.NewLeaveTypeHandler(service.NewLeaveTypeService(repo))

	leaveTypes := app.Group("/leave-types")
	{
		leaveTypes.Get("/metadata", h.Metadata)
		leaveTypes.Get("/", middleware.RBACMiddleware(data.PERM_LeaveTypeRead), h.List)
		leaveTypes.Get("/export", middleware.RBACMiddleware(data.PERM_LeaveTypeExport), h.Export)
		leaveTypes.Get("/:id", middleware.RBACMiddleware(data.PERM_LeaveTypeRead), h.Detail)
		leaveTypes.Post("/", middleware.RBACMiddleware(data.PERM_LeaveTypeCreate), h.Create)
		leaveTypes.Put("/:id", middleware.RBACMiddleware(data.PERM_LeaveTypeUpdate), h.Update)
		leaveTypes.Delete("/:id", middleware.RBACMiddleware(data.PERM_LeaveTypeDelete), h.Delete)
	}
}
