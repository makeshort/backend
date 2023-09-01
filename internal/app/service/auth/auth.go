package auth

import (
	"backend/internal/app/service/hash"
	"backend/internal/app/service/token"
	"backend/internal/repository"
	"github.com/gin-gonic/gin"
	"net/mail"
)

type UserRegisterDTO struct {
	Email    string
	Username string
	Password string
}

type UserLoginDTO struct {
	Email    string
	Password string
}

type Service struct {
	repo   *repository.Repository
	hasher *hash.Hasher
	token  *token.Manager
}

func New(repo *repository.Repository, hasher *hash.Hasher, token *token.Manager) *Service {
	return &Service{
		repo:   repo,
		hasher: hasher,
		token:  token,
	}
}

func (s *Service) Register(ctx *gin.Context, userDTO UserRegisterDTO) (string, error) {
	isEmailValid := checkEmailValidity(userDTO.Email)
	if !isEmailValid {
		return "", ErrInvalidEmail
	}

	passwordHash := s.hasher.Create(userDTO.Password)

	id, err := s.repo.User.Create(ctx, userDTO.Email, userDTO.Username, passwordHash)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) Login(ctx *gin.Context, userDTO UserLoginDTO) (*token.Pair, error) {
	passwordHash := s.hasher.Create(userDTO.Password)

	user, err := s.repo.User.GetByCredentials(ctx, userDTO.Email, passwordHash)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.token.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	err = s.repo.Session.Create(ctx, tokenPair.RefreshToken, user.ID, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *Service) Logout(ctx *gin.Context, refreshToken string) error {
	return s.repo.Session.Close(ctx, refreshToken)
}

func (s *Service) Refresh(ctx *gin.Context, refreshToken string) (*token.Pair, error) {
	session, err := s.repo.Session.Get(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	err = s.repo.Session.Close(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.token.GenerateTokenPair(session.UserID)
	if err != nil {
		return nil, err
	}

	err = s.repo.Session.Create(ctx, tokenPair.RefreshToken, session.UserID, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func checkEmailValidity(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
