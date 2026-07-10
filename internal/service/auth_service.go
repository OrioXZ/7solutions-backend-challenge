package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/security"
)

var (
	ErrNameRequired       = errors.New("name is required")
	ErrEmailInvalid       = errors.New("email is invalid")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
}

type AuthService struct {
	users          repository.UserRepository
	passwordHasher security.PasswordHasher
	tokenIssuer    security.TokenIssuer
}

func NewAuthService(
	users repository.UserRepository,
	passwordHasher security.PasswordHasher,
	tokenIssuer security.TokenIssuer,
) *AuthService {
	return &AuthService{
		users:          users,
		passwordHasher: passwordHasher,
		tokenIssuer:    tokenIssuer,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrNameRequired
	}

	email := normalizeEmail(input.Email)
	if !isValidEmail(email) {
		return nil, ErrEmailInvalid
	}

	if len(input.Password) < 8 {
		return nil, ErrPasswordTooShort
	}

	passwordHash, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	email := normalizeEmail(input.Email)
	if !isValidEmail(email) || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := s.passwordHasher.Verify(user.PasswordHash, input.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, expiresAt, err := s.tokenIssuer.Issue(user.ID.Hex())
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	parsedEmail, err := mail.ParseAddress(email)
	return err == nil && parsedEmail.Address == email
}
