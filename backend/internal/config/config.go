package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
	Logging  LoggingConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host           string
	Port           int
	AllowedOrigins []string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string // sqlite or postgres
	Path     string // for sqlite
	Host     string // for postgres
	Port     int    // for postgres
	Name     string // for postgres
	User     string // for postgres
	Password string // for postgres
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret     string
	EncryptionKey string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string
	File  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:           getEnv("SERVER_HOST", DefaultServerHost),
			Port:           getEnvAsInt("SERVER_PORT", DefaultServerPort),
			AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174,http://localhost:5175,http://localhost:5176,http://localhost:3000"),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "sqlite"),
			Path:     getEnv("DB_PATH", DefaultSQLitePath),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", DefaultPostgresPort),
			Name:     getEnv("DB_NAME", "snapshots"),
			User:     getEnv("DB_USER", "snapshots"),
			Password: getEnv("DB_PASSWORD", ""),
		},
		Security: SecurityConfig{
			JWTSecret:     getEnv("JWT_SECRET", ""),
			EncryptionKey: getEnv("ENCRYPTION_KEY", ""),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			File:  getEnv("LOG_FILE", "./logs/app.log"),
		},
	}

	// Validate required fields
	if cfg.Security.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.Security.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	if cfg.Security.EncryptionKey == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY is required")
	}
	// Validate encryption key is valid base64 and correct length
	keyBytes, err := base64.StdEncoding.DecodeString(cfg.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be valid base64: %w", err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be 32 bytes (256 bits) when decoded, got %d bytes", len(keyBytes))
	}

	return cfg, nil
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as int or returns default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsSlice gets environment variable as comma-separated slice or returns default value
func getEnvAsSlice(key, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

//
