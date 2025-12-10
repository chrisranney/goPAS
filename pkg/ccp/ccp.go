// Package ccp provides CyberArk Central Credential Provider (CCP) functionality.
// CCP allows applications to retrieve credentials without requiring a user session.
// This is typically used for automated systems and application-to-vault communication.
package ccp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/chrisranney/gopas/pkg/types"
)

// Client represents a CCP client for retrieving credentials.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// ClientConfig holds configuration for creating a CCP client.
type ClientConfig struct {
	// BaseURL is the CyberArk server URL (e.g., https://cyberark.example.com)
	BaseURL string

	// SkipTLSVerify disables TLS certificate verification (not recommended for production)
	SkipTLSVerify bool

	// Timeout is the request timeout duration
	Timeout time.Duration

	// ClientCert and ClientKey for mutual TLS authentication (optional)
	ClientCert string
	ClientKey  string
}

// CredentialRequest represents a request to retrieve credentials from CCP.
type CredentialRequest struct {
	// AppID is the application ID registered in CyberArk (required)
	AppID string

	// Safe is the safe name containing the credential (required)
	Safe string

	// Object is the account/object name (optional, use with Folder)
	Object string

	// Folder is the folder path within the safe (optional)
	Folder string

	// UserName filters by username (optional)
	UserName string

	// Address filters by address/hostname (optional)
	Address string

	// Query is a free-text search query (optional)
	Query string

	// QueryFormat specifies the query format: "Exact" or "Regexp" (optional)
	QueryFormat string

	// Reason for retrieving the credential (optional, for audit)
	Reason string

	// ConnectionTimeout in seconds for the request (optional)
	ConnectionTimeout int
}

// CredentialResponse represents the response from a CCP credential request.
type CredentialResponse struct {
	// Content is the retrieved password/secret
	Content string `json:"Content"`

	// UserName is the account username
	UserName string `json:"UserName"`

	// Address is the target address/hostname
	Address string `json:"Address"`

	// Safe is the safe name
	Safe string `json:"Safe"`

	// Folder is the folder path
	Folder string `json:"Folder"`

	// Name is the account object name
	Name string `json:"Name"`

	// PolicyID is the platform ID
	PolicyID string `json:"PolicyID"`

	// DeviceType is the device type
	DeviceType string `json:"DeviceType"`

	// Properties contains additional account properties
	Properties map[string]string `json:"Properties,omitempty"`

	// PasswordChangeInProcess indicates if password is being changed
	PasswordChangeInProcess types.FlexibleBool `json:"PasswordChangeInProcess"`

	// CreationMethod indicates how the account was created
	CreationMethod string `json:"CreationMethod,omitempty"`
}

// ErrorResponse represents an error response from CCP.
type ErrorResponse struct {
	ErrorCode string `json:"ErrorCode"`
	ErrorMsg  string `json:"ErrorMsg"`
}

// NewClient creates a new CCP client.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	// Default timeout
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipTLSVerify,
	}

	// Load client certificate if provided (for mutual TLS)
	if cfg.ClientCert != "" && cfg.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.ClientCert, cfg.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    cfg.BaseURL,
	}, nil
}

// GetCredential retrieves a credential from CCP.
func (c *Client) GetCredential(ctx context.Context, req CredentialRequest) (*CredentialResponse, error) {
	if req.AppID == "" {
		return nil, fmt.Errorf("AppID is required")
	}
	if req.Safe == "" {
		return nil, fmt.Errorf("Safe is required")
	}

	// Build the CCP URL
	endpoint := fmt.Sprintf("%s/AIMWebService/api/Accounts", c.baseURL)

	// Build query parameters
	params := url.Values{}
	params.Set("AppID", req.AppID)
	params.Set("Safe", req.Safe)

	if req.Object != "" {
		params.Set("Object", req.Object)
	}
	if req.Folder != "" {
		params.Set("Folder", req.Folder)
	}
	if req.UserName != "" {
		params.Set("UserName", req.UserName)
	}
	if req.Address != "" {
		params.Set("Address", req.Address)
	}
	if req.Query != "" {
		params.Set("Query", req.Query)
	}
	if req.QueryFormat != "" {
		params.Set("QueryFormat", req.QueryFormat)
	}
	if req.Reason != "" {
		params.Set("Reason", req.Reason)
	}
	if req.ConnectionTimeout > 0 {
		params.Set("ConnectionTimeout", fmt.Sprintf("%d", req.ConnectionTimeout))
	}

	fullURL := endpoint + "?" + params.Encode()

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.ErrorMsg != "" {
			return nil, fmt.Errorf("CCP error (%s): %s", errResp.ErrorCode, errResp.ErrorMsg)
		}
		return nil, fmt.Errorf("CCP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var credResp CredentialResponse
	if err := json.Unmarshal(body, &credResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &credResp, nil
}

// GetPassword is a convenience method to retrieve just the password.
func (c *Client) GetPassword(ctx context.Context, req CredentialRequest) (string, error) {
	cred, err := c.GetCredential(ctx, req)
	if err != nil {
		return "", err
	}
	return cred.Content, nil
}

// GetLoginCredentials retrieves credentials suitable for logging into CyberArk PVWA.
// This is useful for retrieving a vaulted CyberArk admin credential to authenticate with.
func (c *Client) GetLoginCredentials(ctx context.Context, req CredentialRequest) (username, password string, err error) {
	cred, err := c.GetCredential(ctx, req)
	if err != nil {
		return "", "", err
	}
	return cred.UserName, cred.Content, nil
}
