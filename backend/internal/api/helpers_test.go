package api

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success with data",
			status:     http.StatusOK,
			data:       map[string]string{"message": "success"},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"success"}`,
		},
		{
			name:       "success with nil data",
			status:     http.StatusNoContent,
			data:       nil,
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
		{
			name:       "error status",
			status:     http.StatusBadRequest,
			data:       map[string]string{"error": "bad request"},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"bad request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondJSON(w, tt.status, tt.data)

			if w.Code != tt.wantStatus {
				t.Errorf("respondJSON() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.wantBody {
					t.Errorf("respondJSON() body = %v, want %v", body, tt.wantBody)
				}
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("respondJSON() Content-Type = %v, want application/json", contentType)
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		message    string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "bad request error",
			status:     http.StatusBadRequest,
			message:    "Invalid input",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"Invalid input"}`,
		},
		{
			name:       "internal server error",
			status:     http.StatusInternalServerError,
			message:    "Database error",
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"Database error"}`,
		},
		{
			name:       "not found error",
			status:     http.StatusNotFound,
			message:    "Resource not found",
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"Resource not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondError(w, tt.status, tt.message)

			if w.Code != tt.wantStatus {
				t.Errorf("respondError() status = %v, want %v", w.Code, tt.wantStatus)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.wantBody {
				t.Errorf("respondError() body = %v, want %v", body, tt.wantBody)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		status     int
		message    string
		wantStatus int
		wantBody   string
		wantLog    string
	}{
		{
			name:       "error with message",
			err:        errors.New("database connection failed"),
			status:     http.StatusInternalServerError,
			message:    "Failed to connect to database",
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"Failed to connect to database"}`,
			wantLog:    "Failed to connect to database: database connection failed",
		},
		{
			name:       "nil error with message",
			err:        nil,
			status:     http.StatusBadRequest,
			message:    "Invalid request",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"Invalid request"}`,
			wantLog:    "Invalid request",
		},
		{
			name:       "not found error",
			err:        errors.New("record not found"),
			status:     http.StatusNotFound,
			message:    "System not found",
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"System not found"}`,
			wantLog:    "System not found: record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var logBuf bytes.Buffer
			log.SetOutput(&logBuf)
			defer log.SetOutput(nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)

			handleError(w, r, tt.err, tt.status, tt.message)

			// Check HTTP response
			if w.Code != tt.wantStatus {
				t.Errorf("handleError() status = %v, want %v", w.Code, tt.wantStatus)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.wantBody {
				t.Errorf("handleError() body = %v, want %v", body, tt.wantBody)
			}

			// Check log output
			logOutput := logBuf.String()
			if !strings.Contains(logOutput, tt.wantLog) {
				t.Errorf("handleError() log = %v, want to contain %v", logOutput, tt.wantLog)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name   string
		userID interface{}
		want   int
	}{
		{
			name:   "valid user ID",
			userID: 123,
			want:   123,
		},
		{
			name:   "zero user ID",
			userID: 0,
			want:   0,
		},
		{
			name:   "no user ID in context",
			userID: nil,
			want:   0,
		},
		{
			name:   "invalid type in context",
			userID: "not an int",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/test", nil)
			if tt.userID != nil {
				ctx := context.WithValue(r.Context(), userIDKey, tt.userID)
				r = r.WithContext(ctx)
			}

			got := getUserIDFromContext(r)
			if got != tt.want {
				t.Errorf("getUserIDFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUsernameFromContext(t *testing.T) {
	tests := []struct {
		name     string
		username interface{}
		want     string
	}{
		{
			name:     "valid username",
			username: "testuser",
			want:     "testuser",
		},
		{
			name:     "empty username",
			username: "",
			want:     "",
		},
		{
			name:     "no username in context",
			username: nil,
			want:     "",
		},
		{
			name:     "invalid type in context",
			username: 123,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/test", nil)
			if tt.username != nil {
				ctx := context.WithValue(r.Context(), usernameKey, tt.username)
				r = r.WithContext(ctx)
			}

			got := getUsernameFromContext(r)
			if got != tt.want {
				t.Errorf("getUsernameFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

//
