package url

import (
	"github.com/jmoiron/sqlx"
)

type Url struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Url {
	return &Url{db: db}
}
