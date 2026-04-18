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

func DailyReportRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewDailyReportRepository(db)
	svc := service.NewDailyReportService(repo)
	h := handler.NewDailyReportHandler(svc)

	reports := app.Group("/daily-reports")
	{
		reports.Get("/", middleware.RBACMiddleware(data.PERM_DailyReportRead), h.List)
		reports.Get("/:id", middleware.RBACMiddleware(data.PERM_DailyReportRead), h.Detail)
		reports.Post("/", middleware.RBACMiddleware(data.PERM_DailyReportCreate), h.Create)
		reports.Delete("/:id", middleware.RBACMiddleware(data.PERM_DailyReportDelete), h.Delete)
	}
}
