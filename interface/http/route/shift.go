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

func ShiftRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewShiftRepository(db)
	txManager := repository.NewTxManager(db)
	h := handler.NewShiftHandler(service.NewShiftService(repo, txManager))

	shifts := app.Group("/shifts")
	{
		shifts.Get("/metadata", h.Metadata)
		shifts.Get("/", middleware.RBACMiddleware(data.PERM_ShiftRead), h.ListTemplates)
		shifts.Get("/:id", middleware.RBACMiddleware(data.PERM_ShiftRead), h.DetailTemplate)
		shifts.Post("/", middleware.RBACMiddleware(data.PERM_ShiftCreate), h.CreateTemplate)
		shifts.Put("/:id", middleware.RBACMiddleware(data.PERM_ShiftUpdate), h.UpdateTemplate)
		shifts.Delete("/:id", middleware.RBACMiddleware(data.PERM_ShiftDelete), h.DeleteTemplate)
		shifts.Get("/:id/details", middleware.RBACMiddleware(data.PERM_ShiftRead), h.ListDetails)
	}

	schedules := app.Group("/schedules")
	{
		schedules.Get("/", middleware.RBACMiddleware(data.PERM_ShiftRead), h.ListSchedules)
		schedules.Get("/:id", middleware.RBACMiddleware(data.PERM_ShiftRead), h.DetailSchedule)
		schedules.Post("/", middleware.RBACMiddleware(data.PERM_ShiftCreate), h.CreateSchedule)
		schedules.Put("/:id", middleware.RBACMiddleware(data.PERM_ShiftUpdate), h.UpdateSchedule)
		schedules.Delete("/:id", middleware.RBACMiddleware(data.PERM_ShiftDelete), h.DeleteSchedule)
	}
}
