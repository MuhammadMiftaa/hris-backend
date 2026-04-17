package handler

import (
	"strconv"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type PositionHandler struct {
	service service.PositionService
}

func NewPositionHandler(service service.PositionService) *PositionHandler {
	return &PositionHandler{service: service}
}

func (h *PositionHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Position metadata",
		Data:       result,
	})
}

func (h *PositionHandler) List(c *fiber.Ctx) error {
	var departmentID *uint
	if deptStr := c.Query("department_id"); deptStr != "" {
		id, err := strconv.ParseUint(deptStr, 10, 32)
		if err == nil {
			v := uint(id)
			departmentID = &v
		}
	}

	result, err := h.service.GetAllPositions(c.Context(), departmentID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Position list",
		Data:       result,
	})
}

func (h *PositionHandler) Create(c *fiber.Ctx) error {
	var input dto.CreatePositionRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreatePosition(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Position created",
		Data:       result,
	})
}

func (h *PositionHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdatePositionRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdatePosition(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Position updated",
		Data:       result,
	})
}

func (h *PositionHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeletePosition(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Position deleted",
	})
}
