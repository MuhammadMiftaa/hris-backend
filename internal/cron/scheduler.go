package cron

import (
	"context"
	"time"

	logger "hris-backend/config/log"
	"hris-backend/internal/service"
	"hris-backend/internal/utils"
)

// Scheduler menjalankan cron jobs secara periodik
type Scheduler struct {
	cronSvc service.CronService
	quit    chan struct{}
}

func NewScheduler(cronSvc service.CronService) *Scheduler {
	return &Scheduler{
		cronSvc: cronSvc,
		quit:    make(chan struct{}),
	}
}

// Start — mulai scheduler di goroutine terpisah
func (s *Scheduler) Start() {
	go s.runDailyJobs()
	go s.runHourlyJobs()
	go s.runWeeklyJobs()
	go s.runPushSender()
	logger.Info("cron: scheduler started")
}

// Stop — hentikan scheduler dengan graceful
func (s *Scheduler) Stop() {
	close(s.quit)
	logger.Info("cron: scheduler stopped")
}

// ============================================================================
// DAILY JOBS (23:50 WIB)
// ============================================================================

func (s *Scheduler) runDailyJobs() {
	now := utils.NowWIB()
	next := nextRunTime(now, 23, 50)

	timer := time.NewTimer(next.Sub(now))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			s.runJobs()
			now = utils.NowWIB()
			next = nextRunTime(now, 23, 50)
			timer.Reset(next.Sub(now))

		case <-s.quit:
			return
		}
	}
}

func (s *Scheduler) runJobs() {
	ctx := context.Background()
	today := utils.TodayDate()

	logger.Info("cron: running daily jobs", map[string]any{"date": today})

	// 1. Tandai absent
	if err := s.cronSvc.RunDailyAbsentMark(ctx, today); err != nil {
		logger.Error("cron: absent mark failed", map[string]any{
			"date":  today,
			"error": err.Error(),
		})
	}

	// 2. Tunggu sebentar, lalu tandai mutabaah missing
	time.Sleep(5 * time.Second)

	if err := s.cronSvc.RunDailyMutabaahMark(ctx, today); err != nil {
		logger.Error("cron: mutabaah mark failed", map[string]any{
			"date":  today,
			"error": err.Error(),
		})
	}

	// 3. Tunggu sebentar, lalu tandai daily report missing
	time.Sleep(5 * time.Second)

	if err := s.cronSvc.RunDailyReportMark(ctx, today); err != nil {
		logger.Error("cron: daily report mark failed", map[string]any{
			"date":  today,
			"error": err.Error(),
		})
	}

	// 4. Generate tomorrow's push reminders
	time.Sleep(5 * time.Second)

	if err := s.cronSvc.GenerateDailyPushReminders(ctx, today); err != nil {
		logger.Error("cron: generate push reminders failed", map[string]any{
			"date":  today,
			"error": err.Error(),
		})
	}
}

// ============================================================================
// WEEKLY JOBS (Monday 00:01 WIB)
// ============================================================================

func (s *Scheduler) runWeeklyJobs() {
	now := utils.NowWIB()
	next := nextRunTimeWeekly(now, time.Monday, 0, 1)

	timer := time.NewTimer(next.Sub(now))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			ctx := context.Background()
			logger.Info("cron: running weekly jobs", map[string]any{"time": utils.NowWIB()})
			if err := s.cronSvc.RunWeeklyPrayerTimeSync(ctx); err != nil {
				logger.Error("cron: weekly prayer time sync failed", map[string]any{
					"error": err.Error(),
				})
			}
			now = utils.NowWIB()
			next = nextRunTimeWeekly(now, time.Monday, 0, 1)
			timer.Reset(next.Sub(now))

		case <-s.quit:
			return
		}
	}
}

// ============================================================================
// HOURLY JOBS (12:00 & 18:00 WIB)
// ============================================================================

func (s *Scheduler) runHourlyJobs() {
	// Check every minute for exact times
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkHourlyJobs()
		case <-s.quit:
			return
		}
	}
}

func (s *Scheduler) checkHourlyJobs() {
	now := utils.NowWIB()
	today := utils.TodayDate()

	// 12:00 WIB — Mutabaah reminder (first)
	if now.Hour() == 12 && now.Minute() == 0 {
		ctx := context.Background()
		if err := s.cronSvc.SendMutabaahReminders(ctx, today); err != nil {
			logger.Error("cron: mutabaah reminder (12:00) failed", map[string]any{
				"date":  today,
				"error": err.Error(),
			})
		}
	}

	// 18:00 WIB — Mutabaah reminder (second)
	if now.Hour() == 18 && now.Minute() == 0 {
		ctx := context.Background()
		if err := s.cronSvc.SendMutabaahReminders(ctx, today); err != nil {
			logger.Error("cron: mutabaah reminder (18:00) failed", map[string]any{
				"date":  today,
				"error": err.Error(),
			})
		}
	}
}

// ============================================================================
// PUSH SENDER (every 1 minute)
// ============================================================================

func (s *Scheduler) runPushSender() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := s.cronSvc.SendPendingNotifications(ctx); err != nil {
				logger.Error("cron: send pending notifications failed", map[string]any{
					"error": err.Error(),
				})
			}
		case <-s.quit:
			return
		}
	}
}

// nextRunTime menghitung waktu berikutnya pada jam:menit yang ditentukan
// Jika sekarang sudah lewat, maka besok
func nextRunTime(now time.Time, hour, minute int) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if now.After(next) || now.Equal(next) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// nextRunTimeWeekly menghitung waktu berikutnya pada hari dan jam:menit yang ditentukan
func nextRunTimeWeekly(now time.Time, weekday time.Weekday, hour, minute int) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	// Adjust next until it matches the correct weekday and is in the future
	for next.Weekday() != weekday || now.After(next) || now.Equal(next) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
