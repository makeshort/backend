package token

import (
	"backend/internal/config"
	"backend/internal/storage"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"time"
)

type Service struct {
	log           *slog.Logger
	storage       storage.Storage
	config        *config.Config
	accessSecret  []byte
	refreshSecret []byte
}

type Claims struct {
	jwt.StandardClaims
	UserID string `json:"id"`
}

type Pair struct {
	AccessToken  string
	RefreshToken string
}

func New(log *slog.Logger, st storage.Storage, config *config.Config) *Service {
	return &Service{
		log:          log,
		storage:      st,
		config:       config,
		accessSecret: []byte(config.Token.Access.Secret),
	}
}

func (s *Service) GenerateTokenPair(userID string) (*Pair, error) {
	accessToken, err := s.generateJWT(userID, s.config.Token.Access.TTL, s.accessSecret)
	if err != nil {
		return nil, err
	}

	refreshToken := s.generateRefreshToken()

	return &Pair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) ParseJWT(rawToken string) (*Claims, error) {
	return s.parseToken(rawToken, s.accessSecret)
}

func (s *Service) parseToken(rawToken string, signingKey []byte) (*Claims, error) {
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

func (s *Service) generateJWT(userID string, timeToLive time.Duration, signingKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(timeToLive).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userID,
	})
	return token.SignedString(signingKey)
}

func (s *Service) generateRefreshToken() string {
	return uuid.NewString()
}
