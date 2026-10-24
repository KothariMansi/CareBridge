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

	"github.com/gin-gonic/gin"

	"github.com/KothariMansi/carebridge/internal/config"
	"github.com/KothariMansi/carebridge/internal/platform/database"
	"github.com/KothariMansi/carebridge/internal/platform/logging"
)

const (
	serviceName     = "carebridge-api"
	shutdownTimeout = 15 * time.Second
)

func main() {
	if err := run(); err != nil {
		// The logger may not exist yet (for example when the config is
		// invalid), so write to stderr directly.
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(".env")
	if err != nil {
		return err
	}

	logger := logging.New(os.Stdout, serviceName, cfg.LogLevel, cfg.AppEnv)
	slog.SetDefault(logger)
	logger.Info("configuration loaded", "config", cfg) // the password is redacted

	// ctx is cancelled on Ctrl+C (SIGINT) or SIGTERM (what Docker and
	// Kubernetes send when stopping a container). It is created before the
	// database connection so Ctrl+C also interrupts a slow startup.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	// Deferred functions run in reverse order: the HTTP server is shut down
	// first (below), and only then is the pool closed.
	defer pool.Close()
	logger.Info("database connected")

	gin.SetMode(gin.ReleaseMode)
	router, err := newRouter(logger, pool)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.Addr(),
		Handler: router,

		// Without timeouts, a slow or malicious client can hold connections
		// open forever.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,

		// Send net/http's own error messages through our JSON logger.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", srv.Addr)
		// ListenAndServe always returns a non-nil error. ErrServerClosed
		// just means Shutdown was called, which is the normal way to stop.
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		stop() // restore default behaviour: a second Ctrl+C now kills the process
	}

	logger.Info("shutdown signal received, draining in-flight requests")

	// A fresh context: ctx is already cancelled and would abort Shutdown
	// immediately.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("server stopped cleanly")
	return nil
}
