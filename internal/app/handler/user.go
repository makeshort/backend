package handler

import (
	"backend/internal/app/middleware"
	"backend/internal/app/request"
	"backend/internal/app/response"
	"backend/internal/lib/logger/sl"
	"backend/internal/service/repository"
	"backend/internal/service/repository/postgres/user"
	"backend/pkg/requestid"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"net/mail"
)

// GetUser       Get user's information.
// @Summary      Get user
// @Description  Get user's information
// @Tags         user
// @Param        id path string true "id"
// @Produce      json
// @Success      200  {object}        response.User
// @Failure      404  {object}        response.Error
// @Failure      500  {object}        response.Error
// @Router       /user/{id}           [get]
func (h *Handler) GetUser(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.GetUser"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	id := ctx.Param("id")

	user, err := h.service.Repository.User.GetByID(ctx, id)
	if err != nil {
		log.Error("error occurred while getting user",
			slog.String("id", id),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't get user")
		return
	}

	ctx.JSON(http.StatusOK, response.User{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	})
}

// UpdateUser    Update user entity in  database.
// @Summary      Update me
// @Description  Update user entity in  database
// @Security     AccessToken
// @Tags         user
// @Param        id path string true "id"
// @Produce      json
// @Success      200  {integer}   integer 1
// @Failure      400  {object}    response.Error
// @Failure      401  {object}    response.Error
// @Failure      404  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /user/{id}       [patch]
func (h *Handler) UpdateUser(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.UpdateUser"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	var body request.UserUpdate

	if err := ctx.BindJSON(&body); err != nil {
		log.Debug("error occurred while decode request body", sl.Err(err))
		response.SendInvalidRequestBodyError(ctx)
		return
	}

	userID := ctx.GetString(middleware.ContextUserID)

	updatedUser, err := h.service.Repository.User.Update(ctx, userID, user.DTO{
		Email:        body.Email,
		Username:     body.Username,
		PasswordHash: h.service.Hasher.Create(body.Password),
		TelegramID:   body.TelegramID,
	})
	if errors.Is(err, repository.ErrUserNotFound) {
		log.Info("user not found")
		response.SendError(ctx, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Error("error occurred while updating user",
			slog.String("id", userID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't update user")
		return
	}

	ctx.Status(http.StatusOK)
	log.Info("user updated",
		slog.String("id", userID),
		slog.String("username", updatedUser.Username),
		slog.String("email", updatedUser.Email),
		slog.String("password_hash", updatedUser.PasswordHash),
		slog.String("telegram_id", updatedUser.TelegramID),
	)
}

// DeleteUser    Delete me from database.
// @Summary      Delete me
// @Description  Delete me from database
// @Security     AccessToken
// @Tags         user
// @Param        id path string true "id"
// @Produce      json
// @Success      200  {integer}   integer 1
// @Failure      401  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /user/{id}       [delete]
func (h *Handler) DeleteUser(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.DeleteUser"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	userID := ctx.GetString(middleware.ContextUserID)

	err := h.service.Repository.User.Delete(ctx, userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		log.Info("user not found")
		response.SendError(ctx, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Error("error occurred while deleting user",
			slog.String("id", userID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't delete user")
		return
	}

	ctx.Status(http.StatusOK)
	log.Info("user deleted",
		slog.String("id", userID),
	)
}

// GetUserUrls   Gets all url documents assigned to given UserID.
// @Summary      Get URLs
// @Security     AccessToken
// @Description  Get all URLs created by user
// @Tags         user
// @Param        id path string true "id"
// @Produce      json
// @Success      200  {array}         response.URL
// @Failure      401  {object}        response.Error
// @Failure      500  {object}        response.Error
// @Router       /user/{id}/urls      [get]
func (h *Handler) GetUserUrls(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.GetUserUrls"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	id := ctx.GetString(middleware.ContextUserID)

	urlDocs, err := h.service.Repository.User.GetUrlsList(ctx, id)
	if err != nil {
		log.Error("error occurred while getting user urls",
			slog.String("id", id),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't get urls")
		return
	}

	urls := make([]response.URL, len(urlDocs))
	for i, url := range urlDocs {
		urls[i].ID = url.ID
		urls[i].Url = url.LongURL
		urls[i].Alias = url.ShortURL
		urls[i].Redirects = url.Redirects
	}
	if len(urls) == 0 {
		ctx.Status(http.StatusNoContent)
	}
	ctx.JSON(http.StatusOK, urls)
}

// checkEmailValidity checks is email valid.
func checkEmailValidity(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
