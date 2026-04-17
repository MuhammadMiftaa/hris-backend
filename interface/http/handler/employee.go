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
		return respondError(c, err)
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
		return respondError(c, err)
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
		return respondError(c, err)
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
		return respondBadRequest(c, err.Error())
	}

	newEmployee, newCredentials, err := h.service.CreateEmployee(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Employee created",
		Data: map[string]any{
			"employee":    newEmployee,
			"credentials": newCredentials,
		},
	})
}

func (h *EmployeeHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateEmployeeRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateEmployee(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
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
	if err := h.service.DeleteEmployee(c.Context(), id); err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Employee successfully deleted"})
}

func (h *EmployeeHandler) ResetPassword(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return respondBadRequest(c, "Employee ID is required")
	}

	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "Invalid Request Payload")
	}

	if err := h.service.ResetPassword(c.Context(), id, req); err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Password successfully reset"})
}

// Contacts
func (h *EmployeeHandler) ListContacts(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetContactsByEmployeeID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contact list", Data: result})
}

func (h *EmployeeHandler) CreateContact(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.CreateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateContact(c.Context(), id, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{Status: true, StatusCode: 201, Message: "Contact created", Data: result})
}

func (h *EmployeeHandler) UpdateContact(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.UpdateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.UpdateContact(c.Context(), id, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contact updated", Data: result})
}

func (h *EmployeeHandler) DeleteContact(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteContact(c.Context(), id); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contact deleted"})
}

// Contracts
func (h *EmployeeHandler) ListContracts(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetContractsByEmployeeID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contract list", Data: result})
}

func (h *EmployeeHandler) CreateContract(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.CreateContractRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateContract(c.Context(), id, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{Status: true, StatusCode: 201, Message: "Contract created", Data: result})
}

func (h *EmployeeHandler) UpdateContract(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.UpdateContractRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.UpdateContract(c.Context(), id, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contract updated", Data: result})
}

func (h *EmployeeHandler) DeleteContract(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteContract(c.Context(), id); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contract deleted"})
}
