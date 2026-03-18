package config

import "os"

type Config struct {
	Env           string
	Port          string
	StoreMode     string
	MySQLDSN      string
	AllowedOrigin string
}

func Load() Config {
	return Config{
		Env:           envOrDefault("APP_ENV", "development"),
		Port:          envOrDefault("APP_PORT", "8080"),
		StoreMode:     envOrDefault("APP_STORE_MODE", "memory"),
		MySQLDSN:      envOrDefault("APP_MYSQL_DSN", ""),
		AllowedOrigin: envOrDefault("APP_ALLOWED_ORIGIN", "*"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
