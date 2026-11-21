package example_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"monorepo/backend/internal/module/example"
	"monorepo/backend/internal/shared/apperror"
)

type mockRepo struct{}

func (m *mockRepo) Create(_ context.Context, input example.CreateInput) (*example.Example, error) {
	now := time.Now()
	return &example.Example{
		ID:          "11111111-1111-1111-1111-111111111111",
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (m *mockRepo) FindByID(_ context.Context, _ string) (*example.Example, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockRepo) FindAll(_ context.Context, _ example.Filter) ([]example.Example, error) {
	return []example.Example{}, nil
}

func (m *mockRepo) Update(_ context.Context, _ string, _ example.UpdateInput) (*example.Example, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func TestExampleUsecase_CreateValidation(t *testing.T) {
	uc := example.NewUsecase(&mockRepo{})

	_, err := uc.Create(context.Background(), example.CreateInput{Name: "   "})
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected validation error for empty name, got %v", err)
	}

	item, err := uc.Create(context.Background(), example.CreateInput{
		Name:        "  First Example  ",
		Description: "  Module-based Go Fiber starter  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "First Example" {
		t.Fatalf("expected trimmed name, got %q", item.Name)
	}
	if item.Description != "Module-based Go Fiber starter" {
		t.Fatalf("expected trimmed description, got %q", item.Description)
	}
}
