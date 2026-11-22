package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/jwt"
	"monorepo/backend/internal/shared/response"
)

const (
	ContextUserIDKey    = "user_id"
	ContextUserEmailKey = "user_email"
)

func RequireAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := strings.TrimSpace(c.Get("Authorization"))
		if authHeader == "" {
			return response.Unauthorized(c, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			return response.Unauthorized(c, "invalid authorization header format")
		}

		claims, err := jwt.ParseToken(jwtSecret, strings.TrimSpace(parts[1]))
		if err != nil {
			if errors.Is(err, jwt.ErrExpiredToken) {
				return response.Unauthorized(c, "token has expired")
			}
			return response.Unauthorized(c, "invalid token")
		}

		c.Locals(ContextUserIDKey, claims.Subject)
		c.Locals(ContextUserEmailKey, claims.Email)

		return c.Next()
	}
}
