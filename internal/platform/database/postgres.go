package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres wraps the shared PostgreSQL connection pool.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres creates and verifies a PostgreSQL pool.
func NewPostgres(databaseURL string) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

// Ping verifies the database connection.
func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Close closes the PostgreSQL pool.
func (p *Postgres) Close() {
	if p == nil || p.pool == nil {
		return
	}

	p.pool.Close()
}
