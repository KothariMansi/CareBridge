package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KothariMansi/carebridge/internal/platform/logging"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error {
	return f.err
}

func newTestRouter(t *testing.T) (*gin.Engine, *bytes.Buffer) {
	t.Helper()
	return newTestRouterWith(t, fakePinger{})
}

// newTestRouter returns a router whose log output is captured in a buffer.
func newTestRouterWith(t *testing.T, db fakePinger) (*gin.Engine, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	logger := logging.New(&buf, "carebridge-api", "debug", "test")
	r, err := newRouter(logger, db)
	if err != nil {
		t.Fatalf("new Router: %v", err)
	}
	return r, &buf
}

// do makes a request to the given handler.
func do(r http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

type errorEnvelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) errorEnvelope {
	t.Helper()
	var env errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("response is not the standard error JSON: %v (body: %s)", err, rec.Body.String())
	}
	return env
}

func TestHealthz(t *testing.T) {
	r, _ := newTestRouter(t)
	rec := do(r, http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want prefix %q", ct, "application/json")
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
	if id := rec.Header().Get("X-Request-ID"); !strings.HasPrefix(id, "req_") {
		t.Errorf("X-Request-ID = %q, want a generated req_ ID", id)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf(`body["status"] = %q, want ok`, body["status"])
	}
}

func TestReadyz(t *testing.T) {
	r, _ := newTestRouter(t)
	rec := do(r, http.MethodGet, "/readyz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if body["status"] != "ready" {
		t.Errorf(`body["status"] = %q, want "ready"`, body["status"])
	}
}

func TestReadyDatabaseDown(t *testing.T) {
	r, logs := newTestRouterWith(t, fakePinger{err: errors.New("dial tcp 10.0.0.5:5432: connection refused")})
	rec := do(r, http.MethodGet, "/readyz", nil)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if env := decodeError(t, rec); env.Error.Code != "NOT_READY" {
		t.Errorf("code = %q, want NOT_READY", env.Error.Code)
	}
	if strings.Contains(rec.Body.String(), "10.0.0.5") {
		t.Errorf("response leaks the database error: %s", rec.Body.String())
	}
	if !strings.Contains(logs.String(), "readiness check failed") {
		t.Errorf("the failure should be logged server-side, logs: %s", logs.String())
	}
}

func TestLivenessIgnoresDatabase(t *testing.T) {
	r, _ := newTestRouterWith(t, fakePinger{err: errors.New("db down")})
	if rec := do(r, http.MethodGet, "/healthz", nil); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 even when the database is down", rec.Code)
	}
}

func TestRequestId(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
		wantEcho bool
	}{
		{name: "safe client ID is reused", incoming: "client-id_123.abc", wantEcho: true},
		{name: "Id with spaces is replaced", incoming: "id with space", wantEcho: false},
		{name: "Id with a newline is replaced", incoming: "newline\n", wantEcho: false},
		{name: "overlong Id is replaced", incoming: strings.Repeat("a", 65), wantEcho: false},
		{name: "missing Id is generated", incoming: "", wantEcho: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := newTestRouter(t)
			headers := map[string]string{}
			if tc.incoming != "" {
				headers["X-Request-ID"] = tc.incoming
			}

			rec := do(r, http.MethodGet, "/healthz", headers)
			got := rec.Header().Get("X-Request-ID")
			if tc.wantEcho && got != tc.incoming {
				t.Errorf("X-Request-ID = %q, want the client's %q", got, tc.incoming)
			}
			if !tc.wantEcho && !strings.HasPrefix(got, "req_") {
				t.Errorf("X-Request-ID = %q, want a generated req_ ID", got)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	r, _ := newTestRouter(t)
	rec := do(r, http.MethodPost, "/healthz", nil)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); !strings.Contains(allow, "GET") {
		t.Errorf("Allow = %q, want it to list GET", allow)
	}
	env := decodeError(t, rec)
	if env.Error.Code != "METHOD_NOT_ALLOWED" {
		t.Errorf("code = %q, want METHOD_NOT_ALLOWED", env.Error.Code)
	}
	if env.Error.RequestID == "" || env.Error.RequestID != rec.Header().Get("X-Request-ID") {
		t.Errorf("body request_id %q must match the X-Request-ID header %q", env.Error.RequestID, rec.Header().Get("X-Request-ID"))
	}
}

func TestNotFound(t *testing.T) {
	r, _ := newTestRouter(t)
	rec := do(r, http.MethodGet, "/not-found", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if env := decodeError(t, rec); env.Error.Code != "NOT_FOUND" {
		t.Errorf("code = %q, want NOT_FOUND", env.Error.Code)
	}
}

func TestPanicIsRecoveredWithoutLeakingDetails(t *testing.T) {
	r, logs := newTestRouter(t)
	r.GET("/just-a-method", func(c *gin.Context) {
		panic("internal detail: db password is justapassword")
	})
	rec := do(r, http.MethodGet, "/just-a-method", nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "justapassword") {
		t.Errorf("response leaks the panic value: %s", rec.Body.String())
	}
	if env := decodeError(t, rec); env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("code = %q, want INTERNAL_ERROR", env.Error.Code)
	}
	if !strings.Contains(logs.String(), "panic recovered") {
		t.Errorf("the panic should be logged server-side, logs %s", logs.String())
	}
}

func TestAccessLog(t *testing.T) {
	r, logs := newTestRouter(t)

	rec := do(r, http.MethodGet, "/healthz", map[string]string{"X-Request-ID": "trace-42"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var entry map[string]any
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("log line is not JSON: %q", line)
		}
		if m["msg"] == "http request" {
			entry = m
		}
	}
	if entry == nil {
		t.Fatalf("no access log line found in: %s", logs.String())
	}

	want := map[string]any{
		"request_id": "trace-42",
		"method":     "GET",
		"route":      "/healthz",
		"status":     float64(200), // JSON numbers decode as float64
		"service":    "carebridge-api",
		"level":      "DEBUG", // probes are logged at debug level
	}
	for k, v := range want {
		if entry[k] != v {
			t.Errorf("log field %q = %v, want %v", k, entry[k], v)
		}
	}
}
