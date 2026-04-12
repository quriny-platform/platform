// Package database manages the PostgreSQL connection pool and provides helpers
// for database lifecycle operations (connect, ping, close). It wraps the pgxpool
// library to give Quriny a single, consistent entry point for database access.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultConnectTimeout is the maximum time we wait for the initial connection
// to succeed during startup.
const DefaultConnectTimeout = 10 * time.Second

// DB wraps a pgxpool.Pool and provides Quriny-specific database operations.
type DB struct {
	Pool *pgxpool.Pool
}

// Connect parses the connection URL, creates a connection pool, and verifies
// the database is reachable with a ping. Returns an error if the database
// cannot be reached within the timeout.
//
// Example connURL: "postgres://quriny:quriny@localhost:5432/quriny?sslmode=disable"
func Connect(ctx context.Context, connURL string) (*DB, error) {
	// Parse the connection URL into pgxpool config.
	config, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	// Create the connection pool. pgxpool manages multiple connections
	// automatically and is safe for concurrent use.
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Verify the database is reachable before returning.
	pingCtx, cancel := context.WithTimeout(ctx, DefaultConnectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close releases all connections in the pool. Call this when the server shuts
// down to avoid leaking database connections.
func (db *DB) Close() {
	db.Pool.Close()
}
