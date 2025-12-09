// Package config provides tests for pasctl configuration management.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.DefaultAuthType != "cyberark" {
		t.Errorf("Default().DefaultAuthType = %v, want cyberark", cfg.DefaultAuthType)
	}
	if cfg.OutputFormat != "table" {
		t.Errorf("Default().OutputFormat = %v, want table", cfg.OutputFormat)
	}
	if cfg.HistorySize != 1000 {
		t.Errorf("Default().HistorySize = %v, want 1000", cfg.HistorySize)
	}
	if cfg.InsecureSSL != false {
		t.Error("Default().InsecureSSL should be false")
	}
	if cfg.Timeout != 30 {
		t.Errorf("Default().Timeout = %v, want 30", cfg.Timeout)
	}
	if cfg.CCP != nil {
		t.Error("Default().CCP should be nil")
	}
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		DefaultServer:   "https://cyberark.example.com",
		DefaultAuthType: "ldap",
		OutputFormat:    "json",
		HistorySize:     500,
		InsecureSSL:     true,
		Timeout:         60,
	}

	if cfg.DefaultServer != "https://cyberark.example.com" {
		t.Errorf("DefaultServer = %v, want https://cyberark.example.com", cfg.DefaultServer)
	}
	if cfg.DefaultAuthType != "ldap" {
		t.Errorf("DefaultAuthType = %v, want ldap", cfg.DefaultAuthType)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %v, want json", cfg.OutputFormat)
	}
	if cfg.HistorySize != 500 {
		t.Errorf("HistorySize = %v, want 500", cfg.HistorySize)
	}
	if !cfg.InsecureSSL {
		t.Error("InsecureSSL should be true")
	}
	if cfg.Timeout != 60 {
		t.Errorf("Timeout = %v, want 60", cfg.Timeout)
	}
}

