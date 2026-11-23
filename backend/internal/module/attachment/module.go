package attachment

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"monorepo/backend/internal/shared/middleware"
	"monorepo/backend/internal/shared/storage"
)

type Module struct {
	Usecase   Usecase
	handler   *Handler
	jwtSecret string
}

func NewModule(pool *pgxpool.Pool, store storage.ObjectStore, tickets TicketFinder, comments CommentFinder, jwtSecret string) *Module {
	repo := NewRepository(pool)
	uc := NewUsecase(repo, store, tickets, comments)
	h := NewHandler(uc)

	return &Module{
		Usecase:   uc,
		handler:   h,
		jwtSecret: jwtSecret,
	}
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	group := router.Group("/attachments", middleware.RequireAuth(m.jwtSecret))
	group.Post("/", m.handler.Upload)
	group.Get("/", m.handler.List)
	group.Get("/:id/file", m.handler.GetFile)
}
