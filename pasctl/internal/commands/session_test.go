// Package commands provides tests for the session commands including CCP support.
package commands

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"pasctl/internal/config"
	"pasctl/internal/output"
)

func TestConnectCommand_Name(t *testing.T) {
	cmd := &ConnectCommand{}
	if cmd.Name() != "connect" {
		t.Errorf("Name() = %v, want connect", cmd.Name())
	}
}

func TestConnectCommand_Description(t *testing.T) {
	cmd := &ConnectCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
	if !strings.Contains(strings.ToLower(desc), "connect") {
		t.Errorf("Description() should mention connect, got: %v", desc)
	}
}

func TestConnectCommand_Usage(t *testing.T) {
	cmd := &ConnectCommand{}
	usage := cmd.Usage()

	// Check for essential content
	requiredContent := []string{
		"connect",
		"--user",
		"--auth",
		"--insecure",
		"--ccp",
	}

	for _, content := range requiredContent {
		if !strings.Contains(usage, content) {
			t.Errorf("Usage() should contain %q", content)
		}
	}
}

func TestConnectCommand_CCPFlagInUsage(t *testing.T) {
	cmd := &ConnectCommand{}
	usage := cmd.Usage()

	// Verify CCP-specific documentation is present
	ccpContent := []string{
		"--ccp",
		"CCP",
		"Central Credential Provider",
		"connect --ccp",
	}

	for _, content := range ccpContent {
		if !strings.Contains(usage, content) {
			t.Errorf("Usage() should contain CCP-related content %q", content)
		}
	}
}

func TestConnectCommand_Execute_CCPNotConfigured(t *testing.T) {
	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	// Try to connect with --ccp when CCP is not configured
	err := cmd.Execute(execCtx, []string{"--ccp"})
	if err == nil {
		t.Error("Execute() with --ccp but no CCP config should return error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("Error should mention CCP not configured, got: %v", err)
	}
}

func TestConnectCommand_Execute_CCPDisabled(t *testing.T) {
	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	// Configure CCP but leave it disabled
	execCtx.Config.CCP = &config.CCPConfig{
		Enabled: false,
		AppID:   "TestApp",
		Safe:    "TestSafe",
	}

	err := cmd.Execute(execCtx, []string{"--ccp"})
	if err == nil {
		t.Error("Execute() with --ccp but CCP disabled should return error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("Error should mention CCP not configured, got: %v", err)
	}
}

func TestConnectCommand_Execute_CCPMissingAppID(t *testing.T) {
	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	// Configure CCP without AppID
	execCtx.Config.CCP = &config.CCPConfig{
		Enabled: true,
		AppID:   "",
		Safe:    "TestSafe",
	}

	err := cmd.Execute(execCtx, []string{"--ccp"})
	if err == nil {
		t.Error("Execute() with --ccp but missing AppID should return error")
	}
}

func TestConnectCommand_Execute_CCPMissingSafe(t *testing.T) {
	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	// Configure CCP without Safe
	execCtx.Config.CCP = &config.CCPConfig{
		Enabled: true,
		AppID:   "TestApp",
		Safe:    "",
	}

	err := cmd.Execute(execCtx, []string{"--ccp"})
	if err == nil {
		t.Error("Execute() with --ccp but missing Safe should return error")
	}
}

func TestConnectCommand_Execute_CCPNoCCPURL(t *testing.T) {
	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	// Configure CCP without URL and no default server
	execCtx.Config.CCP = &config.CCPConfig{
		Enabled: true,
		AppID:   "TestApp",
		Safe:    "TestSafe",
		CCPURL:  "",
	}
	execCtx.Config.DefaultServer = ""

	err := cmd.Execute(execCtx, []string{"--ccp"})
	if err == nil {
		t.Error("Execute() with --ccp but no URL should return error")
	}
	if !strings.Contains(err.Error(), "URL") {
		t.Errorf("Error should mention missing URL, got: %v", err)
	}
}

func TestConnectCommand_Execute_CCPWithMockServer(t *testing.T) {
	ccpCalled := false

	// Create mock server that handles both CCP and auth endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/AIMWebService/api/Accounts") {
			// CCP request
			ccpCalled = true

			// Verify required parameters
			query := r.URL.Query()
			if query.Get("AppID") != "TestApp" {
				t.Errorf("AppID = %v, want TestApp", query.Get("AppID"))
			}
			if query.Get("Safe") != "TestSafe" {
				t.Errorf("Safe = %v, want TestSafe", query.Get("Safe"))
			}

			// Return mock credential
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Content":  "testpassword",
				"UserName": "testuser",
				"Safe":     "TestSafe",
			})
			return
		}

		// Auth request - return unauthorized to simulate auth failure
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ErrorCode": "PASWS001E",
			"ErrorMsg":  "Authentication failed",
		})
	}))
	defer server.Close()

	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	// Configure CCP to use mock server
	execCtx.Config.CCP = &config.CCPConfig{
		Enabled: true,
		AppID:   "TestApp",
		Safe:    "TestSafe",
		CCPURL:  server.URL,
	}
	execCtx.Config.DefaultServer = server.URL
	execCtx.Config.InsecureSSL = true

	// This will fail at authentication stage (expected) but should pass CCP retrieval
	err := cmd.Execute(execCtx, []string{"--ccp"})

	// Verify CCP was called
	if !ccpCalled {
		t.Error("CCP endpoint was not called")
	}

	// The error should be about authentication, not about CCP configuration
	if err != nil && strings.Contains(err.Error(), "not configured") {
		t.Errorf("Error should not be about CCP configuration: %v", err)
	}
	// We expect authentication to fail with a mock server
	if err == nil {
		t.Error("Expected authentication error with mock server")
	}
}

