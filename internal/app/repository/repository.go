package repository

import (
	"backend/internal/app/repository/postgres/url"
	"backend/internal/app/repository/postgres/user"
	"context"
	"github.com/jmoiron/sqlx"
)

type User interface {
	Create(ctx context.Context, email string, username string, passwordHash string) (string, error)
}

type Url interface {
}

type Session interface {
}

type Repository struct {
	User
	Url
	Session
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		User: user.New(db),
		Url:  url.New(db),
	}
}
