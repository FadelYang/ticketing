package comment

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

func NewModule(pool *pgxpool.Pool, tickets TicketFinder, users UserFinder, jwtSecret string) *Module {
	repo := NewRepository(pool)
	uc := NewUsecase(repo, tickets, users)
	h := NewHandler(uc)

	return &Module{
		Usecase:   uc,
		handler:   h,
		jwtSecret: jwtSecret,
	}
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	authMiddleware := middleware.RequireAuth(m.jwtSecret)

	ticketComments := router.Group("/tickets/:ticketId/comments", authMiddleware)
	ticketComments.Get("/", m.handler.ListByTicket)
	ticketComments.Post("/", m.handler.Create)

	comments := router.Group("/comments", authMiddleware)
	comments.Get("/:id", m.handler.GetByID)
}
