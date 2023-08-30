package middleware

import (
	"backend/internal/app/response"
	"backend/internal/app/service"
	"backend/internal/app/service/repository"
	repoUser "backend/internal/app/service/repository/postgres/user"
	"backend/internal/config"
	"backend/internal/lib/logger/sl"
	"backend/pkg/requestid"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
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

// New returns a new instance of Middleware.
func New(cfg *config.Config, log *slog.Logger, service *service.Service) *Middleware {
	return &Middleware{
		config:  cfg,
		log:     log,
		service: service,
	}
}

// UserIdentity parse access token in Authorization header and set UserID in context.
func (m *Middleware) UserIdentity(ctx *gin.Context) {
	log := m.log.With(
		slog.String("op", "handler.UserIdentity"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	header := ctx.GetHeader(HeaderAuthorization)
	if header == "" {
		log.Debug("auth header is empty")
		response.SendAuthFailedError(ctx)
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		log.Debug("auth header is invalid")
		response.SendAuthFailedError(ctx)
		return
	}

	if len(headerParts[1]) == 0 {
		log.Debug("access token is empty")
		response.SendAuthFailedError(ctx)
		return
	}

	claims, err := m.service.TokenManager.ParseJWT(headerParts[1])
	if err != nil {
		log.Debug("can't parse token", sl.Err(err))
		response.SendAuthFailedError(ctx)
		return
	}

	tokenType := headerParts[0]
	switch tokenType {
	case "Bearer":
		ctx.Set(ContextUserID, claims.ID)
		ctx.Next()
	case "Telegram":
		user, err := m.service.Repository.User.GetByTelegramID(ctx, claims.ID)
		if repoUser.IsErrUserNotExists(err) {
			log.Debug("user not found", slog.String("telegram_id", claims.ID))
			response.SendAuthFailedError(ctx)
			return
		}
		if err != nil {
			log.Error("error occurred while getting user", sl.Err(err), slog.String("telegram_id", claims.ID))
			response.SendError(ctx, http.StatusInternalServerError, "can't get user")
			return
		}
		ctx.Set(ContextUserID, user.ID)
		ctx.Next()
	default:
		response.SendAuthFailedError(ctx)
		return
	}
}

// CheckOwner middleware checks if user owning URL with ID from parameter.
func (m *Middleware) CheckOwner(ctx *gin.Context) {
	log := m.log.With(
		slog.String("op", "middleware.CheckOwner"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	urlID := ctx.Param("id")
	userID := ctx.GetString(ContextUserID)

	url, err := m.service.Repository.Url.GetByID(ctx, urlID)
	if errors.Is(err, repository.ErrURLNotFound) {
		log.Debug("url not found",
			slog.String("id", urlID),
		)
		response.SendError(ctx, http.StatusNotFound, "url with this id not found")
		return
	}
	if err != nil {
		log.Error("error occurred while getting url",
			slog.String("id", urlID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't get url")
		return
	}

	if url.UserID != userID {
		response.SendError(ctx, http.StatusForbidden, "not your url")
		return
	}
	ctx.Next()
}

// CheckMe middleware checks if UserID from context (authenticated user id) completely equals to ID from parameter.
func (m *Middleware) CheckMe(ctx *gin.Context) {
	if ctx.GetString(ContextUserID) != ctx.Param("id") {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	ctx.Next()
}

// RequestLog logs every request with parameters: method, path, client_ip, remote_addr, user_agent, status and duration.
func (m *Middleware) RequestLog(ctx *gin.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, "/api/docs/") { // ignore logging swagger documentation
		return
	}

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
			slog.String("duration", fmt.Sprintf("%dus", time.Since(startTime).Microseconds())),
		)
	}()
}
