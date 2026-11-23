package comment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"monorepo/backend/internal/module/comment"
	"monorepo/backend/internal/module/ticket"
	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
)

type mockCommentRepo struct {
	comments map[string]comment.Comment
}

func newMockCommentRepo() *mockCommentRepo {
	return &mockCommentRepo{
		comments: map[string]comment.Comment{
			"c-public": {
				ID:         "c-public",
				TicketID:   "t-1",
				AuthorID:   "req-1",
				Body:       "Public comment",
				IsInternal: false,
				CreatedAt:  time.Now().Add(-2 * time.Hour),
			},
			"c-internal": {
				ID:         "c-internal",
				TicketID:   "t-1",
				AuthorID:   "agent-1",
				Body:       "Internal note",
				IsInternal: true,
				CreatedAt:  time.Now().Add(-1 * time.Hour),
			},
		},
	}
}

func (m *mockCommentRepo) Create(_ context.Context, input comment.CreateInput) (*comment.Comment, error) {
	item := comment.Comment{
		ID:         "c-new",
		TicketID:   input.TicketID,
		AuthorID:   input.AuthorID,
		Body:       input.Body,
		IsInternal: input.IsInternal,
		CreatedAt:  time.Now(),
	}
	m.comments[item.ID] = item
	return &item, nil
}

func (m *mockCommentRepo) FindByID(_ context.Context, id string) (*comment.Comment, error) {
	item, ok := m.comments[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func (m *mockCommentRepo) FindAll(_ context.Context, filter comment.Filter) (*comment.ListResult, error) {
	items := make([]comment.Comment, 0)
	for _, c := range m.comments {
		if c.TicketID != filter.TicketID {
			continue
		}
		if !filter.IncludeInternal && c.IsInternal {
			continue
		}
		items = append(items, c)
	}
	return &comment.ListResult{
		Items:      items,
		Total:      int64(len(items)),
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: 1,
	}, nil
}

type mockTicketFinder struct {
	tickets map[string]*ticket.Ticket
}

func (m *mockTicketFinder) GetByID(_ context.Context, id string) (*ticket.Ticket, error) {
	t, ok := m.tickets[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	return t, nil
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

func newTestUsecase() comment.Usecase {
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
	return comment.NewUsecase(newMockCommentRepo(), tickets, users)
}

func TestCommentUsecase_CreateRules(t *testing.T) {
	uc := newTestUsecase()

	t.Run("empty body returns validation error", func(t *testing.T) {
		_, err := uc.Create(context.Background(), "req-1", user.RoleRequester, comment.CreateInput{
			TicketID: "t-1",
			Body:     "   ",
		})
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("requester cannot create internal comment", func(t *testing.T) {
		_, err := uc.Create(context.Background(), "req-1", user.RoleRequester, comment.CreateInput{
			TicketID:   "t-1",
			Body:       "Trying internal",
			IsInternal: true,
		})
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("requester cannot comment on another requester's ticket", func(t *testing.T) {
		_, err := uc.Create(context.Background(), "req-1", user.RoleRequester, comment.CreateInput{
			TicketID: "t-2",
			Body:     "Not my ticket",
		})
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("agent can create internal comment on any ticket", func(t *testing.T) {
		created, err := uc.Create(context.Background(), "agent-1", user.RoleAgent, comment.CreateInput{
			TicketID:   "t-1",
			Body:       "Investigating",
			IsInternal: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created.IsInternal || created.AuthorID != "agent-1" || created.Body != "Investigating" {
			t.Fatalf("unexpected created comment: %+v", created)
		}
	})
}

func TestCommentUsecase_VisibilityAndMutationRules(t *testing.T) {
	uc := newTestUsecase()

	t.Run("requester list excludes internal comments", func(t *testing.T) {
		res, err := uc.ListByTicket(context.Background(), "req-1", user.RoleRequester, comment.Filter{TicketID: "t-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 1 || res.Items[0].ID != "c-public" {
			t.Fatalf("expected only c-public, got %+v", res.Items)
		}
	})

	t.Run("agent list includes internal comments", func(t *testing.T) {
		res, err := uc.ListByTicket(context.Background(), "agent-1", user.RoleAgent, comment.Filter{TicketID: "t-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 2 {
			t.Fatalf("expected 2 comments, got %d", len(res.Items))
		}
	})

	t.Run("requester cannot get internal comment by id", func(t *testing.T) {
		_, err := uc.GetByID(context.Background(), "c-internal", "req-1", user.RoleRequester)
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}
