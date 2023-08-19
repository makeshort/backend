package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Success struct {
	Message string `json:"message"`
}

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

type URLCreated struct {
	Url   string `json:"url"`
	Alias string `json:"alias"`
}

type Token struct {
	Token string `json:"token"`
}

type Session struct {
	SessionID string `bson:"session_id"`
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