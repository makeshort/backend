package repository

import (
	"backend/internal/app/service/repository/postgres/url"
	"backend/internal/app/service/repository/postgres/user"
	"context"
	"github.com/jmoiron/sqlx"
)

type User interface {
	Create(ctx context.Context, email string, username string, passwordHash string) (string, error)
	Delete(ctx context.Context, uuid string) error
}

type Url interface {
}

type Session interface {
}

type Repository struct {
	User    User
	Url     Url
	Session Session
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		User: user.New(db),
		Url:  url.New(db),
	}
}
