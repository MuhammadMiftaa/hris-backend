package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type LeaveTypeHandler struct {
	service service.LeaveTypeService
}

func NewLeaveTypeHandler(service service.LeaveTypeService) *LeaveTypeHandler {
	return &LeaveTypeHandler{service: service}
}

func (h *LeaveTypeHandler) Metadata(c *fiber.Ctx) error {
	result := h.service.GetMetadata(c.Context())

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type metadata",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) List(c *fiber.Ctx) error {
	var params dto.LeaveTypeListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllLeaveTypes(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type list",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetLeaveTypeByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type detail",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateLeaveTypeRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateLeaveType(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Leave type created",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateLeaveTypeRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateLeaveType(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type updated",
		Data:       result,
	})
}

func (h *LeaveTypeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteLeaveType(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Leave type deleted",
	})
}

func (h *LeaveTypeHandler) Export(c *fiber.Ctx) error {
	var params dto.LeaveTypeListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.ExportLeaveTypes(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	switch exportReq.Format {
	case dto.ExportCSV:
		headers := []string{"ID", "Nama", "Kategori", "Wajib Dokumen", "Jenis Dokumen", "Maks/Request", "Maks/Tahun", "Potongan"}
		var rows [][]string
		for _, lt := range result.Data {
			reqDoc := "Tidak"
			if lt.RequiresDocument {
				reqDoc = "Ya"
			}
			docType := "-"
			if lt.RequiresDocumentType != nil {
				docType = *lt.RequiresDocumentType
			}
			maxReq := "-"
			if lt.MaxDurationPerRequest != nil {
				maxReq = fmt.Sprintf("%.2f", *lt.MaxDurationPerRequest)
			}
			maxYear := "-"
			if lt.MaxTotalDurationPerYear != nil {
				maxYear = fmt.Sprintf("%.2f", *lt.MaxTotalDurationPerYear)
			}
			rows = append(rows, []string{
				fmt.Sprintf("%d", lt.ID), lt.Name, string(lt.Category), reqDoc, docType, maxReq, maxYear, fmt.Sprintf("%.2f", lt.DeductDays),
			})
		}
		data, err := utils.WriteCSV(headers, rows)
		if err != nil {
			return respondError(c, err)
		}
		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", "attachment; filename=jenis_cuti.csv")
		return c.Send(data)

	case dto.ExportPDF:
		headers := []string{"Nama", "Kategori", "Wajib Dokumen", "Maks/Request", "Maks/Tahun", "Potongan"}
		var rows [][]string
		for _, lt := range result.Data {
			reqDoc := "Tidak"
			if lt.RequiresDocument {
				reqDoc = "Ya"
			}
			maxReq := "-"
			if lt.MaxDurationPerRequest != nil {
				maxReq = fmt.Sprintf("%.2f", *lt.MaxDurationPerRequest)
			}
			maxYear := "-"
			if lt.MaxTotalDurationPerYear != nil {
				maxYear = fmt.Sprintf("%.2f", *lt.MaxTotalDurationPerYear)
			}
			rows = append(rows, []string{
				lt.Name, string(lt.Category), reqDoc, maxReq, maxYear, fmt.Sprintf("%.2f", lt.DeductDays),
			})
		}
		html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
			Title: "Daftar Jenis Cuti", Date: time.Now().Format("02 Jan 2006"),
			Headers: headers, Rows: rows, TotalData: len(result.Data),
		})
		if err != nil {
			return respondError(c, err)
		}
		pdf, err := utils.GeneratePDF(html)
		if err != nil {
			return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
		}
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "attachment; filename=jenis_cuti.pdf")
		return c.Send(pdf)
	default:
		return respondBadRequest(c, "format must be csv or pdf")
	}
}
