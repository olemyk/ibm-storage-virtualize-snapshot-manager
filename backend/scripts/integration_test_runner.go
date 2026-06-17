package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Test configuration
const (
	BaseURL  = "https://localhost"
	Username = "admin"
	Password = "admin123"

	SVC01IP   = "10.33.7.80"
	SVC01Port = 7443
	SVC02IP   = "10.33.7.81"
	SVC02Port = 7443
	SVCUser   = "snapshotmanager"
	SVCPass   = "snapshotmanager"

	VG1Name = "snapshotmanager_vg_svc1_01"
	VG2Name = "snapshotmanager_vg_svc2_01"
)

type TestClient struct {
	httpClient *http.Client
	token      string
	csrfToken  string
	logFile    *os.File
}

func NewTestClient() *TestClient {
	// Create HTTP client that accepts self-signed certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	logFile, err := os.Create("integration_test.log")
	if err != nil {
		log.Fatal("Failed to create log file:", err)
	}

	return &TestClient{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
		logFile: logFile,
	}
}

func (c *TestClient) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] %s\n", timestamp, msg)

	fmt.Print(logMsg)
	c.logFile.WriteString(logMsg)
}

func (c *TestClient) Close() {
	c.logFile.Close()
}

func (c *TestClient) makeRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return respBody, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Test Steps

func (c *TestClient) Step1_Login() error {
	c.Log("STEP 1: Login as admin")

	body := map[string]string{
		"username": Username,
		"password": Password,
	}

	respBody, err := c.makeRequest("POST", "/api/auth/login", body)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}

	c.token = result["token"].(string)
	c.csrfToken = result["csrf_token"].(string)
	c.Log("✓ Login successful, token obtained")
	return nil
}

func (c *TestClient) Step2_AddSystems() (int, int, error) {
	c.Log("STEP 2: Add storage systems SVC01 and SVC02")

	// Add SVC01
	svc01Body := map[string]interface{}{
		"name":            "SVC01",
		"ip_address":      SVC01IP,
		"port":            SVC01Port,
		"username":        SVCUser,
		"password":        SVCPass,
		"skip_tls_verify": true,
	}

	respBody, err := c.makeRequest("POST", "/api/systems", svc01Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to add SVC01: %w", err)
	}

	var svc01Result map[string]interface{}
	json.Unmarshal(respBody, &svc01Result)
	svc01ID := int(svc01Result["id"].(float64))
	c.Log("✓ SVC01 added (ID: %d, IP: %s)", svc01ID, SVC01IP)

	// Add SVC02
	svc02Body := map[string]interface{}{
		"name":            "SVC02",
		"ip_address":      SVC02IP,
		"port":            SVC02Port,
		"username":        SVCUser,
		"password":        SVCPass,
		"skip_tls_verify": true,
	}

	respBody, err = c.makeRequest("POST", "/api/systems", svc02Body)
	if err != nil {
		return svc01ID, 0, fmt.Errorf("failed to add SVC02: %w", err)
	}

	var svc02Result map[string]interface{}
	json.Unmarshal(respBody, &svc02Result)
	svc02ID := int(svc02Result["id"].(float64))
	c.Log("✓ SVC02 added (ID: %d, IP: %s)", svc02ID, SVC02IP)

	return svc01ID, svc02ID, nil
}

