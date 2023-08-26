package repository

import (
	"backend/internal/app/service/repository/postgres/url"
	"backend/internal/app/service/repository/postgres/user"
	"context"
	"github.com/jmoiron/sqlx"
)

type User interface {
	Create(ctx context.Context, email string, username string, passwordHash string) (string, error)
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByCredentials(ctx context.Context, email string, passwordHash string) (user.User, error)
	Delete(ctx context.Context, id string) error
}

type Url interface {
}

type Repository struct {
	User *user.Postgres
	Url  *url.Postgres
	// Session Session
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		User: user.New(db),
		Url:  url.New(db),
	}
}
