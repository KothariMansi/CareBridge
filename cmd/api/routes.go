package main

import (
	"fmt"
	"log/slog"

	"github.com/KothariMansi/carebridge/internal/platform/httpx"
	"github.com/gin-gonic/gin"
)

// newRouter builds the Gin Engine and register every route and middleware.
func newRouter(logger *slog.Logger) (*gin.Engine, error) {
	r := gin.New()

	// Trust no proxy header, so the client IP is the real TCP peer. This also removes Gin's "you trusted all proxies" warning.
	// When the API later runs behind a reverse proxy, list that proxy's address here.
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	r.HandleMethodNotAllowed = true // wrong method will give 405 (with an Allow header), not 404
	r.RedirectTrailingSlash = false // an API should answer, not redirect
	r.NoRoute(httpx.NotFound)
	r.NoMethod(httpx.MethodNotAllowed)

	// Order matters. RequestID goes first so everything after it can read the
	// ID. Recovery goes last so a panic is turned into a 500 *inside*
	// AccessLog, which then logs the real status.
	r.Use(httpx.RequestID(), httpx.AccessLog(logger), httpx.Recovery(logger))

	r.GET("/healthz", httpx.Healthz)

	return r, nil

}
