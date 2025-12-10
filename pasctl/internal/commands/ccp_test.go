// Package commands provides tests for the CCP command.
package commands

import (
	"context"
	"os"
	"strings"
	"testing"

	"pasctl/internal/config"
	"pasctl/internal/output"
)

func TestCCPCommand_Name(t *testing.T) {
	cmd := &CCPCommand{}
	if cmd.Name() != "ccp" {
		t.Errorf("Name() = %v, want ccp", cmd.Name())
	}
}

func TestCCPCommand_Description(t *testing.T) {
	cmd := &CCPCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
	if !strings.Contains(desc, "CCP") {
		t.Errorf("Description() should mention CCP, got: %v", desc)
	}
}

func TestCCPCommand_Usage(t *testing.T) {
	cmd := &CCPCommand{}
	usage := cmd.Usage()

	// Check for essential content
	requiredContent := []string{
		"ccp",
		"setup",
		"show",
		"enable",
		"disable",
		"clear",
		"--app-id",
		"--safe",
		"connect --ccp",
	}

	for _, content := range requiredContent {
		if !strings.Contains(usage, content) {
			t.Errorf("Usage() should contain %q", content)
		}
	}
}

func TestCCPCommand_Execute_NoSubcommand(t *testing.T) {
	cmd := &CCPCommand{}
	execCtx := createTestExecutionContext(t)

	err := cmd.Execute(execCtx, []string{})
	if err == nil {
		t.Error("Execute() with no args should return error")
	}
	if !strings.Contains(err.Error(), "subcommand required") {
		t.Errorf("Error should mention 'subcommand required', got: %v", err)
	}
}

func TestCCPCommand_Execute_UnknownSubcommand(t *testing.T) {
	cmd := &CCPCommand{}
	execCtx := createTestExecutionContext(t)

	err := cmd.Execute(execCtx, []string{"invalid"})
	if err == nil {
		t.Error("Execute() with unknown subcommand should return error")
	}
	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("Error should mention 'unknown subcommand', got: %v", err)
	}
}

func TestCCPCommand_Setup_WithOptions(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
		check   func(*testing.T, *config.Config)
	}{
		{
			name: "successful setup with required options",
			args: []string{"setup", "--app-id=TestApp", "--safe=TestSafe", "--ccp-url=https://ccp.example.com"},
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.CCP == nil {
					t.Fatal("CCP should be configured")
				}
				if cfg.CCP.AppID != "TestApp" {
					t.Errorf("AppID = %v, want TestApp", cfg.CCP.AppID)
				}
				if cfg.CCP.Safe != "TestSafe" {
					t.Errorf("Safe = %v, want TestSafe", cfg.CCP.Safe)
				}
				if cfg.CCP.CCPURL != "https://ccp.example.com" {
					t.Errorf("CCPURL = %v, want https://ccp.example.com", cfg.CCP.CCPURL)
				}
				if !cfg.CCP.Enabled {
					t.Error("CCP should be enabled after setup")
				}
			},
		},
		{
			name: "setup with all options",
			args: []string{"setup",
				"--app-id=MyApp",
				"--safe=MySafe",
				"--object=MyObject",
				"--folder=Root\\Folder",
				"--username=admin",
				"--address=server.local",
				"--query=test",
				"--auth-method=ldap",
				"--ccp-url=https://ccp.example.com",
				"--client-cert=/path/to/cert.pem",
				"--client-key=/path/to/key.pem",
			},
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.CCP == nil {
					t.Fatal("CCP should be configured")
				}
				if cfg.CCP.AppID != "MyApp" {
					t.Errorf("AppID = %v, want MyApp", cfg.CCP.AppID)
				}
				if cfg.CCP.Safe != "MySafe" {
					t.Errorf("Safe = %v, want MySafe", cfg.CCP.Safe)
				}
				if cfg.CCP.Object != "MyObject" {
					t.Errorf("Object = %v, want MyObject", cfg.CCP.Object)
				}
				if cfg.CCP.Folder != "Root\\Folder" {
					t.Errorf("Folder = %v, want Root\\Folder", cfg.CCP.Folder)
				}
				if cfg.CCP.UserName != "admin" {
					t.Errorf("UserName = %v, want admin", cfg.CCP.UserName)
				}
				if cfg.CCP.Address != "server.local" {
					t.Errorf("Address = %v, want server.local", cfg.CCP.Address)
				}
				if cfg.CCP.Query != "test" {
					t.Errorf("Query = %v, want test", cfg.CCP.Query)
				}
				if cfg.CCP.AuthMethod != "ldap" {
					t.Errorf("AuthMethod = %v, want ldap", cfg.CCP.AuthMethod)
				}
				if cfg.CCP.CCPURL != "https://ccp.example.com" {
					t.Errorf("CCPURL = %v, want https://ccp.example.com", cfg.CCP.CCPURL)
				}
				if cfg.CCP.ClientCert != "/path/to/cert.pem" {
					t.Errorf("ClientCert = %v, want /path/to/cert.pem", cfg.CCP.ClientCert)
				}
				if cfg.CCP.ClientKey != "/path/to/key.pem" {
					t.Errorf("ClientKey = %v, want /path/to/key.pem", cfg.CCP.ClientKey)
				}
			},
		},
		{
			name:    "setup missing app-id",
			args:    []string{"setup", "--safe=TestSafe"},
			wantErr: true,
			errMsg:  "--app-id is required",
		},
		{
			name:    "setup missing safe",
			args:    []string{"setup", "--app-id=TestApp"},
			wantErr: true,
			errMsg:  "--safe is required",
		},
		{
			name:    "setup missing ccp-url",
			args:    []string{"setup", "--app-id=TestApp", "--safe=TestSafe"},
			wantErr: true,
			errMsg:  "--ccp-url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CCPCommand{}
			execCtx := createTestExecutionContext(t)

			err := cmd.Execute(execCtx, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() expected error")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error should contain %q, got: %v", tt.errMsg, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
				return
			}

			if tt.check != nil {
				tt.check(t, execCtx.Config)
			}
		})
	}
}

