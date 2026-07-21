package config

import (
	"fmt"
	"os"

	"github.com/Harkaso/gauthly/internal/tenant"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const (
	defaultPort     = ":8080"
	defaultTenantID = "00000000-0000-4000-8000-000000000000"
)

// Config holds the settings the service needs to start.
type Config struct {
	DatabaseURL     string
	LogFormat       string
	Port            string
	DefaultTenantID uuid.UUID
	TenantMode      tenant.Mode
}

// Load reads the configuration from the environment. It fails if a required
// variable is missing or if a value cannot be parsed. Optional values fall
// back to their default.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	var err error

	cfg.DatabaseURL, err = requireString("DB_URL")
	if err != nil {
		return nil, fmt.Errorf("require DB_URL: %w", err)
	}

	cfg.LogFormat = lookupString("LOG_FORMAT", "text")

	cfg.Port = lookupString("PORT", defaultPort)

	cfg.DefaultTenantID, err = uuid.Parse(lookupString("DEFAULT_TENANT_ID", defaultTenantID))
	if err != nil {
		return nil, fmt.Errorf("parse DEFAULT_TENANT_ID: %w", err)
	}

	cfg.TenantMode, err = tenant.ParseMode(lookupString("TENANT_MODE", "B2B"))
	if err != nil {
		return nil, fmt.Errorf("parse TENANT_MODE: %w", err)
	}

	return cfg, nil
}

func requireString(key string) (string, error) {
	val := os.Getenv(key)

	if val == "" {
		return "", fmt.Errorf("required environment variable %q not set", key)
	}

	return val, nil
}

func lookupString(key, def string) string {
	val := os.Getenv(key)

	if val == "" {
		return def
	}

	return val
}
