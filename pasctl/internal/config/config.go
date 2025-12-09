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
