package health

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	handler *Handler
}

func NewModule(pool *pgxpool.Pool) *Module {
	return &Module{
		handler: NewHandler(pool),
	}
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	router.Get("/health", m.handler.Check)
}
