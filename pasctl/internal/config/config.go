// Package config provides configuration management for pasctl.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the pasctl configuration settings.
type Config struct {
	DefaultServer   string `json:"default_server,omitempty"`
	DefaultAuthType string `json:"default_auth_type,omitempty"`
	OutputFormat    string `json:"output_format,omitempty"`
	HistorySize     int    `json:"history_size,omitempty"`
	InsecureSSL     bool   `json:"insecure_ssl,omitempty"`
	Timeout         int    `json:"timeout_seconds,omitempty"`

	// CCP (Central Credential Provider) settings for default login
	// Note: Passwords are NEVER stored - they are retrieved from CCP at runtime
	CCP *CCPConfig `json:"ccp,omitempty"`
}

// CCPConfig holds CCP configuration for default login.
// This allows automatic login using credentials retrieved from CyberArk CCP.
type CCPConfig struct {
	// Enabled indicates whether CCP default login is enabled
	Enabled bool `json:"enabled,omitempty"`

	// CCPURL is the CCP server URL for retrieving credentials
	// (e.g., https://ccp.cyberark.example.com)
	CCPURL string `json:"ccp_url,omitempty"`

	// PVWAURL is the PVWA server URL for authentication after retrieving credentials
	// (e.g., https://pvwa.cyberark.example.com)
	// If empty, uses DefaultServer
	PVWAURL string `json:"pvwa_url,omitempty"`

	// AppID is the application ID registered in CyberArk for CCP access (required)
	AppID string `json:"app_id,omitempty"`

	// Safe is the safe containing the login credential (required)
	Safe string `json:"safe,omitempty"`

	// Object is the account object name (optional, use with Folder or UserName/Address)
	Object string `json:"object,omitempty"`

	// Folder is the folder path within the safe (optional)
	Folder string `json:"folder,omitempty"`

	// UserName filters by username (optional)
	UserName string `json:"username,omitempty"`

	// Address filters by address/hostname (optional)
	Address string `json:"address,omitempty"`

	// Query is a free-text search query (optional)
	Query string `json:"query,omitempty"`

	// AuthMethod is the auth method to use after retrieving credentials
	// Defaults to DefaultAuthType if not specified
	AuthMethod string `json:"auth_method,omitempty"`

	// ClientCert path for mutual TLS authentication with CCP (optional)
	ClientCert string `json:"client_cert,omitempty"`

	// ClientKey path for mutual TLS authentication with CCP (optional)
	ClientKey string `json:"client_key,omitempty"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		DefaultAuthType: "cyberark",
		OutputFormat:    "table",
		HistorySize:     1000,
		InsecureSSL:     false,
		Timeout:         30,
	}
}

// ConfigDir returns the configuration directory path.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pasctl"), nil
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// HistoryPath returns the path to the history file.
func HistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pasctl_history"), nil
}

// Load loads the configuration from disk.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Default(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Default(), err
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return Default(), err
	}

	return cfg, nil
}

// Save saves the configuration to disk.
func (c *Config) Save() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate output format
	switch c.OutputFormat {
	case "table", "json", "yaml":
		// valid
	default:
		c.OutputFormat = "table"
	}

	// Validate history size
	if c.HistorySize <= 0 {
		c.HistorySize = 1000
	}

	// Validate timeout
	if c.Timeout <= 0 {
		c.Timeout = 30
	}

	return nil
}

// IsCCPEnabled returns true if CCP default login is configured and enabled.
func (c *Config) IsCCPEnabled() bool {
	return c.CCP != nil && c.CCP.Enabled && c.CCP.AppID != "" && c.CCP.Safe != ""
}

// GetCCPURL returns the CCP URL for credential retrieval.
// Returns empty string if not configured (CCP URL must be explicitly set).
func (c *Config) GetCCPURL() string {
	if c.CCP != nil && c.CCP.CCPURL != "" {
		return c.CCP.CCPURL
	}
	return ""
}

// GetPVWAURL returns the PVWA URL for authentication after CCP credential retrieval.
// Falls back to DefaultServer if not explicitly set in CCP config.
func (c *Config) GetPVWAURL() string {
	if c.CCP != nil && c.CCP.PVWAURL != "" {
		return c.CCP.PVWAURL
	}
	return c.DefaultServer
}

// GetCCPAuthMethod returns the auth method for CCP login.
func (c *Config) GetCCPAuthMethod() string {
	if c.CCP != nil && c.CCP.AuthMethod != "" {
		return c.CCP.AuthMethod
	}
	return c.DefaultAuthType
}
