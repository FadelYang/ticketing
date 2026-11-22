package ticket

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/middleware"
	"monorepo/backend/internal/shared/response"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(string)
	userRole, _ := c.Locals(middleware.ContextUserRoleKey).(string)

	filter := Filter{
		Search:      c.Query("search"),
		Status:      c.Query("status"),
		Priority:    c.Query("priority"),
		Category:    c.Query("category"),
		RequesterID: c.Query("requester_id"),
		AssigneeID:  c.Query("assignee_id"),
	}

	switch strings.ToLower(strings.TrimSpace(userRole)) {
	case user.RoleRequester:
		filter.RequesterID = userID
	case user.RoleAgent, user.RoleAdmin:
		// Agents and admins can view all tickets.
	default:
		return response.Forbidden(c, "insufficient permissions")
	}

	items, err := h.usecase.List(c.UserContext(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusOK, items)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(string)
	userRole, _ := c.Locals(middleware.ContextUserRoleKey).(string)
	id := c.Params("id")

	role := strings.ToLower(strings.TrimSpace(userRole))
	if role != user.RoleRequester && role != user.RoleAgent && role != user.RoleAdmin {
		return response.Forbidden(c, "insufficient permissions")
	}

	item, err := h.usecase.GetByID(c.UserContext(), id)
	if err != nil {
		return response.Error(c, err)
	}

	if role == user.RoleRequester && item.RequesterID != userID {
		return response.Forbidden(c, "you can only access your own tickets")
	}

	return response.Success(c, fiber.StatusOK, item)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var input CreateInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid JSON request body")
	}

	item, err := h.usecase.Create(c.UserContext(), input)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusCreated, item)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var input UpdateInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid JSON request body")
	}

	item, err := h.usecase.Update(c.UserContext(), id, input)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusOK, item)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.usecase.Delete(c.UserContext(), id); err != nil {
		return response.Error(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
