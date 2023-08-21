package token

import (
	"backend/internal/storage"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/exp/slog"
	"time"
)

const (
	accessTokenTTL  = 30 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
)

type Service struct {
	log        *slog.Logger
	storage    storage.Storage
	signingKey []byte
}

type Claims struct {
	jwt.StandardClaims
	UserID string `json:"id"`
}

type Pair struct {
	AccessToken  string
	RefreshToken string
}

func New(log *slog.Logger, st storage.Storage, signingKey string) *Service {
	return &Service{
		log:        log,
		storage:    st,
		signingKey: []byte(signingKey),
	}
}

func (s *Service) GenerateTokenPair(userID string) (*Pair, error) {
	accessToken, err := s.generateToken(userID, accessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(userID, refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &Pair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) Parse(rawToken string) (claims *Claims, err error) {
	parsedToken, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return s.signingKey, nil
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

func (s *Service) generateToken(userID string, timeToLive time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(timeToLive).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userID,
	})
	return token.SignedString(s.signingKey)
}
