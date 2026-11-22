package auth

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

func (h *Handler) Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid JSON request body")
	}

	result, err := h.usecase.Register(c.UserContext(), input)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusCreated, result)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid JSON request body")
	}

	result, err := h.usecase.Login(c.UserContext(), input)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusOK, result)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	result, err := h.usecase.Logout(c.UserContext())
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusOK, result)
}
