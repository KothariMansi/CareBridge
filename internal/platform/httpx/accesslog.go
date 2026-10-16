package httpx

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLog writes one structured log line per request, after it finishes.
// It logs the route pattern ("/docters/:id"), not the raw path, to keep log
// cardinality low and to keep identifier out of the log. It never logs
// header or bodies. Health probes are logged at debug level so they do not flood the log.
func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		route := routeOf(c)

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		case isProbe(route):
			level = slog.LevelDebug
		}

		logger.LogAttrs(c.Request.Context(), level, "http request",
			slog.String("request_id", GetRequestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("route", route),
			slog.Int("status", status),
			slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			slog.Int("bytes", max(c.Writer.Size(), 0)),
			slog.String("client_ip", c.ClientIP()),
		)

	}
}

// routeOf returns the route pattern ("/docters/:id") for the request, or "unmatched" if no route matched (404).
func routeOf(c *gin.Context) string {
	if p := c.FullPath(); p != "" {
		return p
	}
	return "unmatched"
}

// isProbe returns true if the route is a health or readiness probe, which are logged at debug level.
func isProbe(route string) bool {
	return route == "/healthz" || route == "/readyz"
}
