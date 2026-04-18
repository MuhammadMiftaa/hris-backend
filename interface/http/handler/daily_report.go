package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type DailyReportHandler struct {
	service service.DailyReportService
}

func NewDailyReportHandler(service service.DailyReportService) *DailyReportHandler {
	return &DailyReportHandler{service: service}
}

// List — GET /daily-reports
func (h *DailyReportHandler) List(c *fiber.Ctx) error {
	var params dto.DailyReportListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	res, err := h.service.GetAll(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "daily reports list",
		Data:       res,
	})
}

// Detail — GET /daily-reports/:id
func (h *DailyReportHandler) Detail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid daily report ID")
	}

	res, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "daily report detail",
		Data:       res,
	})
}

// Create — POST /daily-reports
func (h *DailyReportHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateDailyReportRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.Create(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "daily report created",
		Data:       res,
	})
}

// Delete — DELETE /daily-reports/:id
func (h *DailyReportHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid daily report ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "daily report deleted",
	})
}
