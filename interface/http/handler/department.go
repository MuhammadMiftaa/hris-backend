package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type DepartmentHandler struct {
	service service.DepartmentService
}

func NewDepartmentHandler(service service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

func (h *DepartmentHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department metadata",
		Data:       result,
	})
}

func (h *DepartmentHandler) List(c *fiber.Ctx) error {
	var params dto.DepartmentListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllDepartments(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department list",
		Data:       result,
	})
}

func (h *DepartmentHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetDepartmentByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department detail",
		Data:       result,
	})
}

func (h *DepartmentHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateDepartmentRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateDepartment(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Department created",
		Data:       result,
	})
}

func (h *DepartmentHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateDepartmentRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateDepartment(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department updated",
		Data:       result,
	})
}

func (h *DepartmentHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteDepartment(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department deleted",
	})
}
