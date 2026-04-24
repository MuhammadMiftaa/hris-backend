package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// GetEmployeeDashboard — GET /dashboard/employee
func (h *DashboardHandler) GetEmployeeDashboard(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	res, err := h.service.GetEmployeeDashboard(c.Context(), account.EmployeeID, account.IsTrainer)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "employee dashboard",
		Data:       res,
	})
}

// GetHRDDashboard — GET /dashboard/hrd
func (h *DashboardHandler) GetHRDDashboard(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	res, err := h.service.GetHRDDashboard(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "hrd dashboard",
		Data:       res,
	})
}

// GetRankings — GET /dashboard/rankings
func (h *DashboardHandler) GetRankings(c *fiber.Ctx) error {
	res, err := h.service.GetRankings(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "dashboard rankings",
		Data:       res,
	})
}

// GetMetadata — GET /dashboard/metadata
func (h *DashboardHandler) GetMetadata(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	res, err := h.service.GetDashboardMetadata(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "dashboard metadata",
		Data:       res,
	})
}
