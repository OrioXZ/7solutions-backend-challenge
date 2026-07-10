package service

import (
	"context"
	"errors"
	"testing"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type userRepositoryStub struct {
	create func(context.Context, *domain.User) error
	get    func(context.Context, bson.ObjectID) (*domain.User, error)
	list   func(context.Context) ([]domain.User, error)
	update func(context.Context, bson.ObjectID, domain.UserUpdate) (*domain.User, error)
	delete func(context.Context, bson.ObjectID) error
}

func (s userRepositoryStub) Create(ctx context.Context, user *domain.User) error {
	if s.create == nil {
		return nil
	}
	return s.create(ctx, user)
}
func (s userRepositoryStub) FindByID(ctx context.Context, id bson.ObjectID) (*domain.User, error) {
	if s.get == nil {
		return nil, repository.ErrUserNotFound
	}
	return s.get(ctx, id)
}
func (s userRepositoryStub) FindByEmail(context.Context, string) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}
func (s userRepositoryStub) List(ctx context.Context) ([]domain.User, error) {
	if s.list == nil {
		return []domain.User{}, nil
	}
	return s.list(ctx)
}
func (s userRepositoryStub) Update(ctx context.Context, id bson.ObjectID, update domain.UserUpdate) (*domain.User, error) {
	if s.update == nil {
		return nil, repository.ErrUserNotFound
	}
	return s.update(ctx, id, update)
}
func (s userRepositoryStub) Delete(ctx context.Context, id bson.ObjectID) error {
	if s.delete == nil {
		return repository.ErrUserNotFound
	}
	return s.delete(ctx, id)
}
func (s userRepositoryStub) Count(context.Context) (int64, error) { return 0, nil }

type userPasswordHasherStub struct {
	hash string
	err  error
}

func (s userPasswordHasherStub) Hash(string) (string, error) { return s.hash, s.err }
func (s userPasswordHasherStub) Verify(string, string) error { return nil }

func TestUserServiceCreateNormalizesAndHashesInput(t *testing.T) {
	var created *domain.User
	repo := userRepositoryStub{
		create: func(_ context.Context, user *domain.User) error {
			created = user
			return nil
		},
	}
	service := NewUserService(repo, userPasswordHasherStub{hash: "hashed-password"})

	user, err := service.Create(context.Background(), CreateUserInput{
		Name:     "  Alice  ",
		Email:    "  ALICE@example.com  ",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created == nil {
		t.Fatal("repository Create() was not called")
	}
	if user.Name != "Alice" || user.Email != "alice@example.com" {
		t.Fatalf("unexpected normalized user: %+v", user)
	}
	if user.PasswordHash != "hashed-password" {
		t.Fatalf("PasswordHash = %q, want hashed-password", user.PasswordHash)
	}
}

func TestUserServiceGetByIDRejectsInvalidID(t *testing.T) {
	service := NewUserService(userRepositoryStub{}, userPasswordHasherStub{})

	_, err := service.GetByID(context.Background(), "invalid")
	if !errors.Is(err, ErrUserIDInvalid) {
		t.Fatalf("GetByID() error = %v, want %v", err, ErrUserIDInvalid)
	}
}

func TestUserServiceListReturnsRepositoryUsers(t *testing.T) {
	expected := []domain.User{{ID: bson.NewObjectID(), Name: "Alice"}}
	service := NewUserService(userRepositoryStub{
		list: func(context.Context) ([]domain.User, error) { return expected, nil },
	}, userPasswordHasherStub{})

	users, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(users) != 1 || users[0].ID != expected[0].ID {
		t.Fatalf("List() = %+v, want %+v", users, expected)
	}
}

func TestUserServiceUpdateNormalizesFields(t *testing.T) {
	id := bson.NewObjectID()
	name := "  Bob  "
	email := "  BOB@example.com  "
	service := NewUserService(userRepositoryStub{
		update: func(_ context.Context, gotID bson.ObjectID, update domain.UserUpdate) (*domain.User, error) {
			if gotID != id {
				t.Fatalf("ID = %s, want %s", gotID.Hex(), id.Hex())
			}
			if update.Name == nil || *update.Name != "Bob" {
				t.Fatalf("Name update = %#v", update.Name)
			}
			if update.Email == nil || *update.Email != "bob@example.com" {
				t.Fatalf("Email update = %#v", update.Email)
			}
			return &domain.User{ID: id, Name: *update.Name, Email: *update.Email}, nil
		},
	}, userPasswordHasherStub{})

	user, err := service.Update(context.Background(), id.Hex(), UpdateUserInput{Name: &name, Email: &email})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if user.Name != "Bob" || user.Email != "bob@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserServiceUpdateRequiresField(t *testing.T) {
	service := NewUserService(userRepositoryStub{}, userPasswordHasherStub{})

	_, err := service.Update(context.Background(), bson.NewObjectID().Hex(), UpdateUserInput{})
	if !errors.Is(err, ErrUpdateRequired) {
		t.Fatalf("Update() error = %v, want %v", err, ErrUpdateRequired)
	}
}

func TestUserServiceDeletePassesParsedID(t *testing.T) {
	id := bson.NewObjectID()
	called := false
	service := NewUserService(userRepositoryStub{
		delete: func(_ context.Context, gotID bson.ObjectID) error {
			called = true
			if gotID != id {
				t.Fatalf("ID = %s, want %s", gotID.Hex(), id.Hex())
			}
			return nil
		},
	}, userPasswordHasherStub{})

	if err := service.Delete(context.Background(), id.Hex()); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !called {
		t.Fatal("repository Delete() was not called")
	}
}
