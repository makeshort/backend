package handler

import (
	"backend/internal/http-server/constraints"
	"backend/internal/http-server/request"
	"backend/internal/http-server/response"
	"backend/internal/lib/logger/sl"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"net/http"
)

// CreateSession  Creates a session
// @Summary       Create
// @Description   create a session
// @Tags          session
// @Accept        json
// @Produce       json
// @Param         input body       request.UserLogIn true "Account credentials"
// @Success       200  {object}    response.Token
// @Failure       400  {object}    response.Error
// @Failure       500  {object}    response.Error
// @Router        /api/session     [post]
func (h *Handler) CreateSession(ctx *gin.Context) {
	var body request.UserLogIn

	if err := ctx.BindJSON(&body); err != nil {
		h.log.Error("failed to decode request body", sl.Err(err))
		response.InvalidRequestBody(ctx)
		return
	}

	user, err := h.storage.GetUser(ctx, body.Email, h.hasher.Create(body.Password))
	if err != nil {
		response.SendError(ctx, http.StatusBadRequest, "User not found")
		return
	}

	sessionID, err := h.storage.CreateSession(ctx, user.ID)
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "Can't create session")
	}

	ctx.JSON(http.StatusOK, response.Session{SessionID: sessionID.Hex()})
}

// CloseSession  Close a session
// @Summary      Close
// @Security     SessionIDAuth
// @Description  close a session
// @Tags         session
// @Produce      json
// @Success      200  {object}    response.Success
// @Failure      401  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /api/session     [delete]
func (h *Handler) CloseSession(ctx *gin.Context) {
	hexSessionID := ctx.GetHeader(constraints.HeaderSessionID)

	sessionID, err := primitive.ObjectIDFromHex(hexSessionID)
	if err != nil {
		h.log.Error("can't parse hex session id to primitive.ObjectID")
		response.SendError(ctx, http.StatusUnauthorized, "Session ID is invalid")
		return
	}

	err = h.storage.DeleteSession(ctx, sessionID)
	if err != nil {
		h.log.Error("can't close session", slog.String("hex_session_id", hexSessionID), sl.Err(err))
		response.SendError(ctx, http.StatusInternalServerError, "Can't log out")
		return
	}

	ctx.Status(http.StatusOK)
}
