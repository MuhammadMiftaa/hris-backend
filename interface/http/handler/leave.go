package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type LeaveHandler struct {
	service service.LeaveService
}

func NewLeaveHandler(service service.LeaveService) *LeaveHandler {
	return &LeaveHandler{service: service}
}

// Metadata — GET /leave-requests/metadata
func (h *LeaveHandler) Metadata(c *fiber.Ctx) error {
	res, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave metadata",
		Data:       res,
	})
}

// ListBalances — GET /leave-balances
func (h *LeaveHandler) ListBalances(c *fiber.Ctx) error {
	var params dto.LeaveBalanceListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	res, err := h.service.GetAllBalances(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave balances",
		Data:       res,
	})
}

// ListRequests — GET /leave-requests
func (h *LeaveHandler) ListRequests(c *fiber.Ctx) error {
	var params dto.LeaveRequestListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	account := getAccountFromCtx(c)
	res, err := h.service.GetAllRequests(c.Context(), account.RoleLevel, params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave requests list",
		Data:       res,
	})
}

// DetailRequest — GET /leave-requests/:id
func (h *LeaveHandler) DetailRequest(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid leave request ID")
	}

	res, err := h.service.GetRequestByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave request detail",
		Data:       res,
	})
}

// Create — POST /leave-requests
func (h *LeaveHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateLeaveRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.CreateRequest(c.Context(), account.EmployeeID, account.RoleLevel, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "leave request created",
		Data:       res,
	})
}

// Approve — PUT /leave-requests/:id/approve
func (h *LeaveHandler) Approve(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid leave request ID")
	}

	var req dto.ApproveLeaveRequest
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
		Message:    "leave request approved",
		Data:       res,
	})
}

// Reject — PUT /leave-requests/:id/reject
func (h *LeaveHandler) Reject(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid leave request ID")
	}

	var req dto.RejectLeaveRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	if req.Notes == "" {
		return respondBadRequest(c, "rejection notes are required")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.RejectRequest(c.Context(), account.EmployeeID, uint(id), req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "leave request rejected",
		Data:       res,
	})
}

// ExportBalances — GET /leave-balances/export
func (h *LeaveHandler) ExportBalances(c *fiber.Ctx) error {
	var params dto.LeaveBalanceListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}
	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	res, err := h.service.GetAllBalances(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	switch exportReq.Format {
	case dto.ExportCSV:
		headers := []string{"Nama Pegawai", "Departemen", "Tipe Cuti", "Tahun", "Terpakai (Hari)", "Remaining (Hari)"}
		var rows [][]string
		for _, b := range res.Data {
			deptName := "-"
			if b.DepartmentName != nil {
				deptName = *b.DepartmentName
			}
			empName := "-"
			if b.EmployeeName != nil {
				empName = *b.EmployeeName
			}
			leaveName := "-"
			if b.LeaveTypeName != nil {
				leaveName = *b.LeaveTypeName
			}
			rows = append(rows, []string{
				empName, deptName, leaveName, fmt.Sprintf("%d", b.Year),
				fmt.Sprintf("%d", b.UsedDuration), fmt.Sprintf("%d", *b.RemainingDuration),
			})
		}
		data, err := utils.WriteCSV(headers, rows)
		if err != nil {
			return respondError(c, err)
		}
		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", "attachment; filename=saldo_cuti.csv")
		return c.Send(data)

	case dto.ExportPDF:
		headers := []string{"Nama Pegawai", "Departemen", "Tipe Cuti", "Tahun", "Terpakai", "Sisa"}
		var rows [][]string
		for _, b := range res.Data {
			deptName := "-"
			if b.DepartmentName != nil {
				deptName = *b.DepartmentName
			}
			empName := "-"
			if b.EmployeeName != nil {
				empName = *b.EmployeeName
			}
			leaveName := "-"
			if b.LeaveTypeName != nil {
				leaveName = *b.LeaveTypeName
			}
			rows = append(rows, []string{
				empName, deptName, leaveName, fmt.Sprintf("%d", b.Year),
				fmt.Sprintf("%d", b.UsedDuration), fmt.Sprintf("%d", *b.RemainingDuration),
			})
		}
		html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
			Title: "Laporan Saldo Cuti", Date: time.Now().Format("02 Jan 2006"),
			Headers: headers, Rows: rows, TotalData: len(res.Data),
		})
		if err != nil {
			return respondError(c, err)
		}
		pdf, err := utils.GeneratePDF(html)
		if err != nil {
			return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
		}
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "attachment; filename=saldo_cuti.pdf")
		return c.Send(pdf)
	default:
		return respondBadRequest(c, "format must be csv or pdf")
	}
}

// ExportRequests — GET /leave-requests/export
func (h *LeaveHandler) ExportRequests(c *fiber.Ctx) error {
	var params dto.LeaveRequestListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}
	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	account := getAccountFromCtx(c)
	allPerPage := 0
	params.PerPage = &allPerPage

	res, err := h.service.GetAllRequests(c.Context(), account.RoleLevel, params)
	if err != nil {
		return respondError(c, err)
	}

	switch exportReq.Format {
	case dto.ExportCSV:
		headers := []string{"Nama Pegawai", "Departemen", "Tipe Cuti", "Mulai", "Selesai", "Total Hari", "Status", "Alasan"}
		var rows [][]string
		for _, req := range res.Data {
			deptName := "-"
			if req.DepartmentName != nil {
				deptName = *req.DepartmentName
			}
			empName := "-"
			if req.EmployeeName != nil {
				empName = *req.EmployeeName
			}
			leaveName := "-"
			if req.LeaveTypeName != nil {
				leaveName = *req.LeaveTypeName
			}
			reason := "-"
			if req.Reason != nil {
				reason = *req.Reason
			}
			rows = append(rows, []string{
				empName, deptName, leaveName,
				req.StartDate, req.EndDate, fmt.Sprintf("%d", req.TotalDays),
				req.Status, reason,
			})
		}
		data, err := utils.WriteCSV(headers, rows)
		if err != nil {
			return respondError(c, err)
		}
		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", "attachment; filename=pengajuan_cuti.csv")
		return c.Send(data)

	case dto.ExportPDF:
		headers := []string{"Nama Pegawai", "Departemen", "Tipe Cuti", "Mulai", "Selesai", "Hari", "Status"}
		var rows [][]string
		for _, req := range res.Data {
			deptName := "-"
			if req.DepartmentName != nil {
				deptName = *req.DepartmentName
			}
			empName := "-"
			if req.EmployeeName != nil {
				empName = *req.EmployeeName
			}
			leaveName := "-"
			if req.LeaveTypeName != nil {
				leaveName = *req.LeaveTypeName
			}
			rows = append(rows, []string{
				empName, deptName, leaveName,
				req.StartDate, req.EndDate, fmt.Sprintf("%d", req.TotalDays),
				req.Status,
			})
		}
		html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
			Title: "Laporan Pengajuan Cuti", Date: time.Now().Format("02 Jan 2006"),
			Headers: headers, Rows: rows, TotalData: len(res.Data),
		})
		if err != nil {
			return respondError(c, err)
		}
		pdf, err := utils.GeneratePDF(html)
		if err != nil {
			return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
		}
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "attachment; filename=pengajuan_cuti.pdf")
		return c.Send(pdf)
	default:
		return respondBadRequest(c, "format must be csv or pdf")
	}
}
