package repository

import (
	"context"
	"errors"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// UserRepository defines persistence operations used by the application layer.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, id bson.ObjectID, update domain.UserUpdate) (*domain.User, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	Count(ctx context.Context) (int64, error)
}
