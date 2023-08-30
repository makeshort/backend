package token

import (
	"backend/internal/config"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

type Manager struct {
	config *config.Config
}

type DefaultClaims struct {
	jwt.StandardClaims
	ID string `json:"id"`
}

type Pair struct {
	AccessToken  string
	RefreshToken string
}

// New returns a new instance of Manager.
func New(config *config.Config) *Manager {
	return &Manager{config: config}
}

// GenerateTokenPair generates a new token pair, containing access and refresh tokens.
func (m *Manager) GenerateTokenPair(userID string) (*Pair, error) {
	accessToken, err := m.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken := m.generateRefreshToken()

	return &Pair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateAccessToken generates new access token with `id` field.
func (m *Manager) generateAccessToken(ID string) (string, error) {
	claims := DefaultClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(m.config.Token.Access.TTL).Unix(),
		},
		ID: ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.config.Token.Access.Secret))
}

// ParseAccessToken parses JWT into DefaultClaims.
func (m *Manager) ParseAccessToken(rawToken string) (*DefaultClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(rawToken, &DefaultClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(m.config.Token.Access.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*DefaultClaims)
	if !ok {
		return nil, errors.New("token claims are not of type *DefaultClaims")
	}

	return claims, nil
}

// generateRefreshToken generates a random refresh token.
func (m *Manager) generateRefreshToken() string {
	return uuid.NewString()
}
