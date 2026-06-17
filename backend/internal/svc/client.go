package svc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/config"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/models"
)

// Client represents an IBM Storage Virtualize REST API client
type Client struct {
	tokenCache      map[int]*TokenInfo
	mu              sync.RWMutex
	clusterTokens   map[string]int         // Track token count per cluster (by IP:Port)
	systemToCluster map[int]string         // Map system ID to cluster identifier
	authRateLimiter map[string][]time.Time // Track auth requests per cluster for rate limiting
	cmdRateLimiter  map[string][]time.Time // Track command requests per cluster for rate limiting
}

// TokenInfo holds token and expiry information
type TokenInfo struct {
	Token     string
	ExpiresAt time.Time
}

// NewClient creates a new SVC client
func NewClient() *Client {
	return &Client{
		tokenCache:      make(map[int]*TokenInfo),
		clusterTokens:   make(map[string]int),
		systemToCluster: make(map[int]string),
		authRateLimiter: make(map[string][]time.Time),
		cmdRateLimiter:  make(map[string][]time.Time),
	}
}

// getHTTPClient returns an HTTP client configured for the given system
func (c *Client) getHTTPClient(skipTLSVerify bool, timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTLSVerify,
			},
		},
	}
}

// checkAuthRateLimit checks if we can make an auth request (max 3/sec per cluster)
func (c *Client) checkAuthRateLimit(clusterID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	oneSecondAgo := now.Add(-1 * time.Second)

	// Clean up old entries
	var recent []time.Time
	for _, t := range c.authRateLimiter[clusterID] {
		if t.After(oneSecondAgo) {
			recent = append(recent, t)
		}
	}

	// Check if we've hit the limit (3 requests per second)
	if len(recent) >= 3 {
		return fmt.Errorf("auth rate limit exceeded for cluster %s (max 3 requests/second)", clusterID)
	}

	// Add current request
	recent = append(recent, now)
	c.authRateLimiter[clusterID] = recent

	return nil
}

// checkCommandRateLimit checks if we can make a command request (max 10/sec per cluster)
func (c *Client) checkCommandRateLimit(clusterID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	oneSecondAgo := now.Add(-1 * time.Second)

	// Clean up old entries
	var recent []time.Time
	for _, t := range c.cmdRateLimiter[clusterID] {
		if t.After(oneSecondAgo) {
			recent = append(recent, t)
		}
	}

	// Check if we've hit the limit (10 requests per second)
	if len(recent) >= 10 {
		// Wait a bit and retry once
		c.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		c.mu.Lock()

		// Re-check after wait
		recent = nil
		for _, t := range c.cmdRateLimiter[clusterID] {
			if t.After(now.Add(-1 * time.Second)) {
				recent = append(recent, t)
			}
		}

		if len(recent) >= 10 {
			return fmt.Errorf("command rate limit exceeded for cluster %s (max 10 requests/second)", clusterID)
		}
	}

	// Add current request
	recent = append(recent, now)
	c.cmdRateLimiter[clusterID] = recent

	return nil
}

// Authenticate authenticates with IBM SVC and returns a token
func (c *Client) Authenticate(system *models.StorageSystem, password string) (string, error) {
	clusterID := fmt.Sprintf("%s:%d", system.IPAddress, system.Port)

	// Check auth rate limit (3 requests/second per cluster)
	if err := c.checkAuthRateLimit(clusterID); err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s:%d/rest/v1/auth", system.IPAddress, system.Port)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Username", system.Username)
	req.Header.Set("X-Auth-Password", password)

	httpClient := c.getHTTPClient(system.SkipTLSVerify, 30*time.Second)
	// Retry with exponential backoff on rate limit
	var resp *http.Response
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Create new request for each attempt to avoid reusing request body
		retryReq, reqErr := http.NewRequest("POST", url, nil)
		if reqErr != nil {
			return "", fmt.Errorf("failed to create retry request: %w", reqErr)
		}
		retryReq.Header.Set("Content-Type", "application/json")
		retryReq.Header.Set("X-Auth-Username", system.Username)
		retryReq.Header.Set("X-Auth-Password", password)

		resp, err = httpClient.Do(retryReq)
		if err != nil {
			return "", fmt.Errorf("failed to authenticate: %w", err)
		}

		if resp.StatusCode != 429 {
			break
		}

		resp.Body.Close()

		if attempt < maxRetries {
			delay := baseDelay * time.Duration(1<<uint(attempt)) // Exponential: 1s, 2s, 4s
			log.Printf("Rate limit hit (429), retrying in %v (attempt %d/%d)", delay, attempt+1, maxRetries)
			time.Sleep(delay)
		} else {
			return "", fmt.Errorf("rate limit exceeded after %d retries (429)", maxRetries)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return "", fmt.Errorf("invalid credentials (403)")
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Decode JWT to get expiry time
	expiresAt, err := c.getTokenExpiry(result.Token)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	// Cache the token (let IBM SVC enforce the 4-token limit, not client-side)
	c.mu.Lock()
	clusterID = fmt.Sprintf("%s:%d", system.IPAddress, system.Port)

	// Track cluster association for this system
	c.systemToCluster[system.ID] = clusterID

	c.tokenCache[system.ID] = &TokenInfo{
		Token:     result.Token,
		ExpiresAt: expiresAt,
	}
	c.mu.Unlock()

	return result.Token, nil
}

