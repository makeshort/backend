package user

import (
	"context"
	"fmt"
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

	query := fmt.Sprintf("INSERT INTO users (email, username, password_hash) values ($1, $2, $3) RETURNING id")
	row := p.db.QueryRowContext(ctx, query, email, username, passwordHash)

	if row.Err() != nil {
		return "", row.Err()
	}

	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (p *Postgres) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM users WHERE id = $1")
	row := p.db.QueryRowContext(ctx, query, id)

	return row.Err()
}

func (p *Postgres) GetByID(ctx context.Context, id string) (User, error) {
	var user User

	query := fmt.Sprintf("SELECT * FROM users WHERE id = $1")

	err := p.db.GetContext(ctx, &user, query, id)

	return user, err
}

func (p *Postgres) GetByCredentials(ctx context.Context, email string, passwordHash string) (User, error) {
	var user User

	query := fmt.Sprintf("SELECT * FROM users WHERE email = $1 AND password_hash = $2")

	err := p.db.GetContext(ctx, &user, query, email, passwordHash)

	return user, err
}
