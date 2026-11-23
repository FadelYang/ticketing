package main

import (
	"context"
	"fmt"
	"log"

	"monorepo/backend/internal/config"
	"monorepo/backend/internal/database"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	res, err := database.Seed(ctx, pool)
	if err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	fmt.Printf("Seed complete:\n")
	fmt.Printf("  Users resolved:    %d requesters, %d agents, %d admins\n",
		res.ExistingRequesters, res.ExistingAgents, res.ExistingAdmins)
	if res.UsersInserted > 0 {
		fmt.Printf("  Default users:     %d inserted (password: Ticketing@123)\n", res.UsersInserted)
	}
	fmt.Printf("  Tickets inserted:  %d\n", res.TicketsInserted)
	fmt.Printf("  Comments inserted: %d\n", res.CommentsInserted)
}
