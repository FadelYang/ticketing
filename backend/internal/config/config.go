package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv      string
	AppPort     string
	DatabaseURL string
	CORSOrigins string
}

func Load() *Config {
	return &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnv("APP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://appuser:apppassword@localhost:5432/appdb?sslmode=disable"),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
