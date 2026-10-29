package appserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthCheckRoute(t *testing.T) {
	// Create a new router
	router := http.NewServeMux()
	router = addHealthCheckRoute(router)

	// Create a test request for the health check endpoint
	req, err := http.NewRequest("GET", "/-/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a recorder to capture the response
	rec := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rec, req)

	// Check the status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Check the response body
	if body := rec.Body.String(); body != "OK" {
		t.Errorf("Expected body to be 'OK', got %q", body)
	}
}

func TestHealthCheckRouteWriteError(t *testing.T) {
	// Create a custom response writer that fails on Write
	failWriter := &failingResponseWriter{}

	// Create a new router with health check
	router := http.NewServeMux()
	router = addHealthCheckRoute(router)

	// Create a test request for the health check endpoint
	req, err := http.NewRequest("GET", "/-/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Serve the request with the failing writer
	router.ServeHTTP(failWriter, req)

	// Check that the error handler was invoked
	if failWriter.statusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, failWriter.statusCode)
	}
	if !strings.Contains(failWriter.errorMsg, "health check failed") {
		t.Errorf("Expected error message to contain 'health check failed', got %q", failWriter.errorMsg)
	}
}

// Custom writer that fails on Write
type failingResponseWriter struct {
	statusCode int
	errorMsg   string
}

func (f *failingResponseWriter) Header() http.Header {
	return http.Header{}
}

func (f *failingResponseWriter) Write([]byte) (int, error) {
	return 0, http.ErrBodyNotAllowed
}

func (f *failingResponseWriter) WriteHeader(statusCode int) {
	f.statusCode = statusCode
}

// To capture error responses
func (f *failingResponseWriter) Error(msg string, code int) {
	f.statusCode = code
	f.errorMsg = msg
}
