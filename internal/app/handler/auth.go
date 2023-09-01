package handler

import (
	"backend/internal/app/request"
	"backend/internal/app/response"
	"backend/internal/app/service/auth"
	"backend/internal/lib/logger/sl"
	"backend/internal/repository"
	"backend/internal/repository/postgres/user"
	"backend/pkg/requestid"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

// Register      Creates a user in database.
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
	log := h.log.With(
		slog.String("op", "handler.Register"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	var body request.UserCreate

	if err := ctx.BindJSON(&body); err != nil {
		log.Debug("error occurred while decode request body", sl.Err(err))
		response.SendInvalidRequestBodyError(ctx)
		return
	}

	id, err := h.service.Auth.Register(ctx, auth.UserRegisterDTO{
		Email:    body.Email,
		Username: body.Username,
		Password: body.Password,
	})
	if errors.Is(err, user.ErrUserAlreadyExists) || errors.Is(err, auth.ErrInvalidEmail) {
		log.Debug(err.Error(),
			slog.String("email", body.Username),
			slog.String("username", body.Username),
		)
		response.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		log.Error("error occurred while registering user",
			sl.Err(err),
			slog.String("email", body.Email),
		)
		response.SendSomethingWentWrong(ctx)
		return
	}

	log.Info("user successfully created",
		slog.String("id", id),
		slog.String("username", body.Username),
		slog.String("email", body.Email),
	)

	ctx.JSON(http.StatusCreated, response.User{
		ID:       id,
		Email:    body.Email,
		Username: body.Username,
	})
}

// Login          Creates a session.
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
	log := h.log.With(
		slog.String("op", "handler.Login"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	var body request.UserLogin

	if err := ctx.BindJSON(&body); err != nil {
		log.Debug("error occurred while decode request body", sl.Err(err))
		response.SendInvalidRequestBodyError(ctx)
		return
	}

	tokenPair, err := h.service.Auth.Login(ctx, auth.UserLoginDTO{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		log.Error("error occurred while user logging in", sl.Err(err))
		response.SendSomethingWentWrong(ctx)
		return
	}

	ctx.SetCookie(h.config.Cookie.RefreshToken.Name, tokenPair.RefreshToken, int(h.config.Token.Refresh.TTL.Seconds()), h.config.Cookie.RefreshToken.Path, h.config.Cookie.RefreshToken.Domain, false, true)

	ctx.JSON(http.StatusOK, response.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

// Logout        Delete session from database.
// @Summary      User logout
// @Description  Delete session from database
// @Tags         auth
// @Produce      json
// @Success      200  {integer}   integer 1
// @Failure      401  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /auth/session    [delete]
func (h *Handler) Logout(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.Logout"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	refreshToken, err := ctx.Cookie(h.config.Cookie.RefreshToken.Name)
	if err != nil {
		log.Debug("no refresh token cookie found",
			sl.Err(err),
		)
		ctx.JSON(http.StatusBadRequest, response.Error{Message: "no refresh token cookie found"})
		return
	}
	ctx.SetCookie(h.config.Cookie.RefreshToken.Name, "", -1, h.config.Cookie.RefreshToken.Path, h.config.Cookie.RefreshToken.Domain, false, true)

	err = h.service.Auth.Logout(ctx, refreshToken)
	if errors.Is(err, repository.ErrRefreshSessionNotFound) {
		log.Debug("refresh session not found")
		response.SendError(ctx, http.StatusNotFound, "refresh session not found")
		return
	}
	if err != nil {
		response.SendError(ctx, http.StatusInternalServerError, "can't delete refresh session")
		log.Error("error occurred while deleting refresh session",
			sl.Err(err),
		)
		return
	}

	ctx.Status(http.StatusOK)
}

// RefreshTokens     Create a new token pair.
// @Summary          Token refresh
// @Description      Create a new token pair
// @Tags             auth
// @Produce          json
// @Success          200  {object}    response.TokenPair
// @Failure          403  {object}    response.Error
// @Failure          500  {object}    response.Error
// @Router           /auth/refresh    [post]
func (h *Handler) RefreshTokens(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.RefreshTokens"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	refreshToken, err := ctx.Cookie(h.config.Cookie.RefreshToken.Name)
	if err != nil {
		log.Debug("no refresh token cookie found",
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusForbidden, "no refresh token cookie found")
		return
	}

	tokenPair, err := h.service.Auth.Refresh(ctx, refreshToken)
	if errors.Is(err, auth.ErrInvalidRefreshToken) {
		log.Error(err.Error())
		response.SendError(ctx, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	if err != nil {
		log.Error("error occurred while refreshing session", sl.Err(err))
		response.SendSomethingWentWrong(ctx)
		return
	}

	log.Info("refresh session successfully created")

	ctx.SetCookie(h.config.Cookie.RefreshToken.Name, tokenPair.RefreshToken, int(h.config.Token.Refresh.TTL.Seconds()), h.config.Cookie.RefreshToken.Path, h.config.Cookie.RefreshToken.Domain, false, true)

	ctx.JSON(http.StatusOK, response.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}
