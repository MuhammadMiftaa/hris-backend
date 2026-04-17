package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type LeaveTypeHandler struct {
	service service.LeaveTypeService
}

func NewLeaveTypeHandler(service service.LeaveTypeService) *LeaveTypeHandler {
	return &LeaveTypeHandler{service: service}
}

func (h *LeaveTypeHandler) Metadata(c *fiber.Ctx) error {
	result := h.service.GetMetadata(c.Context())

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type metadata",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) List(c *fiber.Ctx) error {
	result, err := h.service.GetAllLeaveTypes(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type list",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetLeaveTypeByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type detail",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateLeaveTypeRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateLeaveType(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Leave type created",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateLeaveTypeRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateLeaveType(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type updated",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteLeaveType(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type deleted",
	})
}
