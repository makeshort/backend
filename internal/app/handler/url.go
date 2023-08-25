package handler

import (
	"backend/internal/app/middleware"
	"backend/internal/app/request"
	"backend/internal/app/response"
	"backend/internal/app/service/storage"
	"backend/internal/lib/logger/sl"
	"backend/internal/lib/random"
	"backend/pkg/requestid"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"net/http"
	neturl "net/url"
)

const AliasLength = 6

// CreateUrl     Creates a URL in database, assigned to user.
// @Summary      Create URL
// @Description  Creates a URL in database, assigned to user
// @Security     AccessToken
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        input body       request.URL true "Url data"
// @Success      201  {object}    response.UrlCreated
// @Failure      400  {object}    response.Error
// @Failure      401  {object}    response.Error
// @Failure      409  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /url             [post]
func (h *Handler) CreateUrl(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.CreateUrl"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	var body request.URL

	if err := ctx.BindJSON(&body); err != nil {
		log.Debug("error occurred while decode request body", sl.Err(err))
		response.SendInvalidRequestBodyError(ctx)
		return
	}

	parsedUrl, isUrlValid := validateUrl(body.Url)
	if !isUrlValid {
		log.Error("provided url is in invalid format",
			slog.String("url", body.Url),
		)
		response.SendError(ctx, http.StatusBadRequest, "url is invalid")
		return
	}

	alias := body.Alias
	if alias == "" {
		alias = random.Generate(AliasLength)
	}

	id := ctx.GetString(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Error("error occurred while parsing user id from hex to ObjectID",
			slog.String("id", id),
			sl.Err(err),
		)
		response.SendAuthFailedError(ctx)
		return
	}

	urlID, err := h.service.Storage.CreateURL(ctx, parsedUrl, alias, userID)
	if errors.Is(err, storage.ErrAliasAlreadyExists) {
		log.Debug("alias already exists",
			slog.String("alias", alias),
		)
		response.SendError(ctx, http.StatusConflict, "alias already exists")
		return
	}
	if err != nil {
		log.Error("error occurred while saving url to database",
			slog.String("url", parsedUrl),
			slog.String("alias", alias),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't save url")
		return
	}

	ctx.JSON(http.StatusCreated, response.UrlCreated{
		ID:    urlID.Hex(),
		Url:   body.Url,
		Alias: alias,
	})
	log.Info("url saved",
		slog.String("id", urlID.Hex()),
		slog.String("url", parsedUrl),
		slog.String("alias", alias),
	)
}

// UpdateUrl     Updates an URL.
// @Summary      Update URL
// @Description  Updates an url
// @Security     AccessToken
// @Tags         url
// @Param        id path string true "id"
// @Produce      json
// @Param        input body       request.URL true "Url data"
// @Success      200  {integer}     integer 1
// @Failure      401  {object}      response.Error
// @Failure      403  {object}      response.Error
// @Failure      404  {object}      response.Error
// @Failure      500  {object}      response.Error
// @Router       /url/{id}          [patch]
func (h *Handler) UpdateUrl(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.UpdateUrl"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	hexUrlID := ctx.Param("id")
	urlID, err := primitive.ObjectIDFromHex(hexUrlID)
	if err != nil {
		log.Error("error occurred while parsing url id from hex to ObjectID",
			slog.String("id", hexUrlID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "url id is invalid")
		return
	}

	var body request.UrlUpdate

	if err = ctx.BindJSON(&body); err != nil {
		log.Debug("error occurred while decode request body", sl.Err(err))
		response.SendInvalidRequestBodyError(ctx)
		return
	}

	parsedUrl, isUrlValid := validateUrl(body.Url)
	if body.Url != "" && !isUrlValid {
		log.Error("provided url is in invalid format",
			slog.String("url", body.Url),
		)
		response.SendError(ctx, http.StatusBadRequest, "url is invalid")
		return
	}

	err = h.service.Storage.UpdateUrl(ctx, urlID, body.Alias, parsedUrl)
	if err != nil {
		log.Error("error occurred while updating url",
			slog.String("id", hexUrlID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't update url")
		return
	}

	ctx.JSON(http.StatusOK, response.UrlUpdated{
		ID:    hexUrlID,
		Url:   body.Url,
		Alias: body.Alias,
	})
}

// DeleteUrl     Deletes a URL.
// @Summary      Delete URL
// @Description  Deletes an url from database
// @Security     AccessToken
// @Tags         url
// @Param        id path string true "id"
// @Produce      json
// @Success      200  {integer}     integer 1
// @Failure      401  {object}      response.Error
// @Failure      403  {object}      response.Error
// @Failure      404  {object}      response.Error
// @Failure      500  {object}      response.Error
// @Router       /url/{id}          [delete]
func (h *Handler) DeleteUrl(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.DeleteUrl"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	hexUrlID := ctx.Param("id")

	urlID, err := primitive.ObjectIDFromHex(hexUrlID)
	if err != nil {
		log.Error("error occurred while parsing user id from hex to ObjectID",
			slog.String("id", hexUrlID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusNotFound, "no url with this id")
		return
	}

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

	url, err := h.service.Storage.GetUrlByID(ctx, urlID)
	if errors.Is(err, storage.ErrURLNotFound) {
		log.Debug("url doesn't exists",
			slog.String("id", hexUrlID),
		)
		response.SendError(ctx, http.StatusNotFound, "url not found")
		return
	}
	if err != nil {
		log.Error("error while getting url",
			slog.String("id", hexUrlID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't get url")
		return
	}

	if url.UserID != userID {
		log.Debug("now url's owner",
			slog.String("id", hexUrlID),
			slog.String("alias", url.Alias),
			slog.String("user_id", userID.Hex()),
			slog.String("owner_id", url.UserID.Hex()),
		)
		response.SendError(ctx, http.StatusForbidden, "url was not created by you")
		return
	}

	err = h.service.Storage.DeleteURL(ctx, url.Alias)
	if errors.Is(err, storage.ErrURLNotFound) {
		log.Debug("no url to delete",
			slog.String("alias", url.Alias),
		)
		response.SendError(ctx, http.StatusNotFound, "no url to delete")
		return
	}
	if err != nil {
		log.Error("error occurred while deleting url",
			slog.String("id", hexUrlID),
			slog.String("alias", url.Alias),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "failed to delete url")
		return
	}

	ctx.Status(http.StatusOK)
	log.Info("url deleted successfully",
		slog.String("id", hexUrlID),
		slog.String("alias", url.Alias),
	)
}

// Redirect redirects user from /{alias} to URL assigned to this alias.
func (h *Handler) Redirect(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.DeleteUrl"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	alias := ctx.Param("alias")

	url, err := h.service.Storage.GetUrlByAlias(ctx, alias)
	if errors.Is(err, storage.ErrURLNotFound) {
		response.SendError(ctx, http.StatusNotFound, "url not found")
		return
	}
	if err != nil {
		log.Error("error occurred while getting url",
			slog.String("alias", alias),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't found url")
		return
	}

	err = h.service.Storage.IncrementRedirectsCounter(ctx, url.Alias)
	if err != nil {
		log.Error("error while incrementing requests counter",
			slog.String("alias", alias),
			sl.Err(err),
		)
	}

	log.Debug("redirected",
		slog.String("url", url.Link),
		slog.String("alias", alias),
	)
	ctx.Redirect(http.StatusPermanentRedirect, url.Link)
}

// validateUrl validates URL and return validated email and boolean is email valid.
func validateUrl(rawUrl string) (string, bool) {
	parsedUrl, err := neturl.ParseRequestURI(rawUrl)
	if err != nil {
		return "", false
	}
	return parsedUrl.String(), true
}
