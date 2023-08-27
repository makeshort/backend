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
	ExpiresAt time.Time
}

// New returns a new instance of *Redis.
func New(client *redis.Client, cfg *config.Config) *Redis {
	return &Redis{
		client: client,
		config: cfg,
	}
}

// Create creates a new session in redis storage.
// If session with provided refresh token already exists, function will return an ErrRefreshTokenAlreadyExists.
func (r *Redis) Create(ctx context.Context, refreshToken string, userID string, ip string, userAgent string) error {
	exists, err := r.client.Exists(ctx, refreshToken).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		return ErrRefreshTokenAlreadyExists
	}

	session := Session{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(r.config.Token.Refresh.TTL),
	}

	marshalledSession, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, refreshToken, marshalledSession, r.config.Token.Refresh.TTL).Err()
}

// Close deletes a session from redis storage by refresh token.
// If the session does not exist, the function will return an ErrSessionNotExists.
func (r *Redis) Close(ctx context.Context, refreshToken string) error {
	exists, err := r.client.Exists(ctx, refreshToken).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrSessionNotExists
	}
	return r.client.Del(ctx, refreshToken).Err()
}

// Get returns a Session from redis storage by refresh token.
// If the session does not exist, the function will return an ErrSessionNotExists.
func (r *Redis) Get(ctx context.Context, refreshToken string) (Session, error) {
	exists, err := r.client.Exists(ctx, refreshToken).Result()
	if err != nil {
		return Session{}, err
	}
	if exists == 0 {
		return Session{}, ErrSessionNotExists
	}

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