func (c *TestClient) Step3_TestConnections(svc01ID, svc02ID int) error {
	c.Log("STEP 3: Test system connections")

	maxRetries := 3
	retryDelay := 30 * time.Second

	// Test SVC01 with retry logic for token limit
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err = c.makeRequest("POST", fmt.Sprintf("/api/systems/%d/test", svc01ID), nil)
		if err == nil {
			c.Log("✓ SVC01 connection successful")
			break
		}

		// Check if error contains token limit message
		errMsg := err.Error()
		if attempt < maxRetries && (len(errMsg) > 0 && (errMsg[0:8] == "HTTP 400" || errMsg[0:8] == "HTTP 429")) {
			c.Log("⚠ Connection failed (HTTP 400/429 - possibly token limit), waiting %v for tokens to expire (attempt %d/%d)...", retryDelay, attempt, maxRetries)
			time.Sleep(retryDelay)
			continue
		}

		return fmt.Errorf("SVC01 connection test failed: %w", err)
	}

	if err != nil {
		return fmt.Errorf("SVC01 connection test failed after %d attempts: %w", maxRetries, err)
	}

	// Test SVC02 with retry logic for token limit
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err = c.makeRequest("POST", fmt.Sprintf("/api/systems/%d/test", svc02ID), nil)
		if err == nil {
			c.Log("✓ SVC02 connection successful")
			break
		}

		// Check if error contains token limit message
		errMsg := err.Error()
		if attempt < maxRetries && (len(errMsg) > 0 && (errMsg[0:8] == "HTTP 400" || errMsg[0:8] == "HTTP 429")) {
			c.Log("⚠ Connection failed (HTTP 400/429 - possibly token limit), waiting %v for tokens to expire (attempt %d/%d)...", retryDelay, attempt, maxRetries)
			time.Sleep(retryDelay)
			continue
		}

		return fmt.Errorf("SVC02 connection test failed: %w", err)
	}

	if err != nil {
		return fmt.Errorf("SVC02 connection test failed after %d attempts: %w", maxRetries, err)
	}

	return nil
}

func (c *TestClient) Step4_SyncVolumeGroups(svc01ID, svc02ID int) (int, int, error) {
	c.Log("STEP 4: Sync volume groups")

	// Sync SVC01
	_, err := c.makeRequest("POST", fmt.Sprintf("/api/systems/%d/volumegroups/sync", svc01ID), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to sync SVC01 volume groups: %w", err)
	}
	c.Log("✓ SVC01 volume groups synced")

	// Sync SVC02
	_, err = c.makeRequest("POST", fmt.Sprintf("/api/systems/%d/volumegroups/sync", svc02ID), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to sync SVC02 volume groups: %w", err)
	}
	c.Log("✓ SVC02 volume groups synced")

	// Get volume group IDs
	respBody, err := c.makeRequest("GET", fmt.Sprintf("/api/systems/%d/volumegroups", svc01ID), nil)
	if err != nil {
		return 0, 0, err
	}

	var vgs1 []map[string]interface{}
	json.Unmarshal(respBody, &vgs1)

	var vg1ID int
	for _, vg := range vgs1 {
		if vg["vg_name"].(string) == VG1Name {
			vg1ID = int(vg["id"].(float64))
			break
		}
	}

	respBody, err = c.makeRequest("GET", fmt.Sprintf("/api/systems/%d/volumegroups", svc02ID), nil)
	if err != nil {
		return vg1ID, 0, err
	}

	var vgs2 []map[string]interface{}
	json.Unmarshal(respBody, &vgs2)

	var vg2ID int
	for _, vg := range vgs2 {
		if vg["vg_name"].(string) == VG2Name {
			vg2ID = int(vg["id"].(float64))
			break
		}
	}

	if vg1ID == 0 || vg2ID == 0 {
		return vg1ID, vg2ID, fmt.Errorf("volume groups not found (VG1: %d, VG2: %d)", vg1ID, vg2ID)
	}

	c.Log("✓ Volume group IDs: VG1=%d, VG2=%d", vg1ID, vg2ID)
	return vg1ID, vg2ID, nil
}

func (c *TestClient) Step5_CreateSchedules(vg1ID, vg2ID int) (int, int, error) {
	c.Log("STEP 5: Create snapshot schedules")

	// Schedule for VG1 - Daily at 02:00, 2 days retention
	schedule1Body := map[string]interface{}{
		"name":                  "Daily_02_00_2d_retention",
		"cron_expression":       "0 2 * * *",
		"retention_days":        2,
		"safeguarded":           false,
		"snapshot_name_pattern": "snap_{timestamp}",
		"is_active":             true,
	}

	respBody, err := c.makeRequest("POST", fmt.Sprintf("/api/volumegroups/%d/schedules", vg1ID), schedule1Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create schedule for VG1: %w", err)
	}

	var result1 map[string]interface{}
	json.Unmarshal(respBody, &result1)
	schedule1ID := int(result1["id"].(float64))
	c.Log("✓ Schedule created for VG1 (ID: %d)", schedule1ID)

	// Schedule for VG2 - Daily at 03:00, 2 days retention
	schedule2Body := map[string]interface{}{
		"name":                  "Daily_03_00_2d_retention",
		"cron_expression":       "0 3 * * *",
		"retention_days":        2,
		"safeguarded":           false,
		"snapshot_name_pattern": "snap_{timestamp}",
		"is_active":             true,
	}

	respBody, err = c.makeRequest("POST", fmt.Sprintf("/api/volumegroups/%d/schedules", vg2ID), schedule2Body)
	if err != nil {
		return schedule1ID, 0, fmt.Errorf("failed to create schedule for VG2: %w", err)
	}

	var result2 map[string]interface{}
	json.Unmarshal(respBody, &result2)
	schedule2ID := int(result2["id"].(float64))
	c.Log("✓ Schedule created for VG2 (ID: %d)", schedule2ID)

	return schedule1ID, schedule2ID, nil
}

