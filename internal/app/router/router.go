package router

import (
	"backend/internal/app/handler"
	"backend/internal/app/middleware"
	"backend/internal/app/service"
	"backend/internal/config"
	"backend/internal/lib/logger/format"
	"backend/pkg/requestid"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/exp/slog"
)

type Router struct {
	config     *config.Config
	log        *slog.Logger
	handler    *handler.Handler
	middleware *middleware.Middleware
	service    *service.Service
}

// New returns a new instance of Router.
func New(cfg *config.Config, log *slog.Logger, service *service.Service) *Router {
	mw := middleware.New(cfg, log, service)
	h := handler.New(cfg, log, service)
	return &Router{
		config:     cfg,
		log:        log,
		handler:    h,
		middleware: mw,
		service:    service,
	}
}

// InitRoutes create a new routes list for handler.
func (r *Router) InitRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(requestid.New)
	router.Use(r.middleware.RequestLog)

	router.GET("/:alias", r.handler.Redirect)

	api := router.Group("/api")
	{
		api.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		auth := api.Group("/auth")
		{
			auth.POST("/session", r.handler.Login)
			auth.DELETE("/session", r.handler.Logout)
			auth.POST("/signup", r.handler.Register)
			auth.POST("/refresh", r.handler.RefreshTokens)
		}

		url := api.Group("/url", r.middleware.UserIdentity)
		{
			url.POST("/", r.handler.CreateURL)
			// url.PATCH("/:id")
			url.DELETE("/:id", r.middleware.CheckOwner, r.handler.DeleteURL)
		}

		user := api.Group("/user")
		{
			user.GET("/:id", r.handler.GetUser)

			// user.PATCH("/:id")
			user.DELETE("/:id", r.middleware.UserIdentity, r.middleware.CheckMe, r.handler.DeleteUser)
			user.GET("/:id/urls", r.middleware.UserIdentity, r.middleware.CheckMe, r.handler.GetUserUrls)
		}
	}

	r.logRoutes(router.Routes())

	return router
}

// logRoutes logs all routes of Router.
func (r *Router) logRoutes(routes gin.RoutesInfo) {
	for _, route := range routes {
		method := format.CompleteStringToLength(route.Method, 10, ' ')
		path := format.CompleteStringToLength(route.Path, 25, ' ')

		routeLog := fmt.Sprintf("%s%s --> %s", method, path, route.Handler)

		r.log.Debug(routeLog)
	}
}
