package user

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type User struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *User {
	return &User{db: db}
}

func (u *User) Create(ctx context.Context, email string, username string, passwordHash string) (string, error) {
	var id string

	query := fmt.Sprintf("INSERT INTO 'users' (email, username, password_hash) values ($1, $2, $3) RETURNING id")
	row := u.db.QueryRow(query, email, username, passwordHash)

	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}
