package route

import (
	"hris-backend/config/storage"
	"hris-backend/interface/http/handler"
	"hris-backend/interface/http/middleware"
	"hris-backend/internal/service"
	"hris-backend/internal/utils/data"

	"github.com/gofiber/fiber/v2"
)

func DocumentRoutes(app *fiber.App, minio storage.MinioClient) {
	svc := service.NewDocumentService(minio)
	h := handler.NewDocumentHandler(svc)

	docs := app.Group("/documents")
	{
		docs.Post("/upload", middleware.RBACMiddleware(data.PERM_LeaveCreate), h.UploadDocument)
		docs.Get("/download", middleware.RBACMiddleware(data.PERM_LeaveRead), h.GetDownloadURL)
	}
}
