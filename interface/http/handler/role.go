package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

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
	var params dto.RoleListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllRoles(c.Context(), params)
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

func (h *RoleHandler) Export(c *fiber.Ctx) error {
	var params dto.RoleListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAllRoles(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	switch exportReq.Format {
	case dto.ExportCSV:
		return h.exportCSV(c, result.Data)
	case dto.ExportPDF:
		return h.exportPDF(c, result.Data)
	default:
		return respondBadRequest(c, "format must be csv or pdf")
	}
}

func (h *RoleHandler) exportCSV(c *fiber.Ctx, roles []dto.RoleResponse) error {
	headers := []string{"Nama Role", "Level", "Total Izin"}
	var rows [][]string
	for _, r := range roles {
		rows = append(rows, []string{
			r.Name, r.Level, fmt.Sprintf("%d", r.PermissionCount),
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=peran_pengguna.csv")
	return c.Send(data)
}

func (h *RoleHandler) exportPDF(c *fiber.Ctx, roles []dto.RoleResponse) error {
	headers := []string{"Nama Role", "Level", "Total Izin"}
	var rows [][]string
	for _, r := range roles {
		rows = append(rows, []string{
			r.Name, r.Level, fmt.Sprintf("%d", r.PermissionCount),
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Peran Pengguna", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(roles),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=peran_pengguna.pdf")
	return c.Send(pdf)
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
