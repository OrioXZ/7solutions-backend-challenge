package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const usersCollection = "users"

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(database *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{
		collection: database.Collection(usersCollection),
	}
}

func (r *MongoUserRepository) EnsureIndexes(ctx context.Context) error {
	model := mongo.IndexModel{
		Keys: bson.D{{Key: "email", Value: 1}},
		Options: options.Index().
			SetName("uniq_users_email").
			SetUnique(true),
	}

	if _, err := r.collection.Indexes().CreateOne(ctx, model); err != nil {
		return fmt.Errorf("create unique email index: %w", err)
	}

	return nil
}

func (r *MongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	if _, err := r.collection.InsertOne(ctx, user); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrEmailAlreadyExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *MongoUserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.User, error) {
	return r.findOne(ctx, bson.D{{Key: "_id", Value: id}})
}

func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findOne(ctx, bson.D{{Key: "email", Value: email}})
}

func (r *MongoUserRepository) List(ctx context.Context) ([]domain.User, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.D{},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}
	defer cursor.Close(ctx)

	users := make([]domain.User, 0)
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}

	return users, nil
}

func (r *MongoUserRepository) Update(
	ctx context.Context,
	id bson.ObjectID,
	update domain.UserUpdate,
) (*domain.User, error) {
	set := bson.D{}
	if update.Name != nil {
		set = append(set, bson.E{Key: "name", Value: *update.Name})
	}
	if update.Email != nil {
		set = append(set, bson.E{Key: "email", Value: *update.Email})
	}

	if len(set) == 0 {
		return r.FindByID(ctx, id)
	}

	var user domain.User
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: set}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}

func (r *MongoUserRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *MongoUserRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return count, nil
}

func (r *MongoUserRepository) findOne(ctx context.Context, filter any) (*domain.User, error) {
	var user domain.User
	if err := r.collection.FindOne(ctx, filter).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	return &user, nil
}

var _ UserRepository = (*MongoUserRepository)(nil)
