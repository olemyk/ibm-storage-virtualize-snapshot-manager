package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/config"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/models"
	"github.com/ibm-storage-virtualize-snapshot-manager/pkg/crypto"
)

// Service handles authentication operations
type Service struct {
	jwtSecret string
}

// NewService creates a new authentication service
func NewService(jwtSecret string) *Service {
	return &Service{
		jwtSecret: jwtSecret,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(config.JWTExpirationHours) * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is about to expire (within buffer period)
	if claims.ExpiresAt != nil {
		expiryTime := claims.ExpiresAt.Time
		bufferDuration := time.Duration(config.TokenRefreshBufferMinutes) * time.Minute

		if time.Now().Add(bufferDuration).After(expiryTime) {
			return nil, fmt.Errorf("token expires too soon, please refresh")
		}
	}

	return claims, nil
}

// ValidatePasswordComplexity validates password meets complexity requirements
func (s *Service) ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' ||
			char == '^' || char == '&' || char == '*' || char == '(' || char == ')' ||
			char == '-' || char == '_' || char == '+' || char == '=' || char == '{' ||
			char == '}' || char == '[' || char == ']' || char == '|' || char == '\\' ||
			char == ':' || char == ';' || char == '"' || char == '\'' || char == '<' ||
			char == '>' || char == ',' || char == '.' || char == '?' || char == '/':
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// HashPassword hashes a password
func (s *Service) HashPassword(password string) (string, error) {
	return crypto.HashPassword(password)
}

// CheckPassword checks if a password matches a hash
func (s *Service) CheckPassword(password, hash string) bool {
	return crypto.CheckPasswordHash(password, hash)
}

//
