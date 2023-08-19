package handler

import (
	"backend/internal/http-server/constraints"
	"backend/internal/http-server/request"
	"backend/internal/http-server/response"
	al "backend/internal/lib/alias"
	"backend/internal/lib/logger/sl"
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

// CreateURL     Creates a URL in database. Assigned to UserID.
// @Summary      Create
// @Security     SessionIDAuth
// @Description  create an url in database
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        input body       request.URL true "Url data"
// @Success      201  {object}    response.URLCreated
// @Failure      400  {object}    response.Error
// @Failure      401  {object}    response.Error
// @Failure      409  {object}    response.Error
// @Failure      500  {object}    response.Error
// @Router       /api/url         [post]
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

	if body.Alias == "" {
		body.Alias = al.Generate(constraints.AliasLength)
	}

	id := ctx.GetString(constraints.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		h.log.Error("can't parse user id fom hex string to primitive.ObjectID", sl.Err(err))
		response.InvalidAuthToken(ctx)
		return
	}

	_, err = h.storage.CreateURL(ctx, parsedUrl, body.Alias, userID)
	if errors.Is(err, storage.ErrAliasAlreadyExists) {
		h.log.Info("alias already exists", slog.String("alias", body.Alias))
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
		Alias: body.Alias,
	})
	h.log.Info("url saved", slog.String("alias", body.Alias), slog.String("url", parsedUrl))
}

// DeleteURL     Deletes a URL.
// @Summary      Delete
// @Security     SessionIDAuth
// @Description  delete an url from database
// @Tags         url
// @Produce      json
// @Success      201  {object}      response.Success
// @Failure      400  {object}      response.Error
// @Failure      403  {object}      response.Error
// @Failure      404  {object}      response.Error
// @Failure      500  {object}      response.Error
// @Router       /api/url/:alias    [delete]
func (h *Handler) DeleteURL(ctx *gin.Context) {
	alias := ctx.Param("alias")

	id := ctx.GetString(constraints.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		h.log.Error("can't parse user id fom hex string to primitive.ObjectID", sl.Err(err))
		response.InvalidAuthToken(ctx)
		return
	}

	url, err := h.storage.GetURL(ctx, alias)
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

	ctx.JSON(http.StatusOK, response.Success{Message: "url deleted"})
	h.log.Info("url deleted", slog.String("alias", alias))
}

// Redirect      Redirects user from alias to it's url.
// @Summary      Redirect
// @Description  redirect from alias to it's url
// @Tags         url
// @Produce      json
// @Success      308  {integer}   integer 1
// @Failure      404  {object}    response.Error
// @Router       /:alias          [get]
func (h *Handler) Redirect(ctx *gin.Context) {
	alias := ctx.Param("alias")

	url, err := h.storage.GetURL(ctx, alias)
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
	err = h.storage.IncrementUrlCounter(ctx, alias)
	if err != nil {
		h.log.Error("error while incrementing requests counter", slog.String("alias", alias))
	}

	h.log.Info("redirected", slog.String("alias", alias), slog.String("url", url.Link))
}

func validateUrl(rawUrl string) (string, bool) {
	parsedUrl, err := neturl.ParseRequestURI(rawUrl)
	if err != nil {
		return "", false
	}
	return parsedUrl.String(), true
}
