package url

import (
	"github.com/jmoiron/sqlx"
)

type Postgres struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Postgres {
	return &Postgres{db: db}
}
