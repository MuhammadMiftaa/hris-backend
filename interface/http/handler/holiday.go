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

type HolidayHandler struct {
	service service.HolidayService
}

func NewHolidayHandler(service service.HolidayService) *HolidayHandler {
	return &HolidayHandler{service: service}
}

func (h *HolidayHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday metadata", Data: result})
}

func (h *HolidayHandler) List(c *fiber.Ctx) error {
	var params dto.HolidayListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllHolidays(c.Context(), &params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday list", Data: result})
}

func (h *HolidayHandler) Export(c *fiber.Ctx) error {
	var params dto.HolidayListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAllHolidays(c.Context(), &params)
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

func (h *HolidayHandler) exportCSV(c *fiber.Ctx, holidays []dto.HolidayResponse) error {
	headers := []string{"Tahun", "Nama Libur", "Cabang ID", "Tipe", "Tanggal"}
	var rows [][]string
	for _, hol := range holidays {
		rows = append(rows, []string{
			fmt.Sprintf("%d", hol.Year), hol.Name, fmt.Sprintf("%v", hol.BranchID), hol.Type, hol.Date,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=hari_libur.csv")
	return c.Send(data)
}

func (h *HolidayHandler) exportPDF(c *fiber.Ctx, holidays []dto.HolidayResponse) error {
	headers := []string{"Tahun", "Nama Libur", "Cabang ID", "Tipe", "Tanggal"}
	var rows [][]string
	for _, hol := range holidays {
		rows = append(rows, []string{
			fmt.Sprintf("%d", hol.Year), hol.Name, fmt.Sprintf("%v", hol.BranchID), hol.Type, hol.Date,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Hari Libur", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(holidays),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=hari_libur.pdf")
	return c.Send(pdf)
}

func (h *HolidayHandler) Detail(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	result, err := h.service.GetHolidayByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday detail", Data: result})
}

func (h *HolidayHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateHolidayRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.CreateHoliday(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status: true, StatusCode: 201, Message: "Holiday created", Data: result,
	})
}

func (h *HolidayHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	var input dto.UpdateHolidayRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}
	result, err := h.service.UpdateHoliday(c.Context(), uint(id), input)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday updated", Data: result})
}

func (h *HolidayHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondBadRequest(c, "Invalid ID")
	}
	if err := h.service.DeleteHoliday(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Holiday deleted"})
}

func (h *HolidayHandler) SyncFromExternalAPI(c *fiber.Ctx) error {
	var req dto.SyncHolidayRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	res, err := h.service.SyncFromExternalAPI(c.Context(), req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "holiday sync completed",
		Data:       res,
	})
}
