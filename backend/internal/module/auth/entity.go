package auth

import (
	"context"

	"monorepo/backend/internal/module/user"
)

type RegisterInput struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phonenumber"`
	Password    string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	User        user.User `json:"user"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type userWithPassword struct {
	User         user.User
	PasswordHash string
}

type Repository interface {
	CreateUser(ctx context.Context, name, email, phoneNumber, passwordHash string) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*userWithPassword, error)
}

type Usecase interface {
	Register(ctx context.Context, input RegisterInput) (*TokenResponse, error)
	Login(ctx context.Context, input LoginInput) (*TokenResponse, error)
	Logout(ctx context.Context) (*LogoutResponse, error)
}
