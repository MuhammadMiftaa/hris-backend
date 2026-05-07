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
				fmt.Sprintf("%.2f", b.UsedDuration), fmt.Sprintf("%.2f", func() float64 {
					if b.RemainingDuration != nil { return *b.RemainingDuration }; return 0
				}()),
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
				fmt.Sprintf("%.2f", b.UsedDuration), fmt.Sprintf("%.2f", func() float64 {
					if b.RemainingDuration != nil { return *b.RemainingDuration }; return 0
				}()),
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
				req.StartDate, req.EndDate, fmt.Sprintf("%.2f", req.TotalDays),
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
				req.StartDate, req.EndDate, fmt.Sprintf("%.2f", req.TotalDays),
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

// ListEmployeeBalanceSummary — GET /leave-balances/summary
func (h *LeaveHandler) ListEmployeeBalanceSummary(c *fiber.Ctx) error {
	var params dto.EmployeeBalanceSummaryParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}
	res, err := h.service.GetEmployeeBalanceSummary(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "balance summary", Data: res})
}

// ExportEmployeeBalanceSummary — GET /leave-balances/summary/export
func (h *LeaveHandler) ExportEmployeeBalanceSummary(c *fiber.Ctx) error {
	var params dto.EmployeeBalanceSummaryParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}
	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}
	res, err := h.service.ExportEmployeeBalanceSummary(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	headers := []string{"Nama Pegawai", "Departemen", "Jabatan", "Tahun", "Total Jatah", "Total Terpakai", "Sisa"}
	var rows [][]string
	for _, b := range res.Data {
		dept := "-"
		if b.DepartmentName != nil {
			dept = *b.DepartmentName
		}
		job := "-"
		if b.JobPositionTitle != nil {
			job = *b.JobPositionTitle
		}
		rows = append(rows, []string{
			b.EmployeeName, dept, job, fmt.Sprintf("%d", b.Year),
			fmt.Sprintf("%.2f", b.TotalAllocated), fmt.Sprintf("%.2f", b.TotalUsed), fmt.Sprintf("%.2f", b.TotalRemaining),
		})
	}
	switch exportReq.Format {
	case dto.ExportCSV:
		data, err := utils.WriteCSV(headers, rows)
		if err != nil {
			return respondError(c, err)
		}
		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", "attachment; filename=ringkasan_saldo_cuti.csv")
		return c.Send(data)
	case dto.ExportPDF:
		html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
			Title: "Ringkasan Saldo Cuti", Date: time.Now().Format("02 Jan 2006"),
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
		c.Set("Content-Disposition", "attachment; filename=ringkasan_saldo_cuti.pdf")
		return c.Send(pdf)
	default:
		return respondBadRequest(c, "format must be csv or pdf")
	}
}

// GetEmployeeBalanceDetail — GET /leave-balances/employee/:employeeId
func (h *LeaveHandler) GetEmployeeBalanceDetail(c *fiber.Ctx) error {
	eid, err := c.ParamsInt("employeeId")
	if err != nil {
		return respondBadRequest(c, "invalid employee ID")
	}
	year := c.QueryInt("year", 0)
	if year == 0 {
		year = time.Now().Year()
	}
	res, err := h.service.GetEmployeeBalanceDetail(c.Context(), uint(eid), year)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "employee balance detail", Data: res})
}

// UpsertBalance — POST /leave-balances
func (h *LeaveHandler) UpsertBalance(c *fiber.Ctx) error {
	var req dto.UpsertLeaveBalanceRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	if req.EmployeeID == 0 {
		return respondBadRequest(c, "employee_id wajib diisi")
	}
	if req.LeaveTypeID == 0 {
		return respondBadRequest(c, "leave_type_id wajib diisi")
	}
	if req.AllocatedDuration < 0 {
		return respondBadRequest(c, "allocated_duration tidak boleh negatif")
	}
	if req.Year < 2000 || req.Year > 2100 {
		return respondBadRequest(c, "year tidak valid")
	}
	account := getAccountFromCtx(c)
	res, err := h.service.UpsertBalance(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{Status: true, StatusCode: 201, Message: "balance upserted", Data: res})
}

// DeleteBalance — DELETE /leave-balances/:id
func (h *LeaveHandler) DeleteBalance(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid balance ID")
	}
	if err := h.service.DeleteBalance(c.Context(), uint(id)); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "balance deleted"})
}

// AdjustBalance — POST /leave-balances/:id/adjust
func (h *LeaveHandler) AdjustBalance(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid balance ID")
	}
	var req dto.AdjustLeaveBalanceRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	if req.Delta == 0 {
		return respondBadRequest(c, "delta tidak boleh 0")
	}
	account := getAccountFromCtx(c)
	res, err := h.service.AdjustBalance(c.Context(), account.EmployeeID, uint(id), req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "balance adjusted", Data: res})
}

// GetBalanceAdjustments — GET /leave-balances/:id/adjustments
func (h *LeaveHandler) GetBalanceAdjustments(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid balance ID")
	}
	res, err := h.service.GetBalanceAdjustments(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "balance adjustments", Data: res})
}
