package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type DepartmentHandler struct {
	service service.DepartmentService
}

func NewDepartmentHandler(service service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

func (h *DepartmentHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department metadata",
		Data:       result,
	})
}

func (h *DepartmentHandler) List(c *fiber.Ctx) error {
	var params dto.DepartmentListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllDepartments(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department list",
		Data:       result,
	})
}

func (h *DepartmentHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetDepartmentByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department detail",
		Data:       result,
	})
}

func (h *DepartmentHandler) Export(c *fiber.Ctx) error {
	var params dto.DepartmentListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAllDepartments(c.Context(), params)
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

func (h *DepartmentHandler) exportCSV(c *fiber.Ctx, depts []dto.DepartmentResponse) error {
	headers := []string{"Kode", "Nama Departemen", "Cabang"}
	var rows [][]string
	for _, d := range depts {
		branch := "-"
		if d.BranchName != nil {
			branch = *d.BranchName
		}
		rows = append(rows, []string{
			d.Code, d.Name, branch,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=departemen.csv")
	return c.Send(data)
}

func (h *DepartmentHandler) exportPDF(c *fiber.Ctx, depts []dto.DepartmentResponse) error {
	headers := []string{"Kode", "Nama Departemen", "Cabang"}
	var rows [][]string
	for _, d := range depts {
		branch := "-"
		if d.BranchName != nil {
			branch = *d.BranchName
		}
		rows = append(rows, []string{
			d.Code, d.Name, branch,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Departemen", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(depts),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=departemen.pdf")
	return c.Send(pdf)
}

func (h *DepartmentHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateDepartmentRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateDepartment(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Department created",
		Data:       result,
	})
}

func (h *DepartmentHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateDepartmentRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateDepartment(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department updated",
		Data:       result,
	})
}

func (h *DepartmentHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteDepartment(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Department deleted",
	})
}
