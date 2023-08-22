package main

import (
	_ "backend/docs"
	"backend/internal/config"
	"backend/internal/http-server/handler"
	"backend/internal/lib/hash"
	"backend/internal/lib/logger/prettyslog"
	"backend/internal/lib/logger/sl"
	"backend/internal/storage/mongo"
	"backend/internal/token"
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
)

// @title                        URL Shortener App API
// @version                      0.1
// @description                  API Server for URL Shortener Application
// @host                         localhost:8081
// @BasePath                     /api
// @securityDefinitions.apikey   AccessToken
// @in                           header
// @name                         Authorization
func main() {
	cfg := config.MustLoad()
	log := initLogger(cfg.Env)
	hasher := hash.New(cfg.HashSalt)

	gin.SetMode(gin.ReleaseMode)

	log.Info("url shortener rest api server running", slog.String("env", cfg.Env))

	storage := mongo.New(cfg)
	log.Info("mongo client started")
	defer func() {
		err := storage.Client.Disconnect(context.Background())
		if err != nil {
			log.Error("failed to disconnect mongo client", err)
		}
	}()

	tokenService := token.New(log, storage, cfg)
	h := handler.New(log, storage, hasher, tokenService, cfg)

	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      h.InitRoutes(),
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))
	}

	log.Error("server stopped")
}

func initLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case config.EnvLocal:
		log = initPrettyLogger()
	case config.EnvDevelopment:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case config.EnvProduction:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func initPrettyLogger() *slog.Logger {
	opts := prettyslog.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	prettyHandler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(prettyHandler)
}
