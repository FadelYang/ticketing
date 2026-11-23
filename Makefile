.PHONY: help migrate-create migrate-up migrate-down migrate-status seed \
	docker-up docker-down docker-migrate-up docker-migrate-down docker-migrate-status docker-seed

MIGRATIONS_DIR := ./migrations
steps ?= 1

help:
	@echo "Available commands:"
	@echo "  make migrate-create name=<name>       Create timestamp-based .up.sql and .down.sql migration files"
	@echo "  make migrate-up                       Apply all pending migrations (local Go + Postgres)"
	@echo "  make migrate-down [steps=1]           Roll back latest migration(s) (local Go + Postgres)"
	@echo "  make migrate-status                   Show applied and pending migrations"
	@echo "  make seed                             Seed sample users, tickets, and comments (local)"
	@echo "  make docker-up                        Build and start all containers via Docker Compose"
	@echo "  make docker-down                      Stop and remove Docker Compose containers"
	@echo "  make docker-migrate-up                Apply pending migrations inside backend container"
	@echo "  make docker-migrate-down [steps=1]    Roll back migration(s) inside backend container"
	@echo "  make docker-migrate-status            Show migration status inside backend container"
	@echo "  make docker-seed                      Seed sample users, tickets, and comments inside container"

migrate-create:
ifndef name
	$(error Please provide a migration name, e.g.: make migrate-create name=create_users_table)
endif
	cd backend && go run ./cmd/migrate -dir $(MIGRATIONS_DIR) create $(name)

migrate-up:
	cd backend && go run ./cmd/migrate -dir $(MIGRATIONS_DIR) up

migrate-down:
	cd backend && go run ./cmd/migrate -dir $(MIGRATIONS_DIR) down $(steps)

migrate-status:
	cd backend && go run ./cmd/migrate -dir $(MIGRATIONS_DIR) status

seed:
	cd backend && go run ./cmd/seed

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-migrate-up:
	docker compose exec backend /app/migrate up

docker-migrate-down:
	docker compose exec backend /app/migrate down $(steps)

docker-migrate-status:
	docker compose exec backend /app/migrate status

docker-seed:
	docker compose exec backend /app/seed
