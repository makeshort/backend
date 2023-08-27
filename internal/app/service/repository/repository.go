package repository

import (
	"backend/internal/app/service/repository/postgres/url"
	"backend/internal/app/service/repository/postgres/user"
	"backend/internal/app/service/repository/redis/session"
	"backend/internal/config"
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type User interface {
	Create(ctx context.Context, email string, username string, passwordHash string) (string, error)
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByCredentials(ctx context.Context, email string, passwordHash string) (user.User, error)
	GetUrlsList(ctx context.Context, id string) ([]url.URL, error)
	Delete(ctx context.Context, id string) error
}

type Url interface {
	Create(ctx context.Context, longUrl string, shortUrl string, userID string) (string, error)
	GetByID(ctx context.Context, id string) (url.URL, error)
	GetByShortUrl(ctx context.Context, shortUrl string) (url.URL, error)
	IncrementRedirectsCounter(ctx context.Context, id string) error
	Update(ctx context.Context, id string, shortUrl string, longUrl string) (url.URL, error)
	Delete(ctx context.Context, id string) error
}

type Session interface {
	Create(ctx context.Context, refreshToken string, userID string, ip string, userAgent string) error
	Close(ctx context.Context, refreshToken string) error
	Get(ctx context.Context, refreshToken string) (session.Session, error)
}

type Repository struct {
	User    *user.Postgres
	Url     *url.Postgres
	Session *session.Redis
}

func New(postgresDB *sqlx.DB, redisDB *redis.Client, cfg *config.Config) *Repository {
	return &Repository{
		User:    user.New(postgresDB),
		Url:     url.New(postgresDB),
		Session: session.New(redisDB, cfg),
	}
}

var (
	ErrURLNotFound            = errors.New("url not found")
	ErrAliasAlreadyExists     = errors.New("alias already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
)
