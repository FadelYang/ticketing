package user

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

func NewModule(pool *pgxpool.Pool, jwtSecret string) *Module {
	repo := NewRepository(pool)
	uc := NewUsecase(repo)
	h := NewHandler(uc)

	return &Module{
		Usecase:   uc,
		handler:   h,
		jwtSecret: jwtSecret,
	}
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	group := router.Group("/users")
	group.Get("/", m.handler.List)
	group.Post("/", m.handler.Create)
	group.Get("/:id", m.handler.GetByID)
	group.Put("/:id", m.handler.Update)
	group.Patch("/:id/role", middleware.RequireAuth(m.jwtSecret), middleware.RequireRoles(RoleAdmin), m.handler.UpdateRole)
	group.Delete("/:id", m.handler.Delete)
}
