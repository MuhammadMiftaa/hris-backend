package route

import (
	"hris-backend/interface/http/handler"
	"hris-backend/interface/http/middleware"
	"hris-backend/internal/redis"
	"hris-backend/internal/repository"
	"hris-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AuthRoutes(app *fiber.App, db *gorm.DB, rdb redis.Redis) {
	h := handler.NewAuthHandler(service.NewAuthService(repository.NewAuthRepository(db), rdb))

	auth := app.Group("/auth")
	{
		auth.Post("/login", h.Login)
		auth.Post("/refresh", h.Refresh)
		auth.Post("/logout", middleware.AuthMiddleware(rdb), h.Logout)
	}
}
