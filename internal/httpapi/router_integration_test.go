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
	"github.com/OrioXZ/7solutions-backend-challenge/internal/security"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/service"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type routerHealthCheckerStub struct {
	ping func(context.Context) error
}

func (s routerHealthCheckerStub) Ping(ctx context.Context) error {
	if s.ping == nil {
		return nil
	}
	return s.ping(ctx)
}

func TestRouterKeepsHealthAndAuthenticationRoutesPublic(t *testing.T) {
	userID := bson.NewObjectID()
	authService := authenticationServiceStub{
		registerFn: func(context.Context, service.RegisterInput) (*domain.User, error) {
			return &domain.User{
				ID:        userID,
				Name:      "Alice",
				Email:     "alice@example.com",
				CreatedAt: time.Now().UTC(),
			}, nil
		},
		loginFn: func(context.Context, service.LoginInput) (*service.LoginResult, error) {
			return &service.LoginResult{
				AccessToken: "header.payload.signature",
				TokenType:   "Bearer",
				ExpiresAt:   time.Now().UTC().Add(time.Hour),
			}, nil
		},
	}
	validator := tokenValidatorStub{
		validate: func(string) (*security.TokenClaims, error) {
			t.Fatal("public routes must not validate a token")
			return nil, nil
		},
	}
	router := NewRouter(routerHealthCheckerStub{}, authService, userManagementServiceStub{}, validator)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{
			name:       "register",
			method:     http.MethodPost,
			path:       "/api/v1/auth/register",
			body:       `{"name":"Alice","email":"alice@example.com","password":"password123"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "login",
			method:     http.MethodPost,
			path:       "/api/v1/auth/login",
			body:       `{"email":"alice@example.com","password":"password123"}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}
}

func TestRouterProtectsEveryUserRoute(t *testing.T) {
	validator := tokenValidatorStub{
		validate: func(string) (*security.TokenClaims, error) {
			t.Fatal("validator must not run when Authorization header is missing")
			return nil, nil
		},
	}
	router := NewRouter(
		routerHealthCheckerStub{},
		authenticationServiceStub{},
		userManagementServiceStub{},
		validator,
	)

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/users"},
		{method: http.MethodGet, path: "/api/v1/users"},
		{method: http.MethodGet, path: "/api/v1/users/507f1f77bcf86cd799439011"},
		{method: http.MethodPatch, path: "/api/v1/users/507f1f77bcf86cd799439011"},
		{method: http.MethodDelete, path: "/api/v1/users/507f1f77bcf86cd799439011"},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.path, nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want %d", test.method, test.path, response.Code, http.StatusUnauthorized)
		}
	}
}

func TestRouterAllowsValidTokenToReachUserList(t *testing.T) {
	userID := bson.NewObjectID()
	users := userManagementServiceStub{
		list: func(context.Context) ([]domain.User, error) {
			return []domain.User{
				{
					ID:           userID,
					Name:         "Alice",
					Email:        "alice@example.com",
					PasswordHash: "must-not-leak",
					CreatedAt:    time.Now().UTC(),
				},
			}, nil
		},
	}
	validator := tokenValidatorStub{
		validate: func(token string) (*security.TokenClaims, error) {
			if token != "valid-token" {
				t.Fatalf("token = %q, want valid-token", token)
			}
			return &security.TokenClaims{Subject: userID.Hex()}, nil
		},
	}
	router := NewRouter(routerHealthCheckerStub{}, authenticationServiceStub{}, users, validator)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 || body.Data[0]["id"] != userID.Hex() {
		t.Fatalf("unexpected response: %#v", body.Data)
	}
	if _, exists := body.Data[0]["password"]; exists {
		t.Fatal("response exposed password")
	}
	if _, exists := body.Data[0]["password_hash"]; exists {
		t.Fatal("response exposed password hash")
	}
}

func TestRouterRejectsInvalidBearerTokenBeforeUserHandler(t *testing.T) {
	called := false
	users := userManagementServiceStub{
		list: func(context.Context) ([]domain.User, error) {
			called = true
			return nil, nil
		},
	}
	validator := tokenValidatorStub{
		validate: func(string) (*security.TokenClaims, error) {
			return nil, security.ErrTokenInvalid
		},
	}
	router := NewRouter(routerHealthCheckerStub{}, authenticationServiceStub{}, users, validator)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatal("user handler was called for an invalid token")
	}
}
