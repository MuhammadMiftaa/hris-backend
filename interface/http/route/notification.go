package route

import (
	"hris-backend/config/env"
	"hris-backend/interface/http/handler"
	"hris-backend/internal/repository"
	"hris-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func NotificationRoutes(app *fiber.App, db *gorm.DB) {
	pushRepo := repository.NewPushRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	empRepo := repository.NewEmployeeRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)

	pushSvc := service.NewPushService(
		env.Cfg.Vapid.PrivateKey,
		env.Cfg.Vapid.PublicKey,
		"mailto:wafa@example.com",
	)
	notifSvc := service.NewNotificationService(pushRepo, notifRepo, pushSvc, empRepo, attendRepo)
	h := handler.NewNotificationHandler(notifSvc)

	notifications := app.Group("/notifications")
	{
		notifications.Get("/", h.List)
		notifications.Get("/unread-count", h.UnreadCount)
		notifications.Post("/:id/read", h.MarkAsRead)
		notifications.Post("/read-all", h.MarkAllAsRead)
	}
}
