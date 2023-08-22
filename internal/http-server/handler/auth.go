package handler

import (
	"backend/internal/http-server/request"
	"backend/internal/http-server/response"
	"backend/internal/lib/logger/sl"
	"backend/internal/storage"
	"backend/internal/token"
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

	passwordHash := h.hasher.Create(body.Password)

	userID, err := h.storage.CreateUser(ctx, body.Email, body.Username, passwordHash)
	if errors.Is(err, storage.ErrUserAlreadyExists) {
		h.log.Info("user already exists")
		response.SendError(ctx, http.StatusConflict, "user with this email already exists")
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
// @Param         input body       request.UserLogIn true "Account credentials"
// @Success       200  {object}    response.TokenPair
// @Failure       400  {object}    response.Error
// @Failure       500  {object}    response.Error
// @Router        /auth/session    [post]
func (h *Handler) Login(ctx *gin.Context) {
	var body request.UserLogIn

	if err := ctx.BindJSON(&body); err != nil {
		h.log.Error("failed to decode request body", sl.Err(err))
		response.InvalidRequestBody(ctx)
		return
	}

	user, err := h.storage.GetUserByCredentials(ctx, body.Email, h.hasher.Create(body.Password))
	if err != nil {
		response.SendError(ctx, http.StatusNotFound, "user not found")
		return
	}

	tokenPair, err := h.tokenService.GenerateTokenPair(user.ID.Hex())
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create token")
		return
	}

	_, err = h.storage.CreateRefreshSession(ctx, user.ID, tokenPair.RefreshToken, token.RefreshTokenTTL)
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create refresh session")
		return
	}

	ctx.SetCookie("refresh_token", tokenPair.RefreshToken, int(token.RefreshTokenTTL.Seconds()), "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, response.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

// Logout        Close a session
// @Summary      User logout
// @Description  Close a session
// @Security     AccessToken
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
	}
	ctx.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)

	err = h.storage.DeleteRefreshSession(ctx, refreshToken)
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

// Refresh       Create a new token pair
// @Summary      Token refresh
// @Description  Create a new token pair
// @Tags         auth
// @Param        input body       request.RefreshToken true "Refresh token"
// @Produce      json
// @Success      200  {object}    response.TokenPair
// @Failure      403  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /auth/refresh    [post]
func (h *Handler) Refresh(ctx *gin.Context) {
	var body request.RefreshToken

	if err := ctx.BindJSON(&body); err != nil {
		h.log.Error("failed to decode request body", sl.Err(err))
		response.InvalidRequestBody(ctx)
		return
	}

	isRefreshTokenValid, userID := h.storage.IsRefreshTokenValid(ctx, body.Token)
	if !isRefreshTokenValid {
		response.SendError(ctx, http.StatusForbidden, "invalid refresh token")
		return
	}

	err := h.storage.DeleteRefreshSession(ctx, body.Token)
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't delete refresh session")
	}

	tokenPair, err := h.tokenService.GenerateTokenPair(userID.Hex())
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create token")
		return
	}

	_, err = h.storage.CreateRefreshSession(ctx, userID, tokenPair.RefreshToken, token.RefreshTokenTTL)
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't create refresh session")
		return
	}

	ctx.SetCookie("refresh_token", tokenPair.RefreshToken, int(token.RefreshTokenTTL.Seconds()), "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, response.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}
