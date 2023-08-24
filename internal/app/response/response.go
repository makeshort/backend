package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Error struct {
	Message string `json:"message"`
}

type User struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type URL struct {
	Url       string `json:"url"`
	Alias     string `json:"alias"`
	Redirects int    `json:"redirects"`
}

type UrlCreated struct {
	Url   string `json:"url"`
	Alias string `json:"alias"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// InvalidRequestBody sends an error response with 400 Bad Request status code.
func InvalidRequestBody(ctx *gin.Context) {
	SendError(ctx, http.StatusBadRequest, "invalid request body")
}

// InvalidAuthToken sends an error response with 401 Unauthorized status code.
func InvalidAuthToken(ctx *gin.Context) {
	SendError(ctx, http.StatusUnauthorized, "invalid auth token")
}

// SendError sends an error response with message field.
func SendError(ctx *gin.Context, statusCode int, message string) {
	ctx.AbortWithStatusJSON(statusCode, Error{Message: message})
}
