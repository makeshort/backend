package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Error struct {
	Message string `json:"message"`
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type URL struct {
	ID        string `json:"id"`
	Url       string `json:"url"`
	Alias     string `json:"alias"`
	Redirects int    `json:"redirects"`
}

type UrlCreated struct {
	ID    string `json:"id"`
	Url   string `json:"url"`
	Alias string `json:"alias"`
}

type UrlUpdated struct {
	ID    string `json:"id"`
	Url   string `json:"url,omitempty"`
	Alias string `json:"alias,omitempty"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// SendInvalidRequestBodyError sends an error response with 400 Bad Request status code.
func SendInvalidRequestBodyError(ctx *gin.Context) {
	SendError(ctx, http.StatusBadRequest, "invalid request body")
}

// SendAuthFailedError sends an error response with 401 Unauthorized status code.
func SendAuthFailedError(ctx *gin.Context) {
	SendError(ctx, http.StatusUnauthorized, "auth failed")
}

// SendError sends an error response with some status code and message field.
func SendError(ctx *gin.Context, statusCode int, message string) {
	ctx.AbortWithStatusJSON(statusCode, Error{Message: message})
}
