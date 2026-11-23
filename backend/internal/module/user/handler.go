package user

import (
	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/response"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) List(c *fiber.Ctx) error {
	filter := Filter{
		Search: c.Query("search"),
	}

	items, err := h.usecase.List(c.UserContext(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusOK, items)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	item, err := h.usecase.GetByID(c.UserContext(), id)
	if err != nil {
		return response.Error(c, err)
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

func (h *Handler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")

	var input UpdateRoleInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid JSON request body")
	}

	item, err := h.usecase.UpdateRole(c.UserContext(), id, input.Role)
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
