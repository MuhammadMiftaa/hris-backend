package handler

import (
	"strconv"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type NotificationHandler struct {
	notifSvc service.NotificationService
}

func NewNotificationHandler(notifSvc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifSvc: notifSvc}
}

func (h *NotificationHandler) List(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	var params dto.NotificationListParams
	if err := c.QueryParser(&params); err != nil {
		return respondBadRequest(c, "Parameter query tidak valid")
	}

	result, err := h.notifSvc.GetNotifications(c.Context(), account.EmployeeID, params)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: fiber.StatusOK,
		Message:    "Daftar notifikasi",
		Data:       result,
	})
}

func (h *NotificationHandler) UnreadCount(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	count, err := h.notifSvc.GetUnreadCount(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: fiber.StatusOK,
		Message:    "Jumlah notifikasi belum dibaca",
		Data: dto.UnreadCountResponse{
			Count: count,
		},
	})
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return respondBadRequest(c, "ID notifikasi tidak valid")
	}

	if err := h.notifSvc.MarkAsRead(c.Context(), uint(id), account.EmployeeID); err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: fiber.StatusOK,
		Message:    "Notifikasi telah ditandai dibaca",
	})
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)

	if err := h.notifSvc.MarkAllAsRead(c.Context(), account.EmployeeID); err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: fiber.StatusOK,
		Message:    "Semua notifikasi telah ditandai dibaca",
	})
}
