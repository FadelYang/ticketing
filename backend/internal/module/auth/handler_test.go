package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/jwt"
	"monorepo/backend/internal/shared/middleware"
)

func setupTestApp() *fiber.App {
	const secret = "test-secret"
	uc := NewUsecase(newMockRepo(), secret, 30*time.Minute)
	h := NewHandler(uc)

	app := fiber.New()
	group := app.Group("/api/v1/auth")
	group.Post("/register", middleware.OptionalAuth(secret), h.Register)
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

	// 5. Admin token registering a new agent -> role is set from form
	adminToken, err := jwt.GenerateToken("test-secret", 30*time.Minute, "admin-id", "admin@example.com", "Admin", user.RoleAdmin)
	if err != nil {
		t.Fatalf("generate admin token failed: %v", err)
	}
	agentBody := []byte(`{"name":"AgentBob","email":"agentbob@example.com","phonenumber":"+628999999999","role":"agent","password":"password123"}`)
	agentReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(agentBody))
	agentReq.Header.Set("Content-Type", "application/json")
	agentReq.Header.Set("Authorization", "Bearer "+adminToken)

	agentResp, err := app.Test(agentReq)
	if err != nil {
		t.Fatalf("admin register request failed: %v", err)
	}
	if agentResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", agentResp.StatusCode)
	}

	var agentParsed struct {
		Success bool          `json:"success"`
		Data    TokenResponse `json:"data"`
	}
	if err := json.NewDecoder(agentResp.Body).Decode(&agentParsed); err != nil {
		t.Fatalf("decode agent register response failed: %v", err)
	}
	if agentParsed.Data.User.Role != user.RoleAgent {
		t.Fatalf("expected role agent, got %q", agentParsed.Data.User.Role)
	}

	// 6. Non-admin registering an agent -> 403 Forbidden
	unauthAgentBody := []byte(`{"name":"AgentUnauth","email":"agentunauth@example.com","phonenumber":"+628888888888","role":"agent","password":"password123"}`)
	unauthAgentReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(unauthAgentBody))
	unauthAgentReq.Header.Set("Content-Type", "application/json")

	unauthAgentResp, err := app.Test(unauthAgentReq)
	if err != nil {
		t.Fatalf("unauth agent register request failed: %v", err)
	}
	if unauthAgentResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", unauthAgentResp.StatusCode)
	}
}