func (c *TestClient) Step6_ExecuteSchedules(schedule1ID, schedule2ID int) error {
	c.Log("STEP 6: Execute schedules manually")

	// Execute schedule 1
	_, err := c.makeRequest("POST", fmt.Sprintf("/api/schedules/%d/execute", schedule1ID), nil)
	if err != nil {
		return fmt.Errorf("failed to execute schedule 1: %w", err)
	}
	c.Log("✓ Schedule 1 execution started")

	// Execute schedule 2
	_, err = c.makeRequest("POST", fmt.Sprintf("/api/schedules/%d/execute", schedule2ID), nil)
	if err != nil {
		return fmt.Errorf("failed to execute schedule 2: %w", err)
	}
	c.Log("✓ Schedule 2 execution started")

	// Wait for executions to complete
	c.Log("Waiting 10 seconds for executions to complete...")
	time.Sleep(10 * time.Second)

	// Check execution status
	respBody, err := c.makeRequest("GET", "/api/executions?limit=10", nil)
	if err != nil {
		return fmt.Errorf("failed to get executions: %w", err)
	}

	var executions []map[string]interface{}
	json.Unmarshal(respBody, &executions)

	successCount := 0
	failedCount := 0
	for _, exec := range executions {
		status := exec["status"].(string)
		if status == "success" {
			successCount++
		} else if status == "failed" {
			failedCount++
			if errorMsg, ok := exec["error_message"]; ok && errorMsg != nil {
				c.Log("  ⚠ Execution failed: %v", errorMsg)
			}
		}
	}

	c.Log("✓ Executions completed: %d successful, %d failed out of %d total", successCount, failedCount, len(executions))

	if successCount < 2 {
		return fmt.Errorf("expected at least 2 successful executions, got %d", successCount)
	}

	return nil
}

func (c *TestClient) Step7_AddNTPServer() error {
	c.Log("STEP 7: Add NTP server and sync")

	ntpBody := map[string]interface{}{
		"server_address": "no.pool.ntp.org",
		"is_active":      true,
		"priority":       1,
	}

	_, err := c.makeRequest("POST", "/api/ntp/servers", ntpBody)
	if err != nil {
		return fmt.Errorf("failed to add NTP server: %w", err)
	}
	c.Log("✓ NTP server added")

	// Trigger sync
	_, err = c.makeRequest("POST", "/api/ntp-servers/sync", nil)
	if err != nil {
		c.Log("⚠ NTP sync failed (this is OK if NTP service is not running): %v", err)
	} else {
		c.Log("✓ NTP sync triggered")
	}

	return nil
}

func (c *TestClient) Step8_CheckAuditLogs() error {
	c.Log("STEP 8: Check audit logs")

	respBody, err := c.makeRequest("GET", "/api/audit-logs?limit=20", nil)
	if err != nil {
		return fmt.Errorf("failed to get audit logs: %w", err)
	}

	var result struct {
		Logs   []map[string]interface{} `json:"logs"`
		Total  int                      `json:"total"`
		Limit  int                      `json:"limit"`
		Offset int                      `json:"offset"`
	}
	json.Unmarshal(respBody, &result)

	c.Log("✓ Audit logs retrieved: %d entries (total: %d)", len(result.Logs), result.Total)

	if result.Total == 0 {
		return fmt.Errorf("no audit logs found - audit system may not be working")
	}

	// Check for recent login audit
	foundLogin := false
	for _, log := range result.Logs {
		if log["action"].(string) == "login" {
			foundLogin = true
			break
		}
	}

	if !foundLogin {
		c.Log("⚠ No login audit entry found in recent logs")
	} else {
		c.Log("✓ Login audit entry found")
	}

	return nil
}

