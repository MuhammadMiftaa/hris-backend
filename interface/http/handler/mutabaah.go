package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type MutabaahHandler struct {
	service service.MutabaahService
}

func NewMutabaahHandler(service service.MutabaahService) *MutabaahHandler {
	return &MutabaahHandler{service: service}
}

// GetTodayStatus — GET /mutabaah/today
func (h *MutabaahHandler) GetTodayStatus(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	result, err := h.service.GetTodayStatus(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "today mutabaah status",
		Data:       result,
	})
}

// Submit — POST /mutabaah/submit
func (h *MutabaahHandler) Submit(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var req dto.MutabaahSubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	result, err := h.service.Submit(c.Context(), account.EmployeeID, account.IsTrainer, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "mutabaah berhasil disubmit",
		Data:       result,
	})
}

// Cancel — POST /mutabaah/cancel
func (h *MutabaahHandler) Cancel(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var req dto.MutabaahCancelRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	result, err := h.service.Cancel(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah berhasil dibatalkan",
		Data:       result,
	})
}

// List — GET /mutabaah
func (h *MutabaahHandler) List(c *fiber.Ctx) error {
	var params dto.MutabaahListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllLogs(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah list",
		Data:       result,
	})
}

func (h *MutabaahHandler) Export(c *fiber.Ctx) error {
	var params dto.MutabaahListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAllLogs(c.Context(), params)
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

func (h *MutabaahHandler) exportCSV(c *fiber.Ctx, logs []dto.MutabaahLogResponse) error {
	headers := []string{"Tanggal", "Pegawai", "Target Halaman", "Status Submit", "Waktu Submit"}
	var rows [][]string
	for _, m := range logs {
		status := "Belum"
		if m.IsSubmitted {
			status = "Sudah"
		}
		submittedAt := "-"
		if m.SubmittedAt != nil {
			submittedAt = m.SubmittedAt.Format("15:04")
		}
		rows = append(rows, []string{
			m.LogDate, m.EmployeeName, fmt.Sprintf("%d", m.TargetPages), status, submittedAt,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=mutabaah.csv")
	return c.Send(data)
}

func (h *MutabaahHandler) exportPDF(c *fiber.Ctx, logs []dto.MutabaahLogResponse) error {
	headers := []string{"Tanggal", "Pegawai", "Target", "Status", "Waktu"}
	var rows [][]string
	for _, m := range logs {
		status := "Belum"
		if m.IsSubmitted {
			status = "Sudah"
		}
		submittedAt := "-"
		if m.SubmittedAt != nil {
			submittedAt = m.SubmittedAt.Format("15:04")
		}
		rows = append(rows, []string{
			m.LogDate, m.EmployeeName, fmt.Sprintf("%d", m.TargetPages), status, submittedAt,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Mutabaah (Tahsin/Tahfidz)", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(logs),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=mutabaah.pdf")
	return c.Send(pdf)
}

// HRDCancel — PUT /mutabaah/:id/cancel
func (h *MutabaahHandler) HRDCancel(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid mutabaah ID")
	}

	result, err := h.service.HRDCancel(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah reports canceled by HRD",
		Data:       result,
	})
}

func (h *MutabaahHandler) GetDailyReport(c *fiber.Ctx) error {
	startDate := c.Query("start_date", c.Query("date"))
	endDate := c.Query("end_date", startDate)
	if startDate == "" {
		return respondBadRequest(c, "query start_date atau date wajib diisi (format YYYY-MM-DD)")
	}

	result, err := h.service.GetDailyReport(c.Context(), startDate, endDate)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah daily report",
		Data:       result,
	})
}

func (h *MutabaahHandler) GetMonthlyReport(c *fiber.Ctx) error {
	month := c.QueryInt("month")
	year := c.QueryInt("year")
	if month <= 0 || month > 12 || year <= 0 {
		return respondBadRequest(c, "query month dan year wajib diisi dan valid")
	}

	result, err := h.service.GetMonthlyReport(c.Context(), month, year)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah monthly report",
		Data:       result,
	})
}

// GetCategoryReport — GET /mutabaah/report/category?date=2024-01-01
func (h *MutabaahHandler) GetCategoryReport(c *fiber.Ctx) error {
	date := c.Query("date")
	if date == "" {
		return respondBadRequest(c, "query date wajib diisi (format YYYY-MM-DD)")
	}

	result, err := h.service.GetCategoryReport(c.Context(), date)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah category report",
		Data:       result,
	})
}

