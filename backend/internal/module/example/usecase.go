package example

import (
	"context"
	"fmt"
	"strings"

	"monorepo/backend/internal/shared/apperror"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Create(ctx context.Context, input CreateInput) (*Example, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", apperror.ErrValidation)
	}
	if len(input.Name) > 200 {
		return nil, fmt.Errorf("%w: name must be at most 200 characters", apperror.ErrValidation)
	}

	return u.repo.Create(ctx, input)
}

func (u *usecase) GetByID(ctx context.Context, id string) (*Example, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	return u.repo.FindByID(ctx, id)
}

func (u *usecase) List(ctx context.Context, filter Filter) ([]Example, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	return u.repo.FindAll(ctx, filter)
}

func (u *usecase) Update(ctx context.Context, id string, input UpdateInput) (*Example, error) {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", apperror.ErrValidation)
	}
	if len(input.Name) > 200 {
		return nil, fmt.Errorf("%w: name must be at most 200 characters", apperror.ErrValidation)
	}

	return u.repo.Update(ctx, id, input)
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	return u.repo.Delete(ctx, id)
}
