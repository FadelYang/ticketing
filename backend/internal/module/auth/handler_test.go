package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/middleware"
)

func setupTestApp() *fiber.App {
	const secret = "test-secret"
	uc := NewUsecase(newMockRepo(), secret, 30*time.Minute)
	h := NewHandler(uc)

	app := fiber.New()
	group := app.Group("/api/v1/auth")
	group.Post("/register", h.Register)
	group.Post("/login", h.Login)
	group.Post("/logout", middleware.RequireAuth(secret), h.Logout)
	return app
}

func TestAuthHTTPFlow(t *testing.T) {
	app := setupTestApp()

	// 1. Register
	regBody := []byte(`{"name":"Alice","email":"alice@example.com","phonenumber":"+628123456789","password":"password123"}`)
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")

	regResp, err := app.Test(regReq)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	if regResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", regResp.StatusCode)
	}

	// 2. Login
	loginBody := []byte(`{"email":"alice@example.com","password":"password123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", loginResp.StatusCode)
	}

	var parsed struct {
		Success bool          `json:"success"`
		Data    TokenResponse `json:"data"`
	}
	if err := json.NewDecoder(loginResp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode login response failed: %v", err)
	}
	if !parsed.Success || parsed.Data.AccessToken == "" || parsed.Data.ExpiresIn != 1800 {
		t.Fatalf("unexpected login payload: %+v", parsed)
	}

	// 3. Logout without token -> 401
	unauthLogoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	unauthLogoutResp, err := app.Test(unauthLogoutReq)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	if unauthLogoutResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", unauthLogoutResp.StatusCode)
	}

	// 4. Logout with valid token -> 200
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+parsed.Data.AccessToken)
	logoutResp, err := app.Test(logoutReq)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	if logoutResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", logoutResp.StatusCode)
	}
}
