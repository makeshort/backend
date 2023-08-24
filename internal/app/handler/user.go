package handler

import (
	"backend/internal/app/middleware"
	"backend/internal/app/response"
	"backend/internal/app/service/storage"
	"backend/internal/lib/logger/sl"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"net/http"
	"net/mail"
)

// DeleteMe      Delete me from database
// @Summary      Delete me
// @Description  Delete me from database
// @Security     AccessToken
// @Tags         user
// @Produce      json
// @Success      200  {integer}   integer 1
// @Failure      401  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /user/me         [delete]
func (h *Handler) DeleteMe(ctx *gin.Context) {
	hexUserID := ctx.GetString(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(hexUserID)

	if err != nil {
		h.log.Error("can't parse user ID form hex", slog.String("id", hexUserID), sl.Err(err))
		response.SendError(ctx, http.StatusUnauthorized, "can't parse auth token")
		return
	}

	err = h.service.Storage.DeleteUser(ctx, userID)
	if errors.Is(err, storage.ErrUserNotFound) {
		h.log.Info("user not found")
		response.SendError(ctx, http.StatusInternalServerError, "user not found")
		return
	}

	if err != nil {
		h.log.Error("error while deleting user", sl.Err(err), slog.String("id", hexUserID))
		response.SendError(ctx, http.StatusInternalServerError, "can't delete user")
		return
	}

	ctx.Status(http.StatusOK)
	h.log.Info("user created", slog.String("id", hexUserID))
}

// GetMyURLs     Gets all url documents assigned to given UserID.
// @Summary      Get URLs
// @Security     AccessToken
// @Description  Get all URLs created by user
// @Tags         user
// @Produce      json
// @Success      200  {array}         response.URL
// @Failure      401  {object}        response.Error
// @Failure      500  {object}        response.Error
// @Router       /user/me/urls        [get]
func (h *Handler) GetMyURLs(ctx *gin.Context) {
	hexUserID := ctx.GetString(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(hexUserID)
	if err != nil {
		h.log.Error("can't parse user id fom hex string to primitive.ObjectID", sl.Err(err))
		response.InvalidAuthToken(ctx)
		return
	}

	urlDocs, err := h.service.Storage.GetUserURLs(ctx, userID)
	if err != nil {
		h.log.Error("can't get urls", slog.String("id", hexUserID), sl.Err(err))
		response.SendError(ctx, http.StatusInternalServerError, "can't get urls")
		return
	}

	urls := make([]response.URL, len(urlDocs))
	for i, url := range urlDocs {
		urls[i].Url = url.Link
		urls[i].Alias = url.Alias
		urls[i].Redirects = url.Redirects
	}
	ctx.JSON(http.StatusOK, urls)
}

func checkEmailValidity(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
