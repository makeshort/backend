package handler

import (
	"backend/internal/app/middleware"
	"backend/internal/app/request"
	"backend/internal/app/response"
	"backend/internal/lib/logger/sl"
	"backend/internal/lib/random"
	"backend/internal/service/repository"
	repoUrl "backend/internal/service/repository/postgres/url"
	"backend/pkg/requestid"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
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

	userID := ctx.GetString(middleware.ContextUserID)

	urlID, err := h.service.Repository.Url.Create(ctx, userID, parsedUrl, alias)
	if errors.Is(err, repository.ErrAliasAlreadyExists) {
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
		ID:    urlID,
		Url:   body.Url,
		Alias: alias,
	})
	log.Info("url saved",
		slog.String("id", urlID),
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

	urlID := ctx.Param("id")

	var body request.UrlUpdate

	if err := ctx.BindJSON(&body); err != nil {
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

	url, err := h.service.Repository.Url.Update(ctx, urlID, repoUrl.DTO{
		LongURL:  body.Alias,
		ShortURL: parsedUrl,
	})
	if err != nil {
		log.Error("error occurred while updating url",
			slog.String("id", urlID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't update url")
		return
	}

	ctx.JSON(http.StatusOK, response.UrlUpdated{
		ID:    urlID,
		Url:   url.LongURL,
		Alias: url.ShortURL,
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

	urlID := ctx.Param("id")

	userID := ctx.GetString(middleware.ContextUserID)

	url, err := h.service.Repository.Url.GetByID(ctx, urlID)
	if errors.Is(err, repository.ErrURLNotFound) {
		log.Debug("url doesn't exists",
			slog.String("id", urlID),
		)
		response.SendError(ctx, http.StatusNotFound, "url not found")
		return
	}
	if err != nil {
		log.Error("error while getting url",
			slog.String("id", urlID),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "can't get url")
		return
	}

	if *url.UserID != userID {
		log.Debug("now url's owner",
			slog.String("id", urlID),
			slog.String("alias", url.ShortURL),
			slog.String("user_id", userID),
			slog.String("owner_id", *url.UserID),
		)
		response.SendError(ctx, http.StatusForbidden, "url was not created by you")
		return
	}

	err = h.service.Repository.Url.Delete(ctx, url.ID)
	if errors.Is(err, repository.ErrURLNotFound) {
		log.Debug("no url to delete",
			slog.String("alias", url.ShortURL),
		)
		response.SendError(ctx, http.StatusNotFound, "no url to delete")
		return
	}
	if err != nil {
		log.Error("error occurred while deleting url",
			slog.String("id", urlID),
			slog.String("alias", url.ShortURL),
			sl.Err(err),
		)
		response.SendError(ctx, http.StatusInternalServerError, "failed to delete url")
		return
	}

	ctx.Status(http.StatusOK)
	log.Info("url deleted successfully",
		slog.String("id", urlID),
		slog.String("alias", url.ShortURL),
	)
}

// Redirect redirects user from /{alias} to URL assigned to this alias.
// Redirect      Redirects to an URL.
// @Summary      Redirect to URL
// @Description  Redirects to an URL
// @Tags         url
// @Param        alias path string true "alias"
// @Success      308  {integer}     integer 1
// @Failure      404  {object}      response.Error
// @Failure      500  {object}      response.Error
// @Router       /{alias}           [get]
func (h *Handler) Redirect(ctx *gin.Context) {
	log := h.log.With(
		slog.String("op", "handler.Redirect"),
		slog.String("request_id", requestid.Get(ctx)),
	)

	alias := ctx.Param("alias")

	url, err := h.service.Repository.Url.GetByShortUrl(ctx, alias)
	if errors.Is(err, repository.ErrURLNotFound) {
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

	err = h.service.Repository.Url.IncrementRedirectsCounter(ctx, url.ID)
	if err != nil {
		log.Error("error while incrementing requests counter",
			slog.String("alias", alias),
			sl.Err(err),
		)
	}

	log.Debug("redirected",
		slog.String("url", url.LongURL),
		slog.String("alias", alias),
	)
	ctx.Redirect(http.StatusPermanentRedirect, url.LongURL)
}

// validateUrl validates URL and return validated email and boolean is email valid.
func validateUrl(rawUrl string) (string, bool) {
	parsedUrl, err := neturl.ParseRequestURI(rawUrl)
	if err != nil {
		return "", false
	}
	return parsedUrl.String(), true
}