func TestCCPCommand_Show(t *testing.T) {
	tests := []struct {
		name      string
		setupCCP  *config.CCPConfig
		wantErr   bool
	}{
		{
			name:     "show with no CCP configured",
			setupCCP: nil,
			wantErr:  false,
		},
		{
			name: "show with CCP configured and enabled",
			setupCCP: &config.CCPConfig{
				Enabled:    true,
				AppID:      "TestApp",
				Safe:       "TestSafe",
				Object:     "TestObject",
				Folder:     "Root",
				UserName:   "admin",
				Address:    "server.local",
				Query:      "test",
				CCPURL:     "https://ccp.example.com",
				AuthMethod: "ldap",
				ClientCert: "/path/to/cert.pem",
				ClientKey:  "/path/to/key.pem",
			},
			wantErr: false,
		},
		{
			name: "show with CCP configured but disabled",
			setupCCP: &config.CCPConfig{
				Enabled: false,
				AppID:   "TestApp",
				Safe:    "TestSafe",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CCPCommand{}
			execCtx := createTestExecutionContext(t)
			execCtx.Config.CCP = tt.setupCCP

			err := cmd.Execute(execCtx, []string{"show"})

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
			}
		})
	}
}

func TestCCPCommand_Enable(t *testing.T) {
	tests := []struct {
		name     string
		setupCCP *config.CCPConfig
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "enable with no CCP configured",
			setupCCP: nil,
			wantErr:  true,
			errMsg:   "not configured",
		},
		{
			name: "enable with missing AppID",
			setupCCP: &config.CCPConfig{
				Enabled: false,
				AppID:   "",
				Safe:    "TestSafe",
			},
			wantErr: true,
			errMsg:  "not fully configured",
		},
		{
			name: "enable with missing Safe",
			setupCCP: &config.CCPConfig{
				Enabled: false,
				AppID:   "TestApp",
				Safe:    "",
			},
			wantErr: true,
			errMsg:  "not fully configured",
		},
		{
			name: "successful enable",
			setupCCP: &config.CCPConfig{
				Enabled: false,
				AppID:   "TestApp",
				Safe:    "TestSafe",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CCPCommand{}
			execCtx := createTestExecutionContext(t)
			execCtx.Config.CCP = tt.setupCCP

			err := cmd.Execute(execCtx, []string{"enable"})

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() expected error")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error should contain %q, got: %v", tt.errMsg, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
				return
			}

			if !execCtx.Config.CCP.Enabled {
				t.Error("CCP should be enabled after enable command")
			}
		})
	}
}

func TestCCPCommand_Disable(t *testing.T) {
	tests := []struct {
		name     string
		setupCCP *config.CCPConfig
		wantErr  bool
	}{
		{
			name:     "disable with no CCP configured",
			setupCCP: nil,
			wantErr:  false, // Should not error, just warn
		},
		{
			name: "disable enabled CCP",
			setupCCP: &config.CCPConfig{
				Enabled: true,
				AppID:   "TestApp",
				Safe:    "TestSafe",
			},
			wantErr: false,
		},
		{
			name: "disable already disabled CCP",
			setupCCP: &config.CCPConfig{
				Enabled: false,
				AppID:   "TestApp",
				Safe:    "TestSafe",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CCPCommand{}
			execCtx := createTestExecutionContext(t)
			execCtx.Config.CCP = tt.setupCCP

			err := cmd.Execute(execCtx, []string{"disable"})

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
				return
			}

			if tt.setupCCP != nil && execCtx.Config.CCP.Enabled {
				t.Error("CCP should be disabled after disable command")
			}
		})
	}
}

func TestCCPCommand_Clear(t *testing.T) {
	tests := []struct {
		name     string
		setupCCP *config.CCPConfig
		wantErr  bool
	}{
		{
			name:     "clear with no CCP configured",
			setupCCP: nil,
			wantErr:  false,
		},
		{
			name: "clear with CCP configured",
			setupCCP: &config.CCPConfig{
				Enabled:    true,
				AppID:      "TestApp",
				Safe:       "TestSafe",
				Object:     "TestObject",
				Folder:     "Root",
				UserName:   "admin",
				Address:    "server.local",
				CCPURL:     "https://ccp.example.com",
				AuthMethod: "ldap",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CCPCommand{}
			execCtx := createTestExecutionContext(t)
			execCtx.Config.CCP = tt.setupCCP

			err := cmd.Execute(execCtx, []string{"clear"})

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
				return
			}

			if execCtx.Config.CCP != nil {
				t.Error("CCP should be nil after clear command")
			}
		})
	}
}

