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

func BranchRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewBranchRepository(db)
	h := handler.NewBranchHandler(service.NewBranchService(repo))

	branches := app.Group("/branches")
	{
		branches.Get("/", middleware.RBACMiddleware(data.PERM_BranchRead), h.List)
		branches.Get("/export", middleware.RBACMiddleware(data.PERM_BranchRead), h.Export)
		branches.Get("/:id", middleware.RBACMiddleware(data.PERM_BranchRead), h.Detail)
		branches.Post("/", middleware.RBACMiddleware(data.PERM_BranchCreate), h.Create)
		branches.Put("/:id", middleware.RBACMiddleware(data.PERM_BranchUpdate), h.Update)
		branches.Delete("/:id", middleware.RBACMiddleware(data.PERM_BranchDelete), h.Delete)
	}
}
