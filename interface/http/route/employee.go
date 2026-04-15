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

func EmployeeRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewEmployeeRepository(db)
	txManager := repository.NewTxManager(db)
	h := handler.NewEmployeeHandler(service.NewEmployeeService(repo, txManager))

	employee := app.Group("/employee")
	{
		employee.Get("/metadata", h.Metadata)
		employee.Get("/", h.List, middleware.RBACMiddleware(data.PERM_EmployeeRead))
		employee.Get("/:id", h.Detail, middleware.RBACMiddleware(data.PERM_EmployeeRead))
		employee.Post("/", h.Create, middleware.RBACMiddleware(data.PERM_EmployeeCreate))
		employee.Put("/:id", h.Update, middleware.RBACMiddleware(data.PERM_EmployeeUpdate))
		employee.Delete("/:id", h.Delete, middleware.RBACMiddleware(data.PERM_EmployeeDelete))
	}
}
