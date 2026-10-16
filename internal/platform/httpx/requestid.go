package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RrequestIDHeader is the header used to receive and echo the request ID.
const RequestIDHeader = "X-Request-ID"

const RequestIDkey = "request_id" // key in the gin context for the request ID

type ctxKey struct{}

// RequestID makes sure every request has and ID. It reused the client's X-Request-ID when the value is safe,
// otherwise it generates one. The ID is echoed in the response header and stored in both gin context and
// the request's context.Context, so services can read it without importing gin. The ID is also logged by the logger middleware.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if !validRequestID(id) {
			id = newRequestID()
		}
		c.Set(RequestIDkey, id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKey{}, id))
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// GetRequestID return the request ID stored by the RequestID middleware.
func GetRequestID(c *gin.Context) string {
	return c.GetString(RequestIDkey)
}

// RequestIDFromContext returns the request ID stored in ctx, or "".
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// validRequestID accepts 1 to 64 characters from [A-Za-z0-9._-].
// Anything else is rejected to avoid log injection and other attacks.
// The client can use any ID it wants, but the server will generate a new one if the value is unsafe.
func validRequestID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for i := 0; i < len(id); i++ {
		ch := id[i]
		ok := (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '.' || ch == '_' || ch == '-'
		if !ok {
			return false
		}
	}
	return true
}

func newRequestID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return "req_" + hex.EncodeToString(b[:])
}
