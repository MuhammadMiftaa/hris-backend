package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type AttendanceHandler struct {
	service service.AttendanceService
}

func NewAttendanceHandler(service service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{service: service}
}

// PresignClockPhoto — minta presigned URL untuk upload foto clock in/out
// POST /attendance/presign
func (h *AttendanceHandler) PresignClockPhoto(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var req dto.AttendancePresignRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request")
	}
	if req.Action == "" {
		return respondBadRequest(c, "action is required")
	}

	result, err := h.service.PresignClockPhoto(c.Context(), account.EmployeeID, req.Action)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "presigned URL generated",
		Data:       result,
	})
}

// GetPhotoURL — dapatkan signed download URL untuk foto
// GET /attendance/photo?key=...
func (h *AttendanceHandler) GetPhotoURL(c *fiber.Ctx) error {
	key := c.Query("key")
	if key == "" {
		return respondBadRequest(c, "key is required")
	}

	url, err := h.service.GetPhotoURL(c.Context(), key)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "photo URL",
		Data:       fiber.Map{"url": url},
	})
}

// GetTodayStatus — status presensi hari ini untuk pegawai yang login
// GET /attendance/today
func (h *AttendanceHandler) GetTodayStatus(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	result, err := h.service.GetTodayStatus(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "today attendance status",
		Data:       result,
	})
}

// ClockIn — submit clock in
// POST /attendance/clock-in
func (h *AttendanceHandler) ClockIn(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var req dto.ClockInRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	if req.PhotoKey == "" {
		return respondBadRequest(c, "photo_key is required")
	}
	if req.Latitude == 0 && req.Longitude == 0 {
		return respondBadRequest(c, "latitude dan longitude harus diisi")
	}

	result, err := h.service.ClockIn(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "clock in berhasil",
		Data:       result,
	})
}

// ClockOut — submit clock out
// POST /attendance/clock-out
func (h *AttendanceHandler) ClockOut(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var req dto.ClockOutRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}
	if req.PhotoKey == "" {
		return respondBadRequest(c, "photo_key is required")
	}

	result, err := h.service.ClockOut(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "clock out berhasil",
		Data:       result,
	})
}

// List — admin: daftar semua presensi
// GET /attendance
func (h *AttendanceHandler) List(c *fiber.Ctx) error {
	var params dto.AttendanceListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	// account := getAccountFromCtx(c)
	// if account.RoleLevel == string(model.RoleLevelManager) || account.RoleLevel == string(model.RoleLevelStaff) {
	// 	params.EmployeeID = &account.EmployeeID
	// }

	result, err := h.service.GetAllLogs(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "attendance list",
		Data:       result,
	})
}

// Metadata — GET /attendance/metadata
func (h *AttendanceHandler) Metadata(c *fiber.Ctx) error {
	res, err := h.service.GetMetadata(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "attendance metadata",
		Data:       res,
	})
}

// CreateManual — POST /attendance/manual
func (h *AttendanceHandler) CreateManual(c *fiber.Ctx) error {
	var req dto.CreateManualAttendanceRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.CreateManualAttendance(c.Context(), account.EmployeeID, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "manual attendance created",
		Data:       res,
	})
}

// ListOverrides — GET /attendance-overrides
func (h *AttendanceHandler) ListOverrides(c *fiber.Ctx) error {
	var params dto.OverrideListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, err.Error())
	}

	// account := getAccountFromCtx(c)
	// if account.RoleLevel == string(model.RoleLevelManager) || account.RoleLevel == string(model.RoleLevelStaff) {
	// 	params.EmployeeID = &account.EmployeeID
	// }

	res, err := h.service.GetAllOverrides(c.Context(), params)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "attendance overrides list",
		Data:       res,
	})
}

// DetailOverride — GET /attendance-overrides/:id
func (h *AttendanceHandler) DetailOverride(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid override ID")
	}

	res, err := h.service.GetOverrideByID(c.Context(), uint(id))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "attendance override detail",
		Data:       res,
	})
}

// CreateOverride — POST /attendance-overrides
func (h *AttendanceHandler) CreateOverride(c *fiber.Ctx) error {
	var req dto.CreateOverrideRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.CreateOverride(c.Context(), account.EmployeeID, account.RoleLevel, req)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "attendance override submitted",
		Data:       res,
	})
}

// UpdateOverride — PUT /attendance-overrides/:id
func (h *AttendanceHandler) UpdateOverride(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondBadRequest(c, "invalid override ID")
	}

	var req dto.UpdateOverrideStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "invalid request body")
	}

	account := getAccountFromCtx(c)
	res, err := h.service.UpdateOverrideStatus(c.Context(), account.EmployeeID, account.RoleLevel, uint(id), req)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "attendance override updated",
		Data:       res,
	})
}
