package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// readyTimeout bounds one readiness check, so hung database cannot make the probe hang with it.
const readyTimeout = 5 * time.Second

// Pinger is the interface that wraps the Ping method, which checks if the database is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Readyz is the readiness probe: it reports whether the service can do useful work right now,
// which means that the database is reachable. Unlike /healthz, a failure here is expected during
// a database outage and tells the load balancer to stop sending traffic without restarting the process.
func Readyz(logger *slog.Logger, db Pinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		ctx, cancel := context.WithTimeout(c.Request.Context(), readyTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			// The reason goes to logs only, never to the client.
			logger.ErrorContext(
				c.Request.Context(),
				"readiness check failed",
				slog.String("request_id", GetRequestID(c)),
				slog.Any("error", err),
			)
			AbortWithError(c, http.StatusServiceUnavailable, "NOT_READY", "The service is not ready to accept requests")
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
