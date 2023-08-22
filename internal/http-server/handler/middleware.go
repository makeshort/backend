package handler

import (
	"backend/internal/http-server/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) CheckAuth(ctx *gin.Context) {
	header := ctx.GetHeader(HeaderAuthorization)
	if header == "" {
		response.SendError(ctx, http.StatusUnauthorized, "empty auth header")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		response.SendError(ctx, http.StatusUnauthorized, "invalid auth header")
		return
	}

	if len(headerParts[1]) == 0 {
		response.SendError(ctx, http.StatusUnauthorized, "token is invalid")
		return
	}

	claims, err := h.tokenService.ParseJWT(headerParts[1])
	if err != nil {
		response.SendError(ctx, http.StatusUnauthorized, "token is invalid")
		return
	}

	ctx.Set(ContextUserID, claims.UserID)
}

func (h *Handler) LogRequest(ctx *gin.Context) {
	startTime := time.Now()

	h.log.Info("request handled",
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.Request.URL.Path),
		slog.String("client_ip", ctx.ClientIP()),
	)

	entry := h.log.With(
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
