package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type BusinessTripHandler struct {
	service service.BusinessTripService
}

func NewBusinessTripHandler(service service.BusinessTripService) *BusinessTripHandler {
	return &BusinessTripHandler{service: service}
}

// List — GET /business-trips
func (h *BusinessTripHandler) List(c *fiber.Ctx) error {
	var params dto.BusinessTripListParams
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
		Message:    "business trip requests list",
		Data:       res,
	})
}

// Detail — GET /business-trips/:id
func (h *BusinessTripHandler) Detail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid business trip ID")
	}

	res, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "business trip detail",
		Data:       res,
	})
}

// Create — POST /business-trips
func (h *BusinessTripHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateBusinessTripRequest
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
		Message:    "business trip request created",
		Data:       res,
	})
}

// UpdateStatus — PUT /business-trips/:id
func (h *BusinessTripHandler) UpdateStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid business trip ID")
	}

	var req dto.UpdateBusinessTripStatusRequest
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
		Message:    "business trip request status updated",
		Data:       res,
	})
}

// Delete — DELETE /business-trips/:id
func (h *BusinessTripHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid business trip ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "business trip request deleted",
	})
}
