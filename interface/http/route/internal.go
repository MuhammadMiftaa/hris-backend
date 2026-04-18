package route

import (
	"hris-backend/interface/http/handler"
	"hris-backend/internal/repository"
	"hris-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// InternalRoutes — route untuk operasi internal (cron trigger, ops tooling)
// Sebaiknya dilindungi network-level (tidak expose ke publik) atau
// tambahkan middleware secret key jika perlu
func InternalRoutes(app *fiber.App, db *gorm.DB) {
	attendRepo := repository.NewAttendanceRepository(db)
	mutaRepo := repository.NewMutabaahRepository(db)
	dailyRepo := repository.NewDailyReportRepository(db)
	txManager := repository.NewTxManager(db)
	cronSvc := service.NewCronService(attendRepo, mutaRepo, dailyRepo, txManager)
	cronH := handler.NewCronHandler(cronSvc)

	internal := app.Group("/internal")
	{
		// cron := internal.Group("/cron", func(c *fiber.Ctx) error {
		// 	secret := c.Get("X-Cron-Secret")
		// 	if secret == "" || secret != os.Getenv("CRON_SECRET_KEY") {
		// 		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized cron request"})
		// 	}
		// 	return c.Next()
		// })
		cron := internal.Group("/cron")
		cron.Post("/absent-mark", cronH.TriggerAbsentMark)
		cron.Post("/mutabaah-mark", cronH.TriggerMutabaahMark)
	}
}
