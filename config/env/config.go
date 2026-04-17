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
	}

	Database struct {
		DBHost     string `env:"DB_HOST"`
		DBPort     string `env:"DB_PORT"`
		DBUser     string `env:"DB_USER"`
		DBPassword string `env:"DB_PASSWORD"`
		DBName     string `env:"DB_NAME"`
	}

	Redis struct {
		Address  string `env:"REDIS_ADDRESS"`
		Password string `env:"REDIS_PASSWORD"`
		DB       string `env:"REDIS_DB"`
	}

	Minio struct {
		Host      string `env:"MINIO_HOST"`
		Port      string `env:"MINIO_PORT"`
		AccessKey string `env:"MINIO_ROOT_USER"`
		SecretKey string `env:"MINIO_ROOT_PASSWORD"`
	}

	Config struct {
		Server   Server
		Database Database
		Redis    Redis
		Minio    Minio
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

	lookupEnv("DB_USER", &Cfg.Database.DBUser, &missing)
	lookupEnv("DB_HOST", &Cfg.Database.DBHost, &missing)
	lookupEnv("DB_PORT", &Cfg.Database.DBPort, &missing)
	lookupEnv("DB_NAME", &Cfg.Database.DBName, &missing)
	lookupEnv("DB_PASSWORD", &Cfg.Database.DBPassword, &missing)

	lookupEnv("REDIS_ADDRESS", &Cfg.Redis.Address, &missing)
	Cfg.Redis.Password, _ = os.LookupEnv("REDIS_PASSWORD")
	Cfg.Redis.DB, _ = os.LookupEnv("REDIS_DB")

	lookupEnv("MINIO_HOST", &Cfg.Minio.Host, &missing)
	lookupEnv("MINIO_ROOT_USER", &Cfg.Minio.AccessKey, &missing)
	lookupEnv("MINIO_ROOT_PASSWORD", &Cfg.Minio.SecretKey, &missing)
	if port, ok := os.LookupEnv("MINIO_PORT"); ok {
		Cfg.Minio.Port = port
	} else {
		Cfg.Minio.Port = "9000"
	}

	return missing, nil
}
