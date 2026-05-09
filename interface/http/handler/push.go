package handler

import (
	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"

	"github.com/gofiber/fiber/v2"
)

type PushHandler struct {
	notifSvc service.NotificationService
}

func NewPushHandler(notifSvc service.NotificationService) *PushHandler {
	return &PushHandler{notifSvc: notifSvc}
}

func (h *PushHandler) Subscribe(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)
	var req dto.PushSubscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return respondBadRequest(c, "Format body tidak valid")
	}

	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		return respondBadRequest(c, "Data subscription tidak lengkap")
	}

	if err := h.notifSvc.SaveSubscription(c.Context(), account.EmployeeID, req); err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: fiber.StatusCreated,
		Message:    "Berlangganan notifikasi berhasil",
	})
}

func (h *PushHandler) SubscriptionStatus(c *fiber.Ctx) error {
	account := getAccountFromCtx(c)
	isSubscribed, err := h.notifSvc.GetSubscriptionStatus(c.Context(), account.EmployeeID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: fiber.StatusOK,
		Message:    "Status berlangganan",
		Data: dto.PushSubscriptionResponse{
			IsSubscribed: isSubscribed,
		},
	})
}
