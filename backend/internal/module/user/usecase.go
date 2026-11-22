package user

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"monorepo/backend/internal/shared/apperror"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Create(ctx context.Context, input CreateInput) (*User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)

	if err := validateUserFields(input.Name, input.Email, input.PhoneNumber); err != nil {
		return nil, err
	}

	if input.CreatedBy != nil {
		trimmed := strings.TrimSpace(*input.CreatedBy)
		if trimmed == "" {
			input.CreatedBy = nil
		} else {
			input.CreatedBy = &trimmed
		}
	}

	return u.repo.Create(ctx, input)
}

func (u *usecase) GetByID(ctx context.Context, id string) (*User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	return u.repo.FindByID(ctx, id)
}

func (u *usecase) List(ctx context.Context, filter Filter) ([]User, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	return u.repo.FindAll(ctx, filter)
}

func (u *usecase) Update(ctx context.Context, id string, input UpdateInput) (*User, error) {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)

	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	if err := validateUserFields(input.Name, input.Email, input.PhoneNumber); err != nil {
		return nil, err
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

func validateUserFields(name, email, phoneNumber string) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", apperror.ErrValidation)
	}
	if len(name) > 50 {
		return fmt.Errorf("%w: name must be at most 50 characters", apperror.ErrValidation)
	}

	if email == "" {
		return fmt.Errorf("%w: email is required", apperror.ErrValidation)
	}
	if len(email) > 50 {
		return fmt.Errorf("%w: email must be at most 50 characters", apperror.ErrValidation)
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return fmt.Errorf("%w: email must be a valid email address", apperror.ErrValidation)
	}

	if phoneNumber == "" {
		return fmt.Errorf("%w: phonenumber is required", apperror.ErrValidation)
	}
	if len(phoneNumber) > 50 {
		return fmt.Errorf("%w: phonenumber must be at most 50 characters", apperror.ErrValidation)
	}

	return nil
}
