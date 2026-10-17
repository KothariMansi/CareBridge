package config

import (
	"strings"
	"testing"
)

// setEnv clear every configuration variable and then applies kv.
// viper treat an empty environment variable as unset.
func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for _, k := range []string{keyAppEnv, keyHTTPPort, keyLogLevel, keyDatabaseURL} {
		t.Setenv(k, "")
	}
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

const validDB = "postgres://root:secret@localhost:5656/carebridge?sslmode=disable"

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantErr     []string // every substring must appear in the error
		wantEnv     string
		wantPort    int
		wantLogLvl  string
		wantAddress string
	}{
		{
			name:        "everything valid",
			env:         map[string]string{keyAppEnv: "production", keyHTTPPort: "9000", keyLogLevel: "warn", keyDatabaseURL: validDB},
			wantEnv:     "production",
			wantPort:    9000,
			wantLogLvl:  "warn",
			wantAddress: ":9000",
		},
		{
			name:        "defaults apply when only the database url is set",
			env:         map[string]string{keyDatabaseURL: validDB},
			wantEnv:     "development",
			wantPort:    8080,
			wantLogLvl:  "info",
			wantAddress: ":8080",
		},
		{
			name:        "values are normalized",
			env:         map[string]string{keyAppEnv: " Staging ", keyLogLevel: "DEBUG", keyDatabaseURL: validDB},
			wantEnv:     "staging",
			wantPort:    8080,
			wantLogLvl:  "debug",
			wantAddress: ":8080",
		},
		{
			name:    "missing database url",
			env:     map[string]string{},
			wantErr: []string{"DATABASE_URL is required"},
		},
		{
			name:    "port is not a number",
			env:     map[string]string{keyDatabaseURL: validDB, keyHTTPPort: "nat-a-number"},
			wantErr: []string{"HTTP_PORT"},
		},
		{
			name:    "port is too large",
			env:     map[string]string{keyDatabaseURL: validDB, keyHTTPPort: "70000"},
			wantErr: []string{"HTTP_PORT"},
		},
		{
			name:    "port zero",
			env:     map[string]string{keyHTTPPort: "0", keyDatabaseURL: validDB},
			wantErr: []string{"HTTP_PORT"},
		},
		{
			name:    "invalid log level",
			env:     map[string]string{keyLogLevel: "verbose", keyDatabaseURL: validDB},
			wantErr: []string{"LOG_LEVEL"},
		},
		{
			name:    "invalid environment",
			env:     map[string]string{keyAppEnv: "prod", keyDatabaseURL: validDB},
			wantErr: []string{"APP_ENV"},
		},
		{
			name:    "datanase url with wrong scheme",
			env:     map[string]string{keyDatabaseURL: "mysql://u:p@localhost/db"},
			wantErr: []string{"DATABASE_URL must start with"},
		},
		{
			name:    "database URL without host",
			env:     map[string]string{keyDatabaseURL: "postgres:///carebridge"},
			wantErr: []string{"DATABASE_URL must include a host"},
		},
		{
			name: "all problems are reported together",
			env: map[string]string{
				keyAppEnv: "prod", keyHTTPPort: "abc", keyLogLevel: "verbose",
			},
			wantErr: []string{"APP_ENV", "HTTP_PORT", "LOG_LEVEL", "DATABASE_URL is required"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, tc.env)
			cfg, err := Load("")

			if len(tc.wantErr) > 0 {
				if err == nil {
					t.Fatalf("Load() = nil, want error containing %v", tc.wantErr)
				}
				for _, want := range tc.wantErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q does not contain %q", err.Error(), want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("load() unexpected error: %v", err)
			}
			if cfg.AppEnv != tc.wantEnv {
				t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, tc.wantEnv)
			}
			if cfg.HTTPPort != tc.wantPort {
				t.Errorf("HTTPPort = %d, want %d", cfg.HTTPPort, tc.wantPort)
			}
			if cfg.LogLevel != tc.wantLogLvl {
				t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, tc.wantLogLvl)
			}
			if cfg.Addr() != tc.wantAddress {
				t.Errorf("Addr() = %q, want %q", cfg.Addr(), tc.wantAddress)
			}
		})
	}
}
