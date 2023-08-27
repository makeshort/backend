package handler

import (
	"backend/internal/app/request"
	"backend/internal/app/response"
	"backend/internal/app/service/storage"
	"backend/internal/lib/logger/sl"
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

	isEmailValid := checkEmailValidity(body.Email)
	if !isEmailValid {
		log.Debug("email is invalid",
			slog.String("email", body.Email),
		)
		response.SendError(ctx, http.StatusBadRequest, "email is invalid")
		return
	}

	passwordHash := h.service.Hasher.Create(body.Password)

	id, err := h.service.Repository.User.Create(ctx, body.Email, body.Username, passwordHash)
	if err != nil {
		log.Error("error occurred while creating user", sl.Err(err))
		response.SendError(ctx, http.StatusInternalServerError, "can't create user")
		return
	}

	log.Info("user created",
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

	passwordHash := h.service.Hasher.Create(body.Password)

	user, err := h.service.Repository.User.GetByCredentials(ctx, body.Email, passwordHash)
	if err != nil {
		log.Debug("user not found in database",
			slog.String("email", body.Email),
		)
		response.SendError(ctx, http.StatusBadRequest, "user not found")
		return
	}

	tokenPair, err := h.service.TokenManager.GenerateTokenPair(user.ID)
	if err != nil {
		log.Error("error occurred while generating token pair",
			slog.String("user_id", user.ID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't create token pair")
		return
	}

	err = h.service.Repository.Session.Create(ctx, tokenPair.RefreshToken, user.ID, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		log.Error("error occurred while creating refresh session in database", sl.Err(err))
		response.SendError(ctx, http.StatusInternalServerError, "can't create refresh session")
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

	err = h.service.Repository.Session.Close(ctx, refreshToken)
	if errors.Is(err, storage.ErrRefreshSessionNotFound) {
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

	session, err := h.service.Repository.Session.Get(ctx, refreshToken)
	if err != nil {
		log.Debug("invalid refresh token")
		response.SendError(ctx, http.StatusForbidden, "invalid refresh token")
		return
	}

	err = h.service.Repository.Session.Close(ctx, refreshToken)
	if err != nil {
		log.Error("error occurred while deleting refresh session", sl.Err(err))
		response.SendError(ctx, http.StatusInternalServerError, "can't delete refresh session")
	}

	tokenPair, err := h.service.TokenManager.GenerateTokenPair(session.UserID)
	if err != nil {
		log.Error("error occurred while generating token pair",
			slog.String("user_id", session.UserID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't create token pair")
		return
	}

	err = h.service.Repository.Session.Create(ctx, tokenPair.RefreshToken, session.UserID, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		log.Error("error occurred while creating refresh session",
			slog.String("user_id", session.UserID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't create refresh session")
		return
	}

	log.Info("refresh session successfully created",
		slog.String("user_id", session.UserID),
	)

	ctx.SetCookie(h.config.Cookie.RefreshToken.Name, tokenPair.RefreshToken, int(h.config.Token.Refresh.TTL.Seconds()), h.config.Cookie.RefreshToken.Path, h.config.Cookie.RefreshToken.Domain, false, true)

	ctx.JSON(http.StatusOK, response.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}
