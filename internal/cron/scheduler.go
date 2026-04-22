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
	go s.run()
	logger.Info("cron: scheduler started")
}

// Stop — hentikan scheduler dengan graceful
func (s *Scheduler) Stop() {
	close(s.quit)
	logger.Info("cron: scheduler stopped")
}

func (s *Scheduler) run() {
	// Pertama kali: hitung waktu ke pukul 23:50 berikutnya
	now := utils.NowWIB()
	next := nextRunTime(now, 23, 50)

	timer := time.NewTimer(next.Sub(now))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			s.runJobs()
			// Reset timer ke 23:50 besok
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
