package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type PermissionRequestHandler struct {
	service service.PermissionRequestService
}

func NewPermissionRequestHandler(service service.PermissionRequestService) *PermissionRequestHandler {
	return &PermissionRequestHandler{service: service}
}

// Metadata — GET /permission-requests/metadata
func (h *PermissionRequestHandler) Metadata(c *fiber.Ctx) error {
	res, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "request metadata",
		Data:       res,
	})
}

// List — GET /permission-requests
func (h *PermissionRequestHandler) List(c *fiber.Ctx) error {
	var params dto.PermissionListParams
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
		Message:    "permission requests list",
		Data:       res,
	})
}

// Detail — GET /permission-requests/:id
func (h *PermissionRequestHandler) Detail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid permission request ID")
	}

	res, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "permission request detail",
		Data:       res,
	})
}

// Create — POST /permission-requests
func (h *PermissionRequestHandler) Create(c *fiber.Ctx) error {
	var req dto.CreatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.Create(c.Context(), account.EmployeeID, account.RoleLevel, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "permission request created",
		Data:       res,
	})
}

// UpdateStatus — PUT /permission-requests/:id
func (h *PermissionRequestHandler) UpdateStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid permission request ID")
	}

	var req dto.UpdatePermissionStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.UpdateStatus(c.Context(), account.EmployeeID, uint(id), req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "permission request status updated",
		Data:       res,
	})
}

// Delete — DELETE /permission-requests/:id
func (h *PermissionRequestHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid permission request ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "permission request deleted",
	})
}