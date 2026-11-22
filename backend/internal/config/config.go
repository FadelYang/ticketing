package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv       string
	AppPort      string
	DatabaseURL  string
	CORSOrigins  string
	JWTSecret    string
	JWTExpiresIn time.Duration
}

func Load() *Config {
	return &Config{
		AppEnv:       getEnv("APP_ENV", "development"),
		AppPort:      getEnv("APP_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://appuser:apppassword@localhost:5432/appdb?sslmode=disable"),
		CORSOrigins:  getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiresIn: time.Duration(getEnvInt("JWT_EXPIRATION_MINUTES", 30)) * time.Minute,
	}
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}
