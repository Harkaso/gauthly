package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConns          = 20
	minConns          = 5
	connectTimeout    = 5 * time.Second
	maxConnIdleTime   = 15 * time.Minute
	maxConnLifetime   = 40 * time.Minute
	healthCheckPeriod = 1 * time.Minute
)

func Connect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.ConnConfig.ConnectTimeout = connectTimeout
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.HealthCheckPeriod = healthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
