// Package connections provides CyberArk PSM connection functionality.
// This is equivalent to the Connections functions in psPAS including
// New-PASSession (PSM Connect), Get-PASPSMConnectionParameter, etc.
package connections

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
)

// ConnectionRequest represents a PSM connection request.
type ConnectionRequest struct {
	Reason            string            `json:"reason,omitempty"`
	TicketingSystemName string          `json:"ticketingSystemName,omitempty"`
	TicketID          string            `json:"ticketId,omitempty"`
	ConnectionComponent string          `json:"ConnectionComponent,omitempty"`
	ConnectionParams  map[string]string `json:"ConnectionParams,omitempty"`
}

// ConnectionResponse represents a PSM connection response.
type ConnectionResponse struct {
	PSMConnectURL string `json:"PSMConnectURL,omitempty"`
	RDPFile       string `json:"RDPFile,omitempty"`
}

// Connect initiates a PSM connection to an account.
// This is equivalent to New-PASPSMSession in psPAS.
func Connect(ctx context.Context, sess *session.Session, accountID string, req ConnectionRequest) (*ConnectionResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/PSMConnect", accountID), req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate connection: %w", err)
	}

	var connResp ConnectionResponse
	if err := json.Unmarshal(resp.Body, &connResp); err != nil {
		return nil, fmt.Errorf("failed to parse connection response: %w", err)
	}

	return &connResp, nil
}

// AdHocConnectRequest represents an ad-hoc PSM connection request.
type AdHocConnectRequest struct {
	UserName          string            `json:"userName"`
	Secret            string            `json:"secret"`
	Address           string            `json:"address"`
	PlatformID        string            `json:"platformId"`
	ExtraFields       map[string]string `json:"extraFields,omitempty"`
	PSMConnectPrerequisites *PSMPrerequisites `json:"PSMConnectPrerequisites,omitempty"`
}

// PSMPrerequisites holds PSM connection prerequisites.
type PSMPrerequisites struct {
	ConnectionComponent string `json:"ConnectionComponent,omitempty"`
	ConnectionType      string `json:"ConnectionType,omitempty"`
}

// AdHocConnect initiates an ad-hoc PSM connection without a managed account.
// This is equivalent to New-PASPSMSession -AdHocConnect in psPAS.
func AdHocConnect(ctx context.Context, sess *session.Session, req AdHocConnectRequest) (*ConnectionResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if req.UserName == "" {
		return nil, fmt.Errorf("userName is required")
	}

	if req.Secret == "" {
		return nil, fmt.Errorf("secret is required")
	}

	if req.Address == "" {
		return nil, fmt.Errorf("address is required")
	}

	if req.PlatformID == "" {
		return nil, fmt.Errorf("platformID is required")
	}

	resp, err := sess.Client.Post(ctx, "/Accounts/AdHocConnect", req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate ad-hoc connection: %w", err)
	}

	var connResp ConnectionResponse
	if err := json.Unmarshal(resp.Body, &connResp); err != nil {
		return nil, fmt.Errorf("failed to parse connection response: %w", err)
	}

	return &connResp, nil
}

// ConnectionComponent represents a PSM connection component.
type ConnectionComponent struct {
	PSMConnectorID string `json:"PSMConnectorID"`
	PSMServerID    string `json:"PSMServerID,omitempty"`
}

// GetConnectionComponents retrieves available connection components for a platform.
// This is equivalent to Get-PASConnectionComponent in psPAS.
func GetConnectionComponents(ctx context.Context, sess *session.Session, platformID string) ([]ConnectionComponent, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if platformID == "" {
		return nil, fmt.Errorf("platformID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Platforms/%s/PrivilegedSessionManagement", url.PathEscape(platformID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection components: %w", err)
	}

	var result struct {
		PSMConnectors []ConnectionComponent `json:"PSMConnectors"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse connection components response: %w", err)
	}

	return result.PSMConnectors, nil
}

// PSMServer represents a PSM server.
type PSMServer struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Address     string `json:"Address,omitempty"`
	PSMVersion  string `json:"PSMVersion,omitempty"`
}

// GetPSMServers retrieves available PSM servers.
// This is equivalent to Get-PASPSMServer in psPAS.
func GetPSMServers(ctx context.Context, sess *session.Session) ([]PSMServer, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/PSM/Servers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PSM servers: %w", err)
	}

	var result struct {
		PSMServers []PSMServer `json:"PSMServers"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PSM servers response: %w", err)
	}

	return result.PSMServers, nil
}
