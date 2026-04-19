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

	res, err := h.service.GetEmployeeDashboard(c.Context(), account.EmployeeID)
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
