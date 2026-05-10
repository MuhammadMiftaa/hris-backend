package env

import (
	"os"

	"github.com/joho/godotenv"
)

type (
	Server struct {
		Mode      string `env:"MODE"`
		HTTPPort  string `env:"HTTP_PORT"`
		ClientURL string `env:"CLIENT_URL"`
		ByPass    string `env:"BYPASS"`
	}

	Database struct {
		DBHost     string `env:"DB_HOST"`
		DBPort     string `env:"DB_PORT"`
		DBUser     string `env:"DB_USER"`
		DBPassword string `env:"DB_PASSWORD"`
		DBName     string `env:"DB_NAME"`
	}

	Redis struct {
		Host     string `env:"REDIS_HOST"`
		Port     string `env:"REDIS_PORT"`
		Password string `env:"REDIS_PASSWORD"`
		DB       string `env:"REDIS_DB"`
	}

	Minio struct {
		Host      string `env:"MINIO_HOST"`
		Port      string `env:"MINIO_PORT"`
		Secure    bool   `env:"MINIO_SECURE"`
		AccessKey string `env:"MINIO_ROOT_USER"`
		SecretKey string `env:"MINIO_ROOT_PASSWORD"`
		PublicURL string `env:"MINIO_PUBLIC_URL"`
	}

	Gotenberg struct {
		URL string `env:"GOTENBERG_URL"`
	}

	Vapid struct {
		PrivateKey string `env:"VAPID_PRIVATE_KEY"`
		PublicKey  string `env:"VAPID_PUBLIC_KEY"`
	}

	ExternalAPI struct {
		IndonesiaHolidayAPIKey string `env:"INDONESIA_HOLIDAY_API_KEY"`
		IndonesiaHolidayAPIURL string `env:"INDONESIA_HOLIDAY_API_URL"`
	}

	Config struct {
		Server      Server
		Database    Database
		Redis       Redis
		Minio       Minio
		Gotenberg   Gotenberg
		Vapid       Vapid
		ExternalAPI ExternalAPI
	}
)

var Cfg Config

const errEnvNotSet = " env is not set"

func lookupEnv(key string, dest *string, missing *[]string) {
	if val, ok := os.LookupEnv(key); ok {
		*dest = val
	} else {
		*missing = append(*missing, key+errEnvNotSet)
	}
}

func LoadNative() ([]string, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	var missing []string

	lookupEnv("MODE", &Cfg.Server.Mode, &missing)
	lookupEnv("HTTP_PORT", &Cfg.Server.HTTPPort, &missing)
	lookupEnv("CLIENT_URL", &Cfg.Server.ClientURL, &missing)
	lookupEnv("BYPASS", &Cfg.Server.ByPass, &missing)

	lookupEnv("DB_USER", &Cfg.Database.DBUser, &missing)
	lookupEnv("DB_HOST", &Cfg.Database.DBHost, &missing)
	lookupEnv("DB_PORT", &Cfg.Database.DBPort, &missing)
	lookupEnv("DB_NAME", &Cfg.Database.DBName, &missing)
	lookupEnv("DB_PASSWORD", &Cfg.Database.DBPassword, &missing)

	lookupEnv("REDIS_HOST", &Cfg.Redis.Host, &missing)
	lookupEnv("REDIS_PORT", &Cfg.Redis.Port, &missing)
	lookupEnv("REDIS_PASSWORD", &Cfg.Redis.Password, &missing)
	lookupEnv("REDIS_DB", &Cfg.Redis.DB, &missing)

	lookupEnv("MINIO_HOST", &Cfg.Minio.Host, &missing)
	lookupEnv("MINIO_PORT", &Cfg.Minio.Port, &missing)
	if val, ok := os.LookupEnv("MINIO_SECURE"); ok {
		Cfg.Minio.Secure = val == "true"
	} else {
		Cfg.Minio.Secure = false
	}
	lookupEnv("MINIO_ROOT_USER", &Cfg.Minio.AccessKey, &missing)
	lookupEnv("MINIO_ROOT_PASSWORD", &Cfg.Minio.SecretKey, &missing)
	lookupEnv("MINIO_PUBLIC_URL", &Cfg.Minio.PublicURL, &missing)

	lookupEnv("GOTENBERG_URL", &Cfg.Gotenberg.URL, &missing)

	lookupEnv("VAPID_PRIVATE_KEY", &Cfg.Vapid.PrivateKey, &missing)
	// VAPID_PUBLIC_KEY is optional — will be derived from private key if not set
	if val, ok := os.LookupEnv("VAPID_PUBLIC_KEY"); ok {
		Cfg.Vapid.PublicKey = val
	}

	lookupEnv("INDONESIA_HOLIDAY_API_KEY", &Cfg.ExternalAPI.IndonesiaHolidayAPIKey, &missing)
	lookupEnv("INDONESIA_HOLIDAY_API_URL", &Cfg.ExternalAPI.IndonesiaHolidayAPIURL, &missing)

	return missing, nil
}
