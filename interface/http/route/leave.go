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

func LeaveRoutes(app *fiber.App, db *gorm.DB, minio storage.MinioClient) {
	leaveRepo := repository.NewLeaveRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)
	txManager := repository.NewTxManager(db)
	svc := service.NewLeaveService(leaveRepo, attendRepo, txManager, minio)
	h := handler.NewLeaveHandler(svc)

	app.Get("/leave-requests/metadata", h.Metadata)

	balances := app.Group("/leave-balances")
	{
		balances.Get("/", middleware.RBACMiddleware(data.PERM_LeaveRead), h.ListBalances)
	}

	requests := app.Group("/leave-requests")
	{
		requests.Get("/", middleware.RBACMiddleware(data.PERM_LeaveRead), h.ListRequests)
		requests.Get("/:id", middleware.RBACMiddleware(data.PERM_LeaveRead), h.DetailRequest)
		requests.Post("/", middleware.RBACMiddleware(data.PERM_LeaveCreate), h.Create)
		requests.Put("/:id/approve", middleware.RBACMiddleware(data.PERM_LeaveUpdate), h.Approve)
		requests.Put("/:id/reject", middleware.RBACMiddleware(data.PERM_LeaveUpdate), h.Reject)
	}
}
