package handler

import (
	"backend/internal/http-server/request"
	"backend/internal/http-server/response"
	"backend/internal/lib/logger/sl"
	"backend/internal/lib/random"
	"backend/internal/storage"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"net/http"
	neturl "net/url"
)

type CreateURLRequestBody struct {
	Url   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

// CreateURL     Creates a URL in database, assigned to user
// @Summary      Create URL
// @Description  Creates a URL in database, assigned to user
// @Security     AccessToken
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        input body       request.URL true "Url data"
// @Success      201  {object}    response.URLCreated
// @Failure      400  {object}    response.Error
// @Failure      401  {object}    response.Error
// @Failure      409  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /url             [post]
func (h *Handler) CreateURL(ctx *gin.Context) {
	var body request.URL

	if err := ctx.BindJSON(&body); err != nil {
		h.log.Error("can't decode request body")
		response.InvalidRequestBody(ctx)
		return
	}

	parsedUrl, isUrlValid := validateUrl(body.Url)
	if !isUrlValid {
		h.log.Error("url is invalid", slog.String("url", parsedUrl))
		response.SendError(ctx, http.StatusBadRequest, "url is invalid")
		return
	}

	alias := body.Alias
	if alias == "" {
		alias = random.Generate(AliasLength)
	}

	id := ctx.GetString(ContextUserID)
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		h.log.Error("can't parse user id fom hex string to primitive.ObjectID", sl.Err(err), slog.String("id", id))
		response.InvalidAuthToken(ctx)
		return
	}

	_, err = h.storage.CreateURL(ctx, parsedUrl, alias, userID)
	if errors.Is(err, storage.ErrAliasAlreadyExists) {
		h.log.Info("alias already exists", slog.String("alias", alias))
		response.SendError(ctx, http.StatusConflict, "alias already exists")
		return
	}

	if err != nil {
		h.log.Error("can't save url", sl.Err(err))
		response.SendError(ctx, http.StatusInternalServerError, "can't save url")
		return
	}

	ctx.JSON(http.StatusCreated, response.URLCreated{
		Url:   body.Url,
		Alias: alias,
	})
	h.log.Info("url saved", slog.String("url", parsedUrl), slog.String("alias", alias))
}

// DeleteURL     Deletes a URL
// @Summary      Delete URL
// @Description  Deletes an url from database
// @Security     AccessToken
// @Tags         url
// @Param        alias path string true "alias"
// @Produce      json
// @Success      200  {integer}     integer 1
// @Failure      401  {object}      response.Error
// @Failure      403  {object}      response.Error
// @Failure      404  {object}      response.Error
// @Failure      500  {object}      response.Error
// @Router       /url/{alias}       [delete]
func (h *Handler) DeleteURL(ctx *gin.Context) {
	alias := ctx.Param("alias")

	id := ctx.GetString(ContextUserID)
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		h.log.Error("can't parse user id fom hex string to primitive.ObjectID", sl.Err(err))
		response.InvalidAuthToken(ctx)
		return
	}

	url, err := h.storage.GetUrlByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			h.log.Info("url doesn't exists", slog.String("alias", alias))
		} else {
			h.log.Error("error while getting url", slog.String("alias", alias), sl.Err(err))
		}
		response.SendError(ctx, http.StatusNotFound, "url not found")
		return
	}

	if url.UserID != userID {
		response.SendError(ctx, http.StatusForbidden, "url was not created by you")
		return
	}

	err = h.storage.DeleteURL(ctx, alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			h.log.Info("no url to delete")
			response.SendError(ctx, http.StatusNotFound, "no url to delete")
		} else {
			h.log.Error("error while deleting url", slog.String("alias", alias), sl.Err(err))
			response.SendError(ctx, http.StatusInternalServerError, "failed to delete url")
		}
		return
	}

	ctx.Status(http.StatusOK)
	h.log.Info("url deleted", slog.String("alias", alias))
}

func (h *Handler) Redirect(ctx *gin.Context) {
	alias := ctx.Param("alias")

	url, err := h.storage.GetUrlByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			h.log.Info("url not found", slog.String("alias", alias))
		} else {
			h.log.Error("error while getting url", sl.Err(err))
		}
		response.SendError(ctx, http.StatusNotFound, "url not found")
		return
	}

	ctx.Redirect(http.StatusPermanentRedirect, url.Link)
	err = h.storage.IncrementRedirectsCounter(ctx, alias)
	if err != nil {
		h.log.Error("error while incrementing requests counter", slog.String("alias", alias))
	}

	h.log.Info("redirected", slog.String("url", url.Link), slog.String("alias", alias))
}

func validateUrl(rawUrl string) (string, bool) {
	parsedUrl, err := neturl.ParseRequestURI(rawUrl)
	if err != nil {
		return "", false
	}
	return parsedUrl.String(), true
}
