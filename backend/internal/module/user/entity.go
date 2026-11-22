package user

import (
	"context"
	"time"
)

type User struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phonenumber"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   *string   `json:"created_by"`
}

type Filter struct {
	Search string
}

type CreateInput struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phonenumber"`
	CreatedBy   *string `json:"created_by"`
}

type UpdateInput struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phonenumber"`
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindAll(ctx context.Context, filter Filter) ([]User, error)
	Update(ctx context.Context, id string, input UpdateInput) (*User, error)
	Delete(ctx context.Context, id string) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, filter Filter) ([]User, error)
	Update(ctx context.Context, id string, input UpdateInput) (*User, error)
	Delete(ctx context.Context, id string) error
}
