package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

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

func (h *PermissionRequestHandler) Export(c *fiber.Ctx) error {
	var params dto.PermissionListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAll(c.Context(), params)
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

func (h *PermissionRequestHandler) exportCSV(c *fiber.Ctx, perms []dto.PermissionRequestResponse) error {
	headers := []string{"Tanggal", "Pegawai", "Departemen", "Tipe Izin", "Mulai", "Selesai", "Status"}
	var rows [][]string
	for _, p := range perms {
		empName := "-"
		if p.EmployeeName != nil {
			empName = *p.EmployeeName
		}
		deptName := "-"
		if p.DepartmentName != nil {
			deptName = *p.DepartmentName
		}
		leave := "-"
		if p.LeaveTime != nil {
			leave = *p.LeaveTime
		}
		ret := "-"
		if p.ReturnTime != nil {
			ret = *p.ReturnTime
		}
		rows = append(rows, []string{
			p.Date, empName, deptName, p.PermissionType, leave, ret, p.Status,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=perizinan.csv")
	return c.Send(data)
}

func (h *PermissionRequestHandler) exportPDF(c *fiber.Ctx, perms []dto.PermissionRequestResponse) error {
	headers := []string{"Tanggal", "Pegawai", "Departemen", "Tipe Izin", "Mulai", "Selesai", "Status"}
	var rows [][]string
	for _, p := range perms {
		empName := "-"
		if p.EmployeeName != nil {
			empName = *p.EmployeeName
		}
		deptName := "-"
		if p.DepartmentName != nil {
			deptName = *p.DepartmentName
		}
		leave := "-"
		if p.LeaveTime != nil {
			leave = *p.LeaveTime
		}
		ret := "-"
		if p.ReturnTime != nil {
			ret = *p.ReturnTime
		}
		rows = append(rows, []string{
			p.Date, empName, deptName, p.PermissionType, leave, ret, p.Status,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Perizinan", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(perms),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=perizinan.pdf")
	return c.Send(pdf)
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