package example

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	handler *Handler
}

// NewModule wires Repository -> Usecase -> Handler for the example module.
func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	uc := NewUsecase(repo)
	h := NewHandler(uc)

	return &Module{handler: h}
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	group := router.Group("/examples")
	group.Get("/", m.handler.List)
	group.Post("/", m.handler.Create)
	group.Get("/:id", m.handler.GetByID)
	group.Put("/:id", m.handler.Update)
	group.Delete("/:id", m.handler.Delete)
}
