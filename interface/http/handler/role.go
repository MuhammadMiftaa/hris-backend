package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type RoleHandler struct {
	service service.RoleService
}

func NewRoleHandler(service service.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) Metadata(c *fiber.Ctx) error {
	result := h.service.GetMetadata(c.Context())

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Role metadata",
		Data:       result,
	})
}

func (h *RoleHandler) List(c *fiber.Ctx) error {
	result, err := h.service.GetAllRoles(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Role list",
		Data:       result,
	})
}

func (h *RoleHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetRoleByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Role detail",
		Data:       result,
	})
}

func (h *RoleHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateRoleRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateRole(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Role created",
		Data:       result,
	})
}

func (h *RoleHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateRoleRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateRole(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Role updated",
		Data:       result,
	})
}

func (h *RoleHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteRole(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Role deleted",
	})
}

func (h *RoleHandler) ListPermissions(c *fiber.Ctx) error {
	result, err := h.service.GetAllPermissions(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Permission list",
		Data:       result,
	})
}

func (h *RoleHandler) UpdatePermissions(c *fiber.Ctx) error {
	var input dto.UpdateRolePermissionsRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateRolePermissions(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Role permissions updated",
		Data:       result,
	})
}
