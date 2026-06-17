package api

import (
	"log"
	"net/http"
	"strings"
)

// sanitizeError sanitizes error messages to prevent information disclosure
func sanitizeError(err error, userMessage string) string {
	if err == nil {
		return userMessage
	}

	// Log the actual error for debugging
	log.Printf("Error (sanitized for user): %v", err)

	// Return generic user-friendly message
	return userMessage
}

// respondErrorSafe responds with a sanitized error message
func respondErrorSafe(w http.ResponseWriter, statusCode int, err error, userMessage string) {
	sanitized := sanitizeError(err, userMessage)
	respondError(w, statusCode, sanitized)
}

// isSensitiveError checks if an error message contains sensitive information
func isSensitiveError(errMsg string) bool {
	sensitiveKeywords := []string{
		"sql",
		"database",
		"query",
		"table",
		"column",
		"constraint",
		"syntax",
		"password",
		"token",
		"secret",
		"key",
		"hash",
		"encrypt",
		"decrypt",
	}

	lowerMsg := strings.ToLower(errMsg)
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}

	return false
}

// sanitizeErrorMessage sanitizes an error message if it contains sensitive info
func sanitizeErrorMessage(errMsg string, fallback string) string {
	if isSensitiveError(errMsg) {
		log.Printf("Sanitized sensitive error: %s", errMsg)
		return fallback
	}
	return errMsg
}

//
