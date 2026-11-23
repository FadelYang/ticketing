package app

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"monorepo/backend/internal/config"
	"monorepo/backend/internal/database"
	"monorepo/backend/internal/module/auth"
	"monorepo/backend/internal/module/comment"
	"monorepo/backend/internal/module/example"
	"monorepo/backend/internal/module/health"
	"monorepo/backend/internal/module/ticket"
	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/router"
)

type Container struct {
	Config *config.Config
	DB     *pgxpool.Pool
	Fiber  *fiber.App

	Health  *health.Module
	Example *example.Module
	User    *user.Module
	Auth    *auth.Module
	Ticket  *ticket.Module
	Comment *comment.Module
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

	if err := database.SeedDefaultUsers(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("seed default users: %w", err)
	}

	healthMod := health.NewModule(pool)
	exampleMod := example.NewModule(pool)
	userMod := user.NewModule(pool)
	authMod := auth.NewModule(pool, cfg.JWTSecret, cfg.JWTExpiresIn)
	ticketMod := ticket.NewModule(pool, userMod.Usecase, cfg.JWTSecret)
	commentMod := comment.NewModule(pool, ticketMod.Usecase, userMod.Usecase, cfg.JWTSecret)

	// If a new module needs another module's usecase, pass it directly:
	// orderMod := order.NewModule(pool, exampleMod.Usecase)

	fiberApp := router.Setup(
		cfg,
		healthMod,
		exampleMod,
		userMod,
		authMod,
		ticketMod,
		commentMod,
	)

	return &Container{
		Config:  cfg,
		DB:      pool,
		Fiber:   fiberApp,
		Health:  healthMod,
		Example: exampleMod,
		User:    userMod,
		Auth:    authMod,
		Ticket:  ticketMod,
		Comment: commentMod,
	}, nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
