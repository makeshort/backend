package jsonwebtoken

import (
	"backend/internal/storage"
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/exp/slog"
	"time"
)

const (
	tokenTTL = 7 * 24 * time.Hour
)

type JWT struct {
	SigningKey string
	log        *slog.Logger
	storage    storage.Storage
}

type tokenClaims struct {
	jwt.StandardClaims
	UserID    string `json:"_id"`
	SessionID string `json:"session_id"`
}

// New returns a new JWT instance.
func New(log *slog.Logger, st storage.Storage, signingKey string) *JWT {
	return &JWT{
		log:        log,
		storage:    st,
		SigningKey: signingKey,
	}
}

// Generate creates a new JWT with given credentials.
func (j *JWT) Generate(email string, passwordHash string) (accessToken string, err error) {
	user, err := j.storage.GetUser(context.Background(), email, passwordHash)

	if errors.Is(err, storage.ErrUserNotFound) {
		j.log.Error("user not found", slog.String("email", email), slog.String("password_hash", passwordHash))
		return "", err
	}
	if err != nil {
		j.log.Error("can't get user", slog.String("email", email), slog.String("password_hash", passwordHash))
		return "", err
	}
	sessionID, err := j.storage.CreateSession(context.Background(), user.ID)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID.Hex(),
		sessionID.Hex(),
	})
	return token.SignedString([]byte(j.SigningKey))
}

// Parse parses a JWT to some variables: UserID and SessionID
func (j *JWT) Parse(accessToken string) (hexUserID string, hexSessionID string, err error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(j.SigningKey), nil
	})
	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return "", "", errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserID, claims.SessionID, nil
}
