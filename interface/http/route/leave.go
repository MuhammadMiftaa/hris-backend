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
	leaveTypeRepo := repository.NewLeaveTypeRepository(db)
	attendRepo := repository.NewAttendanceRepository(db)
	txManager := repository.NewTxManager(db)
	svc := service.NewLeaveService(leaveRepo, leaveTypeRepo, attendRepo, txManager, minio)
	h := handler.NewLeaveHandler(svc)

	app.Get("/leave-requests/metadata", h.Metadata)

	balances := app.Group("/leave-balances")
	{
		balances.Get("/", middleware.RBACMiddleware(data.PERM_LeaveBalanceRead), h.ListBalances)
		balances.Get("/export", middleware.RBACMiddleware(data.PERM_LeaveBalanceExport), h.ExportBalances)

		// summary per employee
		balances.Get("/summary", middleware.RBACMiddleware(data.PERM_LeaveBalanceRead), h.ListEmployeeBalanceSummary)
		balances.Get("/summary/export", middleware.RBACMiddleware(data.PERM_LeaveBalanceExport), h.ExportEmployeeBalanceSummary)
		balances.Get("/employee/:employeeId", middleware.RBACMiddleware(data.PERM_LeaveBalanceRead), h.GetEmployeeBalanceDetail)

		// CRUD balance & adjustment
		balances.Post("/", middleware.RBACMiddleware(data.PERM_LeaveBalanceCreate), h.UpsertBalance)
		balances.Delete("/:id", middleware.RBACMiddleware(data.PERM_LeaveBalanceDelete), h.DeleteBalance)
		balances.Post("/:id/adjust", middleware.RBACMiddleware(data.PERM_LeaveBalanceUpdate), h.AdjustBalance)
		balances.Get("/:id/adjustments", middleware.RBACMiddleware(data.PERM_LeaveBalanceRead), h.GetBalanceAdjustments)
	}

	requests := app.Group("/leave-requests")
	{
		requests.Get("/", middleware.RBACMiddleware(data.PERM_LeaveRead), h.ListRequests)
		requests.Get("/export", middleware.RBACMiddleware(data.PERM_LeaveExport), h.ExportRequests)
		requests.Get("/:id", middleware.RBACMiddleware(data.PERM_LeaveRead), h.DetailRequest)
		requests.Post("/", middleware.RBACMiddleware(data.PERM_LeaveCreate), h.Create)
		requests.Put("/:id/approve", middleware.RBACMiddleware(data.PERM_LeaveUpdate), h.Approve)
		requests.Put("/:id/reject", middleware.RBACMiddleware(data.PERM_LeaveUpdate), h.Reject)
	}
}
