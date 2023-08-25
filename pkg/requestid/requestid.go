package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const headerXRequestID = "X-Request-ID"

// New initializes the RequestID middleware.
func New(ctx *gin.Context) {
	rid := ctx.GetHeader(headerXRequestID)
	if rid == "" {
		rid = uuid.NewString()
		ctx.Request.Header.Add(headerXRequestID, rid)
	}

	ctx.Header(headerXRequestID, rid)
	ctx.Next()
}

// Get returns the request identifier.
func Get(c *gin.Context) string {
	return c.Writer.Header().Get(headerXRequestID)
}
