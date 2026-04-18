package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type MutabaahHandler struct {
	service service.MutabaahService
}

func NewMutabaahHandler(service service.MutabaahService) *MutabaahHandler {
	return &MutabaahHandler{service: service}
}

// GetTodayStatus — status mutabaah hari ini untuk pegawai yang login
// GET /mutabaah/today
func (h *MutabaahHandler) GetTodayStatus(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	result, err := h.service.GetTodayStatus(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "today mutabaah status",
		Data:       result,
	})
}

// Submit — submit mutabaah hari ini
// POST /mutabaah/submit
func (h *MutabaahHandler) Submit(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var req dto.MutabaahSubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	if req.Pages < 0 {
		return respondBadRequest(c, "pages tidak boleh negatif")
	}

	result, err := h.service.Submit(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "mutabaah berhasil disubmit",
		Data:       result,
	})
}

// Cancel — batalkan submit mutabaah hari ini
// POST /mutabaah/cancel
func (h *MutabaahHandler) Cancel(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	result, err := h.service.Cancel(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah berhasil dibatalkan",
		Data:       result,
	})
}

// List — admin: daftar semua mutabaah
// GET /mutabaah
func (h *MutabaahHandler) List(c *fiber.Ctx) error {
	var params dto.MutabaahListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllLogs(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah list",
		Data:       result,
	})
}

// HRDCancel — admin: batalkan laporan mutabaah yang telah disubmit
// PUT /mutabaah/:id/cancel
func (h *MutabaahHandler) HRDCancel(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid mutabaah ID")
	}

	result, err := h.service.HRDCancel(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah reports canceled by HRD",
		Data:       result,
	})
}
