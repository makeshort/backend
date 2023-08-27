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

	return r.client.Set(ctx, refreshToken, marshalledSession, r.config.Token.Refresh.TTL).Err()
}

func (r *Redis) Close(ctx context.Context, refreshToken string) error {
	return r.client.Del(ctx, refreshToken).Err()
}

func (r *Redis) Get(ctx context.Context, refreshToken string) (Session, error) {
	marshalledData, err := r.client.Get(ctx, refreshToken).Result()
	if err != nil {
		return Session{}, err
	}

	var session Session
	err = json.Unmarshal([]byte(marshalledData), &session)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}
