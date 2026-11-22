package app

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"monorepo/backend/internal/config"
	"monorepo/backend/internal/database"
	"monorepo/backend/internal/module/example"
	"monorepo/backend/internal/module/health"
	"monorepo/backend/internal/router"
)

type Container struct {
	Config *config.Config
	DB     *pgxpool.Pool
	Fiber  *fiber.App

	Health  *health.Module
	Example *example.Module
}

func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := database.RunMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	healthMod := health.NewModule(pool)
	exampleMod := example.NewModule(pool)

	// If a new module needs another module's usecase, pass it directly:
	// orderMod := order.NewModule(pool, exampleMod.Usecase)

	fiberApp := router.Setup(
		cfg,
		healthMod,
		exampleMod,
	)

	return &Container{
		Config:  cfg,
		DB:      pool,
		Fiber:   fiberApp,
		Health:  healthMod,
		Example: exampleMod,
	}, nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
