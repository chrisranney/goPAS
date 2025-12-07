// Package serverweb provides CyberArk server and web service functionality.
// This is equivalent to the ServerWebServices functions in psPAS including
// Get-PASServer, Get-PASServerWebService, etc.
package serverweb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chrisranney/gopas/internal/session"
)

// ServerInfo represents CyberArk server information.
type ServerInfo struct {
	ServerID         string  `json:"ServerID"`
	ServerName       string  `json:"ServerName"`
	ServicesUsed     string  `json:"ServicesUsed,omitempty"`
	ApplicationsUsed string  `json:"ApplicationsUsed,omitempty"`
	InternalVersion  float64 `json:"InternalVersion"`
	ExternalVersion  string  `json:"ExternalVersion"`
}

// GetServer retrieves CyberArk server information.
// This is equivalent to Get-PASServer in psPAS.
func GetServer(ctx context.Context, sess *session.Session) (*ServerInfo, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
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

// WebServiceStatus represents web service status.
type WebServiceStatus struct {
	IsWebServiceEnabled bool   `json:"IsWebServiceEnabled"`
	WebServiceID        string `json:"WebServiceID,omitempty"`
}

// GetWebServiceStatus retrieves web service status.
// This is equivalent to Get-PASServerWebService in psPAS.
func GetWebServiceStatus(ctx context.Context, sess *session.Session) (*WebServiceStatus, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/WebServices/PIMServices.svc/Verify", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get web service status: %w", err)
	}

	var status WebServiceStatus
	if err := json.Unmarshal(resp.Body, &status); err != nil {
		// If parsing fails, the service is likely up but returned unexpected format
		return &WebServiceStatus{IsWebServiceEnabled: true}, nil
	}

	return &status, nil
}

// APIStatus represents the API status.
type APIStatus struct {
	StatusCode int    `json:"StatusCode"`
	Message    string `json:"Message,omitempty"`
}

// VerifyAPI verifies the API is accessible.
func VerifyAPI(ctx context.Context, sess *session.Session) (*APIStatus, error) {
	if sess == nil {
		return nil, fmt.Errorf("session is required")
	}

	resp, err := sess.Client.Get(ctx, "/Server/Verify", nil)
	if err != nil {
		return &APIStatus{StatusCode: 500, Message: err.Error()}, err
	}

	return &APIStatus{StatusCode: resp.StatusCode, Message: "OK"}, nil
}
