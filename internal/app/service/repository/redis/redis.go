package redis

import (
	"backend/internal/config"
	"context"
	"github.com/redis/go-redis/v9"
)

func New(cfg *config.Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	status := db.Ping(context.Background())

	return db, status.Err()
}
