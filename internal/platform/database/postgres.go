// Package database creates and configures the PostgreSQL connection pool.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool settings. They are constants for now; move them into configuration
// when there is a real reason to tune them per environment.
const (
	maxConns          = 10
	minConns          = 2
	maxConnLifeTime   = 30 * time.Minute
	maxConnIdleTime   = 5 * time.Minute
	healthCheckPeriod = time.Minute
	connectTimeout    = 5 * time.Second // one connection attempt
	pingTimeout       = 5 * time.Second // the startup check
	applicationName   = "carebridge-api"
)

// NewPool creates a connection pool and check that the database is reachable.
// It fails fast: if postgres is unreachable at startup, the process exits and
// the orchestrator (Docker, Kubernetes) will restart it, which is cleaner than
// serving requests that will fail anyway.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		// pgx hides the password in this error, so it is safe to wrap.
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = maxConnLifeTime
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.HealthCheckPeriod = healthCheckPeriod
	cfg.ConnConfig.ConnectTimeout = connectTimeout
	// Shows up in pg_stat_activity, so you can tell which service owns a connection.
	cfg.ConnConfig.RuntimeParams["application_name"] = applicationName

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
