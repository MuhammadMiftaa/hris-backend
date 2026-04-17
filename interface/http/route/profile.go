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

func ProfileRoutes(app *fiber.App, db *gorm.DB, minio storage.MinioClient) {
	repo := repository.NewProfileRepository(db)
	svc := service.NewProfileService(repo, minio)
	h := handler.NewProfileHandler(svc)

	profile := app.Group("/profile")
	{
		// Simple profile (sidebar/header cache)
		profile.Get("/", middleware.RBACMiddleware(data.PERM_ProfileRead), h.GetProfile)
		profile.Put("/", middleware.RBACMiddleware(data.PERM_ProfileUpdate), h.UpdateProfile)

		// Photo management
		profile.Post("/photo", middleware.RBACMiddleware(data.PERM_ProfileUpdate), h.UploadPhoto)
		profile.Delete("/photo", middleware.RBACMiddleware(data.PERM_ProfileUpdate), h.DeletePhoto)

		// Employee profile detail (untuk ProfilePage)
		profile.Get("/employee", middleware.RBACMiddleware(data.PERM_ProfileRead), h.GetEmployeeProfile)
		profile.Get("/employee/contacts", middleware.RBACMiddleware(data.PERM_ProfileRead), h.GetEmployeeContacts)

		// Change password
		profile.Post("/change-password", middleware.RBACMiddleware(data.PERM_ProfileUpdate), h.ChangePassword)
	}
}
