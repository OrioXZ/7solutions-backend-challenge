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

type registrationServiceStub struct {
	register func(ctx context.Context, input service.RegisterInput) (*domain.User, error)
}

func (s registrationServiceStub) Register(
	ctx context.Context,
	input service.RegisterInput,
) (*domain.User, error) {
	return s.register(ctx, input)
}

func TestAuthHandlerRegisterSuccess(t *testing.T) {
	createdAt := time.Date(2026, time.July, 10, 12, 0, 0, 0, time.UTC)
	userID := bson.NewObjectID()

	stub := registrationServiceStub{
		register: func(_ context.Context, input service.RegisterInput) (*domain.User, error) {
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
	stub := registrationServiceStub{
		register: func(_ context.Context, _ service.RegisterInput) (*domain.User, error) {
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
	stub := registrationServiceStub{
		register: func(_ context.Context, _ service.RegisterInput) (*domain.User, error) {
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
