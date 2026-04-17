package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type BranchHandler struct {
	service service.BranchService
}

func NewBranchHandler(service service.BranchService) *BranchHandler {
	return &BranchHandler{service: service}
}

func (h *BranchHandler) List(c *fiber.Ctx) error {
	result, err := h.service.GetAllBranches(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Branch list",
		Data:       result,
	})
}

func (h *BranchHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetBranchByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Branch detail",
		Data:       result,
	})
}

func (h *BranchHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateBranchRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateBranch(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Branch created",
		Data:       result,
	})
}

func (h *BranchHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateBranchRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateBranch(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Branch updated",
		Data:       result,
	})
}

func (h *BranchHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteBranch(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Branch deleted",
	})
}
