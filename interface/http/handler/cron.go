package handler

import (
	"errors"

	"hris-backend/internal/service"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type CronHandler struct {
	service service.CronService
}

func NewCronHandler(service service.CronService) *CronHandler {
	return &CronHandler{service: service}
}

// TriggerAbsentMark — trigger absent mark manual (internal/ops)
// POST /internal/cron/absent-mark
func (h *CronHandler) TriggerAbsentMark(c *fiber.Ctx) error {
	date := c.Query("date", utils.TodayDate())

	if err := h.service.RunDailyAbsentMark(c.Context(), date); err != nil {
		return respondError(c, errors.New("absent mark failed: "+err.Error()))
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "absent mark completed for " + date,
	})
}

// TriggerMutabaahMark — trigger mutabaah missing mark manual (internal/ops)
// POST /internal/cron/mutabaah-mark
func (h *CronHandler) TriggerMutabaahMark(c *fiber.Ctx) error {
	date := c.Query("date", utils.TodayDate())

	if err := h.service.RunDailyMutabaahMark(c.Context(), date); err != nil {
		return respondError(c, errors.New("mutabaah mark failed: "+err.Error()))
	}
	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "mutabaah mark completed for " + date,
	})
}
