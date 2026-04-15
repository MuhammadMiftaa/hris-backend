package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type EmployeeHandler struct {
	service service.EmployeeService
}

func NewEmployeeHandler(service service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{
		service: service,
	}
}

func (h *EmployeeHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee metadata",
		Data:       result,
	})
}

func (h *EmployeeHandler) List(c *fiber.Ctx) error {
	result, err := h.service.GetAllEmployees(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee list",
		Data:       result,
	})
}

func (h *EmployeeHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetEmployeeByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee detail",
		Data:       result,
	})
}

func (h *EmployeeHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateEmployeeRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	result, err := h.service.CreateEmployee(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Employee created",
		Data:       result,
	})
}

func (h *EmployeeHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateEmployeeRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	id := c.Params("id")
	result, err := h.service.UpdateEmployee(c.Context(), id, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee updated",
		Data:       result,
	})
}

func (h *EmployeeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteEmployee(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: fiber.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee deleted",
	})
}
