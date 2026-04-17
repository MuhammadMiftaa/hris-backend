package handler

import (
	"strconv"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type ShiftHandler struct {
	service service.ShiftService
}

func NewShiftHandler(service service.ShiftService) *ShiftHandler {
	return &ShiftHandler{service: service}
}

func (h *ShiftHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift metadata", Data: result})
}

func (h *ShiftHandler) ListTemplates(c *fiber.Ctx) error {
	result, err := h.service.GetAllTemplates(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift template list", Data: result})
}

func (h *ShiftHandler) DetailTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	result, err := h.service.GetTemplateByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift template detail", Data: result})
}

func (h *ShiftHandler) CreateTemplate(c *fiber.Ctx) error {
	var input dto.CreateShiftRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.CreateTemplate(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status: true, StatusCode: 201, Message: "Shift template created", Data: result,
	})
}

func (h *ShiftHandler) UpdateTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	var input dto.UpdateShiftRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.UpdateTemplate(c.Context(), uint(id), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift template updated", Data: result})
}

func (h *ShiftHandler) DeleteTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	if err := h.service.DeleteTemplate(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift template deleted"})
}

func (h *ShiftHandler) ListDetails(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	result, err := h.service.GetDetailsByTemplateID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift details list", Data: result})
}

func (h *ShiftHandler) ListSchedules(c *fiber.Ctx) error {
	var params dto.ScheduleListParams

	if v := c.QueryInt("employee_id", 0); v > 0 {
		uid := uint(v)
		params.EmployeeID = &uid
	}
	if v := c.QueryInt("shift_template_id", 0); v > 0 {
		uid := uint(v)
		params.ShiftTemplateID = &uid
	}
	if v := c.Query("is_active"); v != "" {
		b := v == "true"
		params.IsActive = &b
	}

	result, err := h.service.GetAllSchedules(c.Context(), &params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Schedule list", Data: result})
}

func (h *ShiftHandler) DetailSchedule(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	result, err := h.service.GetScheduleByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Schedule detail", Data: result})
}

func (h *ShiftHandler) CreateSchedule(c *fiber.Ctx) error {
	var input dto.CreateScheduleRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.CreateSchedule(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status: true, StatusCode: 201, Message: "Schedule created", Data: result,
	})
}

func (h *ShiftHandler) UpdateSchedule(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	var input dto.UpdateScheduleRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.UpdateSchedule(c.Context(), uint(id), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Schedule updated", Data: result})
}

func (h *ShiftHandler) DeleteSchedule(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	if err := h.service.DeleteSchedule(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Schedule deleted"})
}
