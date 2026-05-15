package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hris-backend/config/db"
	"hris-backend/config/env"
	logger "hris-backend/config/log"
	redisSetup "hris-backend/config/redis"
	"hris-backend/config/storage"
	"hris-backend/interface/http/router"
	"hris-backend/internal/cron"
	"hris-backend/internal/redis"
	"hris-backend/internal/repository"
	"hris-backend/internal/service"
	"hris-backend/internal/utils"
	"hris-backend/internal/utils/data"
)

func init() {
	var err error
	var missing []string

	if missing, err = env.LoadNative(); err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}
	log.Printf("Environment variables by .env file loaded successfully")

	logger.SetupLogger()

	if len(missing) > 0 {
		for _, envVar := range missing {
			logger.Warn(data.LogEnvVarMissing, map[string]any{
				"service": data.EnvService,
				"env_var": envVar,
			})
		}
	}
}

func shutdownHTTP(app interface{ ShutdownWithTimeout(time.Duration) error }, errors map[string]any) {
	if app == nil {
		return
	}
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		logger.Error(data.LogHTTPServerShutdownFailed, map[string]any{
			"service": data.HTTPServerService,
			"error":   err.Error(),
		})
		errors["http_error"] = true
	}
}

func shutdownRedis(redisInstance redis.Redis, errors map[string]any) {
	if redisInstance == nil {
		return
	}
	if err := redisInstance.Close(); err != nil {
		logger.Error(data.LogRedisCloseFailed, map[string]any{
			"service": data.RedisService,
			"error":   err.Error(),
		})
		errors["cache_error"] = true
	}
}

func shutdownInfra(dbInstance db.DatabaseClient, errors map[string]any) {
	if err := dbInstance.Close(); err != nil {
		logger.Error(data.LogDBCloseFailed, map[string]any{
			"service": data.DatabaseService,
			"error":   err.Error(),
		})
		errors["database_error"] = true
	}
}

func main() {
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

	// ── MinIO ──────────────────────────────────────────────────────
	startTime = time.Now()
	minioClient, err := storage.NewMinioClient(env.Cfg.Minio)
	if err != nil {
		logger.Fatal("minio_setup_failed", map[string]any{
			"service": "minio",
			"error":   err.Error(),
		})
	}
	// Pastikan bucket sudah ada
	if err := minioClient.EnsureBuckets(context.Background()); err != nil {
		logger.Fatal("minio_bucket_setup_failed", map[string]any{
			"service": "minio",
			"error":   err.Error(),
		})
	}
	logger.Info("minio_setup_success", map[string]any{
		"service":  "minio",
		"host":     env.Cfg.Minio.Host,
		"duration": utils.Ms(time.Since(startTime)),
	})

	// ── Repositories ─────────────────────────────────────────────
	attendanceRepo := repository.NewAttendanceRepository(dbInstance.GetDB())

	// ── Notification Service ─────────────────────────────────────
	pushRepo := repository.NewPushRepository(dbInstance.GetDB())
	notifRepo := repository.NewNotificationRepository(dbInstance.GetDB())
	empRepo := repository.NewEmployeeRepository(dbInstance.GetDB())
	pushSvc := service.NewPushService(
		env.Cfg.Vapid.PrivateKey,
		env.Cfg.Vapid.PublicKey,
		"mailto:wafa@example.com",
	)
	notifSvc := service.NewNotificationService(pushRepo, notifRepo, pushSvc, empRepo, attendanceRepo)
	logger.Info("push service initialized", map[string]any{
		"service":     "push_service",
		"private_key": env.Cfg.Vapid.PrivateKey,
		"public_key":  env.Cfg.Vapid.PublicKey,
	})

	// ── Cron Scheduler ────────────────────────────────────────────
	mutabaahRepo := repository.NewMutabaahRepository(dbInstance.GetDB())
	dailyReportRepo := repository.NewDailyReportRepository(dbInstance.GetDB())
	shiftRepo := repository.NewShiftRepository(dbInstance.GetDB())
	txManager := repository.NewTxManager(dbInstance.GetDB())
	cronSvc := service.NewCronService(attendanceRepo, mutabaahRepo, dailyReportRepo, shiftRepo, txManager, notifSvc)
	scheduler := cron.NewScheduler(cronSvc)
	scheduler.Start()

	// ── HTTP Server ────────────────────────────────────────────────
	startTime = time.Now()
	httpServer := router.SetupHTTPServer(dbInstance, redisInstance, minioClient, notifSvc)
	if httpServer != nil {
		go func() {
			if err := httpServer.Listen(":" + env.Cfg.Server.HTTPPort); err != nil && err != http.ErrServerClosed {
				logger.Fatal(data.LogHTTPServerStartFailed, map[string]any{
					"service": data.HTTPServerService,
					"error":   err.Error(),
				})
			}
		}()
		logger.Info(data.LogHTTPServerStarted, map[string]any{
			"service":  data.HTTPServerService,
			"port":     env.Cfg.Server.HTTPPort,
			"duration": utils.Ms(time.Since(startTime)),
		})
	}

	// ── Wait for shutdown signal ───────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(data.LogShutdownSignalReceived, map[string]any{
		"service": data.MainService,
	})

	startTime = time.Now()
	shutdownErrors := map[string]any{"service": data.MainService}

	scheduler.Stop()

	shutdownHTTP(httpServer, shutdownErrors)
	shutdownRedis(redisInstance, shutdownErrors)
	shutdownInfra(dbInstance, shutdownErrors)

	if len(shutdownErrors) > 1 {
		logger.Info(data.LogShutdownCompletedWithErrors, shutdownErrors)
	} else {
		logger.Info(data.LogShutdownCompleted, map[string]any{
			"service":  data.MainService,
			"duration": utils.Ms(time.Since(startTime)),
		})
	}
}
