package url

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"time"
)

type Postgres struct {
	db *sqlx.DB
}

type URL struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	LongURL   string    `db:"long_url"`
	ShortURL  string    `db:"short_url"`
	Redirects int       `db:"redirects"`
	CreatedAt time.Time `db:"created_at"`
}

type DTO struct {
	LongURL  string `db:"long_url"`
	ShortURL string `db:"short_url"`
}

// New returns a new instance of *Postgres.
func New(db *sqlx.DB) *Postgres {
	return &Postgres{db: db}
}

// Create creates a new url in database. If url with provided short url already exists,
// function will return an ErrShortUrlAlreadyExists.
func (p *Postgres) Create(ctx context.Context, longUrl string, shortUrl string, userID string) (string, error) {
	var id string

	query := "INSERT INTO urls (user_id, long_url, short_url) values ($1, $2, $3) RETURNING id"
	err := p.db.GetContext(ctx, &id, query, userID, longUrl, shortUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrShortUrlAlreadyExists
	}

	return id, err
}

// GetByID returns an url by its ID.
// If the url does not exist in database, the function will return an ErrUrlNotFound.
func (p *Postgres) GetByID(ctx context.Context, id string) (URL, error) {
	var url URL

	query := "SELECT * FROM urls WHERE id = $1"

	err := p.db.GetContext(ctx, &url, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return URL{}, ErrUrlNotFound
	}

	return url, err
}

// GetByShortUrl returns an url by its short url.
// If the url does not exist in database, the function will return an ErrUrlNotFound.
func (p *Postgres) GetByShortUrl(ctx context.Context, shortUrl string) (URL, error) {
	var url URL

	query := "SELECT * FROM urls WHERE short_url = $1"

	err := p.db.GetContext(ctx, &url, query, shortUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return URL{}, ErrUrlNotFound
	}

	return url, err
}

// IncrementRedirectsCounter increments url's redirects counter in database.
// If the url does not exist, the function wil return an ErrUrlNotFound.
func (p *Postgres) IncrementRedirectsCounter(ctx context.Context, id string) error {
	query := "UPDATE urls SET redirects = redirects + 1 WHERE id = $1"

	res, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUrlNotFound
	}

	return nil
}

// Update updates an url in database. If url with provided ID does not exist,
// the function will return an ErrUrlNotFound.
// If some fields of DTO are empty, they won't be updated.
func (p *Postgres) Update(ctx context.Context, id string, dto DTO) (URL, error) {
	var url URL

	query := "UPDATE urls SET short_url = CASE WHEN $1::varchar(10) IS NOT NULL AND $1 <> '' THEN $1 ELSE short_url END, long_url = CASE WHEN $2::varchar(2048) IS NOT NULL AND $2 <> '' THEN $2 ELSE long_url END WHERE id = $3 RETURNING *"

	row := p.db.QueryRowContext(ctx, query, dto.ShortURL, dto.LongURL, id)

	err := row.Scan(&url.ID, &url.UserID, &url.LongURL, &url.ShortURL, &url.Redirects, &url.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return URL{}, ErrUrlNotFound
	}

	return url, err
}

// Delete deletes an url from database by its ID.
// If url with provided ID does not exist in database, the function will return an ErrUrlNotFound.
func (p *Postgres) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM urls WHERE id = $1"
	res, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUrlNotFound
	}

	return nil
}