func TestConnectCommand_Execute_CCPWithServerURL(t *testing.T) {
	// Test that server URL can be provided along with --ccp
	ccpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Content":  "testpassword",
			"UserName": "testuser",
		})
	}))
	defer ccpServer.Close()

	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	execCtx.Config.CCP = &config.CCPConfig{
		Enabled: true,
		AppID:   "TestApp",
		Safe:    "TestSafe",
		CCPURL:  ccpServer.URL,
	}
	execCtx.Config.InsecureSSL = true

	// Provide server URL with --ccp flag
	err := cmd.Execute(execCtx, []string{ccpServer.URL, "--ccp"})

	// Should get past CCP retrieval - error should be about auth, not CCP config
	if err != nil && strings.Contains(err.Error(), "not configured") {
		t.Errorf("Should not get CCP configuration error when properly configured: %v", err)
	}
	// We expect authentication to fail, which is fine for this test
}

func TestConnectCommand_Execute_CCPUsesConfigAuthMethod(t *testing.T) {
	ccpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Content":  "testpassword",
			"UserName": "testuser",
		})
	}))
	defer ccpServer.Close()

	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	execCtx.Config.CCP = &config.CCPConfig{
		Enabled:    true,
		AppID:      "TestApp",
		Safe:       "TestSafe",
		CCPURL:     ccpServer.URL,
		AuthMethod: "ldap", // Specify auth method in CCP config
	}
	execCtx.Config.DefaultServer = ccpServer.URL
	execCtx.Config.InsecureSSL = true

	// Execute - we're mainly testing that CCP credential retrieval and config works
	err := cmd.Execute(execCtx, []string{"--ccp"})

	// We expect an auth error (not CCP config error)
	if err != nil && strings.Contains(err.Error(), "not configured") {
		t.Errorf("Should not get CCP configuration error: %v", err)
	}
}

func TestConnectCommand_Execute_AlreadyConnected(t *testing.T) {
	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	// Create a mock session that reports as valid
	// Note: This tests the connection logic, not actual session validity
	// In real tests, you'd mock the session properly

	// For now, test without session (should work)
	err := cmd.Execute(execCtx, []string{"https://example.com", "--ccp"})

	// Should fail on CCP not configured, not on "already connected"
	if err != nil && strings.Contains(err.Error(), "already connected") {
		t.Error("Should not report already connected when session is nil")
	}
}

func TestDisconnectCommand_Name(t *testing.T) {
	cmd := &DisconnectCommand{}
	if cmd.Name() != "disconnect" {
		t.Errorf("Name() = %v, want disconnect", cmd.Name())
	}
}

