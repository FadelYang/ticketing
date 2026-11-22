package ticket_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"monorepo/backend/internal/module/ticket"
	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
)

type mockTicketRepo struct{}

func (m *mockTicketRepo) Create(_ context.Context, input ticket.CreateInput) (*ticket.Ticket, error) {
	now := time.Now()
	return &ticket.Ticket{
		ID:          "11111111-1111-1111-1111-111111111111",
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		Priority:    input.Priority,
		Status:      input.Status,
		RequesterID: input.RequesterID,
		AssigneeID:  input.AssigneeID,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeadlineAt:  input.DeadlineAt,
	}, nil
}

func (m *mockTicketRepo) FindByID(_ context.Context, _ string) (*ticket.Ticket, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockTicketRepo) FindAll(_ context.Context, _ ticket.Filter) ([]ticket.Ticket, error) {
	return []ticket.Ticket{}, nil
}

func (m *mockTicketRepo) Update(_ context.Context, id string, input ticket.UpdateInput) (*ticket.Ticket, error) {
	now := time.Now()
	status := input.Status
	if status == "" {
		status = ticket.StatusOpen
	}
	return &ticket.Ticket{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		Priority:    input.Priority,
		Status:      status,
		RequesterID: input.RequesterID,
		AssigneeID:  input.AssigneeID,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeadlineAt:  input.DeadlineAt,
	}, nil
}

func (m *mockTicketRepo) Delete(_ context.Context, _ string) error {
	return nil
}

type mockUserFinder struct {
	users map[string]*user.User
}

func (m *mockUserFinder) GetByID(_ context.Context, id string) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	return u, nil
}

func newTestUsers() *mockUserFinder {
	return &mockUserFinder{
		users: map[string]*user.User{
			"req-1":   {ID: "req-1", Role: user.RoleRequester},
			"agent-1": {ID: "agent-1", Role: user.RoleAgent},
			"admin-1": {ID: "admin-1", Role: user.RoleAdmin},
		},
	}
}

func TestTicketUsecase_CreateValidation(t *testing.T) {
	uc := ticket.NewUsecase(&mockTicketRepo{}, newTestUsers())
	deadline := time.Now().Add(24 * time.Hour)
	badAssignee := "req-1"
	missingAssignee := "missing"

	tests := []struct {
		name  string
		input ticket.CreateInput
	}{
		{
			name: "empty title",
			input: ticket.CreateInput{
				Title: "   ", Description: "desc", Category: "billing", Priority: "low", RequesterID: "req-1", DeadlineAt: deadline,
			},
		},
		{
			name: "title too long",
			input: ticket.CreateInput{
				Title: strings.Repeat("a", 251), Description: "desc", Category: "billing", Priority: "low", RequesterID: "req-1", DeadlineAt: deadline,
			},
		},
		{
			name: "empty description",
			input: ticket.CreateInput{
				Title: "Help", Description: "  ", Category: "billing", Priority: "low", RequesterID: "req-1", DeadlineAt: deadline,
			},
		},
		{
			name: "empty category",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: " ", Priority: "low", RequesterID: "req-1", DeadlineAt: deadline,
			},
		},
		{
			name: "invalid priority",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: "billing", Priority: "critical", RequesterID: "req-1", DeadlineAt: deadline,
			},
		},
		{
			name: "invalid status",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: "billing", Priority: "low", Status: "archived", RequesterID: "req-1", DeadlineAt: deadline,
			},
		},
		{
			name: "missing deadline",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: "billing", Priority: "low", RequesterID: "req-1",
			},
		},
		{
			name: "missing requester_id",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: "billing", Priority: "low", DeadlineAt: deadline,
			},
		},
		{
			name: "non-existent requester",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: "billing", Priority: "low", RequesterID: "unknown", DeadlineAt: deadline,
			},
		},
		{
			name: "requester with wrong role",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: "billing", Priority: "low", RequesterID: "agent-1", DeadlineAt: deadline,
			},
		},
		{
			name: "non-existent assignee",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: "billing", Priority: "low", RequesterID: "req-1", AssigneeID: &missingAssignee, DeadlineAt: deadline,
			},
		},
		{
			name: "assignee with requester role",
			input: ticket.CreateInput{
				Title: "Help", Description: "desc", Category: "billing", Priority: "low", RequesterID: "req-1", AssigneeID: &badAssignee, DeadlineAt: deadline,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Create(context.Background(), tc.input)
			if !errors.Is(err, apperror.ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestTicketUsecase_CreateDefaultsAndNormalization(t *testing.T) {
	uc := ticket.NewUsecase(&mockTicketRepo{}, newTestUsers())
	deadline := time.Now().Add(24 * time.Hour)
	blankAssignee := "   "

	item, err := uc.Create(context.Background(), ticket.CreateInput{
		Title:       "  Cannot login  ",
		Description: "  Password reset fails  ",
		Category:    "  auth  ",
		Priority:    "  HIGH ",
		RequesterID: " req-1 ",
		AssigneeID:  &blankAssignee,
		DeadlineAt:  deadline,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Title != "Cannot login" {
		t.Fatalf("expected trimmed title, got %q", item.Title)
	}
	if item.Priority != ticket.PriorityHigh {
		t.Fatalf("expected priority high, got %q", item.Priority)
	}
	if item.Status != ticket.StatusOpen {
		t.Fatalf("expected default status open, got %q", item.Status)
	}
	if item.AssigneeID != nil {
		t.Fatalf("expected blank assignee_id to become nil, got %v", *item.AssigneeID)
	}
}

func TestTicketUsecase_CreateWithAgentAndAdminAssignee(t *testing.T) {
	uc := ticket.NewUsecase(&mockTicketRepo{}, newTestUsers())
	deadline := time.Now().Add(24 * time.Hour)

	for _, assigneeID := range []string{"agent-1", "admin-1"} {
		aid := assigneeID
		item, err := uc.Create(context.Background(), ticket.CreateInput{
			Title:       "Issue",
			Description: "Details",
			Category:    "support",
			Priority:    "urgent",
			Status:      "in progress",
			RequesterID: "req-1",
			AssigneeID:  &aid,
			DeadlineAt:  deadline,
		})
		if err != nil {
			t.Fatalf("unexpected error for assignee %s: %v", assigneeID, err)
		}
		if item.Status != ticket.StatusInProgress {
			t.Fatalf("expected normalized status in_progress, got %q", item.Status)
		}
		if item.AssigneeID == nil || *item.AssigneeID != assigneeID {
			t.Fatalf("expected assignee_id %s, got %v", assigneeID, item.AssigneeID)
		}
	}
}

func TestTicketUsecase_UpdateValidation(t *testing.T) {
	uc := ticket.NewUsecase(&mockTicketRepo{}, newTestUsers())
	deadline := time.Now().Add(24 * time.Hour)

	_, err := uc.Update(context.Background(), " ", ticket.UpdateInput{
		Title:       "Issue",
		Description: "Details",
		Category:    "support",
		Priority:    "low",
		RequesterID: "req-1",
		DeadlineAt:  deadline,
	})
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected validation error for empty id, got %v", err)
	}

	_, err = uc.Update(context.Background(), "ticket-1", ticket.UpdateInput{
		Title:       "Issue",
		Description: "Details",
		Category:    "support",
		Priority:    "low",
		Status:      "in_progress",
		RequesterID: "req-1",
		DeadlineAt:  deadline,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
