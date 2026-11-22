package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"monorepo/backend/internal/shared/middleware"
)

type Module struct {
	Usecase   Usecase
	handler   *Handler
	jwtSecret string
}

func NewModule(pool *pgxpool.Pool, jwtSecret string, jwtTTL time.Duration) *Module {
	repo := NewRepository(pool)
	uc := NewUsecase(repo, jwtSecret, jwtTTL)
	h := NewHandler(uc)

	return &Module{
		Usecase:   uc,
		handler:   h,
		jwtSecret: jwtSecret,
	}
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	group := router.Group("/auth")
	group.Post("/register", middleware.OptionalAuth(m.jwtSecret), m.handler.Register)
	group.Post("/login", m.handler.Login)
	group.Post("/logout", middleware.RequireAuth(m.jwtSecret), m.handler.Logout)
}
