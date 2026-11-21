package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"monorepo/backend/internal/config"
	"monorepo/backend/internal/database"
	"monorepo/backend/internal/module/example"
	"monorepo/backend/internal/module/health"
	"monorepo/backend/internal/router"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Register feature modules (add new modules here as the app grows)
	healthModule := health.NewModule(pool)
	exampleModule := example.NewModule(pool)

	app := router.Setup(cfg, healthModule, exampleModule)

	// Graceful shutdown
	go func() {
		addr := ":" + cfg.AppPort
		log.Printf("server starting on %s (%s)", addr, cfg.AppEnv)
		if err := app.Listen(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("error during server shutdown: %v", err)
	}
}
