package service

import (
	"context"
	"errors"
	"strings"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/security"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrUserIDInvalid  = errors.New("user ID is invalid")
	ErrUpdateRequired = errors.New("at least one of name or email is required")
)

type CreateUserInput struct {
	Name     string
	Email    string
	Password string
}

type UpdateUserInput struct {
	Name  *string
	Email *string
}

// UserService contains user-management business rules independently of HTTP and MongoDB.
type UserService struct {
	users          repository.UserRepository
	passwordHasher security.PasswordHasher
}

func NewUserService(users repository.UserRepository, passwordHasher security.PasswordHasher) *UserService {
	return &UserService{
		users:          users,
		passwordHasher: passwordHasher,
	}
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*domain.User, error) {
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

func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	objectID, err := parseUserID(id)
	if err != nil {
		return nil, err
	}

	return s.users.FindByID(ctx, objectID)
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.users.List(ctx)
}

func (s *UserService) Update(ctx context.Context, id string, input UpdateUserInput) (*domain.User, error) {
	objectID, err := parseUserID(id)
	if err != nil {
		return nil, err
	}
	if input.Name == nil && input.Email == nil {
		return nil, ErrUpdateRequired
	}

	update := domain.UserUpdate{}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, ErrNameRequired
		}
		update.Name = &name
	}
	if input.Email != nil {
		email := normalizeEmail(*input.Email)
		if !isValidEmail(email) {
			return nil, ErrEmailInvalid
		}
		update.Email = &email
	}

	return s.users.Update(ctx, objectID, update)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	objectID, err := parseUserID(id)
	if err != nil {
		return err
	}

	return s.users.Delete(ctx, objectID)
}

func parseUserID(id string) (bson.ObjectID, error) {
	objectID, err := bson.ObjectIDFromHex(strings.TrimSpace(id))
	if err != nil {
		return bson.NilObjectID, ErrUserIDInvalid
	}

	return objectID, nil
}
