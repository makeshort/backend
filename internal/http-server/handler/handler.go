package handler

import (
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
	HeaderSessionID     = "SessionID"
	ContextUserID       = "UserID"
	AliasLength         = 6
)

type Handler struct {
	log          *slog.Logger
	storage      storage.Storage
	hasher       *hash.Hasher
	tokenService *token.Service
}

func New(log *slog.Logger, storage storage.Storage, hasher *hash.Hasher, tokenService *token.Service) *Handler {
	return &Handler{log: log, storage: storage, hasher: hasher, tokenService: tokenService}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(h.LogRequest)

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/:alias", h.Redirect)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/session", h.Login)
			auth.DELETE("/session", h.CheckAuth, h.Logout)
			auth.POST("/signup", h.Register)
			auth.POST("/refresh", h.Refresh)
		}

		api.POST("/url", h.CheckAuth, h.CreateURL)
		// api.PATCH("/url/:alias", h.CheckAuth, h.UpdateURL)
		api.DELETE("/url/:alias", h.CheckAuth, h.DeleteURL)
		// api.PATCH("/user/me", h.CheckAuth, h.UpdateMe)
		api.DELETE("/user/me", h.CheckAuth, h.DeleteMe)
		api.GET("/user/me/urls", h.CheckAuth, h.GetMyURLs)
	}

	h.logRoutes(router.Routes())

	return router
}

func (h *Handler) logRoutes(routes gin.RoutesInfo) {
	for _, route := range routes {
		method := format.CompleteStringToLength(route.Method, 10, ' ')
		path := format.CompleteStringToLength(route.Path, 25, ' ')

		routeLog := fmt.Sprintf("%s%s --> %s", method, path, route.Handler)

		h.log.Debug(routeLog)
	}
}
