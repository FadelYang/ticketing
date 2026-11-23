package user

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

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
	input.Role = strings.ToLower(strings.TrimSpace(input.Role))
	if input.Role == "" {
		input.Role = RoleRequester
	}

	if err := validateUserFields(input.Name, input.Email, input.PhoneNumber); err != nil {
		return nil, err
	}
	if !IsValidRole(input.Role) {
		return nil, fmt.Errorf("%w: role must be one of requester, agent, admin", apperror.ErrValidation)
	}

	if input.CreatedBy != nil {
		trimmed := strings.TrimSpace(*input.CreatedBy)
		if trimmed == "" {
			input.CreatedBy = nil
		} else {
			input.CreatedBy = &trimmed
		}
	}

	if strings.TrimSpace(input.Password) != "" {
		if len(input.Password) < 8 {
			return nil, fmt.Errorf("%w: password must be at least 8 characters", apperror.ErrValidation)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		hashStr := string(hash)
		input.PasswordHash = &hashStr
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
	input.Role = strings.ToLower(strings.TrimSpace(input.Role))

	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	if err := validateUserFields(input.Name, input.Email, input.PhoneNumber); err != nil {
		return nil, err
	}
	if input.Role != "" && !IsValidRole(input.Role) {
		return nil, fmt.Errorf("%w: role must be one of requester, agent, admin", apperror.ErrValidation)
	}

	return u.repo.Update(ctx, id, input)
}

func (u *usecase) UpdateRole(ctx context.Context, id string, role string) (*User, error) {
	id = strings.TrimSpace(id)
	role = strings.ToLower(strings.TrimSpace(role))

	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	if !IsValidRole(role) {
		return nil, fmt.Errorf("%w: role must be one of requester, agent, admin", apperror.ErrValidation)
	}

	return u.repo.UpdateRole(ctx, id, role)
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
