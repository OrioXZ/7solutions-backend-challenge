package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddlewareRecordsRequestMetadata(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry["msg"] != "HTTP request" {
		t.Fatalf("msg = %#v, want %q", entry["msg"], "HTTP request")
	}
	if entry["method"] != http.MethodPost {
		t.Fatalf("method = %#v, want %q", entry["method"], http.MethodPost)
	}
	if entry["path"] != "/api/v1/users" {
		t.Fatalf("path = %#v, want %q", entry["path"], "/api/v1/users")
	}
	if entry["status"] != float64(http.StatusCreated) {
		t.Fatalf("status = %#v, want %d", entry["status"], http.StatusCreated)
	}
	if _, exists := entry["duration_ms"]; !exists {
		t.Fatal("duration_ms was not logged")
	}
}

func TestLoggingMiddlewareCapturesImplicitOKStatus(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry["status"] != float64(http.StatusOK) {
		t.Fatalf("status = %#v, want %d", entry["status"], http.StatusOK)
	}
}
