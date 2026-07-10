package service

import (
	"context"
	"errors"
	"testing"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type stubPasswordHasher struct {
	hashResult string
	hashErr    error
}

func (s stubPasswordHasher) Hash(string) (string, error) {
	return s.hashResult, s.hashErr
}

func (s stubPasswordHasher) Verify(string, string) error {
	return nil
}

type mockUserRepository struct {
	createFn func(context.Context, *domain.User) error
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.createFn(ctx, user)
}

func (m *mockUserRepository) FindByID(context.Context, bson.ObjectID) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepository) FindByEmail(context.Context, string) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepository) List(context.Context) ([]domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) Update(context.Context, bson.ObjectID, domain.UserUpdate) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepository) Delete(context.Context, bson.ObjectID) error {
	return repository.ErrUserNotFound
}

func (m *mockUserRepository) Count(context.Context) (int64, error) {
	return 0, nil
}

func TestAuthServiceRegister(t *testing.T) {
	var createdUser *domain.User
	repo := &mockUserRepository{
		createFn: func(_ context.Context, user *domain.User) error {
			createdUser = user
			return nil
		},
	}

	service := NewAuthService(repo, stubPasswordHasher{hashResult: "hashed-password"})

	user, err := service.Register(context.Background(), RegisterInput{
		Name:     "  Alice  ",
		Email:    "  ALICE@example.com  ",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if createdUser == nil {
		t.Fatal("repository Create() was not called")
	}
	if user.Name != "Alice" {
		t.Fatalf("Name = %q, want %q", user.Name, "Alice")
	}
	if user.Email != "alice@example.com" {
		t.Fatalf("Email = %q, want %q", user.Email, "alice@example.com")
	}
	if user.PasswordHash != "hashed-password" {
		t.Fatalf("PasswordHash = %q, want hashed password", user.PasswordHash)
	}
}

func TestAuthServiceRegisterValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   RegisterInput
		wantErr error
	}{
		{
			name:    "missing name",
			input:   RegisterInput{Email: "alice@example.com", Password: "password123"},
			wantErr: ErrNameRequired,
		},
		{
			name:    "invalid email",
			input:   RegisterInput{Name: "Alice", Email: "invalid", Password: "password123"},
			wantErr: ErrEmailInvalid,
		},
		{
			name:    "short password",
			input:   RegisterInput{Name: "Alice", Email: "alice@example.com", Password: "short"},
			wantErr: ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{
				createFn: func(context.Context, *domain.User) error {
					t.Fatal("repository Create() should not be called")
					return nil
				},
			}
			service := NewAuthService(repo, stubPasswordHasher{hashResult: "hashed-password"})

			_, err := service.Register(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Register() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthServiceRegisterRepositoryError(t *testing.T) {
	repo := &mockUserRepository{
		createFn: func(context.Context, *domain.User) error {
			return repository.ErrEmailAlreadyExists
		},
	}
	service := NewAuthService(repo, stubPasswordHasher{hashResult: "hashed-password"})

	_, err := service.Register(context.Background(), RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if !errors.Is(err, repository.ErrEmailAlreadyExists) {
		t.Fatalf("Register() error = %v, want %v", err, repository.ErrEmailAlreadyExists)
	}
}
