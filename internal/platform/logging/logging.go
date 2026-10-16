// package logging builds the application's structured logger.
package logging

import (
	"io"
	"log/slog"
)

// New returns a JSON logger that writes to w at the given level
// ("debug", "info", "warn" or "error"; any else falls back to "info").
// Every record carries the service name and the environment.
func New(w io.Writer, service, level, env string) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: parseLevel(level)})
	return slog.New(handler).With("service", service, "env", env)
}

func parseLevel(s string) slog.Level {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(s)); err != nil {
		return slog.LevelInfo
	}
	return lvl
}
