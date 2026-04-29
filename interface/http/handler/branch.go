package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type BranchHandler struct {
	service service.BranchService
}

func NewBranchHandler(service service.BranchService) *BranchHandler {
	return &BranchHandler{service: service}
}

func (h *BranchHandler) List(c *fiber.Ctx) error {
	var params dto.BranchListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllBranches(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Branch list",
		Data:       result,
	})
}

func (h *BranchHandler) Export(c *fiber.Ctx) error {
	var params dto.BranchListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAllBranches(c.Context(), params)
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

func (h *BranchHandler) exportCSV(c *fiber.Ctx, branches []dto.BranchResponse) error {
	headers := []string{"Kode", "Nama Cabang", "Alamat", "Radius (m)", "Boleh WFH"}
	var rows [][]string
	for _, b := range branches {
		allowWfh := "Tidak"
		if b.AllowWFH {
			allowWfh = "Ya"
		}
		address := "-"
		if b.Address != nil {
			address = *b.Address
		}
		rows = append(rows, []string{
			b.Code, b.Name, address, fmt.Sprintf("%d", b.RadiusMeters), allowWfh,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=cabang.csv")
	return c.Send(data)
}

func (h *BranchHandler) exportPDF(c *fiber.Ctx, branches []dto.BranchResponse) error {
	headers := []string{"Kode", "Nama Cabang", "Alamat", "Radius (m)", "WFH"}
	var rows [][]string
	for _, b := range branches {
		allowWfh := "Tidak"
		if b.AllowWFH {
			allowWfh = "Ya"
		}
		address := "-"
		if b.Address != nil {
			address = *b.Address
		}
		rows = append(rows, []string{
			b.Code, b.Name, address, fmt.Sprintf("%d", b.RadiusMeters), allowWfh,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Cabang", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(branches),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=cabang.pdf")
	return c.Send(pdf)
}

func (h *BranchHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetBranchByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Branch detail",
		Data:       result,
	})
}

func (h *BranchHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateBranchRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateBranch(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Branch created",
		Data:       result,
	})
}

func (h *BranchHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateBranchRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateBranch(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Branch updated",
		Data:       result,
	})
}

func (h *BranchHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteBranch(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Branch deleted",
	})
}
