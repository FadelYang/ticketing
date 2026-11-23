package ticket

import (
	"context"
	"time"

	"monorepo/backend/internal/module/user"
)

const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
)

func IsValidPriority(priority string) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

func IsValidStatus(status string) bool {
	switch status {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}

type Ticket struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	RequesterID string    `json:"requester_id"`
	AssigneeID  *string   `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeadlineAt  time.Time `json:"deadline_at"`
}

const (
	SortNewest = "newest"
	SortOldest = "oldest"

	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

type Filter struct {
	Search      string
	Status      string
	Priority    string
	Category    string
	RequesterID string
	AssigneeID  string
	Sort        string
	Page        int
	Limit       int
}

type ListResult struct {
	Items      []Ticket `json:"items"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}

type StatusCounts struct {
	All        int64 `json:"all"`
	Open       int64 `json:"open"`
	InProgress int64 `json:"in_progress"`
	Resolved   int64 `json:"resolved"`
	Closed     int64 `json:"closed"`
}

type CreateInput struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	RequesterID string    `json:"requester_id"`
	AssigneeID  *string   `json:"assignee_id"`
	DeadlineAt  time.Time `json:"deadline_at"`
}

type UpdateInput struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	RequesterID string    `json:"requester_id"`
	AssigneeID  *string   `json:"assignee_id"`
	DeadlineAt  time.Time `json:"deadline_at"`
}

type PatchInput struct {
	Priority   *string `json:"priority"`
	Status     *string `json:"status"`
	AssigneeID *string `json:"assignee_id"`
	Comment    *string `json:"comment"`
}

type UserFinder interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
}

type CommentCreator interface {
	CreateForTicket(ctx context.Context, actorID, actorRole, ticketID, body string) error
}

type TicketNotifier interface {
	NotifyTicketCreated(t *Ticket)
	NotifyTicketUpdated(t *Ticket)
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (*Ticket, error)
	FindByID(ctx context.Context, id string) (*Ticket, error)
	FindAll(ctx context.Context, filter Filter) (*ListResult, error)
	CountByStatus(ctx context.Context, filter Filter) (*StatusCounts, error)
	Update(ctx context.Context, id string, input UpdateInput) (*Ticket, error)
	Delete(ctx context.Context, id string) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*Ticket, error)
	GetByID(ctx context.Context, id string) (*Ticket, error)
	List(ctx context.Context, filter Filter) (*ListResult, error)
	CountByStatus(ctx context.Context, filter Filter) (*StatusCounts, error)
	Update(ctx context.Context, id string, input UpdateInput) (*Ticket, error)
	Patch(ctx context.Context, id string, actorID string, actorRole string, input PatchInput) (*Ticket, error)
	Delete(ctx context.Context, id string) error
}