func (c *TestClient) Step9_CreateTestUsers() ([]int, error) {
	c.Log("STEP 9: Create test users with different roles")

	users := []struct {
		username string
		password string
		role     string
	}{
		{"test_viewer", "TestViewer123!", "viewer"},
		{"test_operator", "TestOperator123!", "operator"},
		{"test_admin", "TestAdmin123!", "admin"},
	}

	var userIDs []int

	for _, u := range users {
		userBody := map[string]interface{}{
			"username": u.username,
			"password": u.password,
			"email":    u.username + "@test.local",
			"role":     u.role,
		}

		respBody, err := c.makeRequest("POST", "/api/users", userBody)
		if err != nil {
			return userIDs, fmt.Errorf("failed to create user %s: %w", u.username, err)
		}

		var result map[string]interface{}
		json.Unmarshal(respBody, &result)
		userID := int(result["id"].(float64))
		userIDs = append(userIDs, userID)

		c.Log("✓ User created: %s (role: %s, ID: %d)", u.username, u.role, userID)
	}

	return userIDs, nil
}

func (c *TestClient) Step10_TestUserLogin() error {
	c.Log("STEP 10: Test login with different user roles")

	// Save current token
	originalToken := c.token
	originalCSRF := c.csrfToken

	// Test viewer login
	viewerBody := map[string]string{
		"username": "test_viewer",
		"password": "TestViewer123!",
	}

	_, err := c.makeRequest("POST", "/api/auth/login", viewerBody)
	if err != nil {
		return fmt.Errorf("viewer login failed: %w", err)
	}
	c.Log("✓ Viewer login successful")

	// Test operator login
	operatorBody := map[string]string{
		"username": "test_operator",
		"password": "TestOperator123!",
	}

	_, err = c.makeRequest("POST", "/api/auth/login", operatorBody)
	if err != nil {
		return fmt.Errorf("operator login failed: %w", err)
	}
	c.Log("✓ Operator login successful")

	// Restore admin token
	c.token = originalToken
	c.csrfToken = originalCSRF

	return nil
}

func (c *TestClient) Step11_UpdateSchedule(scheduleID int) error {
	c.Log("STEP 11: Update schedule")

	updateBody := map[string]interface{}{
		"name":                  "Daily 02:00 - 3d retention (updated)",
		"cron_expression":       "0 2 * * *",
		"retention_days":        3, // Changed from 2 to 3
		"safeguarded":           false,
		"snapshot_name_pattern": "{schedule_name}_{timestamp}",
		"is_active":             true,
	}

	_, err := c.makeRequest("PUT", fmt.Sprintf("/api/schedules/%d", scheduleID), updateBody)
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	c.Log("✓ Schedule updated (retention changed to 3 days)")
	return nil
}

func (c *TestClient) Step12_DeleteSchedule(scheduleID int) error {
	c.Log("STEP 12: Delete schedule")

	_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/schedules/%d", scheduleID), nil)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	c.Log("✓ Schedule deleted (ID: %d)", scheduleID)
	return nil
}

func (c *TestClient) Step13_DeleteSystem(systemID int) error {
	c.Log("STEP 13: Delete storage system")

	_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/systems/%d", systemID), nil)
	if err != nil {
		return fmt.Errorf("failed to delete system: %w", err)
	}

	c.Log("✓ System deleted (ID: %d)", systemID)
	return nil
}

func (c *TestClient) Step14_Cleanup(userIDs []int, svc01ID int) error {
	c.Log("STEP 14: Cleanup test data")

	// Delete test users
	for _, userID := range userIDs {
		_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/users/%d", userID), nil)
		if err != nil {
			c.Log("⚠ Failed to delete user %d: %v", userID, err)
		} else {
			c.Log("✓ Test user deleted (ID: %d)", userID)
		}
	}

	// Delete remaining system
	_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/systems/%d", svc01ID), nil)
	if err != nil {
		c.Log("⚠ Failed to delete SVC01: %v", err)
	} else {
		c.Log("✓ SVC01 deleted")
	}

	return nil
}

