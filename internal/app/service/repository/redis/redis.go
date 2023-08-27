package redis

import (
	"backend/internal/config"
	"context"
	"github.com/redis/go-redis/v9"
)

func New(cfg config.Redis) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       0,
	})

	status := db.Ping(context.Background())

	return db, status.Err()
}
