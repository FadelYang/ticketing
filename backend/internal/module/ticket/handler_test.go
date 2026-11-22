package ticket_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/module/ticket"
	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
	"monorepo/backend/internal/shared/jwt"
	"monorepo/backend/internal/shared/middleware"
)

type handlerMockRepo struct {
	tickets []ticket.Ticket
}

func (m *handlerMockRepo) Create(_ context.Context, input ticket.CreateInput) (*ticket.Ticket, error) {
	return nil, nil
}

func (m *handlerMockRepo) FindByID(_ context.Context, id string) (*ticket.Ticket, error) {
	for _, t := range m.tickets {
		if t.ID == id {
			copyT := t
			return &copyT, nil
		}
	}
	return nil, apperror.ErrNotFound
}

func (m *handlerMockRepo) FindAll(_ context.Context, filter ticket.Filter) ([]ticket.Ticket, error) {
	result := make([]ticket.Ticket, 0)
	for _, t := range m.tickets {
		if filter.RequesterID != "" && t.RequesterID != filter.RequesterID {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (m *handlerMockRepo) Update(_ context.Context, _ string, _ ticket.UpdateInput) (*ticket.Ticket, error) {
	return nil, nil
}

func (m *handlerMockRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func setupTicketTestApp() (*fiber.App, string) {
	const secret = "test-secret"
	repo := &handlerMockRepo{
		tickets: []ticket.Ticket{
			{ID: "t-1", Title: "Ticket 1", RequesterID: "req-1", Priority: ticket.PriorityLow, Status: ticket.StatusOpen},
			{ID: "t-2", Title: "Ticket 2", RequesterID: "req-2", Priority: ticket.PriorityHigh, Status: ticket.StatusOpen},
		},
	}
	uc := ticket.NewUsecase(repo, newTestUsers())
	h := ticket.NewHandler(uc)

	app := fiber.New()
	group := app.Group("/api/v1/tickets")
	group.Get("/", middleware.RequireAuth(secret), h.List)
	group.Get("/:id", middleware.RequireAuth(secret), h.GetByID)

	return app, secret
}

func TestTicketHandler_GetRules(t *testing.T) {
	app, secret := setupTicketTestApp()

	req1Token, err := jwt.GenerateToken(secret, 30*time.Minute, "req-1", "req1@example.com", "Req1", user.RoleRequester)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	agentToken, err := jwt.GenerateToken(secret, 30*time.Minute, "agent-1", "agent@example.com", "Agent", user.RoleAgent)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	adminToken, err := jwt.GenerateToken(secret, 30*time.Minute, "admin-1", "admin@example.com", "Admin", user.RoleAdmin)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	t.Run("unauthenticated list returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthenticated get by id returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/t-1", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("requester list returns only own tickets even with requester_id query override", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets?requester_id=req-2", nil)
		req.Header.Set("Authorization", "Bearer "+req1Token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var body struct {
			Success bool            `json:"success"`
			Data    []ticket.Ticket `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(body.Data) != 1 || body.Data[0].ID != "t-1" {
			t.Fatalf("expected only t-1, got %+v", body.Data)
		}
	})

	t.Run("requester get own ticket succeeds, other ticket forbidden", func(t *testing.T) {
		reqOwn := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/t-1", nil)
		reqOwn.Header.Set("Authorization", "Bearer "+req1Token)
		respOwn, err := app.Test(reqOwn)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if respOwn.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for own ticket, got %d", respOwn.StatusCode)
		}

		reqOther := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/t-2", nil)
		reqOther.Header.Set("Authorization", "Bearer "+req1Token)
		respOther, err := app.Test(reqOther)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if respOther.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 for other ticket, got %d", respOther.StatusCode)
		}
	})

	for _, tc := range []struct {
		name  string
		token string
	}{
		{name: "agent", token: agentToken},
		{name: "admin", token: adminToken},
	} {
		t.Run(tc.name+" list returns all tickets", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}

			var body struct {
				Success bool            `json:"success"`
				Data    []ticket.Ticket `json:"data"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if len(body.Data) != 2 {
				t.Fatalf("expected 2 tickets, got %d", len(body.Data))
			}
		})

		t.Run(tc.name+" get any ticket by id succeeds", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/t-2", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}
		})
	}
}
