package handler

import (
	"fmt"
	"strconv"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type ShiftHandler struct {
	service service.ShiftService
}

func NewShiftHandler(service service.ShiftService) *ShiftHandler {
	return &ShiftHandler{service: service}
}

func (h *ShiftHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift metadata", Data: result})
}

func (h *ShiftHandler) ListTemplates(c *fiber.Ctx) error {
	var params dto.ShiftTemplateListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllTemplates(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift template list", Data: result})
}

func (h *ShiftHandler) ExportTemplates(c *fiber.Ctx) error {
	var params dto.ShiftTemplateListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAllTemplates(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	switch exportReq.Format {
	case dto.ExportCSV:
		return h.exportTemplatesCSV(c, result.Data)
	case dto.ExportPDF:
		return h.exportTemplatesPDF(c, result.Data)
	default:
		return respondBadRequest(c, "format must be csv or pdf")
	}
}

func (h *ShiftHandler) exportTemplatesCSV(c *fiber.Ctx, templates []dto.ShiftTemplateResponse) error {
	headers := []string{"ID", "Nama Template"}
	var rows [][]string
	for _, t := range templates {
		rows = append(rows, []string{
			fmt.Sprintf("%d", t.ID), t.Name,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=shift_templates.csv")
	return c.Send(data)
}

func (h *ShiftHandler) exportTemplatesPDF(c *fiber.Ctx, templates []dto.ShiftTemplateResponse) error {
	headers := []string{"ID", "Nama Template"}
	var rows [][]string
	for _, t := range templates {
		rows = append(rows, []string{
			fmt.Sprintf("%d", t.ID), t.Name,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Template Shift", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(templates),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=shift_templates.pdf")
	return c.Send(pdf)
}

func (h *ShiftHandler) DetailTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	result, err := h.service.GetTemplateByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift template detail", Data: result})
}

func (h *ShiftHandler) CreateTemplate(c *fiber.Ctx) error {
	var input dto.CreateShiftRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.CreateTemplate(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status: true, StatusCode: 201, Message: "Shift template created", Data: result,
	})
}

func (h *ShiftHandler) UpdateTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	var input dto.UpdateShiftRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.UpdateTemplate(c.Context(), uint(id), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift template updated", Data: result})
}

func (h *ShiftHandler) DeleteTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	if err := h.service.DeleteTemplate(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift template deleted"})
}

func (h *ShiftHandler) ListDetails(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	result, err := h.service.GetDetailsByTemplateID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Shift details list", Data: result})
}

func (h *ShiftHandler) ListSchedules(c *fiber.Ctx) error {
	var params dto.ScheduleListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllSchedules(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Schedule list", Data: result})
}

func (h *ShiftHandler) ExportSchedules(c *fiber.Ctx) error {
	var params dto.ScheduleListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAllSchedules(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	switch exportReq.Format {
	case dto.ExportCSV:
		return h.exportSchedulesCSV(c, result.Data)
	case dto.ExportPDF:
		return h.exportSchedulesPDF(c, result.Data)
	default:
		return respondBadRequest(c, "format must be csv or pdf")
	}
}

func (h *ShiftHandler) exportSchedulesCSV(c *fiber.Ctx, schedules []dto.ScheduleResponse) error {
	headers := []string{"Tanggal Efektif", "Tanggal Selesai", "Pegawai", "Nama Shift", "Status"}
	var rows [][]string
	for _, s := range schedules {
		empName := "-"
		if s.EmployeeName != nil {
			empName = *s.EmployeeName
		}
		shiftName := "-"
		if s.ShiftName != nil {
			shiftName = *s.ShiftName
		}
		end := "-"
		if s.EndDate != nil {
			end = *s.EndDate
		}
		status := "Tidak Aktif"
		if s.IsActive {
			status = "Aktif"
		}
		rows = append(rows, []string{
			s.EffectiveDate, end, empName, shiftName, status,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=jadwal_shift.csv")
	return c.Send(data)
}

func (h *ShiftHandler) exportSchedulesPDF(c *fiber.Ctx, schedules []dto.ScheduleResponse) error {
	headers := []string{"Tanggal Efektif", "Selesai", "Pegawai", "Nama Shift", "Status"}
	var rows [][]string
	for _, s := range schedules {
		empName := "-"
		if s.EmployeeName != nil {
			empName = *s.EmployeeName
		}
		shiftName := "-"
		if s.ShiftName != nil {
			shiftName = *s.ShiftName
		}
		end := "-"
		if s.EndDate != nil {
			end = *s.EndDate
		}
		status := "Tidak Aktif"
		if s.IsActive {
			status = "Aktif"
		}
		rows = append(rows, []string{
			s.EffectiveDate, end, empName, shiftName, status,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Jadwal Shift", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(schedules),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=jadwal_shift.pdf")
	return c.Send(pdf)
}

func (h *ShiftHandler) DetailSchedule(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	result, err := h.service.GetScheduleByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Schedule detail", Data: result})
}

func (h *ShiftHandler) CreateSchedule(c *fiber.Ctx) error {
	var input dto.CreateScheduleRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.CreateSchedule(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status: true, StatusCode: 201, Message: "Schedule created", Data: result,
	})
}

func (h *ShiftHandler) UpdateSchedule(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	var input dto.UpdateScheduleRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.UpdateSchedule(c.Context(), uint(id), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Schedule updated", Data: result})
}

func (h *ShiftHandler) DeleteSchedule(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	if err := h.service.DeleteSchedule(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Schedule deleted"})
}

// CheckTodaySchedule — GET /schedules/my-today
func (h *ShiftHandler) CheckTodaySchedule(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	result, err := h.service.CheckTodaySchedule(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "today schedule status", Data: result})
}

