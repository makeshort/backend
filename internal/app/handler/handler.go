package handler

import (
	"backend/internal/config"
	"backend/internal/service"
	"log/slog"
)

type Handler struct {
	config  *config.Config
	log     *slog.Logger
	service *service.Service
}

// New returns a new instance of Handler.
func New(cfg *config.Config, log *slog.Logger, service *service.Service) *Handler {
	return &Handler{
		config:  cfg,
		log:     log,
		service: service,
	}
}
