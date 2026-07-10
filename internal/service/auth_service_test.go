package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type stubPasswordHasher struct {
	hashResult string
	hashErr    error
	verifyErr  error
	verifyFn   func(hash, password string) error
}

func (s stubPasswordHasher) Hash(string) (string, error) {
	return s.hashResult, s.hashErr
}

func (s stubPasswordHasher) Verify(hash, password string) error {
	if s.verifyFn != nil {
		return s.verifyFn(hash, password)
	}
	return s.verifyErr
}

type stubTokenIssuer struct {
	issueFn func(subject string) (string, time.Time, error)
}

func (s stubTokenIssuer) Issue(subject string) (string, time.Time, error) {
	if s.issueFn == nil {
		return "", time.Time{}, errors.New("unexpected token issue")
	}
	return s.issueFn(subject)
}

type mockUserRepository struct {
	createFn      func(context.Context, *domain.User) error
	findByEmailFn func(context.Context, string) (*domain.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, user)
}

func (m *mockUserRepository) FindByID(context.Context, bson.ObjectID) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findByEmailFn == nil {
		return nil, repository.ErrUserNotFound
	}
	return m.findByEmailFn(ctx, email)
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

	authService := NewAuthService(
		repo,
		stubPasswordHasher{hashResult: "hashed-password"},
		stubTokenIssuer{},
	)

	user, err := authService.Register(context.Background(), RegisterInput{
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
			authService := NewAuthService(
				repo,
				stubPasswordHasher{hashResult: "hashed-password"},
				stubTokenIssuer{},
			)

			_, err := authService.Register(context.Background(), tt.input)
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
	authService := NewAuthService(
		repo,
		stubPasswordHasher{hashResult: "hashed-password"},
		stubTokenIssuer{},
	)

	_, err := authService.Register(context.Background(), RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if !errors.Is(err, repository.ErrEmailAlreadyExists) {
		t.Fatalf("Register() error = %v, want %v", err, repository.ErrEmailAlreadyExists)
	}
}

func TestAuthServiceLogin(t *testing.T) {
	userID := bson.NewObjectID()
	expiresAt := time.Date(2026, time.July, 11, 13, 0, 0, 0, time.UTC)

	repo := &mockUserRepository{
		findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			if email != "alice@example.com" {
				t.Fatalf("FindByEmail() email = %q, want alice@example.com", email)
			}
			return &domain.User{
				ID:           userID,
				Email:        email,
				PasswordHash: "stored-hash",
			}, nil
		},
	}

	hasher := stubPasswordHasher{
		verifyFn: func(hash, password string) error {
			if hash != "stored-hash" || password != "password123" {
				t.Fatalf("Verify() got hash=%q password=%q", hash, password)
			}
			return nil
		},
	}

	issuer := stubTokenIssuer{
		issueFn: func(subject string) (string, time.Time, error) {
			if subject != userID.Hex() {
				t.Fatalf("Issue() subject = %q, want %q", subject, userID.Hex())
			}
			return "signed.jwt.token", expiresAt, nil
		},
	}

	authService := NewAuthService(repo, hasher, issuer)
	result, err := authService.Login(context.Background(), LoginInput{
		Email:    "  ALICE@example.com  ",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.AccessToken != "signed.jwt.token" {
		t.Fatalf("AccessToken = %q, want signed.jwt.token", result.AccessToken)
	}
	if result.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want Bearer", result.TokenType)
	}
	if !result.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %v, want %v", result.ExpiresAt, expiresAt)
	}
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	tests := []struct {
		name   string
		input  LoginInput
		repo   *mockUserRepository
		hasher stubPasswordHasher
	}{
		{
			name:  "invalid email",
			input: LoginInput{Email: "invalid", Password: "password123"},
			repo: &mockUserRepository{
				findByEmailFn: func(context.Context, string) (*domain.User, error) {
					t.Fatal("repository must not be called for invalid email")
					return nil, nil
				},
			},
		},
		{
			name:  "user not found",
			input: LoginInput{Email: "alice@example.com", Password: "password123"},
			repo: &mockUserRepository{
				findByEmailFn: func(context.Context, string) (*domain.User, error) {
					return nil, repository.ErrUserNotFound
				},
			},
		},
		{
			name:  "wrong password",
			input: LoginInput{Email: "alice@example.com", Password: "wrong-password"},
			repo: &mockUserRepository{
				findByEmailFn: func(context.Context, string) (*domain.User, error) {
					return &domain.User{PasswordHash: "stored-hash"}, nil
				},
			},
			hasher: stubPasswordHasher{verifyErr: errors.New("password mismatch")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := NewAuthService(tt.repo, tt.hasher, stubTokenIssuer{})
			_, err := authService.Login(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
			}
		})
	}
}

func TestAuthServiceLoginRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repo := &mockUserRepository{
		findByEmailFn: func(context.Context, string) (*domain.User, error) {
			return nil, wantErr
		},
	}

	authService := NewAuthService(repo, stubPasswordHasher{}, stubTokenIssuer{})
	_, err := authService.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "password123",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Login() error = %v, want %v", err, wantErr)
	}
}

func TestAuthServiceLoginTokenError(t *testing.T) {
	wantErr := errors.New("token signing failed")
	repo := &mockUserRepository{
		findByEmailFn: func(context.Context, string) (*domain.User, error) {
			return &domain.User{
				ID:           bson.NewObjectID(),
				PasswordHash: "stored-hash",
			}, nil
		},
	}
	issuer := stubTokenIssuer{
		issueFn: func(string) (string, time.Time, error) {
			return "", time.Time{}, wantErr
		},
	}

	authService := NewAuthService(repo, stubPasswordHasher{}, issuer)
	_, err := authService.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "password123",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Login() error = %v, want %v", err, wantErr)
	}
}
