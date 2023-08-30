package token

import (
	"backend/internal/config"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

type Manager struct {
	config       *config.Config
	accessSecret []byte
}

type Claims struct {
	jwt.StandardClaims
	ID string `json:"id"`
}

type Pair struct {
	AccessToken  string
	RefreshToken string
}

// New returns a new instance of Manager.
func New(config *config.Config) *Manager {
	return &Manager{
		config:       config,
		accessSecret: []byte(config.Token.Access.Secret),
	}
}

// GenerateTokenPair generates a new token pair with Access Token and Refresh Token.
func (m *Manager) GenerateTokenPair(userID string) (*Pair, error) {
	accessToken, err := m.generateJWT(userID, m.accessSecret)
	if err != nil {
		return nil, err
	}

	refreshToken := m.generateRefreshToken()

	return &Pair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// ParseJWT parses a JWT to Claims.
func (s *Manager) ParseJWT(rawToken string) (*Claims, error) {
	return s.parseToken(rawToken, s.accessSecret)
}

// parseToken parses JWT with signing key.
func (s *Manager) parseToken(rawToken string, signingKey []byte) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		return nil, errors.New("token claims are not of type *tokenClaims")
	}

	return claims, nil
}

// generateJWT generates a new JWT.
func (s *Manager) generateJWT(userID string, signingKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.config.Token.Access.TTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userID,
	})
	return token.SignedString(signingKey)
}

// generateRefreshToken generates a new Refresh Token.
func (s *Manager) generateRefreshToken() string {
	return uuid.NewString()
}
