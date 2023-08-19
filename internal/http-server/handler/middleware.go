package handler

import (
	"backend/internal/http-server/constraints"
	"backend/internal/http-server/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

func (h *Handler) CheckAuth(ctx *gin.Context) {
	hexSessionID := ctx.GetHeader(constraints.HeaderSessionID)
	if hexSessionID == "" {
		response.SendError(ctx, http.StatusUnauthorized, "SessionID header is empty")
		return
	}

	sessionID, err := primitive.ObjectIDFromHex(hexSessionID)
	if err != nil {
		response.SendError(ctx, http.StatusUnauthorized, "Session ID is invalid")
		return
	}

	isSessionActive, userID := h.storage.IsSessionActive(ctx, sessionID)
	if !isSessionActive {
		response.SendError(ctx, http.StatusUnauthorized, "Your session is expired")
		return
	}

	ctx.Set(constraints.ContextUserID, userID.Hex())
}

//func (h *Handler) CheckAuth(ctx *gin.Context) {
//	header := ctx.GetHeader(constraints.HeaderAuthorization)
//	if header == "" {
//		response.SendError(ctx, http.StatusUnauthorized, "empty auth header")
//		return
//	}
//
//	headerParts := strings.Split(header, " ")
//	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
//		response.SendError(ctx, http.StatusUnauthorized, "invalid auth header")
//		return
//	}
//
//	if len(headerParts[1]) == 0 {
//		response.SendError(ctx, http.StatusUnauthorized, "token is invalid")
//		return
//	}
//
//	hexUserID, hexSessionID, err := h.jwt.Parse(headerParts[1])
//	if err != nil {
//		response.SendError(ctx, http.StatusUnauthorized, "token is invalid")
//		return
//	}
//
//	sessionID, err := primitive.ObjectIDFromHex(hexSessionID)
//	if err != nil {
//		response.SendError(ctx, http.StatusUnauthorized, "token is invalid")
//		return
//	}
//
//	isSessionForceClosed := h.storage.IsSessionForceClosed(ctx, sessionID)
//	if isSessionForceClosed {
//		response.SendError(ctx, http.StatusUnauthorized, "token is expired")
//		return
//	}
//
//	ctx.Set(constraints.ContextUserID, hexUserID)
//}

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
