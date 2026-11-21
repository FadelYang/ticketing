package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/apperror"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Success(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

func BadRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}

func Error(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apperror.ErrValidation):
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	case errors.Is(err, apperror.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "internal server error",
		})
	}
}
