package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultPort = ":8080"
)

type Config struct {
	DatabaseURL string
	Port        string
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
