package service

import (
	"backend/internal/service/hash"
	"backend/internal/service/repository"
	"backend/internal/service/token"
)

type Service struct {
	Repository   *repository.Repository
	TokenManager *token.Manager
	Hasher       *hash.Hasher
}

// New returns a new instance of Service.
func New(tokenManager *token.Manager, hasher *hash.Hasher, repo *repository.Repository) *Service {
	return &Service{
		Repository:   repo,
		TokenManager: tokenManager,
		Hasher:       hasher,
	}
}
