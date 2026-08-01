package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter tracks requests per IP address with a sliding window
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	maxReq   int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
// maxReq: maximum requests allowed
// window: time window for rate limiting
func NewRateLimiter(maxReq int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		maxReq:   maxReq,
		window:   window,
	}
}

// Allow returns true if the request from clientIP should be allowed
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get request history for this IP
	history := rl.requests[clientIP]

	// Remove old requests outside the window
	var validReqs []time.Time
	for _, t := range history {
		if t.After(windowStart) {
			validReqs = append(validReqs, t)
		}
	}

	// Check if limit exceeded
	if len(validReqs) >= rl.maxReq {
		rl.requests[clientIP] = validReqs
		return false
	}

	// Add current request
	validReqs = append(validReqs, now)
	rl.requests[clientIP] = validReqs

	return true
}

// RateLimitMiddleware returns a middleware that applies rate limiting to a handler
func RateLimitMiddleware(limiter *RateLimiter, maxRetries int, windowDuration time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !limiter.Allow(clientIP) {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRetries))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(windowDuration.Seconds())))
				http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client's IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		if idx := len(xff) - 1; idx >= 0 {
			for i := 0; i < len(xff); i++ {
				if xff[i] == ',' {
					return xff[:i]
				}
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	return r.RemoteAddr
}
