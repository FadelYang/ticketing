package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
)

type mockRepo struct{}

func (m *mockRepo) Create(_ context.Context, input user.CreateInput) (*user.User, error) {
	now := time.Now()
	return &user.User{
		ID:          "11111111-1111-1111-1111-111111111111",
		Name:        input.Name,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (m *mockRepo) FindByID(_ context.Context, _ string) (*user.User, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockRepo) FindAll(_ context.Context, _ user.Filter) ([]user.User, error) {
	return []user.User{}, nil
}

func (m *mockRepo) Update(_ context.Context, _ string, _ user.UpdateInput) (*user.User, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func TestUserUsecase_CreateValidation(t *testing.T) {
	uc := user.NewUsecase(&mockRepo{})

	tests := []struct {
		name  string
		input user.CreateInput
	}{
		{
			name:  "empty name",
			input: user.CreateInput{Name: "   ", Email: "a@example.com", PhoneNumber: "+123456"},
		},
		{
			name:  "name too long",
			input: user.CreateInput{Name: strings.Repeat("a", 51), Email: "a@example.com", PhoneNumber: "+123456"},
		},
		{
			name:  "invalid email",
			input: user.CreateInput{Name: "Alice", Email: "not-an-email", PhoneNumber: "+123456"},
		},
		{
			name:  "empty phonenumber",
			input: user.CreateInput{Name: "Alice", Email: "a@example.com", PhoneNumber: "  "},
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

func TestUserUsecase_CreateTrimmingAndCreatedBy(t *testing.T) {
	uc := user.NewUsecase(&mockRepo{})

	emptyCreator := "   "
	item, err := uc.Create(context.Background(), user.CreateInput{
		Name:        "  Alice  ",
		Email:       "  alice@example.com  ",
		PhoneNumber: "  +628123456789  ",
		CreatedBy:   &emptyCreator,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Alice" {
		t.Fatalf("expected trimmed name, got %q", item.Name)
	}
	if item.Email != "alice@example.com" {
		t.Fatalf("expected trimmed email, got %q", item.Email)
	}
	if item.PhoneNumber != "+628123456789" {
		t.Fatalf("expected trimmed phonenumber, got %q", item.PhoneNumber)
	}
	if item.CreatedBy != nil {
		t.Fatalf("expected blank created_by to become nil, got %v", *item.CreatedBy)
	}
}

func TestUserUsecase_UpdateValidation(t *testing.T) {
	uc := user.NewUsecase(&mockRepo{})

	_, err := uc.Update(context.Background(), " ", user.UpdateInput{
		Name:        "Alice",
		Email:       "alice@example.com",
		PhoneNumber: "+628123456789",
	})
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected validation error for empty id, got %v", err)
	}

	_, err = uc.Update(context.Background(), "11111111-1111-1111-1111-111111111111", user.UpdateInput{
		Name:        "Alice",
		Email:       "invalid",
		PhoneNumber: "+628123456789",
	})
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected validation error for invalid email, got %v", err)
	}
}
