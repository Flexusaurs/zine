package ratelimit

import (
	"net/http"

	"github.com/go-redis/redis/v8"
)

var (
	rateLimiter *redis.Client
)

// RateLimiter middleware for rate limiting
func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement rate limiting logic
		next.ServeHTTP(w, r)
	})
}

// InitRateLimiter initializes the rate limiter
func InitRateLimiter(client *redis.Client) {
	rateLimiter = client
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	Rate  float64 // requests per second
	Burst int     // max burst size
}

// Default configs
var (
	DefaultConfig = RateLimitConfig{Rate: 10, Burst: 100}
	LoginConfig   = RateLimitConfig{Rate: 0.05, Burst: 5} // 3/min
	PostConfig    = RateLimitConfig{Rate: 0.5, Burst: 10}  // 30/min
)

// RateLimitExceeded error
var RateLimitExceeded = &APIError{
	Code:    "rate_limit_exceeded",
	Message: "rate limit exceeded",
	Status:  http.StatusTooManyRequests,
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}
