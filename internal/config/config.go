// Package config centralizes all environment-driven configuration for JOBHOO.
// Nothing outside this package should call os.Getenv directly — every setting
// the app depends on is declared here so the full configuration surface is
// visible in one place.
package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env         string // "development" | "production"
	Port        string
	DatabaseURL string

	// SessionSecret signs auth session cookies.
	SessionSecret string

	// AIProvider selects which implementation of ai.Provider is wired up at
	// boot (see internal/ai). Swapping providers never touches application
	// code — only this value changes.
	AIProvider string
	AIAPIKey   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:           getEnv("APP_ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		AIProvider:    getEnv("AI_PROVIDER", "mock"),
		AIAPIKey:      getEnv("AI_API_KEY", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.SessionSecret == "" && cfg.Env == "production" {
		return nil, fmt.Errorf("SESSION_SECRET is required in production")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
