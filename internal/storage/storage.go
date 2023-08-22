package storage

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
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

type RefreshSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id"`
	RefreshToken string             `bson:"refresh_token"`
	CreatedAt    primitive.DateTime `bson:"created_at"`
	ExpiresAt    primitive.DateTime `bson:"expires_at"`
}

type Storage interface {
	CreateURL(ctx context.Context, link string, alias string, userID primitive.ObjectID) (primitive.ObjectID, error)
	GetUrlByAlias(ctx context.Context, alias string) (URL, error)
	IncrementRedirectsCounter(ctx context.Context, alias string) error
	DeleteURL(ctx context.Context, alias string) error
	CreateUser(ctx context.Context, email string, username string, passwordHash string) (primitive.ObjectID, error)
	GetUserByCredentials(ctx context.Context, email string, passwordHash string) (User, error)
	GetUserURLs(ctx context.Context, userID primitive.ObjectID) ([]URL, error)
	DeleteUser(ctx context.Context, userID primitive.ObjectID) error
	CreateRefreshSession(ctx context.Context, userID primitive.ObjectID, refreshToken string, timeToLive time.Duration) (primitive.ObjectID, error)
	DeleteRefreshSession(ctx context.Context, refreshToken string) error
	IsRefreshTokenValid(ctx context.Context, refreshToken string) (isRefreshTokenValid bool, ownerID primitive.ObjectID)
}

var (
	ErrURLNotFound            = errors.New("url not found")
	ErrAliasAlreadyExists     = errors.New("alias already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
)
