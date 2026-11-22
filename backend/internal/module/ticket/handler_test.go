package ticket_test

import (
	"bytes"
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

func (m *handlerMockRepo) Update(_ context.Context, id string, input ticket.UpdateInput) (*ticket.Ticket, error) {
	for i, t := range m.tickets {
		if t.ID == id {
			m.tickets[i].Title = input.Title
			m.tickets[i].Description = input.Description
			m.tickets[i].Category = input.Category
			m.tickets[i].Priority = input.Priority
			if input.Status != "" {
				m.tickets[i].Status = input.Status
			}
			m.tickets[i].RequesterID = input.RequesterID
			m.tickets[i].AssigneeID = input.AssigneeID
			m.tickets[i].DeadlineAt = input.DeadlineAt
			m.tickets[i].UpdatedAt = time.Now()
			copyT := m.tickets[i]
			return &copyT, nil
		}
	}
	return nil, apperror.ErrNotFound
}

func (m *handlerMockRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func setupTicketTestApp() (*fiber.App, string) {
	const secret = "test-secret"
	deadline := time.Now().Add(24 * time.Hour)
	repo := &handlerMockRepo{
		tickets: []ticket.Ticket{
			{ID: "t-1", Title: "Ticket 1", Description: "Desc 1", Category: "auth", RequesterID: "req-1", Priority: ticket.PriorityLow, Status: ticket.StatusOpen, DeadlineAt: deadline},
			{ID: "t-2", Title: "Ticket 2", Description: "Desc 2", Category: "billing", RequesterID: "req-2", Priority: ticket.PriorityHigh, Status: ticket.StatusOpen, DeadlineAt: deadline},
		},
	}
	uc := ticket.NewUsecase(repo, newTestUsers())
	h := ticket.NewHandler(uc)

	app := fiber.New()
	group := app.Group("/api/v1/tickets")
	group.Get("/", middleware.RequireAuth(secret), h.List)
	group.Get("/:id", middleware.RequireAuth(secret), h.GetByID)
	group.Patch("/:id", middleware.RequireAuth(secret), h.Patch)

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

func TestTicketHandler_PatchPartialUpdate(t *testing.T) {
	app, secret := setupTicketTestApp()

	agentToken, err := jwt.GenerateToken(secret, 30*time.Minute, "agent-1", "agent@example.com", "Agent", user.RoleAgent)
	if err != nil {
		t.Fatalf("generate agent token: %v", err)
	}
	adminToken, err := jwt.GenerateToken(secret, 30*time.Minute, "admin-1", "admin@example.com", "Admin", user.RoleAdmin)
	if err != nil {
		t.Fatalf("generate admin token: %v", err)
	}

	t.Run("unauthenticated patch returns 401", func(t *testing.T) {
		body := []byte(`{"status":"in_progress"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/t-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("agent can assign unassigned ticket and update status/priority, ignoring other fields", func(t *testing.T) {
		body := []byte(`{"title":"Hacked","status":"in progress","priority":"urgent","assignee_id":"agent-1"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/t-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+agentToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var parsed struct {
			Success bool          `json:"success"`
			Data    ticket.Ticket `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if parsed.Data.Status != ticket.StatusInProgress || parsed.Data.Priority != ticket.PriorityUrgent {
			t.Fatalf("expected updated status/priority, got %+v", parsed.Data)
		}
		if parsed.Data.AssigneeID == nil || *parsed.Data.AssigneeID != "agent-1" {
			t.Fatalf("expected assignee_id agent-1, got %v", parsed.Data.AssigneeID)
		}
		if parsed.Data.Title != "Ticket 1" {
			t.Fatalf("expected title to remain unchanged, got %q", parsed.Data.Title)
		}
	})

	t.Run("non-admin cannot change assignee_id once already assigned", func(t *testing.T) {
		body := []byte(`{"assignee_id":"admin-1"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/t-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+agentToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.StatusCode)
		}
	})

	t.Run("admin can change assignee_id even when already assigned", func(t *testing.T) {
		body := []byte(`{"assignee_id":"admin-1"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/t-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var parsed struct {
			Success bool          `json:"success"`
			Data    ticket.Ticket `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if parsed.Data.AssigneeID == nil || *parsed.Data.AssigneeID != "admin-1" {
			t.Fatalf("expected assignee_id admin-1, got %v", parsed.Data.AssigneeID)
		}
	})

	t.Run("patch non-existent ticket returns 404", func(t *testing.T) {
		body := []byte(`{"status":"resolved"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/missing", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+agentToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("patch with invalid status returns 400", func(t *testing.T) {
		body := []byte(`{"status":"closed"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/t-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+agentToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})
}

func TestTicketHandler_StatusTransitionRules(t *testing.T) {
	const secret = "test-secret"
	deadline := time.Now().Add(24 * time.Hour)
	agent1 := "agent-1"

	makeApp := func(initial ticket.Ticket) *fiber.App {
		repo := &handlerMockRepo{tickets: []ticket.Ticket{initial}}
		uc := ticket.NewUsecase(repo, &mockUserFinder{
			users: map[string]*user.User{
				"req-1":   {ID: "req-1", Role: user.RoleRequester},
				"req-2":   {ID: "req-2", Role: user.RoleRequester},
				"agent-1": {ID: "agent-1", Role: user.RoleAgent},
				"agent-2": {ID: "agent-2", Role: user.RoleAgent},
				"admin-1": {ID: "admin-1", Role: user.RoleAdmin},
			},
		})
		h := ticket.NewHandler(uc)
		app := fiber.New()
		app.Patch("/api/v1/tickets/:id", middleware.RequireAuth(secret), h.Patch)
		return app
	}

	req1Token, _ := jwt.GenerateToken(secret, 30*time.Minute, "req-1", "req1@example.com", "Req1", user.RoleRequester)
	req2Token, _ := jwt.GenerateToken(secret, 30*time.Minute, "req-2", "req2@example.com", "Req2", user.RoleRequester)
	agent1Token, _ := jwt.GenerateToken(secret, 30*time.Minute, "agent-1", "agent1@example.com", "Agent1", user.RoleAgent)
	agent2Token, _ := jwt.GenerateToken(secret, 30*time.Minute, "agent-2", "agent2@example.com", "Agent2", user.RoleAgent)
	adminToken, _ := jwt.GenerateToken(secret, 30*time.Minute, "admin-1", "admin@example.com", "Admin", user.RoleAdmin)

	tests := []struct {
		name       string
		from       string
		assignee   *string
		to         string
		token      string
		wantStatus int
	}{
		// open -> in_progress
		{name: "open->in_progress by assigned agent", from: ticket.StatusOpen, assignee: &agent1, to: ticket.StatusInProgress, token: agent1Token, wantStatus: http.StatusOK},
		{name: "open->in_progress by other agent forbidden", from: ticket.StatusOpen, assignee: &agent1, to: ticket.StatusInProgress, token: agent2Token, wantStatus: http.StatusForbidden},
		{name: "open->in_progress by unassigned agent forbidden", from: ticket.StatusOpen, assignee: nil, to: ticket.StatusInProgress, token: agent1Token, wantStatus: http.StatusForbidden},
		{name: "open->in_progress by admin", from: ticket.StatusOpen, assignee: nil, to: ticket.StatusInProgress, token: adminToken, wantStatus: http.StatusOK},
		{name: "open->in_progress by requester forbidden", from: ticket.StatusOpen, assignee: &agent1, to: ticket.StatusInProgress, token: req1Token, wantStatus: http.StatusForbidden},

		// open -> closed
		{name: "open->closed by requester", from: ticket.StatusOpen, assignee: nil, to: ticket.StatusClosed, token: req1Token, wantStatus: http.StatusOK},
		{name: "open->closed by other requester forbidden", from: ticket.StatusOpen, assignee: nil, to: ticket.StatusClosed, token: req2Token, wantStatus: http.StatusForbidden},
		{name: "open->closed by admin", from: ticket.StatusOpen, assignee: nil, to: ticket.StatusClosed, token: adminToken, wantStatus: http.StatusOK},
		{name: "open->closed by agent forbidden", from: ticket.StatusOpen, assignee: &agent1, to: ticket.StatusClosed, token: agent1Token, wantStatus: http.StatusForbidden},

		// open -> resolved (disallowed)
		{name: "open->resolved disallowed", from: ticket.StatusOpen, assignee: &agent1, to: ticket.StatusResolved, token: adminToken, wantStatus: http.StatusBadRequest},

		// in_progress -> resolved
		{name: "in_progress->resolved by assigned agent", from: ticket.StatusInProgress, assignee: &agent1, to: ticket.StatusResolved, token: agent1Token, wantStatus: http.StatusOK},
		{name: "in_progress->resolved by other agent forbidden", from: ticket.StatusInProgress, assignee: &agent1, to: ticket.StatusResolved, token: agent2Token, wantStatus: http.StatusForbidden},
		{name: "in_progress->resolved by admin", from: ticket.StatusInProgress, assignee: &agent1, to: ticket.StatusResolved, token: adminToken, wantStatus: http.StatusOK},
		{name: "in_progress->resolved by requester forbidden", from: ticket.StatusInProgress, assignee: &agent1, to: ticket.StatusResolved, token: req1Token, wantStatus: http.StatusForbidden},

		// in_progress -> open/closed (disallowed)
		{name: "in_progress->open disallowed", from: ticket.StatusInProgress, assignee: &agent1, to: ticket.StatusOpen, token: adminToken, wantStatus: http.StatusBadRequest},
		{name: "in_progress->closed disallowed", from: ticket.StatusInProgress, assignee: &agent1, to: ticket.StatusClosed, token: adminToken, wantStatus: http.StatusBadRequest},

		// resolved -> closed
		{name: "resolved->closed by requester", from: ticket.StatusResolved, assignee: &agent1, to: ticket.StatusClosed, token: req1Token, wantStatus: http.StatusOK},
		{name: "resolved->closed by agent", from: ticket.StatusResolved, assignee: &agent1, to: ticket.StatusClosed, token: agent2Token, wantStatus: http.StatusOK},
		{name: "resolved->closed by admin", from: ticket.StatusResolved, assignee: &agent1, to: ticket.StatusClosed, token: adminToken, wantStatus: http.StatusOK},

		// resolved -> open
		{name: "resolved->open by requester", from: ticket.StatusResolved, assignee: &agent1, to: ticket.StatusOpen, token: req1Token, wantStatus: http.StatusOK},
		{name: "resolved->open by agent forbidden", from: ticket.StatusResolved, assignee: &agent1, to: ticket.StatusOpen, token: agent1Token, wantStatus: http.StatusForbidden},
		{name: "resolved->open by admin forbidden", from: ticket.StatusResolved, assignee: &agent1, to: ticket.StatusOpen, token: adminToken, wantStatus: http.StatusForbidden},

		// resolved -> in_progress (disallowed)
		{name: "resolved->in_progress disallowed", from: ticket.StatusResolved, assignee: &agent1, to: ticket.StatusInProgress, token: adminToken, wantStatus: http.StatusBadRequest},

		// closed -> open (disallowed)
		{name: "closed->open disallowed", from: ticket.StatusClosed, assignee: &agent1, to: ticket.StatusOpen, token: req1Token, wantStatus: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := makeApp(ticket.Ticket{
				ID:          "t-1",
				Title:       "Ticket 1",
				Description: "Desc 1",
				Category:    "auth",
				Priority:    ticket.PriorityLow,
				Status:      tc.from,
				RequesterID: "req-1",
				AssigneeID:  tc.assignee,
				DeadlineAt:  deadline,
			})

			body := []byte(`{"status":"` + tc.to + `"}`)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/t-1", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tc.token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, resp.StatusCode)
			}
		})
	}
}
