package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/jwt"
	"monorepo/backend/internal/shared/middleware"
)

func TestRequireRoles(t *testing.T) {
	const secret = "test-secret"
	app := fiber.New()
	app.Get("/admin-only", middleware.RequireAuth(secret), middleware.RequireRoles("admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/staff", middleware.RequireAuth(secret), middleware.RequireRoles("agent", "admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	requesterToken, err := jwt.GenerateToken(secret, 30*time.Minute, "u1", "req@example.com", "Req", "requester")
	if err != nil {
		t.Fatalf("unexpected token error: %v", err)
	}
	agentToken, err := jwt.GenerateToken(secret, 30*time.Minute, "u2", "agent@example.com", "Agent", "agent")
	if err != nil {
		t.Fatalf("unexpected token error: %v", err)
	}
	adminToken, err := jwt.GenerateToken(secret, 30*time.Minute, "u3", "admin@example.com", "Admin", "admin")
	if err != nil {
		t.Fatalf("unexpected token error: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		token      string
		wantStatus int
	}{
		{name: "missing token", path: "/admin-only", token: "", wantStatus: http.StatusUnauthorized},
		{name: "requester denied admin", path: "/admin-only", token: requesterToken, wantStatus: http.StatusForbidden},
		{name: "agent denied admin", path: "/admin-only", token: agentToken, wantStatus: http.StatusForbidden},
		{name: "admin allowed admin", path: "/admin-only", token: adminToken, wantStatus: http.StatusOK},
		{name: "requester denied staff", path: "/staff", token: requesterToken, wantStatus: http.StatusForbidden},
		{name: "agent allowed staff", path: "/staff", token: agentToken, wantStatus: http.StatusOK},
		{name: "admin allowed staff", path: "/staff", token: adminToken, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, resp.StatusCode)
			}
		})
	}
}
