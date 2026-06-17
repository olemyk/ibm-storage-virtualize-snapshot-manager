package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

// CSRFTokenManager manages CSRF tokens
type CSRFTokenManager struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token -> expiry time
}

// NewCSRFTokenManager creates a new CSRF token manager
func NewCSRFTokenManager() *CSRFTokenManager {
	manager := &CSRFTokenManager{
		tokens: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go manager.cleanup()

	return manager
}

// GenerateToken generates a new CSRF token
func (m *CSRFTokenManager) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Token valid for 24 hours
	m.tokens[token] = time.Now().Add(24 * time.Hour)

	return token, nil
}

// ValidateToken validates a CSRF token
func (m *CSRFTokenManager) ValidateToken(token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	expiry, exists := m.tokens[token]
	if !exists {
		return false
	}

	return time.Now().Before(expiry)
}

// InvalidateToken removes a token from the store
func (m *CSRFTokenManager) InvalidateToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tokens, token)
}

// cleanup periodically removes expired tokens
func (m *CSRFTokenManager) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()

		for token, expiry := range m.tokens {
			if now.After(expiry) {
				delete(m.tokens, token)
			}
		}
		m.mu.Unlock()
	}
}

// csrfMiddleware creates a middleware that validates CSRF tokens
func (s *Server) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for GET, HEAD, OPTIONS requests (safe methods)
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF check for login endpoint (no token available yet)
		if r.URL.Path == "/api/auth/login" {
			next.ServeHTTP(w, r)
			return
		}

		// Get CSRF token from header
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			respondError(w, http.StatusForbidden, "CSRF token missing")
			return
		}

		// Validate token
		if !s.csrfManager.ValidateToken(token) {
			respondError(w, http.StatusForbidden, "Invalid or expired CSRF token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

//
