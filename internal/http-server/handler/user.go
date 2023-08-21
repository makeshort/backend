package handler

import (
	"backend/internal/http-server/request"
	"backend/internal/http-server/response"
	"backend/internal/lib/logger/sl"
	"backend/internal/storage"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"net/http"
	"net/mail"
)

// Register      Creates a user in database
// @Summary      User registration
// @Description  Creates a user in database
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        input body       request.UserCreate true "User data"
// @Success      201  {object}    response.User
// @Failure      400  {object}    response.Error
// @Failure      409  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /api/user        [post]
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

// DeleteMe      Delete me from database
// @Summary      Delete me
// @Description  Delete me from database
// @Security     SessionIDAuth
// @Tags         user
// @Produce      json
// @Success      200  {integer}   integer 1
// @Failure      401  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /api/user/me     [delete]
func (h *Handler) DeleteMe(ctx *gin.Context) {
	hexUserID := ctx.GetString(ContextUserID)
	userID, err := primitive.ObjectIDFromHex(hexUserID)

	if err != nil {
		h.log.Error("can't parse user ID form hex", slog.String("id", hexUserID), sl.Err(err))
		response.SendError(ctx, http.StatusUnauthorized, "can't parse auth token")
		return
	}

	err = h.storage.DeleteUser(ctx, userID)
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
// @Security     SessionIDAuth
// @Description  Get all URLs created by user
// @Tags         user
// @Produce      json
// @Success      200  {array}         response.URL
// @Failure      401  {object}        response.Error
// @Failure      500  {object}        response.Error
// @Router       /api/user/me/urls    [get]
func (h *Handler) GetMyURLs(ctx *gin.Context) {
	hexUserID := ctx.GetString(ContextUserID)
	userID, err := primitive.ObjectIDFromHex(hexUserID)
	if err != nil {
		h.log.Error("can't parse user id fom hex string to primitive.ObjectID", sl.Err(err))
		response.InvalidAuthToken(ctx)
		return
	}

	urlDocs, err := h.storage.GetUserURLs(ctx, userID)
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
