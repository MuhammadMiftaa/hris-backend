package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type BusinessTripHandler struct {
	service service.BusinessTripService
}

func NewBusinessTripHandler(service service.BusinessTripService) *BusinessTripHandler {
	return &BusinessTripHandler{service: service}
}

// List — GET /business-trips
func (h *BusinessTripHandler) List(c *fiber.Ctx) error {
	var params dto.BusinessTripListParams
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
		Message:    "business trip requests list",
		Data:       res,
	})
}

func (h *BusinessTripHandler) Export(c *fiber.Ctx) error {
	var params dto.BusinessTripListParams
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

func (h *BusinessTripHandler) exportCSV(c *fiber.Ctx, trips []dto.BusinessTripRequestResponse) error {
	headers := []string{"Pegawai", "Departemen", "Tujuan", "Mulai", "Selesai", "Status"}
	var rows [][]string
	for _, t := range trips {
		empName := "-"
		if t.EmployeeName != nil {
			empName = *t.EmployeeName
		}
		deptName := "-"
		if t.DepartmentName != nil {
			deptName = *t.DepartmentName
		}
		rows = append(rows, []string{
			empName, deptName, t.Destination, t.StartDate, t.EndDate, t.Status,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=perjalanan_dinas.csv")
	return c.Send(data)
}

func (h *BusinessTripHandler) exportPDF(c *fiber.Ctx, trips []dto.BusinessTripRequestResponse) error {
	headers := []string{"Pegawai", "Departemen", "Tujuan", "Mulai", "Selesai", "Status"}
	var rows [][]string
	for _, t := range trips {
		empName := "-"
		if t.EmployeeName != nil {
			empName = *t.EmployeeName
		}
		deptName := "-"
		if t.DepartmentName != nil {
			deptName = *t.DepartmentName
		}
		rows = append(rows, []string{
			empName, deptName, t.Destination, t.StartDate, t.EndDate, t.Status,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Perjalanan Dinas", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(trips),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=perjalanan_dinas.pdf")
	return c.Send(pdf)
}

// Detail — GET /business-trips/:id
func (h *BusinessTripHandler) Detail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid business trip ID")
	}

	res, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "business trip detail",
		Data:       res,
	})
}

// Create — POST /business-trips
func (h *BusinessTripHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateBusinessTripRequest
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
		Message:    "business trip request created",
		Data:       res,
	})
}

// UpdateStatus — PUT /business-trips/:id
func (h *BusinessTripHandler) UpdateStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid business trip ID")
	}

	var req dto.UpdateBusinessTripStatusRequest
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
		Message:    "business trip request status updated",
		Data:       res,
	})
}

// Delete — DELETE /business-trips/:id
func (h *BusinessTripHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid business trip ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "business trip request deleted",
	})
}
