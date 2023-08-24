package service

import (
	"backend/internal/app/service/hash"
	"backend/internal/app/service/storage"
	"backend/internal/app/service/token"
)

type Service struct {
	Storage      storage.Storage
	TokenManager *token.Manager
	Hasher       *hash.Hasher
}

func New(storage storage.Storage, tokenManager *token.Manager, hasher *hash.Hasher) *Service {
	return &Service{
		Storage:      storage,
		TokenManager: tokenManager,
		Hasher:       hasher,
	}
}
