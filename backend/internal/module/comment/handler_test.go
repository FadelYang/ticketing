package comment_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/module/comment"
	"monorepo/backend/internal/module/ticket"
	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/jwt"
	"monorepo/backend/internal/shared/middleware"
)

func setupCommentTestApp() (*fiber.App, string) {
	secret := "test-secret"
	repo := newMockCommentRepo()
	tickets := &mockTicketFinder{
		tickets: map[string]*ticket.Ticket{
			"t-1": {ID: "t-1", RequesterID: "req-1"},
			"t-2": {ID: "t-2", RequesterID: "req-2"},
		},
	}
	users := &mockUserFinder{
		users: map[string]*user.User{
			"req-1":   {ID: "req-1", Role: user.RoleRequester},
			"req-2":   {ID: "req-2", Role: user.RoleRequester},
			"agent-1": {ID: "agent-1", Role: user.RoleAgent},
			"admin-1": {ID: "admin-1", Role: user.RoleAdmin},
		},
	}
	uc := comment.NewUsecase(repo, tickets, users)
	h := comment.NewHandler(uc)

	app := fiber.New()
	auth := middleware.RequireAuth(secret)

	ticketComments := app.Group("/api/v1/tickets/:ticketId/comments", auth)
	ticketComments.Get("/", h.ListByTicket)
	ticketComments.Post("/", h.Create)

	comments := app.Group("/api/v1/comments", auth)
	comments.Get("/:id", h.GetByID)

	return app, secret
}

func TestCommentHandler_CRUDAndAuth(t *testing.T) {
	app, secret := setupCommentTestApp()

	req1Token, _ := jwt.GenerateToken(secret, 30*time.Minute, "req-1", "req1@example.com", "Req1", user.RoleRequester)
	agentToken, _ := jwt.GenerateToken(secret, 30*time.Minute, "agent-1", "agent@example.com", "Agent", user.RoleAgent)

	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/t-1/comments", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("requester lists only public comments on own ticket", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/t-1/comments", nil)
		req.Header.Set("Authorization", "Bearer "+req1Token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var body struct {
			Success bool               `json:"success"`
			Data    comment.ListResult `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(body.Data.Items) != 1 || body.Data.Items[0].IsInternal {
			t.Fatalf("expected 1 public comment, got %+v", body.Data.Items)
		}
	})

	t.Run("requester creating internal comment returns 403", func(t *testing.T) {
		payload := []byte(`{"body":"secret","is_internal":true}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/t-1/comments", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+req1Token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.StatusCode)
		}
	})

	t.Run("agent creates internal comment returns 201", func(t *testing.T) {
		payload := []byte(`{"body":"Internal triage note","is_internal":true}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/t-1/comments", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+agentToken)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})
}
