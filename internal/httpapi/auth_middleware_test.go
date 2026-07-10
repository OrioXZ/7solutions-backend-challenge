package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/security"
)

type tokenValidatorStub struct {
	validate func(token string) (*security.TokenClaims, error)
}

func (s tokenValidatorStub) Validate(token string) (*security.TokenClaims, error) {
	return s.validate(token)
}

func TestAuthenticationMiddlewareRejectsMissingAuthorizationHeader(t *testing.T) {
	called := false
	validator := tokenValidatorStub{
		validate: func(string) (*security.TokenClaims, error) {
			t.Fatal("validator must not be called")
			return nil, nil
		},
	}

	handler := AuthenticationMiddleware(validator)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatal("protected handler must not be called")
	}
}

func TestAuthenticationMiddlewareRejectsMalformedBearerHeader(t *testing.T) {
	validator := tokenValidatorStub{
		validate: func(string) (*security.TokenClaims, error) {
			t.Fatal("validator must not be called")
			return nil, nil
		},
	}

	handler := AuthenticationMiddleware(validator)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler must not be called")
	}))

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Token abc")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticationMiddlewareRejectsInvalidToken(t *testing.T) {
	validator := tokenValidatorStub{
		validate: func(token string) (*security.TokenClaims, error) {
			if token != "invalid-token" {
				t.Fatalf("token = %q, want %q", token, "invalid-token")
			}
			return nil, security.ErrTokenInvalid
		},
	}

	handler := AuthenticationMiddleware(validator)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler must not be called")
	}))

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"UNAUTHORIZED"`) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestAuthenticationMiddlewareStoresAuthenticatedSubject(t *testing.T) {
	validator := tokenValidatorStub{
		validate: func(token string) (*security.TokenClaims, error) {
			if token != "valid-token" {
				t.Fatalf("token = %q, want %q", token, "valid-token")
			}
			return &security.TokenClaims{Subject: "user-123"}, nil
		},
	}

	handler := AuthenticationMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subject, ok := AuthenticatedUserID(r.Context())
		if !ok {
			t.Fatal("authenticated subject missing from context")
		}
		if subject != "user-123" {
			t.Fatalf("subject = %q, want %q", subject, "user-123")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestAuthenticationMiddlewareHidesValidatorErrorDetails(t *testing.T) {
	validator := tokenValidatorStub{
		validate: func(string) (*security.TokenClaims, error) {
			return nil, errors.New("sensitive validation detail")
		},
	}

	handler := AuthenticationMiddleware(validator)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler must not be called")
	}))

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if strings.Contains(response.Body.String(), "sensitive validation detail") {
		t.Fatalf("response leaked validator error: %s", response.Body.String())
	}
}
