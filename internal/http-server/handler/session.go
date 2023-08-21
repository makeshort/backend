package handler

import (
	"backend/internal/http-server/request"
	"backend/internal/http-server/response"
	"backend/internal/lib/logger/sl"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
		response.SendError(ctx, http.StatusNotFound, "user not found")
		return
	}

	tokenPair, err := h.tokenService.GenerateTokenPair(user.ID.Hex())
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create token")
		return
	}

	cookieMaxAge := 30 * 24 * time.Hour

	ctx.SetCookie("refresh_token", tokenPair.RefreshToken, int(cookieMaxAge.Seconds()), "/", "localhost", false, true)
	ctx.Header(HeaderAuthorization, fmt.Sprintf("Bearer %s", tokenPair.AccessToken))

	ctx.Status(http.StatusOK)
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
	ctx.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	ctx.Header(HeaderAuthorization, "")

	// TODO: Add refresh / access token to blacklist

	ctx.Status(http.StatusOK)
}
