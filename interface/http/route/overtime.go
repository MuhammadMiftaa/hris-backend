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
		ots.Get("/metadata", h.Metadata)
		ots.Get("/", middleware.RBACMiddleware(data.PERM_OvertimeRead), h.List)
		ots.Get("/export", middleware.RBACMiddleware(data.PERM_OvertimeExport), h.Export)
		ots.Get("/:id", middleware.RBACMiddleware(data.PERM_OvertimeRead), h.Detail)
		ots.Post("/", middleware.RBACMiddleware(data.PERM_OvertimeCreate), h.Create)
		ots.Put("/:id/approve", middleware.RBACMiddleware(data.PERM_OvertimeApprove), h.Approve)
		ots.Put("/:id/reject", middleware.RBACMiddleware(data.PERM_OvertimeApprove), h.Reject)
		// ots.Delete("/:id", middleware.RBACMiddleware(data.PERM_OvertimeDelete), h.Delete)
	}
}
