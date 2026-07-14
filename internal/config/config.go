package config

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const (
	defaultPort     = ":8080"
	defaultTenantID = "00000000-0000-4000-8000-000000000000"
)

type Config struct {
	DatabaseURL     string
	Port            string
	DefaultTenantID uuid.UUID
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	var err error

	cfg.DatabaseURL, err = requireString("DB_URL")
	if err != nil {
		return nil, err
	}

	cfg.Port = lookupString("PORT", defaultPort)

	cfg.DefaultTenantID, err = uuid.Parse(lookupString("DEFAULT_TENANT_ID", defaultTenantID))
	if err != nil {
		return nil, err
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
