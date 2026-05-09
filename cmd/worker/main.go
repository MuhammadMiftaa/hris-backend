package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"hris-backend/config/db"
	"hris-backend/config/env"
	logger "hris-backend/config/log"
	redisSetup "hris-backend/config/redis"
	"hris-backend/internal/redis"
	"hris-backend/internal/repository"
	"hris-backend/internal/service"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"
	"hris-backend/internal/utils/data"
)

func init() {
	var err error
	var missing []string

	if missing, err = env.LoadNative(); err != nil {
		logger.Fatal("failed to load env", map[string]any{"error": err.Error()})
	}
	if len(missing) > 0 {
		for _, envVar := range missing {
			logger.Warn(data.LogEnvVarMissing, map[string]any{
				"service": data.EnvService,
				"env_var": envVar,
			})
		}
	}
	logger.SetupLogger()
}

func main() {
	ctx := context.Background()

	// ── Database ───────────────────────────────────────────────────
	startTime := time.Now()
	dbInstance := db.GetInstance(env.Cfg.Database)
	logger.Info(data.LogDBSetupSuccess, map[string]any{
		"service":  data.DatabaseService,
		"duration": utils.Ms(time.Since(startTime)),
	})

	// ── Redis ──────────────────────────────────────────────────────
	startTime = time.Now()
	redisClient, err := redisSetup.NewRedisClient(redisSetup.RedisConfig{
		Host:     env.Cfg.Redis.Host,
		Port:     env.Cfg.Redis.Port,
		Password: env.Cfg.Redis.Password,
		DB:       redisSetup.ParseRedisDB(env.Cfg.Redis.DB),
	})
	if err != nil {
		logger.Fatal(data.LogRedisSetupFailed, map[string]any{
			"service": data.RedisService,
			"error":   err.Error(),
		})
	}
	redisInstance := redis.NewRedisInstance(redisClient)
	logger.Info(data.LogRedisSetupSuccess, map[string]any{
		"service":  data.RedisService,
		"duration": utils.Ms(time.Since(startTime)),
	})

	// ── Services ───────────────────────────────────────────────────
	pushRepo := repository.NewPushRepository(dbInstance.GetDB())
	notifRepo := repository.NewNotificationRepository(dbInstance.GetDB())

	pushSvc := service.NewPushService(
		env.Cfg.Vapid.PrivateKey,
		env.Cfg.Vapid.PublicKey,
		"mailto:wafa@example.com",
	)

	// ── Worker Loop ────────────────────────────────────────────────
	logger.Info("worker: push notification worker started", map[string]any{
		"service": "push_worker",
	})

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			if err := processPendingNotifications(ctx, redisInstance, notifRepo, pushRepo, pushSvc); err != nil {
				logger.Error("worker: process notifications failed", map[string]any{
					"service": "push_worker",
					"error":   err.Error(),
				})
			}
		case <-quit:
			logger.Info("worker: shutting down gracefully", map[string]any{
				"service": "push_worker",
			})
			return
		}
	}
}

