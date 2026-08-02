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
	Env           string // "development" | "production"
	Port          string
	DatabaseURL   string
	SessionSecret string // signs auth session cookies

	// AI configuration - uses OpenAI provider exclusively
	AIAPIKey string
	// Email configuration
	EmailProvider string // "dev" | "smtp"
	EmailFrom     string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:           getEnv("APP_ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		AIAPIKey:      getEnv("AI_API_KEY", ""),
		EmailProvider: getEnv("EMAIL_PROVIDER", "dev"),
		EmailFrom:     getEnv("EMAIL_FROM", "no-reply@jobhoo.local"),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPass:      getEnv("SMTP_PASS", ""),
	}

	// Support Mailtrap-style env vars for developer testing. If MAIL_HOST is
	// present and explicit SMTP_* vars are not set, copy them across so the
	// app will use SMTP (Mailtrap) even in development.
	if cfg.SMTPHost == "" {
		if mailHost := getEnv("MAIL_HOST", ""); mailHost != "" {
			cfg.SMTPHost = mailHost
			cfg.SMTPPort = getEnv("MAIL_PORT", cfg.SMTPPort)
			cfg.SMTPUser = getEnv("MAIL_USER", cfg.SMTPUser)
			cfg.SMTPPass = getEnv("MAIL_PASS", cfg.SMTPPass)
		}
	}

	// If developer selected EMAIL_PROVIDER=dev but SMTP/Mailtrap creds exist,
	// prefer SMTP so emails are delivered to the Mailtrap inbox for testing.
	if cfg.EmailProvider == "dev" && cfg.SMTPHost != "" {
		cfg.EmailProvider = "smtp"
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
