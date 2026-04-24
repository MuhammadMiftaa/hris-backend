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

func DashboardRoutes(app *fiber.App, db *gorm.DB) {
	dashRepo := repository.NewDashboardRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)
	mutabaahRepo := repository.NewMutabaahRepository(db)
	svc := service.NewDashboardService(dashRepo, attendRepo, mutabaahRepo)
	h := handler.NewDashboardHandler(svc)

	dashboard := app.Group("/dashboard")
	{
		// /dashboard/employee
		dashboard.Get("/employee", middleware.RBACMiddleware(data.PERM_HomeEmployeeRead), h.GetEmployeeDashboard)

		// /dashboard/hrd
		dashboard.Get("/hrd", middleware.RBACMiddleware(data.PERM_HomeAdminRead), h.GetHRDDashboard)

		// /dashboard/rankings
		dashboard.Get("/rankings", middleware.RBACMiddleware(data.PERM_HomeEmployeeRead), h.GetRankings)

		// /dashboard/metadata
		dashboard.Get("/metadata", middleware.RBACMiddleware(data.PERM_HomeEmployeeRead), h.GetMetadata)
	}
}
