package handler

import (
	"backend/internal/app/service"
	"backend/internal/config"
	"golang.org/x/exp/slog"
)

type Handler struct {
	config  *config.Config
	log     *slog.Logger
	service *service.Service
}

func New(cfg *config.Config, log *slog.Logger, service *service.Service) *Handler {
	return &Handler{
		config:  cfg,
		log:     log,
		service: service,
	}
}
