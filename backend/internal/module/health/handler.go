package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"monorepo/backend/internal/shared/response"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

func (h *Handler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.APIResponse{
			Success: false,
			Error:   "database unavailable",
		})
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"status":   "ok",
		"database": "connected",
	})
}
