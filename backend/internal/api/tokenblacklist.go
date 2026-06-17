package api

import (
	"sync"
	"time"
)

// TokenBlacklist manages revoked JWT tokens
type TokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token -> expiry time
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	// Start cleanup goroutine to prevent memory leaks
	go bl.cleanup()

	return bl
}

// Add adds a token to the blacklist with its expiry time
func (bl *TokenBlacklist) Add(token string, expiryTime time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	bl.tokens[token] = expiryTime
}

// IsBlacklisted checks if a token is blacklisted
func (bl *TokenBlacklist) IsBlacklisted(token string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	expiry, exists := bl.tokens[token]
	if !exists {
		return false
	}

	// Token is blacklisted if it hasn't expired yet
	return time.Now().Before(expiry)
}

// cleanup periodically removes expired tokens from blacklist
func (bl *TokenBlacklist) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		bl.mu.Lock()
		now := time.Now()

		for token, expiry := range bl.tokens {
			if now.After(expiry) {
				delete(bl.tokens, token)
			}
		}
		bl.mu.Unlock()
	}
}

//
