package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User represents a persisted application user.
type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Email        string        `bson:"email" json:"email"`
	PasswordHash string        `bson:"password_hash" json:"-"`
	CreatedAt    time.Time     `bson:"created_at" json:"created_at"`
}

// UserUpdate contains the fields that may be changed.
type UserUpdate struct {
	Name  *string
	Email *string
}
