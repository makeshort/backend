package handler

import (
	"backend/internal/app/request"
	"backend/internal/app/response"
	"backend/internal/app/service/storage"
	"backend/internal/lib/logger/sl"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
	"net/http"
)

// Register      Creates a user in database
// @Summary      User registration
// @Description  Creates a user in database
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body         request.UserCreate true "User data"
// @Success      201  {object}      response.User
// @Failure      400  {object}      response.Error
// @Failure      409  {object}      response.Error
// @Failure      500  {object}      response.Error
// @Router       /auth/signup       [post]
func (h *Handler) Register(ctx *gin.Context) {
	var body request.UserCreate

	if err := ctx.BindJSON(&body); err != nil {
		h.log.Error("failed to decode body")
		response.InvalidRequestBody(ctx)
		return
	}

	isEmailValid := checkEmailValidity(body.Email)
	if !isEmailValid {
		h.log.Info("email is invalid")
		response.SendError(ctx, http.StatusBadRequest, "email is invalid")
		return
	}

	passwordHash := h.service.Hasher.Create(body.Password)

	userID, err := h.service.Storage.CreateUser(ctx, body.Email, body.Username, passwordHash)
	if errors.Is(err, storage.ErrUserAlreadyExists) {
		h.log.Info("user already exists")
		response.SendError(ctx, http.StatusConflict, "user with this email or username already exists")
		return
	}

	if err != nil {
		h.log.Error("error while saving user", sl.Err(err), slog.String("email", body.Email))
		response.SendError(ctx, http.StatusInternalServerError, "can't save user")
		return
	}

	ctx.JSON(http.StatusCreated, response.User{Email: body.Email, Username: body.Username})
	h.log.Info("user created", slog.String("id", userID.Hex()), slog.String("email", body.Email), slog.String("password_hash", passwordHash))
}

// Login          Creates a session
// @Summary       User login
// @Description   Creates a session
// @Tags          auth
// @Accept        json
// @Produce       json
// @Param         input body       request.UserLogin true "Account credentials"
// @Success       200  {object}    response.TokenPair
// @Failure       400  {object}    response.Error
// @Failure       500  {object}    response.Error
// @Router        /auth/session    [post]
func (h *Handler) Login(ctx *gin.Context) {
	var body request.UserLogin

	if err := ctx.BindJSON(&body); err != nil {
		h.log.Error("failed to decode request body", sl.Err(err))
		response.InvalidRequestBody(ctx)
		return
	}

	user, err := h.service.Storage.GetUserByCredentials(ctx, body.Email, h.service.Hasher.Create(body.Password))
	if err != nil {
		response.SendError(ctx, http.StatusNotFound, "user not found")
		return
	}

	tokenPair, err := h.service.TokenManager.GenerateTokenPair(user.ID.Hex())
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create token")
		return
	}

	_, err = h.service.Storage.CreateRefreshSession(ctx, user.ID, tokenPair.RefreshToken, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create refresh session")
		return
	}

	ctx.SetCookie(h.config.Cookie.RefreshToken.Name, tokenPair.RefreshToken, int(h.config.Token.Refresh.TTL.Seconds()), h.config.Cookie.RefreshToken.Path, h.config.Cookie.RefreshToken.Domain, false, true)

	ctx.JSON(http.StatusOK, response.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

// Logout        Delete session from database
// @Summary      User logout
// @Description  Delete session from database
// @Tags         auth
// @Produce      json
// @Success      200  {integer}   integer 1
// @Failure      401  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /auth/session    [delete]
func (h *Handler) Logout(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Message: "refresh token is invalid"})
		return
	}
	ctx.SetCookie(h.config.Cookie.RefreshToken.Name, "", -1, h.config.Cookie.RefreshToken.Path, h.config.Cookie.RefreshToken.Domain, false, true)

	err = h.service.Storage.DeleteRefreshSession(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, storage.ErrRefreshSessionNotFound) {
			response.SendError(ctx, http.StatusNotFound, "refresh session not found")
			return
		}
		response.SendError(ctx, http.StatusInternalServerError, "can't delete refresh session")
		return
	}

	ctx.Status(http.StatusOK)
}

// RefreshTokens     Create a new token pair
// @Summary          Token refresh
// @Description      Create a new token pair
// @Tags             auth
// @Produce          json
// @Success          200  {object}    response.TokenPair
// @Failure          403  {object}    response.Error
// @Failure          500  {object}    response.Error
// @Router           /auth/refresh    [post]
func (h *Handler) RefreshTokens(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie(h.config.Cookie.RefreshToken.Name)
	if err != nil {
		response.SendError(ctx, http.StatusForbidden, "no refresh token cookie")
		return
	}

	isRefreshTokenValid, userID := h.service.Storage.IsRefreshTokenValid(ctx, refreshToken)
	if !isRefreshTokenValid {
		response.SendError(ctx, http.StatusForbidden, "invalid refresh token")
		return
	}

	err = h.service.Storage.DeleteRefreshSession(ctx, refreshToken)
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't delete refresh session")
	}

	tokenPair, err := h.service.TokenManager.GenerateTokenPair(userID.Hex())
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create token")
		return
	}

	_, err = h.service.Storage.CreateRefreshSession(ctx, userID, tokenPair.RefreshToken, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create refresh session")
		return
	}

	ctx.SetCookie(h.config.Cookie.RefreshToken.Name, tokenPair.RefreshToken, int(h.config.Token.Refresh.TTL.Seconds()), h.config.Cookie.RefreshToken.Path, h.config.Cookie.RefreshToken.Domain, false, true)

	ctx.JSON(http.StatusOK, response.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}