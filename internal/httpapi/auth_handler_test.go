package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/service"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type authenticationServiceStub struct {
	registerFn func(ctx context.Context, input service.RegisterInput) (*domain.User, error)
	loginFn    func(ctx context.Context, input service.LoginInput) (*service.LoginResult, error)
}

func (s authenticationServiceStub) Register(
	ctx context.Context,
	input service.RegisterInput,
) (*domain.User, error) {
	return s.registerFn(ctx, input)
}

func (s authenticationServiceStub) Login(
	ctx context.Context,
	input service.LoginInput,
) (*service.LoginResult, error) {
	return s.loginFn(ctx, input)
}

func TestAuthHandlerRegisterSuccess(t *testing.T) {
	createdAt := time.Date(2026, time.July, 10, 12, 0, 0, 0, time.UTC)
	userID := bson.NewObjectID()

	stub := authenticationServiceStub{
		registerFn: func(_ context.Context, input service.RegisterInput) (*domain.User, error) {
			if input.Name != "Alice" || input.Email != "alice@example.com" || input.Password != "password123" {
				t.Fatalf("unexpected register input: %+v", input)
			}

			return &domain.User{
				ID:           userID,
				Name:         "Alice",
				Email:        "alice@example.com",
				PasswordHash: "must-not-be-returned",
				CreatedAt:    createdAt,
			}, nil
		},
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/register",
		strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"password123"}`),
	)
	response := httptest.NewRecorder()

	NewAuthHandler(stub).Register(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", body["data"])
	}
	if data["id"] != userID.Hex() {
		t.Fatalf("expected id %q, got %#v", userID.Hex(), data["id"])
	}
	if _, exists := data["password"]; exists {
		t.Fatal("response must not expose password")
	}
	if _, exists := data["password_hash"]; exists {
		t.Fatal("response must not expose password hash")
	}
}

func TestAuthHandlerRegisterDuplicateEmail(t *testing.T) {
	stub := authenticationServiceStub{
		registerFn: func(_ context.Context, _ service.RegisterInput) (*domain.User, error) {
			return nil, repository.ErrEmailAlreadyExists
		},
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/register",
		strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"password123"}`),
	)
	response := httptest.NewRecorder()

	NewAuthHandler(stub).Register(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, response.Code)
	}
	if !strings.Contains(response.Body.String(), `"code":"EMAIL_ALREADY_EXISTS"`) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestAuthHandlerRegisterRejectsInvalidJSON(t *testing.T) {
	called := false
	stub := authenticationServiceStub{
		registerFn: func(_ context.Context, _ service.RegisterInput) (*domain.User, error) {
			called = true
			return nil, nil
		},
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/register",
		strings.NewReader(`{"name":`),
	)
	response := httptest.NewRecorder()

	NewAuthHandler(stub).Register(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if called {
		t.Fatal("registration service must not be called for invalid JSON")
	}
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	expiresAt := time.Date(2026, time.July, 11, 13, 0, 0, 0, time.UTC)
	stub := authenticationServiceStub{
		loginFn: func(_ context.Context, input service.LoginInput) (*service.LoginResult, error) {
			if input.Email != "alice@example.com" || input.Password != "password123" {
				t.Fatalf("unexpected login input: %+v", input)
			}
			return &service.LoginResult{
				AccessToken: "signed.jwt.token",
				TokenType:   "Bearer",
				ExpiresAt:   expiresAt,
			}, nil
		},
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{"email":"alice@example.com","password":"password123"}`),
	)
	response := httptest.NewRecorder()

	NewAuthHandler(stub).Login(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body struct {
		Data loginResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.AccessToken != "signed.jwt.token" {
		t.Fatalf("access token = %q, want signed.jwt.token", body.Data.AccessToken)
	}
	if body.Data.TokenType != "Bearer" {
		t.Fatalf("token type = %q, want Bearer", body.Data.TokenType)
	}
	if !body.Data.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expires at = %v, want %v", body.Data.ExpiresAt, expiresAt)
	}
}

func TestAuthHandlerLoginInvalidCredentials(t *testing.T) {
	stub := authenticationServiceStub{
		loginFn: func(_ context.Context, _ service.LoginInput) (*service.LoginResult, error) {
			return nil, service.ErrInvalidCredentials
		},
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{"email":"alice@example.com","password":"wrong-password"}`),
	)
	response := httptest.NewRecorder()

	NewAuthHandler(stub).Login(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
	if !strings.Contains(response.Body.String(), `"code":"INVALID_CREDENTIALS"`) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestAuthHandlerLoginRejectsInvalidJSON(t *testing.T) {
	called := false
	stub := authenticationServiceStub{
		loginFn: func(_ context.Context, _ service.LoginInput) (*service.LoginResult, error) {
			called = true
			return nil, nil
		},
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{"email":`),
	)
	response := httptest.NewRecorder()

	NewAuthHandler(stub).Login(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if called {
		t.Fatal("login service must not be called for invalid JSON")
	}
}