func TestCCPConfig_Struct(t *testing.T) {
	ccpCfg := CCPConfig{
		Enabled:    true,
		CCPURL:     "https://ccp.example.com",
		AppID:      "MyApp",
		Safe:       "MySafe",
		Object:     "MyAccount",
		Folder:     "Root\\Folder",
		UserName:   "admin",
		Address:    "server.example.com",
		Query:      "admin",
		AuthMethod: "ldap",
		ClientCert: "/path/to/cert.pem",
		ClientKey:  "/path/to/key.pem",
	}

	if !ccpCfg.Enabled {
		t.Error("Enabled should be true")
	}
	if ccpCfg.CCPURL != "https://ccp.example.com" {
		t.Errorf("CCPURL = %v, want https://ccp.example.com", ccpCfg.CCPURL)
	}
	if ccpCfg.AppID != "MyApp" {
		t.Errorf("AppID = %v, want MyApp", ccpCfg.AppID)
	}
	if ccpCfg.Safe != "MySafe" {
		t.Errorf("Safe = %v, want MySafe", ccpCfg.Safe)
	}
	if ccpCfg.Object != "MyAccount" {
		t.Errorf("Object = %v, want MyAccount", ccpCfg.Object)
	}
	if ccpCfg.Folder != "Root\\Folder" {
		t.Errorf("Folder = %v, want Root\\Folder", ccpCfg.Folder)
	}
	if ccpCfg.UserName != "admin" {
		t.Errorf("UserName = %v, want admin", ccpCfg.UserName)
	}
	if ccpCfg.Address != "server.example.com" {
		t.Errorf("Address = %v, want server.example.com", ccpCfg.Address)
	}
	if ccpCfg.Query != "admin" {
		t.Errorf("Query = %v, want admin", ccpCfg.Query)
	}
	if ccpCfg.AuthMethod != "ldap" {
		t.Errorf("AuthMethod = %v, want ldap", ccpCfg.AuthMethod)
	}
	if ccpCfg.ClientCert != "/path/to/cert.pem" {
		t.Errorf("ClientCert = %v, want /path/to/cert.pem", ccpCfg.ClientCert)
	}
	if ccpCfg.ClientKey != "/path/to/key.pem" {
		t.Errorf("ClientKey = %v, want /path/to/key.pem", ccpCfg.ClientKey)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		wantFormat     string
		wantHistory    int
		wantTimeout    int
	}{
		{
			name: "valid config unchanged",
			config: &Config{
				OutputFormat: "table",
				HistorySize:  1000,
				Timeout:      30,
			},
			wantFormat:  "table",
			wantHistory: 1000,
			wantTimeout: 30,
		},
		{
			name: "json format unchanged",
			config: &Config{
				OutputFormat: "json",
				HistorySize:  1000,
				Timeout:      30,
			},
			wantFormat:  "json",
			wantHistory: 1000,
			wantTimeout: 30,
		},
		{
			name: "yaml format unchanged",
			config: &Config{
				OutputFormat: "yaml",
				HistorySize:  1000,
				Timeout:      30,
			},
			wantFormat:  "yaml",
			wantHistory: 1000,
			wantTimeout: 30,
		},
		{
			name: "invalid format corrected to table",
			config: &Config{
				OutputFormat: "invalid",
				HistorySize:  1000,
				Timeout:      30,
			},
			wantFormat:  "table",
			wantHistory: 1000,
			wantTimeout: 30,
		},
		{
			name: "zero history corrected",
			config: &Config{
				OutputFormat: "table",
				HistorySize:  0,
				Timeout:      30,
			},
			wantFormat:  "table",
			wantHistory: 1000,
			wantTimeout: 30,
		},
		{
			name: "negative history corrected",
			config: &Config{
				OutputFormat: "table",
				HistorySize:  -100,
				Timeout:      30,
			},
			wantFormat:  "table",
			wantHistory: 1000,
			wantTimeout: 30,
		},
		{
			name: "zero timeout corrected",
			config: &Config{
				OutputFormat: "table",
				HistorySize:  1000,
				Timeout:      0,
			},
			wantFormat:  "table",
			wantHistory: 1000,
			wantTimeout: 30,
		},
		{
			name: "negative timeout corrected",
			config: &Config{
				OutputFormat: "table",
				HistorySize:  1000,
				Timeout:      -10,
			},
			wantFormat:  "table",
			wantHistory: 1000,
			wantTimeout: 30,
		},
		{
			name: "all invalid values corrected",
			config: &Config{
				OutputFormat: "xml",
				HistorySize:  -1,
				Timeout:      -5,
			},
			wantFormat:  "table",
			wantHistory: 1000,
			wantTimeout: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}

			if tt.config.OutputFormat != tt.wantFormat {
				t.Errorf("After Validate(), OutputFormat = %v, want %v", tt.config.OutputFormat, tt.wantFormat)
			}
			if tt.config.HistorySize != tt.wantHistory {
				t.Errorf("After Validate(), HistorySize = %v, want %v", tt.config.HistorySize, tt.wantHistory)
			}
			if tt.config.Timeout != tt.wantTimeout {
				t.Errorf("After Validate(), Timeout = %v, want %v", tt.config.Timeout, tt.wantTimeout)
			}
		})
	}
}

