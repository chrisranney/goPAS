// Package authentication provides SAML authentication functionality.
// This is equivalent to New-PASSession SAML authentication in psPAS.
package authentication

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chrisranney/gopas/internal/session"
)

// SAMLSessionOptions holds options for SAML authentication.
type SAMLSessionOptions struct {
	// BaseURL is the CyberArk server URL (required)
	BaseURL string

	// SAMLResponse is the SAML response token from the IdP
	SAMLResponse string

	// UseIntegratedAuth uses Windows Integrated Authentication
	UseIntegratedAuth bool

	// IDPLoginURL is the Identity Provider login URL
	IDPLoginURL string
}

// SAMLLoginRequest represents the SAML login request body.
type SAMLLoginRequest struct {
	SAMLResponse      string `json:"SAMLResponse,omitempty"`
	ConcurrentSession bool   `json:"concurrentSession,omitempty"`
}

// SAMLAuthResponse represents the SAML auth response.
type SAMLAuthResponse struct {
	LogonResult string `json:"CyberArkLogonResult"`
}

// NewSAMLSession creates a new session using SAML authentication.
// This is equivalent to New-PASSession -SAMLAuth in psPAS.
func NewSAMLSession(ctx context.Context, opts SAMLSessionOptions) (*session.Session, error) {
	if opts.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	if opts.SAMLResponse == "" && !opts.UseIntegratedAuth {
		return nil, fmt.Errorf("SAMLResponse is required when not using integrated auth")
	}

	// Create a new session
	sess, err := session.NewSession(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Create login request
	loginReq := SAMLLoginRequest{
		SAMLResponse: opts.SAMLResponse,
	}

	// Perform SAML authentication
	resp, err := sess.Client.Post(ctx, "/Auth/SAML/Logon", loginReq)
	if err != nil {
		return nil, fmt.Errorf("SAML authentication failed: %w", err)
	}

	// Parse the response
	var loginResp SAMLAuthResponse
	if err := json.Unmarshal(resp.Body, &loginResp); err != nil {
		// Try to parse as plain string
		token := string(resp.Body)
		token = trimQuotes(token)
		loginResp.LogonResult = token
	}

	if loginResp.LogonResult == "" {
		return nil, fmt.Errorf("no authentication token received")
	}

	// Set the session as authenticated
	sess.SetAuthenticated("SAML User", loginResp.LogonResult, "SAML")

	// Get server version
	if err := fetchServerVersion(ctx, sess); err != nil {
		_ = err
	}

	return sess, nil
}
