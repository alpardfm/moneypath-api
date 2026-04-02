package config

import (
	"fmt"
	"os"
)

// Config contains the runtime configuration loaded from environment variables.
type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
}

// Load reads the application configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
