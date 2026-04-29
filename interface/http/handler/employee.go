package handler

import (
	"fmt"
	"time"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type EmployeeHandler struct {
	service service.EmployeeService
}

func NewEmployeeHandler(service service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{
		service: service,
	}
}

func (h *EmployeeHandler) Metadata(c *fiber.Ctx) error {
	result, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee metadata",
		Data:       result,
	})
}

func (h *EmployeeHandler) List(c *fiber.Ctx) error {
	var params dto.EmployeeListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.GetAllEmployees(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee list",
		Data:       result,
	})
}

func (h *EmployeeHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetEmployeeByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee detail",
		Data:       result,
	})
}

func (h *EmployeeHandler) Export(c *fiber.Ctx) error {
	var params dto.EmployeeListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	var exportReq dto.ExportRequest
	if err := c.QueryParser(&exportReq); err != nil {
		return respondBadRequest(c, err.Error())
	}

	allPerPage := 0
	params.PerPage = &allPerPage

	result, err := h.service.GetAllEmployees(c.Context(), params)
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

func (h *EmployeeHandler) exportCSV(c *fiber.Ctx, employees []dto.Employee) error {
	headers := []string{"NIP", "Nama Lengkap", "Status", "Departemen", "Jabatan", "Cabang"}
	var rows [][]string
	for _, e := range employees {
		deptName := "-"
		if e.DepartmentName != nil {
			deptName = *e.DepartmentName
		}
		posName := "-"
		if e.JobPositionTitle != nil {
			posName = *e.JobPositionTitle
		}
		branch := "-"
		if e.BranchName != nil {
			branch = *e.BranchName
		}
		status := "Non-Active"
		if e.IsActive {
			status = "Active"
		}

		rows = append(rows, []string{
			e.EmployeeNumber, e.FullName, status, deptName, posName, branch,
		})
	}
	data, err := utils.WriteCSV(headers, rows)
	if err != nil {
		return respondError(c, err)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=pegawai.csv")
	return c.Send(data)
}

func (h *EmployeeHandler) exportPDF(c *fiber.Ctx, employees []dto.Employee) error {
	headers := []string{"NIP", "Nama Lengkap", "Status", "Departemen", "Jabatan", "Cabang"}
	var rows [][]string
	for _, e := range employees {
		deptName := "-"
		if e.DepartmentName != nil {
			deptName = *e.DepartmentName
		}
		posName := "-"
		if e.JobPositionTitle != nil {
			posName = *e.JobPositionTitle
		}
		branch := "-"
		if e.BranchName != nil {
			branch = *e.BranchName
		}
		status := "Non-Active"
		if e.IsActive {
			status = "Active"
		}

		rows = append(rows, []string{
			e.EmployeeNumber, e.FullName, status, deptName, posName, branch,
		})
	}
	html, err := utils.RenderPDFHTML(utils.PDFTemplateData{
		Title: "Daftar Pegawai", Date: time.Now().Format("02 Jan 2006"),
		Headers: headers, Rows: rows, TotalData: len(employees),
	})
	if err != nil {
		return respondError(c, err)
	}
	pdf, err := utils.GeneratePDF(html)
	if err != nil {
		return respondError(c, fmt.Errorf("gagal generate PDF: %w", err))
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=pegawai.pdf")
	return c.Send(pdf)
}

func (h *EmployeeHandler) Create(c *fiber.Ctx) error {
	var input dto.CreateEmployeeRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	newEmployee, newCredentials, err := h.service.CreateEmployee(c.Context(), input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Employee created",
		Data: map[string]any{
			"employee":    newEmployee,
			"credentials": newCredentials,
		},
	})
}

func (h *EmployeeHandler) Update(c *fiber.Ctx) error {
	var input dto.UpdateEmployeeRequest
	if err := c.BodyParser(&input); err != nil {
		return respondBadRequest(c, err.Error())
	}

	id := c.Params("id")
	result, err := h.service.UpdateEmployee(c.Context(), id, input)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Employee updated",
		Data:       result,
	})
}

func (h *EmployeeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteEmployee(c.Context(), id); err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Employee successfully deleted"})
}

func (h *EmployeeHandler) ResetPassword(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return respondBadRequest(c, "Employee ID is required")
	}

	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "Invalid Request Payload")
	}

	if err := h.service.ResetPassword(c.Context(), id, req); err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Password successfully reset"})
}

// Contacts
func (h *EmployeeHandler) ListContacts(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetContactsByEmployeeID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contact list", Data: result})
}

func (h *EmployeeHandler) CreateContact(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.CreateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateContact(c.Context(), id, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{Status: true, StatusCode: 201, Message: "Contact created", Data: result})
}

func (h *EmployeeHandler) UpdateContact(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.UpdateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.UpdateContact(c.Context(), id, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contact updated", Data: result})
}

func (h *EmployeeHandler) DeleteContact(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteContact(c.Context(), id); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contact deleted"})
}

// Contracts
func (h *EmployeeHandler) ListContracts(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.service.GetContractsByEmployeeID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contract list", Data: result})
}

func (h *EmployeeHandler) CreateContract(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.CreateContractRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.CreateContract(c.Context(), id, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{Status: true, StatusCode: 201, Message: "Contract created", Data: result})
}

func (h *EmployeeHandler) UpdateContract(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.UpdateContractRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, err.Error())
	}

	result, err := h.service.UpdateContract(c.Context(), id, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contract updated", Data: result})
}

func (h *EmployeeHandler) DeleteContract(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteContract(c.Context(), id); err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{Status: true, StatusCode: 200, Message: "Contract deleted"})
}
