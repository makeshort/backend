package app

import (
	"backend/internal/app/router"
	"backend/internal/app/service"
	"backend/internal/app/service/hash"
	"backend/internal/app/service/repository"
	"backend/internal/app/service/repository/postgres"
	"backend/internal/app/service/repository/redis"
	"backend/internal/app/service/token"
	"backend/internal/config"
	"backend/internal/lib/logger/prettyslog"
	"backend/internal/lib/logger/sl"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// App is a main app struct.
type App struct {
	config *config.Config
	log    *slog.Logger
	hasher *hash.Hasher
}

// New returns a new instance of App.
func New(cfg *config.Config) *App {
	log := initLogger(cfg.Env)
	hasher := hash.New(cfg.HashSalt)
	return &App{
		config: cfg,
		log:    log,
		hasher: hasher,
	}
}

// Run runs the server.
func (a *App) Run() {
	gin.SetMode(gin.ReleaseMode)

	a.log.Info("make.short backend running", slog.String("env", a.config.Env))

	tokenManager := token.New(a.config)
	postgresDB, err := postgres.New(a.config.Db)
	if err != nil {
		a.log.Error("error occurred while connecting to postgres", sl.Err(err))
		os.Exit(1)
	}

	redisDB, err := redis.New(a.config)
	if err != nil {
		a.log.Error("error occurred while connecting to redis", sl.Err(err))
		os.Exit(1)
	}

	repo := repository.New(postgresDB, redisDB, a.config)
	srv := service.New(tokenManager, a.hasher, repo)
	r := router.New(a.config, a.log, srv)

	server := &http.Server{
		Addr:         a.config.Server.Address,
		Handler:      r.InitRoutes(),
		ReadTimeout:  a.config.Server.Timeout,
		WriteTimeout: a.config.Server.Timeout,
		IdleTimeout:  a.config.Server.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				a.log.Error("failed to start server", sl.Err(err))
			}
		}
	}()

	a.log.Info("server started", slog.String("address", server.Addr))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	a.log.Info("server shutting down")

	err = server.Shutdown(context.Background())
	if err != nil {
		a.log.Error("error occurred on server shutting down", sl.Err(err))
	}

	a.log.Info("server stopped")

	err = redisDB.Close()
	if err != nil {
		a.log.Error("error occurred on redis connection closing down", sl.Err(err))
	}

	a.log.Info("redis connection closed")

	err = postgresDB.Close()
	if err != nil {
		a.log.Error("error occurred on postgresql connection closing down", sl.Err(err))
	}

	a.log.Info("postgres connection closed")
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
