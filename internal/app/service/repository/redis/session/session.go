package session

import (
	"backend/internal/config"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type Redis struct {
	client *redis.Client
	config *config.Config
}

type Session struct {
	UserID    string
	IP        string
	UserAgent string
	CreatedAt time.Time
}

func New(client *redis.Client, cfg *config.Config) *Redis {
	return &Redis{
		client: client,
		config: cfg,
	}
}

func (r *Redis) Create(ctx context.Context, refreshToken string, userID string, ip string, userAgent string) error {
	session := Session{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	marshalledSession, err := json.Marshal(session)
	if err != nil {
		return err
	}

	cmd := r.client.Set(ctx, refreshToken, marshalledSession, r.config.Token.Refresh.TTL)
	return cmd.Err()
}

func (r *Redis) Close(ctx context.Context, refreshToken string) error {
	cmd := r.client.Del(ctx, refreshToken)
	return cmd.Err()
}

func (r *Redis) Get(ctx context.Context, refreshToken string) (Session, error) {
	storedData, err := r.client.Get(ctx, refreshToken).Result()
	if err != nil {
		return Session{}, err
	}

	// Unmarshal the JSON data back into your struct
	var retrievedData Session
	err = json.Unmarshal([]byte(storedData), &retrievedData)
	if err != nil {
		return Session{}, err
	}

	return retrievedData, nil
}
