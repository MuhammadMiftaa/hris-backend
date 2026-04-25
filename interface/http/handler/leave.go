package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type LeaveHandler struct {
	service service.LeaveService
}

func NewLeaveHandler(service service.LeaveService) *LeaveHandler {
	return &LeaveHandler{service: service}
}

// Metadata — GET /leave-requests/metadata
func (h *LeaveHandler) Metadata(c *fiber.Ctx) error {
	res, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave metadata",
		Data:       res,
	})
}

// ListBalances — GET /leave-balances
func (h *LeaveHandler) ListBalances(c *fiber.Ctx) error {
	var params dto.LeaveBalanceListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	res, err := h.service.GetAllBalances(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave balances",
		Data:       res,
	})
}

// ListRequests — GET /leave-requests
func (h *LeaveHandler) ListRequests(c *fiber.Ctx) error {
	var params dto.LeaveRequestListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	res, err := h.service.GetAllRequests(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave requests list",
		Data:       res,
	})
}

// DetailRequest — GET /leave-requests/:id
func (h *LeaveHandler) DetailRequest(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid leave request ID")
	}

	res, err := h.service.GetRequestByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave request detail",
		Data:       res,
	})
}

// Create — POST /leave-requests
func (h *LeaveHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateLeaveRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.CreateRequest(c.Context(), account.EmployeeID, account.RoleLevel, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "leave request created",
		Data:       res,
	})
}

// Approve — PUT /leave-requests/:id/approve
func (h *LeaveHandler) Approve(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid leave request ID")
	}

	var req dto.ApproveLeaveRequest
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
		Message:    "leave request approved",
		Data:       res,
	})
}

// Reject — PUT /leave-requests/:id/reject
func (h *LeaveHandler) Reject(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid leave request ID")
	}

	var req dto.RejectLeaveRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	if req.Notes == "" {
		return respondBadRequest(c, "rejection notes are required")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.RejectRequest(c.Context(), account.EmployeeID, uint(id), req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave request rejected",
		Data:       res,
	})
}
