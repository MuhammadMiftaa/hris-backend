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

func HolidayRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewHolidayRepository(db)
	h := handler.NewHolidayHandler(service.NewHolidayService(repo))

	holidays := app.Group("/holidays")
	{
		holidays.Get("/metadata", h.Metadata)
		holidays.Get("/", middleware.RBACMiddleware(data.PERM_HolidayRead), h.List)
		holidays.Get("/:id", middleware.RBACMiddleware(data.PERM_HolidayRead), h.Detail)
		holidays.Post("/", middleware.RBACMiddleware(data.PERM_HolidayCreate), h.Create)
		holidays.Put("/:id", middleware.RBACMiddleware(data.PERM_HolidayUpdate), h.Update)
		holidays.Delete("/:id", middleware.RBACMiddleware(data.PERM_HolidayDelete), h.Delete)
		holidays.Post("/sync", middleware.RBACMiddleware(data.PERM_HolidayCreate), h.SyncFromExternalAPI)
	}
}
