package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Healthz in the liveness probe: it reports that the process is running and able to server requests.
// it must not check the database or any other dependency. Those checks belongs to the readiness probe.
func Healthz(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
