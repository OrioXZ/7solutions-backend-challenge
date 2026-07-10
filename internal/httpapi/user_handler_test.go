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

type userManagementServiceStub struct {
	create func(context.Context, service.CreateUserInput) (*domain.User, error)
	get    func(context.Context, string) (*domain.User, error)
	list   func(context.Context) ([]domain.User, error)
	update func(context.Context, string, service.UpdateUserInput) (*domain.User, error)
	delete func(context.Context, string) error
}

func (s userManagementServiceStub) Create(ctx context.Context, input service.CreateUserInput) (*domain.User, error) {
	return s.create(ctx, input)
}
func (s userManagementServiceStub) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.get(ctx, id)
}
func (s userManagementServiceStub) List(ctx context.Context) ([]domain.User, error) {
	return s.list(ctx)
}
func (s userManagementServiceStub) Update(ctx context.Context, id string, input service.UpdateUserInput) (*domain.User, error) {
	return s.update(ctx, id, input)
}
func (s userManagementServiceStub) Delete(ctx context.Context, id string) error {
	return s.delete(ctx, id)
}

func TestUserHandlerCreateReturnsCreatedUserWithoutPassword(t *testing.T) {
	userID := bson.NewObjectID()
	createdAt := time.Date(2026, time.July, 10, 14, 0, 0, 0, time.UTC)
	stub := userManagementServiceStub{
		create: func(_ context.Context, input service.CreateUserInput) (*domain.User, error) {
			if input.Name != "Alice" || input.Email != "alice@example.com" || input.Password != "password123" {
				t.Fatalf("unexpected input: %+v", input)
			}
			return &domain.User{
				ID:           userID,
				Name:         "Alice",
				Email:        "alice@example.com",
				PasswordHash: "must-not-leak",
				CreatedAt:    createdAt,
			}, nil
		},
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"password123"}`))
	response := httptest.NewRecorder()
	NewUserHandler(stub).Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["id"] != userID.Hex() {
		t.Fatalf("id = %#v, want %q", data["id"], userID.Hex())
	}
	if _, exists := data["password_hash"]; exists {
		t.Fatal("password hash must not be exposed")
	}
}

func TestUserHandlerListReturnsUsers(t *testing.T) {
	stub := userManagementServiceStub{
		list: func(context.Context) ([]domain.User, error) {
			return []domain.User{
				{ID: bson.NewObjectID(), Name: "Alice", Email: "alice@example.com"},
				{ID: bson.NewObjectID(), Name: "Bob", Email: "bob@example.com"},
			}, nil
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	response := httptest.NewRecorder()
	NewUserHandler(stub).List(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var body struct {
		Data []userResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 2 {
		t.Fatalf("len(data) = %d, want 2", len(body.Data))
	}
}

func TestUserHandlerGetByIDMapsNotFound(t *testing.T) {
	stub := userManagementServiceStub{
		get: func(context.Context, string) (*domain.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/users/missing", nil)
	request.SetPathValue("id", "missing")
	response := httptest.NewRecorder()
	NewUserHandler(stub).GetByID(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestUserHandlerUpdatePassesOptionalFields(t *testing.T) {
	userID := bson.NewObjectID()
	stub := userManagementServiceStub{
		update: func(_ context.Context, id string, input service.UpdateUserInput) (*domain.User, error) {
			if id != userID.Hex() {
				t.Fatalf("id = %q, want %q", id, userID.Hex())
			}
			if input.Name == nil || *input.Name != "Updated" || input.Email != nil {
				t.Fatalf("unexpected update input: %+v", input)
			}
			return &domain.User{ID: userID, Name: "Updated", Email: "alice@example.com"}, nil
		},
	}

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+userID.Hex(), strings.NewReader(`{"name":"Updated"}`))
	request.SetPathValue("id", userID.Hex())
	response := httptest.NewRecorder()
	NewUserHandler(stub).Update(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestUserHandlerDeleteReturnsNoContent(t *testing.T) {
	userID := bson.NewObjectID()
	stub := userManagementServiceStub{
		delete: func(_ context.Context, id string) error {
			if id != userID.Hex() {
				t.Fatalf("id = %q, want %q", id, userID.Hex())
			}
			return nil
		},
	}

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+userID.Hex(), nil)
	request.SetPathValue("id", userID.Hex())
	response := httptest.NewRecorder()
	NewUserHandler(stub).Delete(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