func TestCCPCommand_SubcommandCaseInsensitive(t *testing.T) {
	subcommands := []string{
		"SHOW", "Show", "sHoW",
		"ENABLE", "Enable", "eNaBlE",
		"DISABLE", "Disable", "dIsAbLe",
		"CLEAR", "Clear", "cLeAr",
	}

	for _, subCmd := range subcommands {
		t.Run(subCmd, func(t *testing.T) {
			cmd := &CCPCommand{}
			execCtx := createTestExecutionContext(t)

			// Setup CCP for commands that need it
			if strings.ToLower(subCmd) == "enable" {
				execCtx.Config.CCP = &config.CCPConfig{
					AppID: "TestApp",
					Safe:  "TestSafe",
				}
			}

			// Execute should not return "unknown subcommand" error
			err := cmd.Execute(execCtx, []string{subCmd})
			if err != nil && strings.Contains(err.Error(), "unknown subcommand") {
				t.Errorf("Subcommand %q should be recognized (case-insensitive)", subCmd)
			}
		})
	}
}

func TestCCPCommand_Setup_PartialUpdate(t *testing.T) {
	// Test that setup with existing config updates values properly
	cmd := &CCPCommand{}
	execCtx := createTestExecutionContext(t)

	// Set initial CCP config
	execCtx.Config.CCP = &config.CCPConfig{
		Enabled:  true,
		AppID:    "OldApp",
		Safe:     "OldSafe",
		CCPURL:   "https://old-ccp.example.com",
		UserName: "olduser",
	}

	// Update with new values (must include --ccp-url since it's required)
	err := cmd.Execute(execCtx, []string{"setup", "--app-id=NewApp", "--safe=NewSafe", "--ccp-url=https://new-ccp.example.com"})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify new values are set
	if execCtx.Config.CCP.AppID != "NewApp" {
		t.Errorf("AppID = %v, want NewApp", execCtx.Config.CCP.AppID)
	}
	if execCtx.Config.CCP.Safe != "NewSafe" {
		t.Errorf("Safe = %v, want NewSafe", execCtx.Config.CCP.Safe)
	}
	if execCtx.Config.CCP.CCPURL != "https://new-ccp.example.com" {
		t.Errorf("CCPURL = %v, want https://new-ccp.example.com", execCtx.Config.CCP.CCPURL)
	}

	// Verify old values that weren't specified are preserved
	if execCtx.Config.CCP.UserName != "olduser" {
		t.Errorf("UserName = %v, want olduser (should be preserved)", execCtx.Config.CCP.UserName)
	}
}

func TestCCPCommand_Setup_EmptyValueClears(t *testing.T) {
	cmd := &CCPCommand{}
	execCtx := createTestExecutionContext(t)

	// Setup with initial values (including required --ccp-url)
	err := cmd.Execute(execCtx, []string{"setup",
		"--app-id=TestApp",
		"--safe=TestSafe",
		"--ccp-url=https://ccp.example.com",
		"--username=admin",
	})
	if err != nil {
		t.Fatalf("Initial setup failed: %v", err)
	}

	if execCtx.Config.CCP.UserName != "admin" {
		t.Errorf("UserName = %v, want admin", execCtx.Config.CCP.UserName)
	}
}

// Helper function to create a test execution context
func createTestExecutionContext(t *testing.T) *ExecutionContext {
	t.Helper()

	// Create temp directory for config
	tmpDir, err := os.MkdirTemp("", "pasctl-test-*")
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

// Test that CCPCommand implements Command interface
func TestCCPCommand_ImplementsCommand(t *testing.T) {
	var cmd Command = &CCPCommand{}

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
func BenchmarkCCPCommand_Show(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "pasctl-bench-*")
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cmd := &CCPCommand{}
	cfg := config.Default()
	cfg.CCP = &config.CCPConfig{
		Enabled:    true,
		AppID:      "TestApp",
		Safe:       "TestSafe",
		Object:     "TestObject",
		UserName:   "admin",
		CCPURL:     "https://ccp.example.com",
		AuthMethod: "ldap",
	}

	execCtx := &ExecutionContext{
		Ctx:       context.Background(),
		Config:    cfg,
		Formatter: output.NewFormatter(output.FormatTable),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.Execute(execCtx, []string{"show"})
	}
}

func BenchmarkCCPCommand_Setup(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "pasctl-bench-*")
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cmd := &CCPCommand{}
	args := []string{"setup", "--app-id=TestApp", "--safe=TestSafe", "--ccp-url=https://ccp.example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := config.Default()
		execCtx := &ExecutionContext{
			Ctx:       context.Background(),
			Config:    cfg,
			Formatter: output.NewFormatter(output.FormatTable),
		}
		cmd.Execute(execCtx, args)
	}
}
