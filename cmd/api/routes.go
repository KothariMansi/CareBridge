package main

import (
	"fmt"

	"github.com/KothariMansi/carebridge/internal/platform/httpx"
	"github.com/gin-gonic/gin"
)

// newRouter builds the Gin Engine and register every route and middleware.
func newRouter() (*gin.Engine, error) {
	r := gin.Default()

	// Trust no proxy header, so the client IP is the real TCP peer. This also removes Gin's "you trusted all proxies" warning.
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	r.HandleMethodNotAllowed = true // wrong method will give 405 (with an Allow header), not 404
	r.RedirectTrailingSlash = false // an API should answer, not redirect
	r.NoRoute(httpx.NotFound)
	r.NoMethod(httpx.MethodNotAllowed)

	r.GET("/healthz", httpx.Healthz)

	return r, nil

}