// ExportDailyReport — GET /mutabaah/report/daily/export?start_date=...&end_date=...&format=csv|pdf
func (h *MutabaahHandler) ExportDailyReport(c *fiber.Ctx) error {
	startDate := c.Query("start_date", c.Query("date"))
	endDate := c.Query("end_date", startDate)
	if startDate == "" {
		return respondBadRequest(c, "query start_date atau date wajib diisi (format YYYY-MM-DD)")
	}
	format := c.Query("format", "csv")

	result, err := h.service.GetDailyReport(c.Context(), startDate, endDate)
	if err != nil {
		return respondError(c, err)
	}

	headers := []string{"Nama", "NIP", "Departemen", "Kategori", "Target (hlm)", "Status", "Waktu Submit"}
	var rows [][]string
	for _, r := range result {
		cat := "Non-Trainer"
		if r.IsTrainer {
			cat = "Trainer"
		}
		status := "Belum"
		if r.IsSubmitted {
			status = "Sudah"
		}
		dept := "-"
		if r.DepartmentName != nil {
			dept = *r.DepartmentName
		}
		submittedAt := "-"
		if r.SubmittedAt != nil {
			submittedAt = *r.SubmittedAt
		}
		rows = append(rows, []string{
			r.EmployeeName, r.EmployeeNumber, dept, cat,
			fmt.Sprintf("%d", r.TargetPages), status, submittedAt,
		})
	}

	switch format {
	case "pdf":
		html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
			Title: "Mutabaah Harian — " + startDate + " s/d " + endDate, Date: time.Now().Format("02 Jan 2006"),
			Headers: headers, Rows: rows, TotalData: len(result),
		})
		if err != nil {
			return respondError(c, err)
		}
		pdf, err := utils.GeneratePDF(html)
		if err != nil {
			return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
		}
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "attachment; filename=mutabaah_harian.pdf")
		return c.Send(pdf)
	default:
		data, err := utils.WriteCSV(headers, rows)
		if err != nil {
			return respondError(c, err)
		}
		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", "attachment; filename=mutabaah_harian.csv")
		return c.Send(data)
	}
}

// ExportMonthlyReport — GET /mutabaah/report/monthly/export?month=...&year=...&format=csv|pdf
func (h *MutabaahHandler) ExportMonthlyReport(c *fiber.Ctx) error {
	month := c.QueryInt("month")
	year := c.QueryInt("year")
	if month <= 0 || month > 12 || year <= 0 {
		return respondBadRequest(c, "query month dan year wajib diisi dan valid")
	}
	format := c.Query("format", "csv")

	result, err := h.service.GetMonthlyReport(c.Context(), month, year)
	if err != nil {
		return respondError(c, err)
	}

	headers := []string{"Nama", "Departemen", "Kategori", "Hari Wajib", "Hari Submit", "% Kepatuhan"}
	var rows [][]string
	for _, r := range result {
		cat := "Non-Trainer"
		if r.IsTrainer {
			cat = "Trainer"
		}
		dept := "-"
		if r.DepartmentName != nil {
			dept = *r.DepartmentName
		}
		rows = append(rows, []string{
			r.EmployeeName, dept, cat,
			fmt.Sprintf("%d", r.TotalWorkingDays),
			fmt.Sprintf("%d", r.TotalSubmitted),
			fmt.Sprintf("%.1f%%", r.CompliancePercentage),
		})
	}

	switch format {
	case "pdf":
		html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
			Title:   fmt.Sprintf("Mutabaah Bulanan — %d/%d", month, year),
			Date:    time.Now().Format("02 Jan 2006"),
			Headers: headers, Rows: rows, TotalData: len(result),
		})
		if err != nil {
			return respondError(c, err)
		}
		pdf, err := utils.GeneratePDF(html)
		if err != nil {
			return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
		}
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "attachment; filename=mutabaah_bulanan.pdf")
		return c.Send(pdf)
	default:
		data, err := utils.WriteCSV(headers, rows)
		if err != nil {
			return respondError(c, err)
		}
		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", "attachment; filename=mutabaah_bulanan.csv")
		return c.Send(data)
	}
}
