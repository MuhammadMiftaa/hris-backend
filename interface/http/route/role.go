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

func RoleRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewRoleRepository(db)
	txManager := repository.NewTxManager(db)
	h := handler.NewRoleHandler(service.NewRoleService(repo, txManager))

	roles := app.Group("/roles")
	{
		roles.Get("/metadata", h.Metadata)
		roles.Get("/", middleware.RBACMiddleware(data.PERM_RoleRead), h.List)
		roles.Get("/export", middleware.RBACMiddleware(data.PERM_RoleRead), h.Export)
		roles.Get("/:id", middleware.RBACMiddleware(data.PERM_RoleRead), h.Detail)
		roles.Post("/", middleware.RBACMiddleware(data.PERM_RoleCreate), h.Create)
		roles.Put("/:id", middleware.RBACMiddleware(data.PERM_RoleUpdate), h.Update)
		roles.Delete("/:id", middleware.RBACMiddleware(data.PERM_RoleDelete), h.Delete)
		roles.Put("/:id/permissions", middleware.RBACMiddleware(data.PERM_RoleUpdate), h.UpdatePermissions)
	}

	// Permission list (standalone)
	app.Get("/permissions", middleware.RBACMiddleware(data.PERM_RoleRead), h.ListPermissions)
}
