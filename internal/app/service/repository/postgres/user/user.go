package user

import (
	"backend/internal/app/service/repository/postgres/url"
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"time"
)

type Postgres struct {
	db *sqlx.DB
}

type User struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

func New(db *sqlx.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) Create(ctx context.Context, email string, username string, passwordHash string) (string, error) {
	var id string

	query := "INSERT INTO users (email, username, password_hash) values ($1, $2, $3) RETURNING id"
	row := p.db.QueryRowContext(ctx, query, email, username, passwordHash)

	if row.Err() != nil {
		return "", row.Err()
	}

	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUserAlreadyExists
	}

	return id, err
}

func (p *Postgres) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM users WHERE id = $1"
	row := p.db.QueryRowContext(ctx, query, id)
	if errors.Is(row.Err(), sql.ErrNoRows) {
		return ErrUserNotExists
	}

	return row.Err()
}

func (p *Postgres) GetByID(ctx context.Context, id string) (User, error) {
	var user User

	query := "SELECT * FROM users WHERE id = $1"

	err := p.db.GetContext(ctx, &user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrUserNotExists
	}

	return user, err
}

func (p *Postgres) GetByCredentials(ctx context.Context, email string, passwordHash string) (User, error) {
	var user User

	query := "SELECT * FROM users WHERE email = $1 AND password_hash = $2"

	err := p.db.GetContext(ctx, &user, query, email, passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrUserNotExists
	}

	return user, err
}

func (p *Postgres) GetUrls(ctx context.Context, id string) ([]url.URL, error) {
	var urls []url.URL

	query := "SELECT * FROM urls WHERE user_id = $1"

	err := p.db.SelectContext(ctx, &urls, query, id)

	return urls, err
}
