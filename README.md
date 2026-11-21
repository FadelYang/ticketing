# Go Fiber + Vue 3 Pinia Docker Monorepo

A clean full-stack monorepo with a **module-based Go + Fiber + pgx/v5 (raw SQL)** backend, **timestamp-based SQL migrations with Makefile**, **Vite + Vue 3 (Composition API) + Pinia + Tailwind CSS** frontend, and **PostgreSQL**, orchestrated with **Docker Compose**.

## Project Structure

```text
.
├── docker-compose.yml                  # Orchestrates postgres, backend, and frontend
├── Makefile                            # Database migration & Docker shortcuts
├── .env.example                        # Environment variable template
├── backend/                            # Module-based Go + Fiber + pgx/v5 API
│   ├── cmd/
│   │   ├── api/main.go                 # API entrypoint & module registration
│   │   └── migrate/main.go             # Timestamp-based migration CLI (create/up/down/status)
│   ├── migrations/                     # Timestamp-based SQL migrations (YYYYMMDDHHMMSS_name.{up,down}.sql)
│   │   ├── 20261121212100_create_examples_table.up.sql
│   │   ├── 20261121212100_create_examples_table.down.sql
│   │   └── embed.go
│   ├── internal/
│   │   ├── config/                     # Environment config loader
│   │   ├── database/                   # pgxpool connection pool & migration runner
│   │   ├── router/                     # Fiber router, middleware & module route mounting
│   │   ├── shared/
│   │   │   ├── apperror/               # Shared domain errors (ErrNotFound, ErrValidation)
│   │   │   └── response/               # Shared JSON response & error helpers
│   │   └── module/                     # Feature modules (add new domains here)
│   │       ├── health/
│   │       │   ├── handler.go
│   │       │   └── module.go
│   │       └── example/                # Starter CRUD module (Handler -> Usecase -> Repository + Query)
│   │           ├── entity.go           # Entity, DTOs, Repository & Usecase interfaces
│   │           ├── query.go            # Raw SQL queries (separated from repository)
│   │           ├── repository.go       # pgx/v5 data access implementation
│   │           ├── usecase.go          # Business logic & validation
│   │           ├── usecase_test.go     # Unit tests
│   │           ├── handler.go          # Fiber HTTP handlers
│   │           └── module.go           # Wires repo -> usecase -> handler & registers routes
│   └── Dockerfile
└── frontend/                           # Vite + Vue 3 + Pinia + Tailwind CSS
    ├── nginx.conf                      # Serves SPA & proxies /api/* to backend container
    ├── Dockerfile
    └── src/
        ├── assets/                     # Tailwind CSS entry
        ├── components/
        │   ├── common/                 # Reusable UI primitives (BaseButton, BaseBadge, EmptyState)
        │   ├── layout/                 # App shell (AppNavbar)
        │   └── example/                # Example feature components (ExampleForm, ExampleCard, ExampleList)
        ├── composables/                # Reusable Composition API logic (useHealth)
        ├── router/                     # Vue Router config
        ├── services/                   # HTTP client & API modules (api.ts, exampleService.ts)
        ├── stores/                     # Pinia Setup Stores (exampleStore.ts)
        ├── types/                      # TypeScript interfaces (example.ts)
        └── views/                      # Page views (HomeView.vue)
```

## Quick Start (Docker Compose)

Run PostgreSQL, Go Backend, and Vue Frontend together (pending migrations are applied automatically on backend startup):

```bash
docker compose up --build
```

Services will be available at:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1/health
- **PostgreSQL**: `localhost:5432` (`appuser` / `apppassword` / `appdb`)

To stop services (add `-v` to reset the database volume):

```bash
docker compose down
```

## Database Migrations (Makefile)

Migrations use UTC timestamp filenames (`YYYYMMDDHHMMSS_<name>.up.sql` and `.down.sql`) tracked in the `schema_migrations` table.

```bash
# Create a new timestamp-based migration pair (.up.sql & .down.sql)
make migrate-create name=create_users_table

# Apply all pending migrations (against local/configured DATABASE_URL)
make migrate-up

# Roll back the latest migration (or specify steps=2)
make migrate-down
make migrate-down steps=2

# Check applied vs. pending migrations
make migrate-status

# Or run migrations inside the running Docker backend container
make docker-migrate-up
make docker-migrate-down steps=1
make docker-migrate-status
```

## Adding a New Backend Module

1. Create a new folder under `backend/internal/module/<name>/` containing:
   - `entity.go` (entity, DTOs, `Repository` & `Usecase` interfaces)
   - `query.go` (raw SQL constants)
   - `repository.go` (`pgxpool` queries)
   - `usecase.go` (business logic)
   - `handler.go` (Fiber handlers)
   - `module.go` (`NewModule(pool)` + `RegisterRoutes(router)`)
2. Register the module in `backend/cmd/api/main.go`:
   ```go
   newModule := mymodule.NewModule(pool)
   app := router.Setup(cfg, healthModule, exampleModule, newModule)
   ```

## API Endpoints

| Method | Endpoint               | Description                          |
| ------ | ---------------------- | ------------------------------------ |
| GET    | `/api/v1/health`       | Health & database connectivity check |
| GET    | `/api/v1/examples`     | List examples (`?search=`)           |
| GET    | `/api/v1/examples/:id` | Get example by ID                    |
| POST   | `/api/v1/examples`     | Create an example                    |
| PUT    | `/api/v1/examples/:id` | Update an example                    |
| DELETE | `/api/v1/examples/:id` | Delete an example                    |
