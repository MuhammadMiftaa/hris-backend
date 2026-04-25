package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type OvertimeHandler struct {
	service service.OvertimeService
}

func NewOvertimeHandler(service service.OvertimeService) *OvertimeHandler {
	return &OvertimeHandler{service: service}
}

// List — GET /overtime-requests
func (h *OvertimeHandler) List(c *fiber.Ctx) error {
	var params dto.OvertimeListParams
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
		Message:    "overtime requests list",
		Data:       res,
	})
}

// Detail — GET /overtime-requests/:id
func (h *OvertimeHandler) Detail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid overtime ID")
	}

	res, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "overtime request detail",
		Data:       res,
	})
}

// Create — POST /overtime-requests
func (h *OvertimeHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateOvertimeRequest
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
		Message:    "overtime request created",
		Data:       res,
	})
}

// Approve — PUT /overtime-requests/:id/approve
func (h *OvertimeHandler) Approve(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid overtime ID")
	}

	var req dto.ApproveOvertimeRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.ApproveRequest(c.Context(), account.EmployeeID, uint(id), req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "overtime request approved",
		Data:       res,
	})
}

// Reject — PUT /overtime-requests/:id/reject
func (h *OvertimeHandler) Reject(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid overtime ID")
	}

	var req dto.RejectOvertimeRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.RejectRequest(c.Context(), account.EmployeeID, uint(id), req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "overtime request rejected",
		Data:       res,
	})
}

// Delete — DELETE /overtime-requests/:id
func (h *OvertimeHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid overtime ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "overtime request deleted",
	})
}
