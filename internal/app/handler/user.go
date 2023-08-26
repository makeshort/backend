package handler

import (
	"backend/internal/app/middleware"
	"backend/internal/app/response"
	"backend/internal/app/service/storage"
	"backend/internal/lib/logger/sl"
	"backend/pkg/requestid"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Error("error occurred while parsing user id from hex to ObjectID",
			slog.String("id", id),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusNotFound, "id is invalid")
		return
	}

	user, err := h.service.Storage.GetUserByID(ctx, userID)
	if errors.Is(err, storage.ErrUserNotFound) {
		log.Debug("user not found",
			slog.String("id", id),
		)
		response.SendError(ctx, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Error("user not found",
			slog.String("id", id),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't get user")
		return
	}

	ctx.JSON(http.StatusOK, response.User{
		ID:       user.ID.Hex(),
		Email:    user.Email,
		Username: user.Username,
	})
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

	hexUserID := ctx.GetString(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(hexUserID)
	if err != nil {
		log.Error("error occurred while parsing user id from hex to ObjectID",
			slog.String("id", hexUserID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusUnauthorized, "can't parse auth token")
		return
	}

	err = h.service.Storage.DeleteUser(ctx, userID)
	if errors.Is(err, storage.ErrUserNotFound) {
		log.Info("user not found")
		response.SendError(ctx, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Error("error occurred while deleting user",
			slog.String("id", hexUserID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't delete user")
		return
	}

	ctx.Status(http.StatusOK)
	log.Info("user deleted",
		slog.String("id", hexUserID),
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

	hexUserID := ctx.GetString(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(hexUserID)
	if err != nil {
		log.Error("error occurred while parsing user id from hex to ObjectID",
			slog.String("id", hexUserID),
			sl.Err(err),
		)
		response.SendAuthFailedError(ctx)
		return
	}

	urlDocs, err := h.service.Storage.GetUserURLs(ctx, userID)
	if err != nil {
		log.Error("error occurred while getting user urls",
			slog.String("id", hexUserID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't get urls")
		return
	}

	urls := make([]response.URL, len(urlDocs))
	for i, url := range urlDocs {
		urls[i].ID = url.ID.Hex()
		urls[i].Url = url.Link
		urls[i].Alias = url.Alias
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
