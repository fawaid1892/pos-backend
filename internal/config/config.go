package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL           string
	SQLitePath            string
	JWTSecret             string
	JWTExpiryHours        int
	JWTRefreshExpiryHours int
	ServerPort            string
	ElectricURL           string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://localhost:5432/pos_multi_branch?sslmode=disable"),
		SQLitePath:     getEnv("SQLITE_PATH", "./data/sync.db"),
		JWTSecret:      getEnv("JWT_SECRET", "default-secret-change-me"),
		JWTExpiryHours:        getEnvInt("JWT_EXPIRY_HOURS", 24),
		JWTRefreshExpiryHours: getEnvInt("JWT_REFRESH_EXPIRY_HOURS", 720),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		ElectricURL:    getEnv("ELECTRIC_URL", "https://api.electric-sql.cloud/v1/sources/svc-itchy-porcupine-kbtcobz2v0/shapes"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
