package handler

import (
	"strconv"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type HolidayHandler struct {
	service service.HolidayService
}

func NewHolidayHandler(service service.HolidayService) *HolidayHandler {
	return &HolidayHandler{service: service}
}

func (h *HolidayHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status: false, StatusCode: 500, Message: err.Error(),
		})
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday metadata", Data: result})
}

func (h *HolidayHandler) List(c *fiber.Ctx) error {
	var params dto.HolidayListParams

	if v := c.QueryInt("year", 0); v > 0 {
		params.Year = &v
	}
	if v := c.Query("type"); v != "" {
		params.Type = &v
	}
	if v := c.QueryInt("branch_id", 0); v > 0 {
		uid := uint(v)
		params.BranchID = &uid
	}

	result, err := h.service.GetAllHolidays(c.Context(), &params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status: false, StatusCode: 500, Message: err.Error(),
		})
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday list", Data: result})
}

func (h *HolidayHandler) Detail(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status: false, StatusCode: 400, Message: "Invalid ID",
		})
	}
	result, err := h.service.GetHolidayByID(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status: false, StatusCode: 500, Message: err.Error(),
		})
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday detail", Data: result})
}

func (h *HolidayHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateHolidayRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status: false, StatusCode: 400, Message: err.Error(),
		})
	}
	result, err := h.service.CreateHoliday(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status: false, StatusCode: 500, Message: err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status: true, StatusCode: 201, Message: "Holiday created", Data: result,
	})
}

func (h *HolidayHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status: false, StatusCode: 400, Message: "Invalid ID",
		})
	}
	var input dto.UpdateHolidayRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status: false, StatusCode: 400, Message: err.Error(),
		})
	}
	result, err := h.service.UpdateHoliday(c.Context(), uint(id), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status: false, StatusCode: 500, Message: err.Error(),
		})
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday updated", Data: result})
}

func (h *HolidayHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status: false, StatusCode: 400, Message: "Invalid ID",
		})
	}
	if err := h.service.DeleteHoliday(c.Context(), uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status: false, StatusCode: 500, Message: err.Error(),
		})
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday deleted"})
}
