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

// TestClient handles HTTP requests to the API
type TestClient struct {
	BaseURL    string
	Token      string
	CSRFToken  string
	HTTPClient *http.Client
}

// NewTestClient creates a new test client
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

// Log prints a timestamped message
func (c *TestClient) Log(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
}

// makeRequest makes an HTTP request to the API
func (c *TestClient) makeRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if c.CSRFToken != "" {
		req.Header.Set("X-CSRF-Token", c.CSRFToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Step1_Login logs in as admin
func (c *TestClient) Step1_Login() error {
	c.Log("STEP 1: Login as admin")

	loginBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}

	respBody, err := c.makeRequest("POST", "/api/auth/login", loginBody)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	var loginResp struct {
		Token     string `json:"token"`
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	c.Token = loginResp.Token
	c.CSRFToken = loginResp.CSRFToken
	c.Log("✓ Login successful, token obtained")
	return nil
}

// Step2_AddStorageSystems adds two IBM SVC systems
func (c *TestClient) Step2_AddStorageSystems() (int, int, error) {
	c.Log("STEP 2: Add storage systems SVC01 and SVC02")

	// Add SVC01
	svc1Body := map[string]interface{}{
		"name":            "SVC01_Manual",
		"ip_address":      "10.33.7.80",
		"port":            7443,
		"username":        "snapshotmanager",
		"password":        "snapshotmanager",
		"skip_tls_verify": true,
	}

	respBody, err := c.makeRequest("POST", "/api/systems", svc1Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to add SVC01: %w", err)
	}

	var svc1Resp map[string]interface{}
	json.Unmarshal(respBody, &svc1Resp)
	svc1ID := int(svc1Resp["id"].(float64))
	c.Log("✓ SVC01_Manual added (ID: %d, IP: 10.33.7.80)", svc1ID)

	// Add SVC02
	svc2Body := map[string]interface{}{
		"name":            "SVC02_Manual",
		"ip_address":      "10.33.7.81",
		"port":            7443,
		"username":        "snapshotmanager",
		"password":        "snapshotmanager",
		"skip_tls_verify": true,
	}

	respBody, err = c.makeRequest("POST", "/api/systems", svc2Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to add SVC02: %w", err)
	}

	var svc2Resp map[string]interface{}
	json.Unmarshal(respBody, &svc2Resp)
	svc2ID := int(svc2Resp["id"].(float64))
	c.Log("✓ SVC02_Manual added (ID: %d, IP: 10.33.7.81)", svc2ID)

	return svc1ID, svc2ID, nil
}

// Step3_TestConnections tests connections to both systems
func (c *TestClient) Step3_TestConnections(svc1ID, svc2ID int) error {
	c.Log("STEP 3: Test system connections")

	// Test SVC01
	_, err := c.makeRequest("POST", fmt.Sprintf("/api/systems/%d/test", svc1ID), nil)
	if err != nil {
		return fmt.Errorf("SVC01 connection test failed: %w", err)
	}
	c.Log("✓ SVC01_Manual connection successful")

	// Test SVC02
	_, err = c.makeRequest("POST", fmt.Sprintf("/api/systems/%d/test", svc2ID), nil)
	if err != nil {
		return fmt.Errorf("SVC02 connection test failed: %w", err)
	}
	c.Log("✓ SVC02_Manual connection successful")

	return nil
}

// Step4_SyncVolumeGroups syncs volume groups from both systems
func (c *TestClient) Step4_SyncVolumeGroups(svc1ID, svc2ID int) (int, int, error) {
	c.Log("STEP 4: Sync volume groups")

	// Sync SVC01
	_, err := c.makeRequest("POST", fmt.Sprintf("/api/systems/%d/volumegroups/sync", svc1ID), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to sync SVC01 volume groups: %w", err)
	}
	c.Log("✓ SVC01_Manual volume groups synced")

	// Sync SVC02
	_, err = c.makeRequest("POST", fmt.Sprintf("/api/systems/%d/volumegroups/sync", svc2ID), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to sync SVC02 volume groups: %w", err)
	}
	c.Log("✓ SVC02_Manual volume groups synced")

	// Get volume groups for SVC01
	respBody, err := c.makeRequest("GET", fmt.Sprintf("/api/systems/%d/volumegroups", svc1ID), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get SVC01 volume groups: %w", err)
	}

	var vg1List []map[string]interface{}
	json.Unmarshal(respBody, &vg1List)

	var vg1ID int
	for _, vg := range vg1List {
		if vg["vg_name"].(string) == "snapshotmanager_vg_svc1_01" {
			vg1ID = int(vg["id"].(float64))
			break
		}
	}

	// Get volume groups for SVC02
	respBody, err = c.makeRequest("GET", fmt.Sprintf("/api/systems/%d/volumegroups", svc2ID), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get SVC02 volume groups: %w", err)
	}

	var vg2List []map[string]interface{}
	json.Unmarshal(respBody, &vg2List)

	var vg2ID int
	for _, vg := range vg2List {
		if vg["vg_name"].(string) == "snapshotmanager_vg_svc2_01" {
			vg2ID = int(vg["id"].(float64))
			break
		}
	}

	if vg1ID == 0 || vg2ID == 0 {
		return 0, 0, fmt.Errorf("could not find target volume groups")
	}

	c.Log("✓ Volume group IDs: VG1=%d, VG2=%d", vg1ID, vg2ID)
	return vg1ID, vg2ID, nil
}

// Step5_CreateSchedules creates snapshot schedules
func (c *TestClient) Step5_CreateSchedules(vg1ID, vg2ID int) (int, int, error) {
	c.Log("STEP 5: Create snapshot schedules")

	// Create schedule for VG1
	schedule1Body := map[string]interface{}{
		"volume_group_id":       vg1ID,
		"name":                  "Manual_Test_Daily_02_00",
		"cron_expression":       "0 2 * * *",
		"retention_days":        2,
		"safeguarded":           false,
		"snapshot_name_pattern": "manual_snap_{timestamp}",
		"is_active":             true,
	}

	respBody, err := c.makeRequest("POST", fmt.Sprintf("/api/volumegroups/%d/schedules", vg1ID), schedule1Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create schedule for VG1: %w", err)
	}

	var schedule1Resp map[string]interface{}
	json.Unmarshal(respBody, &schedule1Resp)
	schedule1ID := int(schedule1Resp["id"].(float64))
	c.Log("✓ Schedule created for VG1 (ID: %d)", schedule1ID)

	// Create schedule for VG2
	schedule2Body := map[string]interface{}{
		"volume_group_id":       vg2ID,
		"name":                  "Manual_Test_Daily_03_00",
		"cron_expression":       "0 3 * * *",
		"retention_days":        2,
		"safeguarded":           false,
		"snapshot_name_pattern": "manual_snap_{timestamp}",
		"is_active":             true,
	}

	respBody, err = c.makeRequest("POST", fmt.Sprintf("/api/volumegroups/%d/schedules", vg2ID), schedule2Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create schedule for VG2: %w", err)
	}

	var schedule2Resp map[string]interface{}
	json.Unmarshal(respBody, &schedule2Resp)
	schedule2ID := int(schedule2Resp["id"].(float64))
	c.Log("✓ Schedule created for VG2 (ID: %d)", schedule2ID)

	return schedule1ID, schedule2ID, nil
}

// Step6_ExecuteSchedules executes the schedules manually
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
		}
	}

	c.Log("✓ Executions completed: %d successful, %d failed out of %d total", successCount, failedCount, len(executions))
	return nil
}

