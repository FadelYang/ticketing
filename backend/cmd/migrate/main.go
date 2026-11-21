package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"
	"time"

	"monorepo/backend/internal/config"
	"monorepo/backend/internal/database"
	"monorepo/backend/migrations"
)

func main() {
	dir := flag.String("dir", "./migrations", "path to migrations directory")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]

	if cmd == "create" {
		if len(args) < 2 {
			log.Fatal("missing migration name: usage: migrate create <name>")
		}
		upPath, downPath, err := database.CreateMigrationFiles(*dir, args[1], time.Now())
		if err != nil {
			log.Fatalf("create migration failed: %v", err)
		}
		fmt.Printf("Created migration files:\n  %s\n  %s\n", upPath, downPath)
		return
	}

	cfg := config.Load()
	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	var fsys fs.FS = migrations.Files
	if info, statErr := os.Stat(*dir); statErr == nil && info.IsDir() {
		fsys = os.DirFS(*dir)
	}

	migrator := database.NewMigrator(pool, fsys)

	switch cmd {
	case "up":
		applied, err := migrator.Up(ctx)
		if err != nil {
			log.Fatalf("migrate up failed: %v", err)
		}
		if len(applied) == 0 {
			fmt.Println("No pending migrations.")
			return
		}
		for _, v := range applied {
			fmt.Printf("Applied: %s\n", v)
		}

	case "down":
		steps := 1
		if len(args) >= 2 {
			parsed, err := strconv.Atoi(args[1])
			if err != nil || parsed <= 0 {
				log.Fatalf("invalid steps %q: must be a positive integer", args[1])
			}
			steps = parsed
		}
		rolledBack, err := migrator.Down(ctx, steps)
		if err != nil {
			log.Fatalf("migrate down failed: %v", err)
		}
		if len(rolledBack) == 0 {
			fmt.Println("No applied migrations to roll back.")
			return
		}
		for _, v := range rolledBack {
			fmt.Printf("Rolled back: %s\n", v)
		}

	case "status":
		statuses, err := migrator.Status(ctx)
		if err != nil {
			log.Fatalf("migrate status failed: %v", err)
		}
		if len(statuses) == 0 {
			fmt.Println("No migration files found.")
			return
		}
		for _, s := range statuses {
			if s.Applied && s.AppliedAt != nil {
				fmt.Printf("[APPLIED] %s (%s)\n", s.Version, s.AppliedAt.Format(time.RFC3339))
			} else {
				fmt.Printf("[PENDING] %s\n", s.Version)
			}
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate [-dir ./migrations] <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  create <name>   Create timestamp-based .up.sql and .down.sql files")
	fmt.Println("  up              Apply all pending migrations")
	fmt.Println("  down [steps]    Roll back the latest migration(s) (default: 1)")
	fmt.Println("  status          Show applied and pending migrations")
}