func processPendingNotifications(
	ctx context.Context,
	redisInstance redis.Redis,
	notifRepo repository.NotificationRepository,
	pushRepo repository.PushRepository,
	pushSvc service.PushService,
) error {
	// Scan all keys matching push:notify:*
	keys, err := redisInstance.Scan(ctx, "push:notify:*", 100)
	if err != nil {
		return fmt.Errorf("scan redis keys: %w", err)
	}

	now := time.Now().Unix()

	for _, key := range keys {
		// Extract notification ID from key
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			continue
		}
		notifID, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			continue
		}

		// Get stored timestamp
		value, err := redisInstance.Get(ctx, key)
		if err != nil {
			continue
		}

		scheduledAt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}

		// Check if it's time to send
		if scheduledAt > now {
			continue
		}

		// Fetch notification from DB
		notif, err := notifRepo.GetByID(ctx, nil, uint(notifID))
		if err != nil {
			_ = redisInstance.Delete(ctx, key)
			continue
		}

		// Skip if already sent or failed permanently
		if notif.PushStatus == "sent" || notif.PushAttempts >= 3 {
			_ = redisInstance.Delete(ctx, key)
			continue
		}

		// Check conditional logic for clock reminders
		if shouldSkipNotification(ctx, notifRepo, notif) {
			_ = redisInstance.Delete(ctx, key)
			_ = notifRepo.UpdatePushStatus(ctx, nil, notif.ID, "skipped")
			continue
		}

		// Get subscriptions for employee
		subs, err := pushRepo.GetActiveSubscriptionsByEmployeeID(ctx, nil, notif.EmployeeID)
		if err != nil {
			logger.Error("worker: get subscriptions failed", map[string]any{
				"notification_id": notif.ID,
				"error":           err.Error(),
			})
			handleRetry(ctx, redisInstance, notifRepo, key, notif.ID)
			continue
		}

		if len(subs) == 0 {
			// No subscriptions — mark as failed (can't deliver)
			_ = redisInstance.Delete(ctx, key)
			_ = notifRepo.UpdatePushStatus(ctx, nil, notif.ID, "no_subscription")
			continue
		}

		// Build payload
		actionURL := ""
		if notif.ActionURL != nil {
			actionURL = *notif.ActionURL
		}
		actionTab := ""
		if notif.ActionTab != nil {
			actionTab = *notif.ActionTab
		}
		payload := service.BuildPushPayload(notif.Title, notif.Body, actionURL, actionTab, notif.ID)

		// Send to all subscriptions
		anySuccess := false
		for _, sub := range subs {
			if err := pushSvc.SendPush(ctx, sub, payload); err != nil {
				logger.Warn("worker: push send failed", map[string]any{
					"notification_id": notif.ID,
					"endpoint":        sub.Endpoint,
					"error":           err.Error(),
				})
				// If subscription expired, deactivate it
				if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "410") {
					_ = pushRepo.DeactivateSubscription(ctx, nil, sub.Endpoint)
				}
			} else {
				anySuccess = true
			}
		}

		if anySuccess {
			_ = redisInstance.Delete(ctx, key)
			_ = notifRepo.UpdatePushStatus(ctx, nil, notif.ID, "sent")
		} else {
			handleRetry(ctx, redisInstance, notifRepo, key, notif.ID)
		}
	}

	return nil
}

func handleRetry(
	ctx context.Context,
	redisInstance redis.Redis,
	notifRepo repository.NotificationRepository,
	key string,
	notifID uint,
) {
	if err := notifRepo.IncrementPushAttempts(ctx, nil, notifID); err != nil {
		logger.Error("worker: increment attempts failed", map[string]any{
			"notification_id": notifID,
			"error":           err.Error(),
		})
	}

	// Check max attempts
	notif, err := notifRepo.GetByID(ctx, nil, notifID)
	if err != nil {
		_ = redisInstance.Delete(ctx, key)
		return
	}

	if notif.PushAttempts >= 3 {
		_ = redisInstance.Delete(ctx, key)
		_ = notifRepo.UpdatePushStatus(ctx, nil, notifID, "failed")
		logger.Warn("worker: max attempts reached", map[string]any{
			"notification_id": notifID,
		})
	} else {
		// Reschedule for 2 minutes later
		newTime := time.Now().Add(2 * time.Minute).Unix()
		_ = redisInstance.Set(ctx, key, fmt.Sprintf("%d", newTime), 0)
		logger.Info("worker: retry scheduled", map[string]any{
			"notification_id": notifID,
			"attempt":         notif.PushAttempts,
			"retry_at":        newTime,
		})
	}
}

func shouldSkipNotification(ctx context.Context, notifRepo repository.NotificationRepository, notif *model.Notification) bool {
	// Clock in reminder: skip if already clocked in
	if notif.Type == "clock_in_reminder" {
		// This would require attendance repository — for now we always send
		// Full implementation requires injecting attendance repo into worker
		return false
	}
	// Clock out reminder: skip if not clocked in or already clocked out
	if notif.Type == "clock_out_reminder" {
		return false
	}
	return false
}
