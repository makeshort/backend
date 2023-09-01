package user

import (
	"backend/internal/repository/postgres/url"
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
	TelegramID   string    `db:"telegram_id"`
	CreatedAt    time.Time `db:"created_at"`
}

type DTO struct {
	Email        string `db:"email"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	TelegramID   string `db:"telegram_id"`
}

// New returns a new instance of *Postgres with *sqlx.DB field.
func New(db *sqlx.DB) *Postgres {
	return &Postgres{db: db}
}

// Create creates a new user in database.
// If the user with unique fields already exists in database, the function will return an ErrUserAlreadyExists.
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

// Update updates a user by his ID in database.
// If the user does not exist in database, the function will return ErrUserNotExists.
// If some fields of DTO are empty, they won't be updated.
func (p *Postgres) Update(ctx context.Context, id string, dto DTO) (User, error) {
	var user User

	query := "UPDATE users SET email = CASE WHEN $1::varchar(255) IS NOT NULL AND $1 <> '' THEN $1 ELSE email END, username = CASE WHEN $2::varchar(50) IS NOT NULL AND $2 <> '' THEN $2 ELSE username END, password_hash = CASE WHEN $3::varchar(255) IS NOT NULL AND $3 <> '' THEN $3 ELSE password_hash END, telegram_id = CASE WHEN $4::varchar(20) IS NOT NULL AND $4 <> '' THEN $4 ELSE telegram_id END WHERE id = $5 RETURNING *"

	row := p.db.QueryRowContext(ctx, query, dto.Email, dto.Username, dto.PasswordHash, dto.TelegramID, id)

	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.TelegramID, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotExists
	}

	return user, err
}

// Delete deletes a user by his ID from database.
// If the user does not exist in database, the function will return ErrUserNotExists.
func (p *Postgres) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM users WHERE id = $1"
	res, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotExists
	}

	return nil
}

// GetByID gets a user from database by his ID, and return as User.
// If the user does not exist in database, the function will return an ErrUserNotExists.
func (p *Postgres) GetByID(ctx context.Context, id string) (User, error) {
	var user User

	query := "SELECT * FROM users WHERE id = $1"

	err := p.db.GetContext(ctx, &user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrUserNotExists
	}

	return user, err
}

// GetByTelegramID gets a user from database by his telegram ID, and return as User.
// If the user does not exist in database, the function will return an ErrUserNotExists.
func (p *Postgres) GetByTelegramID(ctx context.Context, telegramID string) (User, error) {
	var user User

	query := "SELECT * FROM users WHERE telegram_id = $1"

	err := p.db.GetContext(ctx, &user, query, telegramID)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrUserNotExists
	}

	return user, err
}

// GetByCredentials gets a user by his credential info such as email and hashed password.
// If the user does not exist in database, the function will return an ErrUserNotExists.
func (p *Postgres) GetByCredentials(ctx context.Context, email string, passwordHash string) (User, error) {
	var user User

	query := "SELECT * FROM users WHERE email = $1 AND password_hash = $2"

	err := p.db.GetContext(ctx, &user, query, email, passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrUserNotExists
	}

	return user, err
}

// GetUrlsList gets all urls from database, that assigned to provided user ID.
// If urls with this user ID do not exist in database, the function will return just an empty array.
func (p *Postgres) GetUrlsList(ctx context.Context, id string) ([]url.URL, error) {
	var urls []url.URL

	query := "SELECT * FROM urls WHERE user_id = $1"

	err := p.db.SelectContext(ctx, &urls, query, id)

	return urls, err
}
