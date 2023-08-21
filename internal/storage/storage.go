package storage

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type URL struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Link      string             `bson:"link"`
	Alias     string             `bson:"alias"`
	Redirects int                `bson:"redirects"`
	CreatedAt primitive.DateTime `bson:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at"`
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	Username     string             `bson:"username"`
	PasswordHash string             `bson:"password_hash"`
	CreatedAt    primitive.DateTime `bson:"created_at"`
	UpdatedAt    primitive.DateTime `bson:"updated_at"`
}

type Session struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	CreatedAt primitive.DateTime `bson:"created_at"`
	ExpiredAt primitive.DateTime `bson:"expired_at"`
}

type Storage interface {
	CreateURL(ctx context.Context, link string, alias string, userID primitive.ObjectID) (primitive.ObjectID, error)
	GetURL(ctx context.Context, alias string) (URL, error)
	IncrementUrlCounter(ctx context.Context, alias string) error
	DeleteURL(ctx context.Context, alias string) error
	CreateUser(ctx context.Context, email string, username string, passwordHash string) (primitive.ObjectID, error)
	GetUser(ctx context.Context, email string, passwordHash string) (User, error)
	GetUserURLs(ctx context.Context, userID primitive.ObjectID) ([]URL, error)
	DeleteUser(ctx context.Context, userID primitive.ObjectID) error
}

var (
	ErrURLNotFound        = errors.New("url not found")
	ErrAliasAlreadyExists = errors.New("alias already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
)
