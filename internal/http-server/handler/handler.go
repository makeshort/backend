package handler

import (
	"backend/internal/config"
	"backend/internal/lib/hash"
	"backend/internal/lib/logger/format"
	"backend/internal/storage"
	"backend/internal/token"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/exp/slog"
)

const (
	HeaderAuthorization = "Authorization"
	ContextUserID       = "UserID"
	AliasLength         = 6
)

type Handler struct {
	log          *slog.Logger
	storage      storage.Storage
	hasher       *hash.Hasher
	tokenService *token.Service
	config       *config.Config
}

// New returns a new instance of Handler
func New(log *slog.Logger, storage storage.Storage, hasher *hash.Hasher, tokenService *token.Service, config *config.Config) *Handler {
	return &Handler{log: log, storage: storage, hasher: hasher, tokenService: tokenService, config: config}
}

// InitRoutes create a new routes list for Handler
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(h.RequestLog)

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/:alias", h.Redirect)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/session", h.Login)
			auth.DELETE("/session", h.Logout)
			auth.POST("/signup", h.Register)
			auth.POST("/refresh", h.RefreshTokens)
		}

		url := api.Group("/url", h.UserIdentity)
		{
			url.POST("/", h.CreateURL)
			// url.PATCH("/:alias", h.UpdateURL)
			url.DELETE("/:alias", h.DeleteURL)
		}

		user := api.Group("/user", h.UserIdentity)
		{
			// user.PATCH("/me", h.UpdateMe)
			user.DELETE("/me", h.DeleteMe)
			user.GET("/me/urls", h.GetMyURLs)
		}
	}

	h.logRoutes(router.Routes())

	return router
}

// logRoutes logs all routes for Handler
func (h *Handler) logRoutes(routes gin.RoutesInfo) {
	for _, route := range routes {
		method := format.CompleteStringToLength(route.Method, 10, ' ')
		path := format.CompleteStringToLength(route.Path, 25, ' ')

		routeLog := fmt.Sprintf("%s%s --> %s", method, path, route.Handler)

		h.log.Debug(routeLog)
	}
}
