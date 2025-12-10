// Package platforms provides CyberArk platform management functionality.
// This is equivalent to the Platforms functions in psPAS including
// Get-PASPlatform, Add-PASPlatform, Set-PASPlatform, Remove-PASPlatform, etc.
package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// Platform represents a CyberArk platform.
type Platform struct {
	ID                          types.FlexibleID   `json:"id,omitempty"`
	PlatformID                  types.FlexibleID   `json:"platformId,omitempty"`
	Name                        string             `json:"name"`
	Active                      bool               `json:"active"`
	Description                 string             `json:"description,omitempty"`
	SystemType                  string             `json:"systemType,omitempty"`
	PlatformType                string             `json:"platformType,omitempty"`
	CredentialsManagementPolicy *CredentialsPolicy `json:"credentialsManagementPolicy,omitempty"`
	PrivilegedAccessWorkflows   *AccessWorkflows   `json:"privilegedAccessWorkflows,omitempty"`
	PrivilegedSessionManagement *SessionManagement `json:"privilegedSessionManagement,omitempty"`
	AllowedSafes                string             `json:"allowedSafes,omitempty"`
}

// CredentialsPolicy represents credentials management policy.
type CredentialsPolicy struct {
	Verification              *VerificationPolicy `json:"verification,omitempty"`
	Change                    *ChangePolicy       `json:"change,omitempty"`
	Reconcile                 *ReconcilePolicy    `json:"reconcile,omitempty"`
	SecretUpdateConfiguration *SecretUpdateConfig `json:"secretUpdateConfiguration,omitempty"`
}

// VerificationPolicy represents verification settings.
type VerificationPolicy struct {
	PerformAutomatic          bool `json:"performAutomatic"`
	RequirePasswordEveryXDays int  `json:"requirePasswordEveryXDays,omitempty"`
	AutoOnAdd                 bool `json:"autoOnAdd,omitempty"`
	AllowManual               bool `json:"allowManual"`
}

// ChangePolicy represents password change settings.
type ChangePolicy struct {
	PerformAutomatic          bool `json:"performAutomatic"`
	RequirePasswordEveryXDays int  `json:"requirePasswordEveryXDays,omitempty"`
	AutoOnAdd                 bool `json:"autoOnAdd,omitempty"`
	AllowManual               bool `json:"allowManual"`
}

// ReconcilePolicy represents reconciliation settings.
type ReconcilePolicy struct {
	AutomaticReconcileWhenUnsynced bool `json:"automaticReconcileWhenUnsynced"`
	AllowManual                    bool `json:"allowManual"`
}

// SecretUpdateConfig represents secret update configuration.
type SecretUpdateConfig struct {
	ChangePasswordInResetMode bool `json:"changePasswordInResetMode"`
}

// AccessWorkflows represents privileged access workflows.
type AccessWorkflows struct {
	RequireDualControlPasswordAccessApproval *DualControlPolicy `json:"requireDualControlPasswordAccessApproval,omitempty"`
	EnforceCheckinCheckoutExclusiveAccess    *CheckinCheckout   `json:"enforceCheckinCheckoutExclusiveAccess,omitempty"`
	EnforceOnetimePasswordAccess             *OneTimePassword   `json:"enforceOnetimePasswordAccess,omitempty"`
}

// DualControlPolicy represents dual control settings.
type DualControlPolicy struct {
	IsActive      bool `json:"isActive"`
	IsAnException bool `json:"isAnException,omitempty"`
}

// CheckinCheckout represents check-in/check-out settings.
type CheckinCheckout struct {
	IsActive      bool `json:"isActive"`
	IsAnException bool `json:"isAnException,omitempty"`
}

// OneTimePassword represents one-time password settings.
type OneTimePassword struct {
	IsActive      bool `json:"isActive"`
	IsAnException bool `json:"isAnException,omitempty"`
}

// SessionManagement represents privileged session management settings.
type SessionManagement struct {
	PSMServerID                                    string           `json:"psmServerId,omitempty"`
	PSMServerName                                  string           `json:"psmServerName,omitempty"`
	RequirePrivilegedSessionMonitoringAndIsolation *PSMPolicy       `json:"requirePrivilegedSessionMonitoringAndIsolation,omitempty"`
	RecordAndSaveSessionActivity                   *RecordingPolicy `json:"recordAndSaveSessionActivity,omitempty"`
}

// PSMPolicy represents PSM policy settings.
type PSMPolicy struct {
	IsActive      bool `json:"isActive"`
	IsAnException bool `json:"isAnException,omitempty"`
}

