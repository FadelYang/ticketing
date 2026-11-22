package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
	"monorepo/backend/internal/shared/jwt"
)

type usecase struct {
	repo      Repository
	jwtSecret string
	jwtTTL    time.Duration
}

func NewUsecase(repo Repository, jwtSecret string, jwtTTL time.Duration) Usecase {
	return &usecase{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
	}
}

func (u *usecase) Register(ctx context.Context, input RegisterInput) (*TokenResponse, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)

	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	createdUser, err := u.repo.CreateUser(ctx, input.Name, input.Email, input.PhoneNumber, string(hash))
	if err != nil {
		return nil, err
	}

	return u.buildTokenResponse(createdUser)
}

func (u *usecase) Login(ctx context.Context, input LoginInput) (*TokenResponse, error) {
	input.Email = strings.TrimSpace(input.Email)

	if err := validateEmail(input.Email); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, fmt.Errorf("%w: password is required", apperror.ErrValidation)
	}

	found, err := u.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, fmt.Errorf("%w: invalid email or password", apperror.ErrUnauthorized)
		}
		return nil, err
	}

	if found.PasswordHash == "" {
		return nil, fmt.Errorf("%w: invalid email or password", apperror.ErrUnauthorized)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(found.PasswordHash), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("%w: invalid email or password", apperror.ErrUnauthorized)
	}

	return u.buildTokenResponse(&found.User)
}

func (u *usecase) Logout(_ context.Context) (*LogoutResponse, error) {
	return &LogoutResponse{
		Message: "logged out successfully",
	}, nil
}

func (u *usecase) buildTokenResponse(usr *user.User) (*TokenResponse, error) {
	token, err := jwt.GenerateToken(u.jwtSecret, u.jwtTTL, usr.ID, usr.Email, usr.Name)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(u.jwtTTL.Seconds()),
		User:        *usr,
	}, nil
}

func validateRegisterInput(input RegisterInput) error {
	if input.Name == "" {
		return fmt.Errorf("%w: name is required", apperror.ErrValidation)
	}
	if len(input.Name) > 50 {
		return fmt.Errorf("%w: name must be at most 50 characters", apperror.ErrValidation)
	}

	if err := validateEmail(input.Email); err != nil {
		return err
	}

	if input.PhoneNumber == "" {
		return fmt.Errorf("%w: phonenumber is required", apperror.ErrValidation)
	}
	if len(input.PhoneNumber) > 50 {
		return fmt.Errorf("%w: phonenumber must be at most 50 characters", apperror.ErrValidation)
	}

	if len(input.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", apperror.ErrValidation)
	}
	if len([]byte(input.Password)) > 72 {
		return fmt.Errorf("%w: password must be at most 72 bytes", apperror.ErrValidation)
	}

	return nil
}

func validateEmail(email string) error {
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
	return nil
}
