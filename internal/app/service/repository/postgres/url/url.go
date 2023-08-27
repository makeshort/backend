package url

import (
	"context"
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

func New(db *sqlx.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) Create(ctx context.Context, longUrl string, shortUrl string, userID string) (string, error) {
	var id string

	query := "INSERT INTO urls (user_id, long_url, short_url) values ($1, $2, $3) RETURNING id"
	row := p.db.QueryRowContext(ctx, query, userID, longUrl, shortUrl)

	if row.Err() != nil {
		return "", row.Err()
	}

	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (p *Postgres) GetByID(ctx context.Context, id string) (URL, error) {
	var user URL

	query := "SELECT * FROM urls WHERE id = $1"

	err := p.db.GetContext(ctx, &user, query, id)

	return user, err
}

func (p *Postgres) GetByShortUrl(ctx context.Context, shortUrl string) (URL, error) {
	var user URL

	query := "SELECT * FROM urls WHERE short_url = $1"

	err := p.db.GetContext(ctx, &user, query, shortUrl)

	return user, err
}

func (p *Postgres) IncrementRedirectsCounter(ctx context.Context, id string) error {
	query := "UPDATE urls SET redirects = redirects + 1 WHERE id = $1"

	_, err := p.db.ExecContext(ctx, query, id)

	return err
}

func (p *Postgres) Update(ctx context.Context, id string, shortUrl string, longUrl string) (URL, error) {
	var url URL

	query := "UPDATE urls SET short_url = CASE WHEN $1::varchar(10) IS NOT NULL AND $1 <> '' THEN $1 ELSE short_url END, long_url = CASE WHEN $2::varchar(2048) IS NOT NULL AND $2 <> '' THEN $2 ELSE long_url END WHERE id = $3 RETURNING *"

	row := p.db.QueryRowContext(ctx, query, shortUrl, longUrl, id)

	if err := row.Scan(&url.ID, &url.UserID, &url.LongURL, &url.ShortURL, &url.Redirects, &url.CreatedAt); err != nil {
		return URL{}, err
	}

	return url, nil
}

func (p *Postgres) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM urls WHERE id = $1"
	_, err := p.db.ExecContext(ctx, query, id)

	return err
}