func TestDisconnectCommand_Description(t *testing.T) {
	cmd := &DisconnectCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestStatusCommand_Name(t *testing.T) {
	cmd := &StatusCommand{}
	if cmd.Name() != "status" {
		t.Errorf("Name() = %v, want status", cmd.Name())
	}
}

func TestStatusCommand_Description(t *testing.T) {
	cmd := &StatusCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestConnectCommand_Execute_InsecureFlag(t *testing.T) {
	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	// Configure CCP with a server that will fail (but not on TLS)
	execCtx.Config.CCP = &config.CCPConfig{
		Enabled: true,
		AppID:   "TestApp",
		Safe:    "TestSafe",
		CCPURL:  "https://localhost:12345", // Will fail to connect
	}
	execCtx.Config.DefaultServer = "https://localhost:12345"

	// Test with --insecure flag - should fail on connection, not TLS cert errors
	err := cmd.Execute(execCtx, []string{"--ccp", "--insecure"})

	// We expect a connection error, not a certificate error
	if err != nil && strings.Contains(err.Error(), "x509") {
		t.Error("--insecure flag should skip certificate verification")
	}
}

func TestConnectCommand_Execute_AuthMethodFlag(t *testing.T) {
	ccpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Content":  "testpassword",
			"UserName": "testuser",
		})
	}))
	defer ccpServer.Close()

	cmd := &ConnectCommand{}
	execCtx := createTestSessionExecutionContext(t)

	execCtx.Config.CCP = &config.CCPConfig{
		Enabled: true,
		AppID:   "TestApp",
		Safe:    "TestSafe",
		CCPURL:  ccpServer.URL,
	}
	execCtx.Config.DefaultServer = ccpServer.URL
	execCtx.Config.InsecureSSL = true

	// Test with explicit auth method (should override CCP config)
	err := cmd.Execute(execCtx, []string{"--ccp", "--auth=radius"})

	// We verify flag parsing works - expect auth error, not config error
	if err != nil && strings.Contains(err.Error(), "not configured") {
		t.Errorf("Should not get CCP configuration error: %v", err)
	}
}

func TestConnectCommand_FlagParsing(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "CCP flag only",
			args: []string{"--ccp"},
		},
		{
			name: "CCP with server URL",
			args: []string{"https://example.com", "--ccp"},
		},
		{
			name: "CCP with insecure",
			args: []string{"--ccp", "--insecure"},
		},
		{
			name: "CCP with auth method",
			args: []string{"--ccp", "--auth=ldap"},
		},
		{
			name: "all flags combined",
			args: []string{"https://example.com", "--ccp", "--insecure", "--auth=ldap"},
		},
		{
			name: "flags in different order",
			args: []string{"--auth=ldap", "--ccp", "https://example.com", "--insecure"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ConnectCommand{}
			execCtx := createTestSessionExecutionContext(t)

			// This tests that flag parsing doesn't panic
			// Will error due to no CCP config, which is expected
			_ = cmd.Execute(execCtx, tt.args)
		})
	}
}

// Helper function to create a test execution context for session tests
func createTestSessionExecutionContext(t *testing.T) *ExecutionContext {
	t.Helper()

	// Create temp directory for config
	tmpDir, err := os.MkdirTemp("", "pasctl-session-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Override HOME for config file operations
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.RemoveAll(tmpDir)
	})

	cfg := config.Default()
	formatter := output.NewFormatter(output.FormatTable)

	return &ExecutionContext{
		Ctx:       context.Background(),
		Session:   nil,
		Config:    cfg,
		Formatter: formatter,
	}
}

// Test that ConnectCommand implements Command interface
func TestConnectCommand_ImplementsCommand(t *testing.T) {
	var cmd Command = &ConnectCommand{}

	if cmd.Name() == "" {
		t.Error("Name() should not return empty string")
	}
	if cmd.Description() == "" {
		t.Error("Description() should not return empty string")
	}
	if cmd.Usage() == "" {
		t.Error("Usage() should not return empty string")
	}
}

// Test that DisconnectCommand implements Command interface
func TestDisconnectCommand_ImplementsCommand(t *testing.T) {
	var cmd Command = &DisconnectCommand{}

	if cmd.Name() == "" {
		t.Error("Name() should not return empty string")
	}
	if cmd.Description() == "" {
		t.Error("Description() should not return empty string")
	}
	if cmd.Usage() == "" {
		t.Error("Usage() should not return empty string")
	}
}

// Test that StatusCommand implements Command interface
func TestStatusCommand_ImplementsCommand(t *testing.T) {
	var cmd Command = &StatusCommand{}

	if cmd.Name() == "" {
		t.Error("Name() should not return empty string")
	}
	if cmd.Description() == "" {
		t.Error("Description() should not return empty string")
	}
	if cmd.Usage() == "" {
		t.Error("Usage() should not return empty string")
	}
}

// Benchmark tests
func BenchmarkConnectCommand_FlagParsing(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "pasctl-bench-*")
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cmd := &ConnectCommand{}
	args := []string{"https://example.com", "--ccp", "--insecure", "--auth=ldap"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := config.Default()
		execCtx := &ExecutionContext{
			Ctx:       context.Background(),
			Config:    cfg,
			Formatter: output.NewFormatter(output.FormatTable),
		}
		// Will error, but we're benchmarking flag parsing
		cmd.Execute(execCtx, args)
	}
}
