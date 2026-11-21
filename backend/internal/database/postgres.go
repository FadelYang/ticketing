package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"monorepo/backend/migrations"
)

func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolCfg.MaxConns = 15
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrator := NewMigrator(pool, migrations.Files)
	applied, err := migrator.Up(ctx)
	if err != nil {
		return err
	}
	for _, version := range applied {
		log.Printf("applied migration: %s", version)
	}
	return nil
}
