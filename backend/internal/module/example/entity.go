package example

import (
	"context"
	"time"
)

type Example struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Filter struct {
	Search string
}

type CreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (*Example, error)
	FindByID(ctx context.Context, id string) (*Example, error)
	FindAll(ctx context.Context, filter Filter) ([]Example, error)
	Update(ctx context.Context, id string, input UpdateInput) (*Example, error)
	Delete(ctx context.Context, id string) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*Example, error)
	GetByID(ctx context.Context, id string) (*Example, error)
	List(ctx context.Context, filter Filter) ([]Example, error)
	Update(ctx context.Context, id string, input UpdateInput) (*Example, error)
	Delete(ctx context.Context, id string) error
}