// GetOrRefreshToken gets a cached token or refreshes it if expired
func (c *Client) GetOrRefreshToken(system *models.StorageSystem, password string) (string, error) {
	c.mu.RLock()
	tokenInfo := c.tokenCache[system.ID]
	c.mu.RUnlock()

	// Check if token exists and is not expired (with buffer time before expiry)
	bufferDuration := time.Duration(config.TokenRefreshBufferMinutes) * time.Minute
	if tokenInfo != nil && time.Now().Before(tokenInfo.ExpiresAt.Add(-bufferDuration)) {
		return tokenInfo.Token, nil
	}

	// Token expired or doesn't exist, authenticate
	return c.Authenticate(system, password)
}

// getTokenExpiry decodes JWT token to extract expiry time
func (c *Client) getTokenExpiry(tokenString string) (time.Time, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, fmt.Errorf("exp claim not found")
	}

	return time.Unix(int64(exp), 0), nil
}

// doRequest performs an authenticated request to IBM SVC with retry logic
func (c *Client) doRequest(system *models.StorageSystem, token, endpoint string, body interface{}) ([]byte, error) {
	clusterID := fmt.Sprintf("%s:%d", system.IPAddress, system.Port)

	// Check command rate limit (10 requests/second per cluster)
	if err := c.checkCommandRateLimit(clusterID); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s:%d/rest/v1%s", system.IPAddress, system.Port, endpoint)

	var jsonBody []byte
	var err error
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	httpClient := c.getHTTPClient(system.SkipTLSVerify, 30*time.Second)

	// Retry with exponential backoff on rate limit
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var reqBody io.Reader
		if jsonBody != nil {
			reqBody = bytes.NewBuffer(jsonBody)
		}

		req, err := http.NewRequest("POST", url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Auth-Token", token)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode == 429 {
			if attempt < maxRetries {
				delay := baseDelay * time.Duration(1<<uint(attempt))
				log.Printf("Rate limit hit (429) on %s, retrying in %v (attempt %d/%d)", endpoint, delay, attempt+1, maxRetries)
				time.Sleep(delay)
				continue
			}
			return nil, fmt.Errorf("rate limit exceeded after %d retries (429)", maxRetries)
		}

		if resp.StatusCode == 401 {
			// Token expired, clear cache so next call will refresh
			c.mu.Lock()
			delete(c.tokenCache, system.ID)
			c.mu.Unlock()
			return nil, fmt.Errorf("token expired (401)")
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

// ListVolumeGroups lists all volume groups from the system
func (c *Client) ListVolumeGroups(system *models.StorageSystem, token string) ([]map[string]interface{}, error) {
	respBody, err := c.doRequest(system, token, "/lsvolumegroup", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// ListVolumeGroupSnapshots lists snapshots for a volume group
func (c *Client) ListVolumeGroupSnapshots(system *models.StorageSystem, token, volumeGroup string) ([]map[string]interface{}, error) {
	// Use filtervalue to filter by volume group name
	body := map[string]interface{}{
		"filtervalue": "volume_group_name=" + volumeGroup,
	}

	respBody, err := c.doRequest(system, token, "/lsvolumegroupsnapshot", body)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// ListVolumesInGroup lists all volumes in a volume group
func (c *Client) ListVolumesInGroup(system *models.StorageSystem, token string, volumeGroupID string) ([]map[string]interface{}, error) {
	body := map[string]interface{}{
		"filtervalue": "volume_group_id=" + volumeGroupID,
	}

	respBody, err := c.doRequest(system, token, "/lsvdisk", body)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// AddSnapshotRequest represents a request to add a snapshot
type AddSnapshotRequest struct {
	VolumeGroup      string  `json:"volumegroup"`
	RetentionDays    *int    `json:"retentiondays,omitempty"`
	RetentionMinutes *int    `json:"retentionminutes,omitempty"`
	Safeguarded      bool    `json:"safeguarded,omitempty"`
	Name             string  `json:"name,omitempty"`
	Pool             *string `json:"pool,omitempty"`
}

// AddSnapshot creates a snapshot for a volume group
func (c *Client) AddSnapshot(system *models.StorageSystem, token string, req AddSnapshotRequest) (map[string]interface{}, error) {
	respBody, err := c.doRequest(system, token, "/addsnapshot", req)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// TestConnection tests the connection to an IBM SVC system
func (c *Client) TestConnection(system *models.StorageSystem, password string) error {
	_, err := c.Authenticate(system, password)
	return err
}

//
