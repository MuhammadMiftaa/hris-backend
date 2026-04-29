package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type DailyReportHandler struct {
	service service.DailyReportService
}

func NewDailyReportHandler(service service.DailyReportService) *DailyReportHandler {
	return &DailyReportHandler{service: service}
}

// List — GET /daily-reports
func (h *DailyReportHandler) List(c *fiber.Ctx) error {
	var params dto.DailyReportListParams
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
		Message:    "daily reports list",
		Data:       res,
	})
}

func (h *DailyReportHandler) Export(c *fiber.Ctx) error {
	var params dto.DailyReportListParams
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

func (h *DailyReportHandler) exportCSV(c *fiber.Ctx, reports []dto.DailyReportResponse) error {
	headers := []string{"Tanggal", "Pegawai", "Status Submit", "Aktivitas"}
	var rows [][]string
	for _, r := range reports {
		empName := "-"
		if r.EmployeeName != nil {
			empName = *r.EmployeeName
		}
		activities := "-"
		if r.Activities != nil {
			activities = *r.Activities
		}
		status := "Belum"
		if r.IsSubmitted {
			status = "Sudah"
		}
		rows = append(rows, []string{
			r.ReportDate, empName, status, activities,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=laporan_harian.csv")
	return c.Send(data)
}

func (h *DailyReportHandler) exportPDF(c *fiber.Ctx, reports []dto.DailyReportResponse) error {
	headers := []string{"Tanggal", "Pegawai", "Status Submit", "Aktivitas"}
	var rows [][]string
	for _, r := range reports {
		empName := "-"
		if r.EmployeeName != nil {
			empName = *r.EmployeeName
		}
		activities := "-"
		if r.Activities != nil {
			activities = *r.Activities
		}
		status := "Belum"
		if r.IsSubmitted {
			status = "Sudah"
		}
		rows = append(rows, []string{
			r.ReportDate, empName, status, activities,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Laporan Harian", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(reports),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=laporan_harian.pdf")
	return c.Send(pdf)
}

// Detail — GET /daily-reports/:id
func (h *DailyReportHandler) Detail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid daily report ID")
	}

	res, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "daily report detail",
		Data:       res,
	})
}

// Create — POST /daily-reports
func (h *DailyReportHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateDailyReportRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.Create(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "daily report created",
		Data:       res,
	})
}

// Update — PUT /daily-reports/:id
func (h *DailyReportHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid daily report ID")
	}

	var req dto.UpdateDailyReportRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.Update(c.Context(), uint(id), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "daily report updated",
		Data:       res,
	})
}

// Delete — DELETE /daily-reports/:id
func (h *DailyReportHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid daily report ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "daily report deleted",
	})
}
