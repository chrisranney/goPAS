// Package systemhealth provides CyberArk system health monitoring functionality.
// This is equivalent to the SystemHealth functions in psPAS including
// Get-PASComponentSummary, Get-PASComponentDetail, etc.
package systemhealth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
)

// ComponentSummary represents a summary of component health.
type ComponentSummary struct {
	ComponentID            string `json:"ComponentID"`
	ComponentName          string `json:"ComponentName"`
	ComponentType          string `json:"ComponentType"`
	Description            string `json:"Description,omitempty"`
	ConnectedComponentID   string `json:"ConnectedComponentID,omitempty"`
	ConnectedComponentName string `json:"ConnectedComponentName,omitempty"`
	IsLoggedOn             bool   `json:"IsLoggedOn"`
	LastLogonDate          int64  `json:"LastLogonDate,omitempty"`
}

// ComponentDetail represents detailed component health information.
type ComponentDetail struct {
	ComponentID            string            `json:"ComponentID"`
	ComponentName          string            `json:"ComponentName"`
	ComponentType          string            `json:"ComponentType"`
	Description            string            `json:"Description,omitempty"`
	ConnectedComponentID   string            `json:"ConnectedComponentID,omitempty"`
	ConnectedComponentName string            `json:"ConnectedComponentName,omitempty"`
	IsLoggedOn             bool              `json:"IsLoggedOn"`
	LastLogonDate          int64             `json:"LastLogonDate,omitempty"`
	ComponentVersion       string            `json:"ComponentVersion,omitempty"`
	ComponentSpecificData  map[string]interface{} `json:"ComponentSpecificData,omitempty"`
}

// ComponentSummaryResponse represents the response from listing component summaries.
type ComponentSummaryResponse struct {
	Components []ComponentSummary `json:"Components"`
}

// ListComponentSummary retrieves a summary of component health status.
// This is equivalent to Get-PASComponentSummary in psPAS.
func ListComponentSummary(ctx context.Context, sess *session.Session) ([]ComponentSummary, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/ComponentsMonitoringSummary", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list component summary: %w", err)
	}

	var result ComponentSummaryResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse component summary response: %w", err)
	}

	return result.Components, nil
}

// GetComponentDetail retrieves detailed information for a specific component.
// This is equivalent to Get-PASComponentDetail in psPAS.
func GetComponentDetail(ctx context.Context, sess *session.Session, componentID string) (*ComponentDetail, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if componentID == "" {
		return nil, fmt.Errorf("componentID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/ComponentsMonitoringDetails/%s", url.PathEscape(componentID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get component detail: %w", err)
	}

	var detail ComponentDetail
	if err := json.Unmarshal(resp.Body, &detail); err != nil {
		return nil, fmt.Errorf("failed to parse component detail response: %w", err)
	}

	return &detail, nil
}

// VaultHealth represents the overall vault health status.
type VaultHealth struct {
	IsHealthy     bool   `json:"IsHealthy"`
	HealthDetails string `json:"HealthDetails,omitempty"`
}

// GetVaultHealth retrieves the overall vault health status.
func GetVaultHealth(ctx context.Context, sess *session.Session) (*VaultHealth, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/ServerHealth", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault health: %w", err)
	}

	var health VaultHealth
	if err := json.Unmarshal(resp.Body, &health); err != nil {
		return nil, fmt.Errorf("failed to parse vault health response: %w", err)
	}

	return &health, nil
}
