package service

import (
	"backend/internal/app/service/auth"
	"backend/internal/app/service/hash"
	"backend/internal/app/service/token"
	"backend/internal/repository"
)

type Service struct {
	Auth         *auth.Service
	Repository   *repository.Repository
	TokenManager *token.Manager
	Hasher       *hash.Hasher
}

// New returns a new instance of Service.
func New(tokenManager *token.Manager, hasher *hash.Hasher, repo *repository.Repository) *Service {
	return &Service{
		Auth:         auth.New(repo, hasher, tokenManager),
		Repository:   repo,
		TokenManager: tokenManager,
		Hasher:       hasher,
	}
}
