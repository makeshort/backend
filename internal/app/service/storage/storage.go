package storage

import (
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

type RefreshSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id"`
	RefreshToken string             `bson:"refresh_token"`
	IP           string             `bson:"ip"`
	UserAgent    string             `bson:"user_agent"`
	CreatedAt    primitive.DateTime `bson:"created_at"`
	ExpiresAt    primitive.DateTime `bson:"expires_at"`
}

var (
	ErrURLNotFound            = errors.New("url not found")
	ErrAliasAlreadyExists     = errors.New("alias already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
)
