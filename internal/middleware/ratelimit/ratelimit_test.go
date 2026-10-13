package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLimitByIP_AllowsRequests(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	LimitByIP(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
