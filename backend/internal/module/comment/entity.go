package comment

import (
	"context"
	"time"

	"monorepo/backend/internal/module/ticket"
	"monorepo/backend/internal/module/user"
)

const (
	SortNewest = "newest"
	SortOldest = "oldest"

	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

type Comment struct {
	ID         string    `json:"id"`
	TicketID   string    `json:"ticket_id"`
	AuthorID   string    `json:"author_id"`
	Body       string    `json:"body"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
}

type Filter struct {
	TicketID        string
	IncludeInternal bool
	Sort            string
	Page            int
	Limit           int
}

type ListResult struct {
	Items      []Comment `json:"items"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type CreateInput struct {
	TicketID   string `json:"-"`
	AuthorID   string `json:"-"`
	Body       string `json:"body"`
	IsInternal bool   `json:"is_internal"`
}

type TicketFinder interface {
	GetByID(ctx context.Context, id string) (*ticket.Ticket, error)
}

type UserFinder interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (*Comment, error)
	FindByID(ctx context.Context, id string) (*Comment, error)
	FindAll(ctx context.Context, filter Filter) (*ListResult, error)
}

type Usecase interface {
	Create(ctx context.Context, actorID string, actorRole string, input CreateInput) (*Comment, error)
	GetByID(ctx context.Context, id string, actorID string, actorRole string) (*Comment, error)
	ListByTicket(ctx context.Context, actorID string, actorRole string, filter Filter) (*ListResult, error)
}
