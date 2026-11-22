package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"monorepo/backend/internal/app"
	"monorepo/backend/internal/config"
)

func main() {
	cfg := config.Load()

	container, err := app.NewContainer(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
	defer container.Close()

	// Graceful shutdown
	go func() {
		addr := ":" + cfg.AppPort
		log.Printf("server starting on %s (%s)", addr, cfg.AppEnv)
		if err := container.Fiber.Listen(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	if err := container.Fiber.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("error during server shutdown: %v", err)
	}
}
