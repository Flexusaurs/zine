package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test middleware helpers
func newTestRequest(method, path string) *http.Request {
	req, _ := http.NewRequest(method, path, nil)
	return req
}

func newTestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

func runMiddlewareTest(t *testing.T, middleware http.HandlerFunc, method, path string, expectedStatus int) {
	t.Helper()

	req := newTestRequest(method, path)
	rec := newTestRecorder()

	middleware(rec, req)

	if rec.Code != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, rec.Code)
	}
}
