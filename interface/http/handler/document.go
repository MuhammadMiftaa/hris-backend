package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type DocumentHandler struct {
	service service.DocumentService
}

func NewDocumentHandler(service service.DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

// UploadDocument — POST /documents/upload
func (h *DocumentHandler) UploadDocument(c *fiber.Ctx) error {
	var req dto.UploadDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request payload")
	}

	account := getAccountFromCtx(c)

	res, err := h.service.UploadDocument(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    res.Message,
		Data:       res,
	})
}

// GetDownloadURL — GET /documents/download
func (h *DocumentHandler) GetDownloadURL(c *fiber.Ctx) error {
	var req dto.DocumentDownloadRequest
	if err := c.QueryParser(&req); err != nil {
		return respondBadRequest(c, "invalid query parameter")
	}

	res, err := h.service.GetDownloadURL(c.Context(), req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "url generated",
		Data:       res,
	})
}
