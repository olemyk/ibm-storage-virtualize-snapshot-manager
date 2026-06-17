package api

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	mu          sync.RWMutex
	attempts    map[string][]time.Time
	maxAttempts int
	window      time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts:    make(map[string][]time.Time),
		maxAttempts: maxAttempts,
		window:      window,
	}

	// Start cleanup goroutine to prevent memory leaks
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get attempts for this IP
	attempts := rl.attempts[ip]

	// Filter out old attempts
	var recentAttempts []time.Time
	for _, t := range attempts {
		if t.After(windowStart) {
			recentAttempts = append(recentAttempts, t)
		}
	}

	// Check if limit exceeded
	if len(recentAttempts) >= rl.maxAttempts {
		return false
	}

	// Add current attempt
	recentAttempts = append(recentAttempts, now)
	rl.attempts[ip] = recentAttempts

	return true
}

// cleanup periodically removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for ip, attempts := range rl.attempts {
			var recentAttempts []time.Time
			for _, t := range attempts {
				if t.After(windowStart) {
					recentAttempts = append(recentAttempts, t)
				}
			}

			if len(recentAttempts) == 0 {
				delete(rl.attempts, ip)
			} else {
				rl.attempts[ip] = recentAttempts
			}
		}
		rl.mu.Unlock()
	}
}

// rateLimitMiddleware creates a middleware that rate limits requests
func (s *Server) rateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP from X-Forwarded-For header (for proxy) or RemoteAddr
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.Header.Get("X-Real-IP")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}

			// Remove port if present
			if idx := len(ip) - 1; idx >= 0 {
				for i := idx; i >= 0; i-- {
					if ip[i] == ':' {
						ip = ip[:i]
						break
					}
				}
			}

			if !limiter.Allow(ip) {
				respondError(w, http.StatusTooManyRequests, "Too many requests. Please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

//
