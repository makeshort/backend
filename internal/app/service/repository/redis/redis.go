package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func New() (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	status := db.Ping(context.Background())

	return db, status.Err()
}