func TestConfig_IsCCPEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want bool
	}{
		{
			name: "CCP nil",
			cfg:  &Config{},
			want: false,
		},
		{
			name: "CCP disabled",
			cfg: &Config{
				CCP: &CCPConfig{
					Enabled: false,
					AppID:   "MyApp",
					Safe:    "MySafe",
				},
			},
			want: false,
		},
		{
			name: "CCP enabled but missing AppID",
			cfg: &Config{
				CCP: &CCPConfig{
					Enabled: true,
					AppID:   "",
					Safe:    "MySafe",
				},
			},
			want: false,
		},
		{
			name: "CCP enabled but missing Safe",
			cfg: &Config{
				CCP: &CCPConfig{
					Enabled: true,
					AppID:   "MyApp",
					Safe:    "",
				},
			},
			want: false,
		},
		{
			name: "CCP fully configured and enabled",
			cfg: &Config{
				CCP: &CCPConfig{
					Enabled: true,
					AppID:   "MyApp",
					Safe:    "MySafe",
				},
			},
			want: true,
		},
		{
			name: "CCP with all optional fields",
			cfg: &Config{
				CCP: &CCPConfig{
					Enabled:    true,
					AppID:      "MyApp",
					Safe:       "MySafe",
					Object:     "Account",
					Folder:     "Root",
					UserName:   "admin",
					Address:    "server",
					AuthMethod: "ldap",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.IsCCPEnabled()
			if result != tt.want {
				t.Errorf("IsCCPEnabled() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestConfig_GetCCPURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "CCP nil returns DefaultServer",
			cfg: &Config{
				DefaultServer: "https://default.example.com",
			},
			want: "https://default.example.com",
		},
		{
			name: "CCP URL empty returns DefaultServer",
			cfg: &Config{
				DefaultServer: "https://default.example.com",
				CCP: &CCPConfig{
					CCPURL: "",
				},
			},
			want: "https://default.example.com",
		},
		{
			name: "CCP URL specified returns CCP URL",
			cfg: &Config{
				DefaultServer: "https://default.example.com",
				CCP: &CCPConfig{
					CCPURL: "https://ccp.example.com",
				},
			},
			want: "https://ccp.example.com",
		},
		{
			name: "Both empty returns empty",
			cfg:  &Config{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetCCPURL()
			if result != tt.want {
				t.Errorf("GetCCPURL() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestConfig_GetCCPAuthMethod(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "CCP nil returns DefaultAuthType",
			cfg: &Config{
				DefaultAuthType: "ldap",
			},
			want: "ldap",
		},
		{
			name: "CCP AuthMethod empty returns DefaultAuthType",
			cfg: &Config{
				DefaultAuthType: "ldap",
				CCP: &CCPConfig{
					AuthMethod: "",
				},
			},
			want: "ldap",
		},
		{
			name: "CCP AuthMethod specified returns CCP AuthMethod",
			cfg: &Config{
				DefaultAuthType: "ldap",
				CCP: &CCPConfig{
					AuthMethod: "cyberark",
				},
			},
			want: "cyberark",
		},
		{
			name: "Both empty returns empty",
			cfg:  &Config{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetCCPAuthMethod()
			if result != tt.want {
				t.Errorf("GetCCPAuthMethod() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pasctl-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create test config
	cfg := &Config{
		DefaultServer:   "https://cyberark.example.com",
		DefaultAuthType: "ldap",
		OutputFormat:    "json",
		HistorySize:     500,
		InsecureSSL:     true,
		Timeout:         60,
		CCP: &CCPConfig{
			Enabled:    true,
			CCPURL:     "https://ccp.example.com",
			AppID:      "TestApp",
			Safe:       "TestSafe",
			Object:     "TestAccount",
			Folder:     "Root",
			UserName:   "admin",
			Address:    "server.local",
			Query:      "test",
			AuthMethod: "cyberark",
			ClientCert: "/path/to/cert.pem",
			ClientKey:  "/path/to/key.pem",
		},
	}

	// Save config
	err = cfg.Save()
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded config matches saved config
	if loaded.DefaultServer != cfg.DefaultServer {
		t.Errorf("Loaded DefaultServer = %v, want %v", loaded.DefaultServer, cfg.DefaultServer)
	}
	if loaded.DefaultAuthType != cfg.DefaultAuthType {
		t.Errorf("Loaded DefaultAuthType = %v, want %v", loaded.DefaultAuthType, cfg.DefaultAuthType)
	}
	if loaded.OutputFormat != cfg.OutputFormat {
		t.Errorf("Loaded OutputFormat = %v, want %v", loaded.OutputFormat, cfg.OutputFormat)
	}
	if loaded.HistorySize != cfg.HistorySize {
		t.Errorf("Loaded HistorySize = %v, want %v", loaded.HistorySize, cfg.HistorySize)
	}
	if loaded.InsecureSSL != cfg.InsecureSSL {
		t.Errorf("Loaded InsecureSSL = %v, want %v", loaded.InsecureSSL, cfg.InsecureSSL)
	}
	if loaded.Timeout != cfg.Timeout {
		t.Errorf("Loaded Timeout = %v, want %v", loaded.Timeout, cfg.Timeout)
	}

	// Verify CCP config
	if loaded.CCP == nil {
		t.Fatal("Loaded CCP should not be nil")
	}
	if loaded.CCP.Enabled != cfg.CCP.Enabled {
		t.Errorf("Loaded CCP.Enabled = %v, want %v", loaded.CCP.Enabled, cfg.CCP.Enabled)
	}
	if loaded.CCP.CCPURL != cfg.CCP.CCPURL {
		t.Errorf("Loaded CCP.CCPURL = %v, want %v", loaded.CCP.CCPURL, cfg.CCP.CCPURL)
	}
	if loaded.CCP.AppID != cfg.CCP.AppID {
		t.Errorf("Loaded CCP.AppID = %v, want %v", loaded.CCP.AppID, cfg.CCP.AppID)
	}
	if loaded.CCP.Safe != cfg.CCP.Safe {
		t.Errorf("Loaded CCP.Safe = %v, want %v", loaded.CCP.Safe, cfg.CCP.Safe)
	}
	if loaded.CCP.Object != cfg.CCP.Object {
		t.Errorf("Loaded CCP.Object = %v, want %v", loaded.CCP.Object, cfg.CCP.Object)
	}
	if loaded.CCP.Folder != cfg.CCP.Folder {
		t.Errorf("Loaded CCP.Folder = %v, want %v", loaded.CCP.Folder, cfg.CCP.Folder)
	}
	if loaded.CCP.UserName != cfg.CCP.UserName {
		t.Errorf("Loaded CCP.UserName = %v, want %v", loaded.CCP.UserName, cfg.CCP.UserName)
	}
	if loaded.CCP.Address != cfg.CCP.Address {
		t.Errorf("Loaded CCP.Address = %v, want %v", loaded.CCP.Address, cfg.CCP.Address)
	}
	if loaded.CCP.Query != cfg.CCP.Query {
		t.Errorf("Loaded CCP.Query = %v, want %v", loaded.CCP.Query, cfg.CCP.Query)
	}
	if loaded.CCP.AuthMethod != cfg.CCP.AuthMethod {
		t.Errorf("Loaded CCP.AuthMethod = %v, want %v", loaded.CCP.AuthMethod, cfg.CCP.AuthMethod)
	}
	if loaded.CCP.ClientCert != cfg.CCP.ClientCert {
		t.Errorf("Loaded CCP.ClientCert = %v, want %v", loaded.CCP.ClientCert, cfg.CCP.ClientCert)
	}
	if loaded.CCP.ClientKey != cfg.CCP.ClientKey {
		t.Errorf("Loaded CCP.ClientKey = %v, want %v", loaded.CCP.ClientKey, cfg.CCP.ClientKey)
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pasctl-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Load should return default config when file doesn't exist
	cfg, err := Load()
	if err != nil {
		t.Errorf("Load() unexpected error for nonexistent file: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Should have default values
	defaultCfg := Default()
	if cfg.DefaultAuthType != defaultCfg.DefaultAuthType {
		t.Errorf("Loaded DefaultAuthType = %v, want default %v", cfg.DefaultAuthType, defaultCfg.DefaultAuthType)
	}
	if cfg.OutputFormat != defaultCfg.OutputFormat {
		t.Errorf("Loaded OutputFormat = %v, want default %v", cfg.OutputFormat, defaultCfg.OutputFormat)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pasctl-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".pasctl")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write invalid JSON
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Load should return default config with error
	cfg, err := Load()
	if err == nil {
		t.Error("Load() expected error for invalid JSON")
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config even with error")
	}

	// Should still have default values
	defaultCfg := Default()
	if cfg.DefaultAuthType != defaultCfg.DefaultAuthType {
		t.Errorf("On error, DefaultAuthType = %v, want default %v", cfg.DefaultAuthType, defaultCfg.DefaultAuthType)
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Errorf("ConfigDir() unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("ConfigDir() should return absolute path, got: %v", dir)
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Errorf("ConfigPath() unexpected error: %v", err)
	}
	if path == "" {
		t.Error("ConfigPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("ConfigPath() should return absolute path, got: %v", path)
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("ConfigPath() should end with config.json, got: %v", path)
	}
}

func TestHistoryPath(t *testing.T) {
	path, err := HistoryPath()
	if err != nil {
		t.Errorf("HistoryPath() unexpected error: %v", err)
	}
	if path == "" {
		t.Error("HistoryPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("HistoryPath() should return absolute path, got: %v", path)
	}
	if filepath.Base(path) != ".pasctl_history" {
		t.Errorf("HistoryPath() should end with .pasctl_history, got: %v", path)
	}
}

func TestConfig_JSONMarshaling(t *testing.T) {
	cfg := &Config{
		DefaultServer:   "https://cyberark.example.com",
		DefaultAuthType: "ldap",
		OutputFormat:    "json",
		HistorySize:     500,
		InsecureSSL:     true,
		Timeout:         60,
		CCP: &CCPConfig{
			Enabled:    true,
			CCPURL:     "https://ccp.example.com",
			AppID:      "TestApp",
			Safe:       "TestSafe",
			AuthMethod: "cyberark",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	// Unmarshal back
	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Verify roundtrip
	if decoded.DefaultServer != cfg.DefaultServer {
		t.Errorf("After roundtrip, DefaultServer = %v, want %v", decoded.DefaultServer, cfg.DefaultServer)
	}
	if decoded.CCP == nil {
		t.Fatal("After roundtrip, CCP should not be nil")
	}
	if decoded.CCP.AppID != cfg.CCP.AppID {
		t.Errorf("After roundtrip, CCP.AppID = %v, want %v", decoded.CCP.AppID, cfg.CCP.AppID)
	}
}

func TestConfig_JSONOmitEmpty(t *testing.T) {
	// Config with only some fields set
	cfg := &Config{
		DefaultServer: "https://cyberark.example.com",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	jsonStr := string(data)

	// Fields should be omitted if empty
	if containsField(jsonStr, "output_format") {
		t.Error("JSON should not contain output_format when empty")
	}
	if containsField(jsonStr, "history_size") {
		t.Error("JSON should not contain history_size when zero")
	}
	if containsField(jsonStr, "ccp") {
		t.Error("JSON should not contain ccp when nil")
	}
}

func TestCCPConfig_JSONOmitEmpty(t *testing.T) {
	cfg := &Config{
		CCP: &CCPConfig{
			Enabled: true,
			AppID:   "TestApp",
			Safe:    "TestSafe",
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	jsonStr := string(data)

	// Optional CCP fields should be omitted if empty
	if containsField(jsonStr, "ccp_url") {
		t.Error("JSON should not contain ccp_url when empty")
	}
	if containsField(jsonStr, "object") {
		t.Error("JSON should not contain object when empty")
	}
	if containsField(jsonStr, "folder") {
		t.Error("JSON should not contain folder when empty")
	}
	if containsField(jsonStr, "client_cert") {
		t.Error("JSON should not contain client_cert when empty")
	}
}

func TestConfig_SaveCreatesDirectory(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pasctl-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := Default()

	// Config directory should not exist yet
	configDir := filepath.Join(tmpDir, ".pasctl")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatal("Config directory should not exist before Save()")
	}

	// Save should create directory
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Config path should be a directory")
	}

	// Verify config file was created
	configPath := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file was not created: %v", err)
	}
}

func TestConfig_SaveFilePermissions(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pasctl-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := Default()

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file permissions (0600 = owner read/write only)
	configPath := filepath.Join(tmpDir, ".pasctl", "config.json")
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Check that file is not world-readable
	perm := info.Mode().Perm()
	if perm&0077 != 0 {
		t.Errorf("Config file permissions should not allow group/other access, got: %o", perm)
	}
}

// Helper function to check if JSON contains a field
func containsField(jsonStr, field string) bool {
	return json.Valid([]byte(jsonStr)) && len(jsonStr) > 0 &&
		(len(field) > 0 && jsonContains(jsonStr, field))
}

func jsonContains(jsonStr, field string) bool {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return false
	}
	_, ok := m[field]
	if ok {
		return true
	}
	// Check nested CCP
	if ccp, ok := m["ccp"].(map[string]interface{}); ok {
		_, ok := ccp[field]
		return ok
	}
	return false
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pasctl-config-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config file
	cfg := Default()
	cfg.CCP = &CCPConfig{
		Enabled: true,
		AppID:   "TestApp",
		Safe:    "TestSafe",
	}
	cfg.Save()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load()
	}
}

func BenchmarkSave(b *testing.B) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pasctl-config-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := Default()
	cfg.CCP = &CCPConfig{
		Enabled: true,
		AppID:   "TestApp",
		Safe:    "TestSafe",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.Save()
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := &Config{
		OutputFormat: "invalid",
		HistorySize:  -1,
		Timeout:      -5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.Validate()
	}
}
