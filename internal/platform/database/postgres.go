package database

import (
	"context"
	"fmt"
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

// Connect opens a connection pool for connString and verifies it with a ping,
// so that an unreachable database is reported at startup rather than on
// the first query. The caller owns the returned pool and must close it.
func Connect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.ConnConfig.ConnectTimeout = connectTimeout
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.HealthCheckPeriod = healthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
