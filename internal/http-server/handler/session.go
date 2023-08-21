package handler

import (
	"backend/internal/http-server/request"
	"backend/internal/http-server/response"
	"backend/internal/lib/logger/sl"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

// Login          Creates a session
// @Summary       User login
// @Description   Creates a session
// @Tags          session
// @Accept        json
// @Produce       json
// @Param         input body       request.UserLogIn true "Account credentials"
// @Success       200  {object}    response.TokenPair
// @Failure       400  {object}    response.Error
// @Failure       500  {object}    response.Error
// @Router        /session         [post]
func (h *Handler) Login(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, response.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

// Logout        Close a session
// @Summary      User logout
// @Description  Close a session
// @Security     AccessToken
// @Tags         session
// @Produce      json
// @Success      200  {integer}   integer 1
// @Failure      401  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /session         [delete]
func (h *Handler) Logout(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Message: "refresh token is invalid"})
	}
	ctx.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	ctx.Header(HeaderAuthorization, "")

	claims, err := h.tokenService.ParseRefreshToken(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Message: "refresh token is invalid"})
	}

	_, err = h.storage.BlacklistToken(ctx, refreshToken, primitive.NewDateTimeFromTime(time.Unix(claims.ExpiresAt, 0)))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error{Message: "can't blacklist refresh token"})
	}

	ctx.Status(http.StatusOK)
}
