// Package config loads and validates the application configuration.
//
// Values come from three layers, highest priority first:
//
//  1. process environment variables (what production uses)
//  2. an optional env file such as ".env" (local development convenience)
//  3. built-in defaults, only where a default is safe
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Configuration keys. Each key is also the name of its environment variable.
const (
	keyAppEnv      = "APP_ENV"
	keyHTTPPort    = "HTTP_PORT"
	keyLogLevel    = "LOG_LEVEL"
	keyDatabaseURL = "DATABASE_URL"
)

// Allowed values for APP_ENV.
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

var (
	validEnvs      = []string{EnvDevelopment, EnvStaging, EnvProduction}
	validLogLevels = []string{"debug", "info", "warn", "error"}
)

// Config is the validated application configuration.
type Config struct {
	AppEnv   string
	HTTPPort int
	LogLevel string

	// DatabaseURL contains credentials. Never log it directly: use
	// RedactedDatabaseURL, or log the whole Config (it implements slog.LogValuer).
	DatabaseURL string
}

// Load reads and validates the configuration.
//
// envFile is optional. When it is empty, or the file does not exist, only
// environment variables and defaults are used, which is the normal situation
// in production. A file that exists but cannot be parsed is an error.
//
// Load reports every problem it finds at once instead of stopping at the first.
func Load(envFile string) (Config, error) {
	v := viper.New()

	// Defaults exist only where a wrong guess is harmless.
	// DATABASE_URL deliberately has none.
	v.SetDefault(keyAppEnv, EnvDevelopment)
	v.SetDefault(keyHTTPPort, "8080")
	v.SetDefault(keyLogLevel, "info")

	v.AutomaticEnv()

	if envFile != "" {
		v.SetConfigFile(envFile)
		v.SetConfigType("env")
		if err := v.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return Config{}, fmt.Errorf("read env file %q: %w", envFile, err)
		}
	}

	cfg := Config{
		AppEnv:      strings.ToLower(strings.TrimSpace(v.GetString(keyAppEnv))),
		LogLevel:    strings.ToLower(strings.TrimSpace(v.GetString(keyLogLevel))),
		DatabaseURL: strings.TrimSpace(v.GetString(keyDatabaseURL)),
	}

	var problems []error

	if !slices.Contains(validEnvs, cfg.AppEnv) {
		problems = append(problems, fmt.Errorf("%s must be one of %s (got %q)",
			keyAppEnv, strings.Join(validEnvs, ", "), cfg.AppEnv))
	}

	port, err := parsePort(v.GetString(keyHTTPPort))
	if err != nil {
		problems = append(problems, fmt.Errorf("%s: %w", keyHTTPPort, err))
	}
	cfg.HTTPPort = port

	if !slices.Contains(validLogLevels, cfg.LogLevel) {
		problems = append(problems, fmt.Errorf("%s must be one of %s (got %q)",
			keyLogLevel, strings.Join(validLogLevels, ", "), cfg.LogLevel))
	}

	if err := validateDatabaseURL(cfg.DatabaseURL); err != nil {
		problems = append(problems, err)
	}

	if len(problems) > 0 {
		return Config{}, fmt.Errorf("invalid configuration:\n%w", errors.Join(problems...))
	}
	return cfg, nil
}

func parsePort(raw string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("must be an integer between 1 and 65535 (got %q)", raw)
	}
	return port, nil
}

func validateDatabaseURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("%x is required", keyDatabaseURL)
	}

	u, err := url.Parse(raw)
	if err != nil {
		// Do not wrap err. *url.Error emends the full URL, password included, to the error message, which is a security risk.
		return fmt.Errorf("%s is not a valid URL", keyDatabaseURL)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf(`%s must start with "postgres://" or "postgresql://"`, keyDatabaseURL)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("%x must include a host", keyDatabaseURL)
	}

	return nil
}

// Addr returns the listen address for http.Server, for example ":8080".
func (c Config) Addr() string {
	return ":" + strconv.Itoa(c.HTTPPort)
}

// RedactedDatabaseURL returns the database URL with the password masked out, suitable for logging.
func (c Config) RedactedDatabaseURL() string {
	u, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return "[invalid]"
	}
	return u.Redacted()
}

// LogValue implements slog.LogValuer, so logging the whole Config never leaks secrets.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("app_env", c.AppEnv),
		slog.Int("http_port", c.HTTPPort),
		slog.String("log_level", c.LogLevel),
		slog.String("database_url", c.RedactedDatabaseURL()),
	)
}

// String implements fmt.Stringer, so fmt.Printf(%v, cfg) is safe as well.
func (c Config) String() string {
	return fmt.Sprintf("Config{app_env=%s http_port=%d log_level=%s database_url=%s}",
		c.AppEnv, c.HTTPPort, c.LogLevel, c.RedactedDatabaseURL())
}
