package ticket

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"monorepo/backend/internal/shared/middleware"
)

type Module struct {
	Usecase   Usecase
	handler   *Handler
	jwtSecret string
}

func NewModule(pool *pgxpool.Pool, users UserFinder, jwtSecret string) *Module {
	repo := NewRepository(pool)
	uc := NewUsecase(repo, users)
	h := NewHandler(uc)

	return &Module{
		Usecase:   uc,
		handler:   h,
		jwtSecret: jwtSecret,
	}
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	group := router.Group("/tickets")
	group.Get("/", middleware.RequireAuth(m.jwtSecret), m.handler.List)
	group.Post("/", m.handler.Create)
	group.Get("/:id", middleware.RequireAuth(m.jwtSecret), m.handler.GetByID)
	group.Put("/:id", m.handler.Update)
	group.Patch("/:id", middleware.RequireAuth(m.jwtSecret), m.handler.Patch)
	group.Delete("/:id", m.handler.Delete)
}
