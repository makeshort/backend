package service

import (
	"backend/internal/app/service/hash"
	"backend/internal/app/service/repository"
	"backend/internal/app/service/storage"
	"backend/internal/app/service/token"
)

type Service struct {
	Storage      storage.Storage
	Repository   *repository.Repository
	TokenManager *token.Manager
	Hasher       *hash.Hasher
}

// New returns a new instance of Service.
func New(storage storage.Storage, tokenManager *token.Manager, hasher *hash.Hasher, repo *repository.Repository) *Service {
	return &Service{
		Storage:      storage,
		Repository:   repo,
		TokenManager: tokenManager,
		Hasher:       hasher,
	}
}
