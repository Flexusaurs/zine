package ratelimit

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// LimitByIP returns a middleware that rate limits by IP address
func LimitByIP(next http.Handler) http.Handler {
	// Default: 100 burst, 10/s
	return httprate.Limit(100, 10*time.Second)(next)
}
