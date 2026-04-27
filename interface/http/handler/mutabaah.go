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

// GetTodayStatus — GET /mutabaah/today
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

// Submit — POST /mutabaah/submit
func (h *MutabaahHandler) Submit(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var req dto.MutabaahSubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	result, err := h.service.Submit(c.Context(), account.EmployeeID, account.IsTrainer, req)
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

// Cancel — POST /mutabaah/cancel
func (h *MutabaahHandler) Cancel(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var req dto.MutabaahCancelRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	result, err := h.service.Cancel(c.Context(), account.EmployeeID, req)
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

// List — GET /mutabaah
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

// HRDCancel — PUT /mutabaah/:id/cancel
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

func (h *MutabaahHandler) GetDailyReport(c *fiber.Ctx) error {
	date := c.Query("date")
	if date == "" {
		return respondBadRequest(c, "query date wajib diisi (format YYYY-MM-DD)")
	}

	result, err := h.service.GetDailyReport(c.Context(), date)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah daily report",
		Data:       result,
	})
}

func (h *MutabaahHandler) GetMonthlyReport(c *fiber.Ctx) error {
	month := c.QueryInt("month")
	year := c.QueryInt("year")
	if month <= 0 || month > 12 || year <= 0 {
		return respondBadRequest(c, "query month dan year wajib diisi dan valid")
	}

	result, err := h.service.GetMonthlyReport(c.Context(), month, year)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah monthly report",
		Data:       result,
	})
}

// GetCategoryReport — GET /mutabaah/report/category?date=2024-01-01
func (h *MutabaahHandler) GetCategoryReport(c *fiber.Ctx) error {
	date := c.Query("date")
	if date == "" {
		return respondBadRequest(c, "query date wajib diisi (format YYYY-MM-DD)")
	}

	result, err := h.service.GetCategoryReport(c.Context(), date)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah category report",
		Data:       result,
	})
}