func (c *TestClient) Step0_Cleanup() error {
	c.Log("STEP 0: Cleanup existing test data")

	// Get all systems
	respBody, err := c.makeRequest("GET", "/api/systems", nil)
	if err != nil {
		c.Log("⚠ Could not get systems for cleanup: %v", err)
		return nil // Don't fail on cleanup
	}

	var systems []map[string]interface{}
	json.Unmarshal(respBody, &systems)

	// Delete ALL systems to ensure clean state
	deletedCount := 0
	for _, sys := range systems {
		name := sys["name"].(string)
		id := int(sys["id"].(float64))
		_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/systems/%d", id), nil)
		if err != nil {
			c.Log("⚠ Could not delete system %s (ID: %d): %v", name, id, err)
		} else {
			c.Log("✓ Deleted existing system: %s (ID: %d)", name, id)
			deletedCount++
		}
	}

	if deletedCount > 0 {
		c.Log("✓ Cleanup complete: deleted %d system(s)", deletedCount)
	} else {
		c.Log("✓ No systems to clean up")
	}

	// Also cleanup NTP servers
	respBody, err = c.makeRequest("GET", "/api/ntp/servers", nil)
	if err == nil {
		var ntpServers []map[string]interface{}
		json.Unmarshal(respBody, &ntpServers)

		for _, ntp := range ntpServers {
			id := int(ntp["id"].(float64))
			_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/ntp/servers/%d", id), nil)
			if err != nil {
				c.Log("⚠ Could not delete NTP server (ID: %d): %v", id, err)
			} else {
				c.Log("✓ Deleted existing NTP server (ID: %d)", id)
			}
		}
	}

	// Cleanup test users
	respBody, err = c.makeRequest("GET", "/api/users", nil)
	if err == nil {
		var users []map[string]interface{}
		json.Unmarshal(respBody, &users)

		for _, user := range users {
			username := user["username"].(string)
			if username == "test_viewer" || username == "test_operator" || username == "test_admin" {
				id := int(user["id"].(float64))
				_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/users/%d", id), nil)
				if err != nil {
					c.Log("⚠ Could not delete test user %s (ID: %d): %v", username, id, err)
				} else {
					c.Log("✓ Deleted existing test user: %s (ID: %d)", username, id)
				}
			}
		}
	}

	return nil
}

func main() {
	client := NewTestClient()
	defer client.Close()

	client.Log("=== IBM Storage Virtualize Snapshot Manager - Integration Test ===")
	client.Log("Test started at: %s", time.Now().Format("2006-01-02 15:04:05"))
	client.Log("")

	// Step 1: Login
	if err := client.Step1_Login(); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 0: Cleanup (after login to have auth token)
	if err := client.Step0_Cleanup(); err != nil {
		client.Log("⚠ Cleanup had errors: %v", err)
	}

	// Step 2: Add systems
	svc01ID, svc02ID, err := client.Step2_AddSystems()
	if err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 3: Test connections
	if err := client.Step3_TestConnections(svc01ID, svc02ID); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 4: Sync volume groups
	vg1ID, vg2ID, err := client.Step4_SyncVolumeGroups(svc01ID, svc02ID)
	if err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 5: Create schedules
	schedule1ID, schedule2ID, err := client.Step5_CreateSchedules(vg1ID, vg2ID)
	if err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 6: Execute schedules
	if err := client.Step6_ExecuteSchedules(schedule1ID, schedule2ID); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 7: Add NTP server
	if err := client.Step7_AddNTPServer(); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 8: Check audit logs
	if err := client.Step8_CheckAuditLogs(); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 9: Create test users
	userIDs, err := client.Step9_CreateTestUsers()
	if err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 10: Test user login
	if err := client.Step10_TestUserLogin(); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 11: Update schedule
	if err := client.Step11_UpdateSchedule(schedule1ID); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 12: Delete schedule
	if err := client.Step12_DeleteSchedule(schedule2ID); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 13: Delete system
	if err := client.Step13_DeleteSystem(svc02ID); err != nil {
		client.Log("❌ FAILED: %v", err)
		return
	}

	// Step 14: Cleanup
	if err := client.Step14_Cleanup(userIDs, svc01ID); err != nil {
		client.Log("⚠ Cleanup had errors: %v", err)
	}

	client.Log("")
	client.Log("=== TEST COMPLETED SUCCESSFULLY ===")
	client.Log("Test finished at: %s", time.Now().Format("2006-01-02 15:04:05"))
	client.Log("Log file: integration_test.log")
}
