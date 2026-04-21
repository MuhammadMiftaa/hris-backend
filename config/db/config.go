package db

import (
	"fmt"
	"sync"
	"time"

	"hris-backend/config/env"
	"hris-backend/config/log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseClient interface {
	GetDB() *gorm.DB
	Close() error
}

type databaseClient struct {
	db *gorm.DB
	mu sync.RWMutex
}

var (
	instance DatabaseClient
	once     sync.Once
)

type ConnectionPoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func DefaultPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 10,
	}
}

func NewDatabaseClient(cfg env.Database, poolCfg ConnectionPoolConfig) (DatabaseClient, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(poolCfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(poolCfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(poolCfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(poolCfg.ConnMaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &databaseClient{
		db: db,
	}, nil
}

func GetInstance(cfg env.Database, poolCfg ...ConnectionPoolConfig) DatabaseClient {
	once.Do(func() {
		var pool ConnectionPoolConfig
		if len(poolCfg) > 0 {
			pool = poolCfg[0]
		} else {
			pool = DefaultPoolConfig()
		}

		client, err := NewDatabaseClient(cfg, pool)
		if err != nil {
			log.Log.Fatalf("Failed to initialize Database client: %v", err)
		}
		instance = client
	})

	return instance
}

func (d *databaseClient) GetDB() *gorm.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.db
}

func (d *databaseClient) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %w", err)
		}

		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}

		d.db = nil
	}

	return nil
}
