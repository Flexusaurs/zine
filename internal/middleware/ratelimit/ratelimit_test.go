package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter_AllowsRequests(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	RateLimiter(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRateLimitExceededError(t *testing.T) {
	err := RateLimitExceeded

	if err.Code != "rate_limit_exceeded" {
		t.Errorf("expected code 'rate_limit_exceeded', got '%s'", err.Code)
	}
	if err.Status != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, err.Status)
	}
}

func TestRateLimitConfigs(t *testing.T) {
	// Verify default config
	if DefaultConfig.Rate != 10 {
		t.Errorf("expected default rate 10, got %f", DefaultConfig.Rate)
	}
	if DefaultConfig.Burst != 100 {
		t.Errorf("expected default burst 100, got %d", DefaultConfig.Burst)
	}

	// Verify login config (5 req per minute = 0.083/s, we use 0.05 for safety)
	if LoginConfig.Rate > 0.1 {
		t.Errorf("login rate too high: %f", LoginConfig.Rate)
	}

	// Verify post config (30 req per minute = 0.5/s)
	if PostConfig.Rate != 0.5 {
		t.Errorf("expected post rate 0.5, got %f", PostConfig.Rate)
	}
	if PostConfig.Burst != 10 {
		t.Errorf("expected post burst 10, got %d", PostConfig.Burst)
	}
}