// RecordingPolicy represents recording policy settings.
type RecordingPolicy struct {
	IsActive      bool `json:"isActive"`
	IsAnException bool `json:"isAnException,omitempty"`
}

// PlatformsResponse represents the response from listing platforms.
type PlatformsResponse struct {
	Platforms []Platform `json:"Platforms"`
	Total     int        `json:"Total,omitempty"`
}

// ListOptions holds options for listing platforms.
type ListOptions struct {
	Search       string
	Active       *bool
	PlatformType string
	SystemType   string
}

// List retrieves platforms from CyberArk.
// This is equivalent to Get-PASPlatform in psPAS.
func List(ctx context.Context, sess *session.Session, opts ListOptions) (*PlatformsResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.Search != "" {
		params.Set("search", opts.Search)
	}
	if opts.Active != nil {
		params.Set("active", strconv.FormatBool(*opts.Active))
	}
	if opts.PlatformType != "" {
		params.Set("platformType", opts.PlatformType)
	}
	if opts.SystemType != "" {
		params.Set("systemType", opts.SystemType)
	}

	resp, err := sess.Client.Get(ctx, "/Platforms", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list platforms: %w", err)
	}

	var result PlatformsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse platforms response: %w", err)
	}

	return &result, nil
}

// Get retrieves a specific platform by ID.
// This is equivalent to Get-PASPlatform -PlatformID in psPAS.
func Get(ctx context.Context, sess *session.Session, platformID string) (*Platform, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if platformID == "" {
		return nil, fmt.Errorf("platformID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Platforms/%s", url.PathEscape(platformID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform: %w", err)
	}

	var platform Platform
	if err := json.Unmarshal(resp.Body, &platform); err != nil {
		return nil, fmt.Errorf("failed to parse platform response: %w", err)
	}

	return &platform, nil
}

// Activate activates a platform.
// This is equivalent to Enable-PASPlatform in psPAS.
func Activate(ctx context.Context, sess *session.Session, platformID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if platformID == "" {
		return fmt.Errorf("platformID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Platforms/%s/activate", url.PathEscape(platformID)), nil)
	if err != nil {
		return fmt.Errorf("failed to activate platform: %w", err)
	}

	return nil
}

// Deactivate deactivates a platform.
// This is equivalent to Disable-PASPlatform in psPAS.
func Deactivate(ctx context.Context, sess *session.Session, platformID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if platformID == "" {
		return fmt.Errorf("platformID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Platforms/%s/deactivate", url.PathEscape(platformID)), nil)
	if err != nil {
		return fmt.Errorf("failed to deactivate platform: %w", err)
	}

	return nil
}

// Delete removes a platform from CyberArk.
// This is equivalent to Remove-PASPlatform in psPAS.
func Delete(ctx context.Context, sess *session.Session, platformID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if platformID == "" {
		return fmt.Errorf("platformID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Platforms/%s", url.PathEscape(platformID)))
	if err != nil {
		return fmt.Errorf("failed to delete platform: %w", err)
	}

	return nil
}

// DuplicateOptions holds options for duplicating a platform.
type DuplicateOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Duplicate creates a copy of an existing platform.
// This is equivalent to Copy-PASPlatform in psPAS.
func Duplicate(ctx context.Context, sess *session.Session, platformID string, opts DuplicateOptions) (*Platform, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if platformID == "" {
		return nil, fmt.Errorf("platformID is required")
	}

	if opts.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Platforms/%s/duplicate", url.PathEscape(platformID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate platform: %w", err)
	}

	var platform Platform
	if err := json.Unmarshal(resp.Body, &platform); err != nil {
		return nil, fmt.Errorf("failed to parse platform response: %w", err)
	}

	return &platform, nil
}

// ExportPlatform exports a platform definition.
// This is equivalent to Export-PASPlatform in psPAS.
func ExportPlatform(ctx context.Context, sess *session.Session, platformID string) ([]byte, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if platformID == "" {
		return nil, fmt.Errorf("platformID is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Platforms/%s/export", url.PathEscape(platformID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to export platform: %w", err)
	}

	return resp.Body, nil
}

// ImportPlatform imports a platform definition.
// This is equivalent to Import-PASPlatform in psPAS.
func ImportPlatform(ctx context.Context, sess *session.Session, platformZip []byte) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if len(platformZip) == 0 {
		return fmt.Errorf("platformZip is required")
	}

	body := map[string]interface{}{
		"ImportFile": platformZip,
	}

	_, err := sess.Client.Post(ctx, "/Platforms/import", body)
	if err != nil {
		return fmt.Errorf("failed to import platform: %w", err)
	}

	return nil
}
