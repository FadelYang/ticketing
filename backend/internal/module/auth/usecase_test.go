package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
	"monorepo/backend/internal/shared/jwt"
)

type mockRepo struct {
	usersByEmail map[string]*userWithPassword
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		usersByEmail: make(map[string]*userWithPassword),
	}
}

func (m *mockRepo) CreateUser(_ context.Context, name, email, phoneNumber, role, passwordHash string) (*user.User, error) {
	key := strings.ToLower(email)
	if _, exists := m.usersByEmail[key]; exists {
		return nil, apperror.ErrConflict
	}

	now := time.Now().UTC()
	u := user.User{
		ID:          "11111111-1111-1111-1111-111111111111",
		Name:        name,
		Email:       email,
		PhoneNumber: phoneNumber,
		Role:        role,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.usersByEmail[key] = &userWithPassword{
		User:         u,
		PasswordHash: passwordHash,
	}
	return &u, nil
}

func (m *mockRepo) FindByEmail(_ context.Context, email string) (*userWithPassword, error) {
	item, ok := m.usersByEmail[strings.ToLower(email)]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	return item, nil
}

func TestRegisterAndLoginSuccess(t *testing.T) {
	const secret = "test-secret"
	uc := NewUsecase(newMockRepo(), secret, 30*time.Minute)

	regResp, err := uc.Register(context.Background(), RegisterInput{
		Name:        "  Alice  ",
		Email:       "  alice@example.com  ",
		PhoneNumber: "  +628123456789  ",
		Password:    "supersecret",
	}, "")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if regResp.TokenType != "Bearer" {
		t.Fatalf("expected Bearer token type, got %q", regResp.TokenType)
	}
	if regResp.ExpiresIn != 1800 {
		t.Fatalf("expected expires_in 1800, got %d", regResp.ExpiresIn)
	}
	if regResp.User.Name != "Alice" || regResp.User.Email != "alice@example.com" || regResp.User.Role != user.RoleRequester {
		t.Fatalf("unexpected user returned: %+v", regResp.User)
	}

	claims, err := jwt.ParseToken(secret, regResp.AccessToken)
	if err != nil {
		t.Fatalf("expected valid jwt token, got %v", err)
	}
	if claims.Subject != regResp.User.ID || claims.Email != "alice@example.com" || claims.Role != user.RoleRequester {
		t.Fatalf("unexpected jwt claims: %+v", claims)
	}

	loginResp, err := uc.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.AccessToken == "" || loginResp.ExpiresIn != 1800 {
		t.Fatalf("unexpected login response: %+v", loginResp)
	}
}

func TestRegisterValidation(t *testing.T) {
	uc := NewUsecase(newMockRepo(), "test-secret", 30*time.Minute)

	tests := []struct {
		name  string
		input RegisterInput
	}{
		{
			name:  "empty name",
			input: RegisterInput{Name: " ", Email: "a@example.com", PhoneNumber: "+6281", Password: "password123"},
		},
		{
			name:  "invalid email",
			input: RegisterInput{Name: "Alice", Email: "invalid", PhoneNumber: "+6281", Password: "password123"},
		},
		{
			name:  "empty phonenumber",
			input: RegisterInput{Name: "Alice", Email: "a@example.com", PhoneNumber: " ", Password: "password123"},
		},
		{
			name:  "short password",
			input: RegisterInput{Name: "Alice", Email: "a@example.com", PhoneNumber: "+6281", Password: "short"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Register(context.Background(), tc.input, "")
			if !errors.Is(err, apperror.ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestRegisterByAdminUsesProvidedRole(t *testing.T) {
	uc := NewUsecase(newMockRepo(), "test-secret", 30*time.Minute)

	resp, err := uc.Register(context.Background(), RegisterInput{
		Name:        "Bob",
		Email:       "bob@example.com",
		PhoneNumber: "+628111111111",
		Role:        "agent",
		Password:    "supersecret",
	}, user.RoleAdmin)
	if err != nil {
		t.Fatalf("admin register failed: %v", err)
	}
	if resp.User.Role != user.RoleAgent {
		t.Fatalf("expected role agent, got %q", resp.User.Role)
	}

	if _, err := uc.Register(context.Background(), RegisterInput{
		Name:        "Charlie",
		Email:       "charlie@example.com",
		PhoneNumber: "+628222222222",
		Role:        "invalid-role",
		Password:    "supersecret",
	}, user.RoleAdmin); !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected validation error for invalid role by admin, got %v", err)
	}
}

func TestRegisterAgentRequiresAdmin(t *testing.T) {
	uc := NewUsecase(newMockRepo(), "test-secret", 30*time.Minute)

	for _, callerRole := range []string{"", user.RoleRequester, user.RoleAgent} {
		_, err := uc.Register(context.Background(), RegisterInput{
			Name:        "AgentCandidate",
			Email:       "candidate@example.com",
			PhoneNumber: "+628333333333",
			Role:        user.RoleAgent,
			Password:    "supersecret",
		}, callerRole)
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Fatalf("expected ErrForbidden for callerRole=%q, got %v", callerRole, err)
		}
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	uc := NewUsecase(newMockRepo(), "test-secret", 30*time.Minute)

	_, err := uc.Register(context.Background(), RegisterInput{
		Name:        "Alice",
		Email:       "alice@example.com",
		PhoneNumber: "+628123456789",
		Password:    "supersecret",
	}, "")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if _, err := uc.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "wrong-password",
	}); !errors.Is(err, apperror.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized for wrong password, got %v", err)
	}

	if _, err := uc.Login(context.Background(), LoginInput{
		Email:    "unknown@example.com",
		Password: "supersecret",
	}); !errors.Is(err, apperror.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized for unknown email, got %v", err)
	}
}
