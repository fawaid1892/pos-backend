package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
<<<<<<< HEAD
=======
	SQLitePath     string
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	JWTSecret      string
	JWTExpiryHours int
	ServerPort     string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://posuser:pospass@localhost:5432/pos_multi_branch?sslmode=disable"),
<<<<<<< HEAD
=======
		SQLitePath:     getEnv("SQLITE_PATH", "./data/sync.db"),
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
		JWTSecret:      getEnv("JWT_SECRET", "default-secret-change-me"),
		JWTExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
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
