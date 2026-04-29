package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

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

func (h *OvertimeHandler) Export(c *fiber.Ctx) error {
	var params dto.OvertimeListParams
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

func (h *OvertimeHandler) exportCSV(c *fiber.Ctx, overtimes []dto.OvertimeRequestResponse) error {
	headers := []string{"Tanggal Overtime", "Pegawai", "Departemen", "Mulai", "Selesai", "Status"}
	var rows [][]string
	for _, o := range overtimes {
		empName := "-"
		if o.EmployeeName != nil {
			empName = *o.EmployeeName
		}
		deptName := "-"
		if o.DepartmentName != nil {
			deptName = *o.DepartmentName
		}
		start := "-"
		if o.PlannedStart != nil {
			start = o.PlannedStart.Format("15:04")
		}
		end := "-"
		if o.PlannedEnd != nil {
			end = o.PlannedEnd.Format("15:04")
		}
		rows = append(rows, []string{
			o.OvertimeDate, empName, deptName, start, end, o.Status,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=overtime.csv")
	return c.Send(data)
}

func (h *OvertimeHandler) exportPDF(c *fiber.Ctx, overtimes []dto.OvertimeRequestResponse) error {
	headers := []string{"Tanggal", "Pegawai", "Departemen", "Mulai", "Selesai", "Status"}
	var rows [][]string
	for _, o := range overtimes {
		empName := "-"
		if o.EmployeeName != nil {
			empName = *o.EmployeeName
		}
		deptName := "-"
		if o.DepartmentName != nil {
			deptName = *o.DepartmentName
		}
		start := "-"
		if o.PlannedStart != nil {
			start = o.PlannedStart.Format("15:04")
		}
		end := "-"
		if o.PlannedEnd != nil {
			end = o.PlannedEnd.Format("15:04")
		}
		rows = append(rows, []string{
			o.OvertimeDate, empName, deptName, start, end, o.Status,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Lembur", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(overtimes),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=overtime.pdf")
	return c.Send(pdf)
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
