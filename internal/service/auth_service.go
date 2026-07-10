package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/security"
)

var (
	ErrNameRequired     = errors.New("name is required")
	ErrEmailInvalid     = errors.New("email is invalid")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type AuthService struct {
	users          repository.UserRepository
	passwordHasher security.PasswordHasher
}

func NewAuthService(
	users repository.UserRepository,
	passwordHasher security.PasswordHasher,
) *AuthService {
	return &AuthService{
		users:          users,
		passwordHasher: passwordHasher,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrNameRequired
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
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
