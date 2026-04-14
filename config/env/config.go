package env

import (
	"os"

	"github.com/joho/godotenv"
)

type (
	Server struct {
		Mode      string `env:"MODE"`
		HTTPPort  string `env:"HTTP_PORT"`
		JWTSecret string `env:"JWT_SECRET"`
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

	Config struct {
		Server   Server
		Database Database
		Redis    Redis
	}
)

var Cfg Config

const errEnvNotSet = " env is not set"

// lookupEnv reads an OS environment variable; if missing it appends a message to missing.
func lookupEnv(key string, dest *string, missing *[]string) {
	if val, ok := os.LookupEnv(key); ok {
		*dest = val
	} else {
		*missing = append(*missing, key+errEnvNotSet)
	}
}

// LoadNative loads configuration from OS environment variables (or a .env file).
func LoadNative() ([]string, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	var missing []string

	lookupEnv("MODE", &Cfg.Server.Mode, &missing)
	lookupEnv("HTTP_PORT", &Cfg.Server.HTTPPort, &missing)
	lookupEnv("JWT_SECRET", &Cfg.Server.JWTSecret, &missing)
	lookupEnv("CLIENT_URL", &Cfg.Server.ClientURL, &missing)

	lookupEnv("DB_USER", &Cfg.Database.DBUser, &missing)
	lookupEnv("DB_HOST", &Cfg.Database.DBHost, &missing)
	lookupEnv("DB_PORT", &Cfg.Database.DBPort, &missing)
	lookupEnv("DB_NAME", &Cfg.Database.DBName, &missing)
	lookupEnv("DB_PASSWORD", &Cfg.Database.DBPassword, &missing)

	lookupEnv("REDIS_ADDRESS", &Cfg.Redis.Address, &missing)
	Cfg.Redis.Password, _ = os.LookupEnv("REDIS_PASSWORD")
	Cfg.Redis.DB, _ = os.LookupEnv("REDIS_DB")

	return missing, nil
}
