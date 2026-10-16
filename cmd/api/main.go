package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KothariMansi/carebridge/internal/config"
	"github.com/KothariMansi/carebridge/internal/platform/logging"
	"github.com/gin-gonic/gin"
)

const (
	serviceName     = "carebridge-api"
	shutdownTimeout = 15 * time.Second
)

func main() {
	if err := run(); err != nil {
		// The logger may not exist yet, so use the standard library to Log errors directly.
		fmt.Fprintln(os.Stderr, "fatal: ", err)
		os.Exit(1)
	}

}

func run() error {
	// ".env" is read from the working directory, so start the server from the
	// repository root. In production there is no .env file and the real
	// environment variables are used.
	cfg, err := config.Load(".env")
	if err != nil {
		return err
	}

	logger := logging.New(os.Stdout, serviceName, cfg.LogLevel, cfg.AppEnv)
	slog.SetDefault(logger)
	logger.Info("configuration loaded", "config", cfg) // password is masked

	// Release mode silences Gin's debug banner and route table.
	gin.SetMode(gin.ReleaseMode)

	router, err := newRouter(logger)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    cfg.Addr(),
		Handler: router,

		// Without timeouts a slow or malicious client can hold a connection open forever.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// ctx is cancelled on Ctrl+C (SIGINT) or SIGTERM, which is what docker and Kubernetes send when stopping a container.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", server.Addr)
		// ListenAndServe always returns a non-nil error. ErrServerClosed only means Shutdown was called, which is the normal way to stop.
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		stop() // restore default behaviour: a second Ctrl+C kills the process
	}

	logger.Info("shutdown signal received, finishing in-flight requests")

	// A fresh context: ctx is already cancelled and would abort Shutdown at once.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("server stopped cleanly")
	return nil
}
