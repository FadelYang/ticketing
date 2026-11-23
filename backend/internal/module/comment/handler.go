package comment

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/middleware"
	"monorepo/backend/internal/shared/response"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) ListByTicket(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(string)
	userRole, _ := c.Locals(middleware.ContextUserRoleKey).(string)

	page := DefaultPage
	if rawPage := strings.TrimSpace(c.Query("page")); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err != nil || parsed < 1 {
			return response.BadRequest(c, "page must be a positive integer")
		}
		page = parsed
	}

	limit := DefaultLimit
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 {
			return response.BadRequest(c, "limit must be a positive integer")
		}
		limit = parsed
	}

	filter := Filter{
		TicketID: c.Params("ticketId"),
		Sort:     c.Query("sort"),
		Page:     page,
		Limit:    limit,
	}

	result, err := h.usecase.ListByTicket(c.UserContext(), userID, userRole, filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusOK, result)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(string)
	userRole, _ := c.Locals(middleware.ContextUserRoleKey).(string)
	id := c.Params("id")

	item, err := h.usecase.GetByID(c.UserContext(), id, userID, userRole)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusOK, item)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(string)
	userRole, _ := c.Locals(middleware.ContextUserRoleKey).(string)

	var input CreateInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid JSON request body")
	}
	input.TicketID = c.Params("ticketId")

	item, err := h.usecase.Create(c.UserContext(), userID, userRole, input)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusCreated, item)
}