func main() {
	// Create log file
	logFile, err := os.Create("integration_test_manual.log")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	client := NewTestClient("https://localhost")

	client.Log("=== IBM Storage Virtualize Snapshot Manager - Manual Integration Test ===")
	client.Log("Test started at: %s", time.Now().Format("2006-01-02 15:04:05"))
	client.Log("")
	client.Log("NOTE: This test does NOT cleanup. You can manually verify results in the UI.")
	client.Log("")

	// Step 1: Login
	if err := client.Step1_Login(); err != nil {
		client.Log("❌ FAILED: %v", err)
		os.Exit(1)
	}

	// Step 2: Add storage systems
	svc1ID, svc2ID, err := client.Step2_AddStorageSystems()
	if err != nil {
		client.Log("❌ FAILED: %v", err)
		os.Exit(1)
	}

	// Step 3: Test connections
	if err := client.Step3_TestConnections(svc1ID, svc2ID); err != nil {
		client.Log("❌ FAILED: %v", err)
		os.Exit(1)
	}

	// Step 4: Sync volume groups
	vg1ID, vg2ID, err := client.Step4_SyncVolumeGroups(svc1ID, svc2ID)
	if err != nil {
		client.Log("❌ FAILED: %v", err)
		os.Exit(1)
	}

	// Step 5: Create schedules
	schedule1ID, schedule2ID, err := client.Step5_CreateSchedules(vg1ID, vg2ID)
	if err != nil {
		client.Log("❌ FAILED: %v", err)
		os.Exit(1)
	}

	// Step 6: Execute schedules
	if err := client.Step6_ExecuteSchedules(schedule1ID, schedule2ID); err != nil {
		client.Log("❌ FAILED: %v", err)
		os.Exit(1)
	}

	client.Log("")
	client.Log("=== TEST COMPLETED SUCCESSFULLY ===")
	client.Log("Test finished at: %s", time.Now().Format("2006-01-02 15:04:05"))
	client.Log("")
	client.Log("Created resources (NOT deleted):")
	client.Log("  - Storage Systems: SVC01_Manual (ID: %d), SVC02_Manual (ID: %d)", svc1ID, svc2ID)
	client.Log("  - Volume Groups: VG1 (ID: %d), VG2 (ID: %d)", vg1ID, vg2ID)
	client.Log("  - Schedules: Manual_Test_Daily_02_00 (ID: %d), Manual_Test_Daily_03_00 (ID: %d)", schedule1ID, schedule2ID)
	client.Log("")
	client.Log("You can now:")
	client.Log("  1. Open https://localhost in your browser")
	client.Log("  2. Login with admin/admin123")
	client.Log("  3. Verify the systems, schedules, and snapshot executions")
	client.Log("  4. Manually delete the test resources when done")
	client.Log("")
	client.Log("Log file: integration_test_manual.log")
}
