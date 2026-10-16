package httpx

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery turns a panic in a handler into a logged error and a generic 500.
// The panic value and stack trace go to the log only, never to the client.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			// net/http uses this value to abort a response on purpose
			if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(err)
			}

			logger.ErrorContext(c.Request.Context(), "panic recovered",
				slog.String("request_id", GetRequestID(c)),
				slog.String("method", c.Request.Method),
				slog.String("route", routeOf(c)),
				slog.String("panic", fmt.Sprint(rec)),
				slog.String("stack", string(debug.Stack())),
			)

			if c.Writer.Written() {
				// The status line is already sent, so only stop the chain.
				c.Abort()
				return
			}
			AbortWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR",
				"An unexpected error occurred.")
		}()
		c.Next()
	}
}
