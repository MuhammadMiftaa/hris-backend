package route

import (
	"hris-backend/config/storage"
	"hris-backend/interface/http/handler"
	"hris-backend/interface/http/middleware"
	"hris-backend/internal/repository"
	"hris-backend/internal/service"
	"hris-backend/internal/utils/data"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func EmployeeRoutes(app *fiber.App, db *gorm.DB, minioClient storage.MinioClient) {
	repo := repository.NewEmployeeRepository(db)
	txManager := repository.NewTxManager(db)
	h := handler.NewEmployeeHandler(service.NewEmployeeService(repo, txManager, minioClient))

	employees := app.Group("/employees")
	{
		employees.Get("/metadata", h.Metadata)
		employees.Get("/", middleware.RBACMiddleware(data.PERM_EmployeeRead), h.List)
		employees.Get("/:id", middleware.RBACMiddleware(data.PERM_EmployeeRead), h.Detail)
		employees.Post("/", middleware.RBACMiddleware(data.PERM_EmployeeCreate), h.Create)
		employees.Put("/:id", middleware.RBACMiddleware(data.PERM_EmployeeUpdate), h.Update)
		employees.Delete("/:id", middleware.RBACMiddleware(data.PERM_EmployeeDelete), h.Delete)
		employees.Patch("/:id/reset-password", middleware.RBACMiddleware(data.PERM_EmployeeUpdate), h.ResetPassword)

		// Contacts
		employees.Get("/:id/contacts", middleware.RBACMiddleware(data.PERM_EmployeeRead), h.ListContacts)
		employees.Post("/:id/contacts", middleware.RBACMiddleware(data.PERM_EmployeeUpdate), h.CreateContact)

		// Contracts
		employees.Get("/:id/contracts", middleware.RBACMiddleware(data.PERM_EmployeeRead), h.ListContracts)
		employees.Post("/:id/contracts", middleware.RBACMiddleware(data.PERM_EmployeeUpdate), h.CreateContract)
	}

	app.Put("/employee-contacts/:id", middleware.RBACMiddleware(data.PERM_EmployeeUpdate), h.UpdateContact)
	app.Delete("/employee-contacts/:id", middleware.RBACMiddleware(data.PERM_EmployeeUpdate), h.DeleteContact)

	app.Put("/contracts/:id", middleware.RBACMiddleware(data.PERM_EmployeeUpdate), h.UpdateContract)
	app.Delete("/contracts/:id", middleware.RBACMiddleware(data.PERM_EmployeeUpdate), h.DeleteContract)
}
