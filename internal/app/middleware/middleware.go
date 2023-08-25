package middleware

import (
	"backend/internal/app/response"
	"backend/internal/app/service"
	"backend/internal/config"
	"backend/pkg/requestid"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
	"strings"
	"time"
)

const (
	HeaderAuthorization = "Authorization"
	ContextUserID       = "UserID"
)

type Middleware struct {
	config  *config.Config
	log     *slog.Logger
	service *service.Service
}

func New(cfg *config.Config, log *slog.Logger, service *service.Service) *Middleware {
	return &Middleware{
		config:  cfg,
		log:     log,
		service: service,
	}
}

// UserIdentity parse access token in Authorization header and set UserID in context
func (m *Middleware) UserIdentity(ctx *gin.Context) {
	header := ctx.GetHeader(HeaderAuthorization)
	if header == "" {
		response.SendAuthFailedError(ctx)
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		response.SendAuthFailedError(ctx)
		return
	}

	if len(headerParts[1]) == 0 {
		response.SendAuthFailedError(ctx)
		return
	}

	claims, err := m.service.TokenManager.ParseJWT(headerParts[1])
	if err != nil {
		response.SendAuthFailedError(ctx)
		return
	}

	ctx.Set(ContextUserID, claims.UserID)
	ctx.Next()
}

// RequestLog logs every request with parameters: method, path, client_ip, remote_addr, user_agent, status and duration
func (m *Middleware) RequestLog(ctx *gin.Context) {
	startTime := time.Now()

	m.log.Info("request handled",
		slog.String("request_id", requestid.Get(ctx)),
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.Request.URL.Path),
		slog.String("client_ip", ctx.ClientIP()),
	)

	entry := m.log.With(
		slog.String("request_id", requestid.Get(ctx)),
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.Request.URL.Path),
		slog.String("remote_addr", ctx.Request.RemoteAddr),
		slog.String("user_agent", ctx.Request.UserAgent()),
	)

	defer func() {
		entry.Info("request completed",
			slog.Int("status", ctx.Writer.Status()),
			// slog.Int("bytes", ctx.Writer.Size()), // TODO: Fix response size
			slog.String("duration", time.Since(startTime).String()))
	}()
}
