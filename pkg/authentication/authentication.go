// Package authentication provides CyberArk authentication functionality.
// This is equivalent to the Authentication functions in psPAS including
// New-PASSession, Close-PASSession, Use-PASSession, etc.
package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chrisranney/gopas/internal/client"
	"github.com/chrisranney/gopas/internal/session"
)

// AuthMethod represents the authentication method to use.
type AuthMethod string

const (
	// AuthMethodCyberArk uses CyberArk native authentication
	AuthMethodCyberArk AuthMethod = "CyberArk"
	// AuthMethodLDAP uses LDAP authentication
	AuthMethodLDAP AuthMethod = "LDAP"
	// AuthMethodRADIUS uses RADIUS authentication
	AuthMethodRADIUS AuthMethod = "RADIUS"
	// AuthMethodWindows uses Windows authentication
	AuthMethodWindows AuthMethod = "Windows"
)

// Credentials holds the authentication credentials.
type Credentials struct {
	Username string
	Password string
}

// SessionOptions holds options for creating a new session.
type SessionOptions struct {
	// BaseURL is the CyberArk server URL (required)
	BaseURL string

	// Credentials for authentication
	Credentials Credentials

	// AuthMethod is the authentication method to use (default: CyberArk)
	AuthMethod AuthMethod

	// ConcurrentSession allows concurrent sessions for the same user
	ConcurrentSession bool

	// SkipVersionCheck skips the version check after authentication
	SkipVersionCheck bool

	// CustomHTTPClient allows using a custom HTTP client
	CustomHTTPClient *http.Client
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	ConcurrentSession bool   `json:"concurrentSession,omitempty"`
}

// LoginResponse represents the login response.
type LoginResponse struct {
	Token string `json:"CyberArkLogonResult,omitempty"`
}

// ServerInfo represents the CyberArk server information.
type ServerInfo struct {
	ServerID         string  `json:"ServerID"`
	ServerName       string  `json:"ServerName"`
	ServicesUsed     string  `json:"ServicesUsed"`
	ApplicationsUsed string  `json:"ApplicationsUsed"`
	InternalVersion  float64 `json:"InternalVersion"`
	ExternalVersion  string  `json:"ExternalVersion"`
}

// NewSession creates a new authenticated session with CyberArk.
// This is equivalent to New-PASSession in psPAS.
func NewSession(ctx context.Context, opts SessionOptions) (*session.Session, error) {
	if opts.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	if opts.Credentials.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if opts.Credentials.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Set default auth method
	if opts.AuthMethod == "" {
		opts.AuthMethod = AuthMethodCyberArk
	}

	// Create a new session
	sess, err := session.NewSession(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Build the authentication endpoint based on method
	authPath := getAuthPath(opts.AuthMethod)

	// Create login request
	loginReq := LoginRequest{
		Username:          opts.Credentials.Username,
		Password:          opts.Credentials.Password,
		ConcurrentSession: opts.ConcurrentSession,
	}

	// Perform authentication
	resp, err := sess.Client.Post(ctx, authPath, loginReq)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Parse the response
	var loginResp LoginResponse
	if err := json.Unmarshal(resp.Body, &loginResp); err != nil {
		// Try to parse as plain string (some versions return just the token)
		token := string(resp.Body)
		token = trimQuotes(token)
		loginResp.Token = token
	}

	if loginResp.Token == "" {
		return nil, fmt.Errorf("no authentication token received")
	}

	// Set the session as authenticated
	sess.SetAuthenticated(opts.Credentials.Username, loginResp.Token, string(opts.AuthMethod))

	// Get server version unless skipped
	if !opts.SkipVersionCheck {
		if err := fetchServerVersion(ctx, sess); err != nil {
			// Log warning but don't fail - version check is optional
			_ = err
		}
	}

	return sess, nil
}

// CloseSession closes the authenticated session.
// This is equivalent to Close-PASSession in psPAS.
func CloseSession(ctx context.Context, sess *session.Session) error {
	if sess == nil || !sess.IsValid() {
		return nil
	}

	// Call logoff endpoint
	_, err := sess.Client.Post(ctx, "/Auth/Logoff", nil)
	if err != nil {
		// Check if it's a 401 (already logged out)
		if apiErr, ok := client.AsAPIError(err); ok && apiErr.IsUnauthorized() {
			sess.Close()
			return nil
		}
		return fmt.Errorf("failed to close session: %w", err)
	}

	sess.Close()
	return nil
}

// GetServerInfo retrieves the CyberArk server information.
// This is equivalent to Get-PASServer in psPAS.
func GetServerInfo(ctx context.Context, sess *session.Session) (*ServerInfo, error) {
	if sess == nil {
		return nil, fmt.Errorf("session is required")
	}

	resp, err := sess.Client.Get(ctx, "/WebServices/PIMServices.svc/Server", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	var info ServerInfo
	if err := json.Unmarshal(resp.Body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse server info: %w", err)
	}

	return &info, nil
}

// GetComponentsHealth retrieves the health status of CyberArk components.
// This is equivalent to Get-PASComponentSummary in psPAS.
func GetComponentsHealth(ctx context.Context, sess *session.Session) ([]ComponentHealth, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/ComponentsMonitoringSummary", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get component health: %w", err)
	}

	var result struct {
		Components []ComponentHealth `json:"Components"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse component health: %w", err)
	}

	return result.Components, nil
}

// ComponentHealth represents the health status of a component.
type ComponentHealth struct {
	ComponentID          string `json:"ComponentID"`
	ComponentName        string `json:"ComponentName"`
	Description          string `json:"Description"`
	ConnectedComponentID string `json:"ConnectedComponentID"`
	IsLoggedOn           bool   `json:"IsLoggedOn"`
	LastLogonDate        int64  `json:"LastLogonDate"`
}

// getAuthPath returns the authentication endpoint path based on the method.
func getAuthPath(method AuthMethod) string {
	switch method {
	case AuthMethodLDAP:
		return "/Auth/LDAP/Logon"
	case AuthMethodRADIUS:
		return "/Auth/RADIUS/Logon"
	case AuthMethodWindows:
		return "/Auth/Windows/Logon"
	default:
		return "/Auth/CyberArk/Logon"
	}
}

// fetchServerVersion fetches and stores the server version.
func fetchServerVersion(ctx context.Context, sess *session.Session) error {
	info, err := GetServerInfo(ctx, sess)
	if err != nil {
		return err
	}

	sess.SetVersion(info.ExternalVersion)
	return nil
}

// trimQuotes removes surrounding quotes from a string.
func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
