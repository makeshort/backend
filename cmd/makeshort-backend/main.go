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
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	log.Info("make.short backend running", slog.String("env", cfg.Env))

	storage := mongo.New(cfg)
	log.Info("mongo client started")

	tokenService := token.New(log, storage, cfg)
	h := handler.New(log, storage, hasher, tokenService, cfg)

	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      h.InitRoutes(),
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("failed to start server", sl.Err(err))
			}
		}
	}()

	log.Info("server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info("server shutting down")

	err := server.Shutdown(context.Background())
	if err != nil {
		log.Error("error occurred on server shutting down: %s", err.Error())
	}

	log.Info("server stopped")

	err = storage.Client.Disconnect(context.Background())
	if err != nil {
		log.Error("error occurred on db shutting down: %s", err.Error())
	}

	log.Info("db disconnected")
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
