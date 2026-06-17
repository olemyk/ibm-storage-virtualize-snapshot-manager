package config

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestLoadConfig_JWTSecretValidation(t *testing.T) {
	tests := []struct {
		name      string
		jwtSecret string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid JWT secret (32 chars)",
			jwtSecret: "12345678901234567890123456789012",
			wantErr:   false,
		},
		{
			name:      "valid JWT secret (64 chars)",
			jwtSecret: "1234567890123456789012345678901234567890123456789012345678901234",
			wantErr:   false,
		},
		{
			name:      "empty JWT secret",
			jwtSecret: "",
			wantErr:   true,
			errMsg:    "JWT_SECRET is required",
		},
		{
			name:      "JWT secret too short (31 chars)",
			jwtSecret: "1234567890123456789012345678901",
			wantErr:   true,
			errMsg:    "JWT_SECRET must be at least 32 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("JWT_SECRET", tt.jwtSecret)
			// Valid encryption key for testing
			validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
			os.Setenv("ENCRYPTION_KEY", validKey)
			defer func() {
				os.Unsetenv("JWT_SECRET")
				os.Unsetenv("ENCRYPTION_KEY")
			}()

			_, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Load() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoadConfig_EncryptionKeyValidation(t *testing.T) {
	validJWT := "12345678901234567890123456789012"

	tests := []struct {
		name          string
		encryptionKey string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "valid encryption key (32 bytes base64)",
			encryptionKey: base64.StdEncoding.EncodeToString(make([]byte, 32)),
			wantErr:       false,
		},
		{
			name:          "empty encryption key",
			encryptionKey: "",
			wantErr:       true,
			errContains:   "ENCRYPTION_KEY is required",
		},
		{
			name:          "invalid base64",
			encryptionKey: "not-valid-base64!!!",
			wantErr:       true,
			errContains:   "must be valid base64",
		},
		{
			name:          "wrong key length (16 bytes)",
			encryptionKey: base64.StdEncoding.EncodeToString(make([]byte, 16)),
			wantErr:       true,
			errContains:   "must be 32 bytes",
		},
		{
			name:          "wrong key length (64 bytes)",
			encryptionKey: base64.StdEncoding.EncodeToString(make([]byte, 64)),
			wantErr:       true,
			errContains:   "must be 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("JWT_SECRET", validJWT)
			os.Setenv("ENCRYPTION_KEY", tt.encryptionKey)
			defer func() {
				os.Unsetenv("JWT_SECRET")
				os.Unsetenv("ENCRYPTION_KEY")
			}()

			_, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Set required fields
	os.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	os.Setenv("ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ENCRYPTION_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check defaults
	if cfg.Server.Host != DefaultServerHost {
		t.Errorf("Server.Host = %v, want %v", cfg.Server.Host, DefaultServerHost)
	}
	if cfg.Server.Port != DefaultServerPort {
		t.Errorf("Server.Port = %v, want %v", cfg.Server.Port, DefaultServerPort)
	}
	if cfg.Database.Type != "sqlite" {
		t.Errorf("Database.Type = %v, want sqlite", cfg.Database.Type)
	}
	if cfg.Database.Path != DefaultSQLitePath {
		t.Errorf("Database.Path = %v, want %v", cfg.Database.Path, DefaultSQLitePath)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

//
