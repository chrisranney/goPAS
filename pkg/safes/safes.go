// Package safes provides CyberArk safe management functionality.
// This is equivalent to the Safes functions in psPAS including
// Get-PASSafe, Add-PASSafe, Set-PASSafe, Remove-PASSafe, etc.
package safes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrisranney/gopas/internal/session"
)

// Safe represents a CyberArk safe.
type Safe struct {
	SafeURLId                 string   `json:"safeUrlId"`
	SafeName                  string   `json:"safeName"`
	SafeNumber                int      `json:"safeNumber"`
	Description               string   `json:"description,omitempty"`
	Location                  string   `json:"location,omitempty"`
	Creator                   *Creator `json:"creator,omitempty"`
	OLACEnabled               bool     `json:"olacEnabled"`
	ManagingCPM               string   `json:"managingCPM,omitempty"`
	NumberOfVersionsRetention *int     `json:"numberOfVersionsRetention,omitempty"`
	NumberOfDaysRetention     int      `json:"numberOfDaysRetention,omitempty"`
	AutoPurgeEnabled          bool     `json:"autoPurgeEnabled"`
	CreationTime              int64    `json:"creationTime"`
	LastModificationTime      int64    `json:"lastModificationTime,omitempty"`
	IsExpiredMember           bool     `json:"isExpiredMember,omitempty"`
	Accounts                  *int     `json:"accounts,omitempty"`
}

// Creator represents the safe creator information.
type Creator struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SafesResponse represents the response from listing safes.
type SafesResponse struct {
	Value    []Safe `json:"value"`
	Count    int    `json:"count"`
	NextLink string `json:"nextLink,omitempty"`
}

// ListOptions holds options for listing safes.
type ListOptions struct {
	Search          string
	Sort            string
	Offset          int
	Limit           int
	IncludeAccounts bool
	ExtendedDetails bool
}

// List retrieves safes from CyberArk.
// This is equivalent to Get-PASSafe in psPAS.
func List(ctx context.Context, sess *session.Session, opts ListOptions) (*SafesResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.Search != "" {
		params.Set("search", opts.Search)
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.IncludeAccounts {
		params.Set("includeAccounts", "true")
	}
	if opts.ExtendedDetails {
		params.Set("extendedDetails", "true")
	}

	resp, err := sess.Client.Get(ctx, "/Safes", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list safes: %w", err)
	}

	var result SafesResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse safes response: %w", err)
	}

	return &result, nil
}

// Get retrieves a specific safe by name.
// This is equivalent to Get-PASSafe -SafeName in psPAS.
func Get(ctx context.Context, sess *session.Session, safeName string) (*Safe, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Safes/%s", url.PathEscape(safeName)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get safe: %w", err)
	}

	var safe Safe
	if err := json.Unmarshal(resp.Body, &safe); err != nil {
		return nil, fmt.Errorf("failed to parse safe response: %w", err)
	}

	return &safe, nil
}

// CreateOptions holds options for creating a safe.
type CreateOptions struct {
	SafeName                  string `json:"safeName"`
	Description               string `json:"description,omitempty"`
	Location                  string `json:"location,omitempty"`
	OLACEnabled               bool   `json:"olacEnabled,omitempty"`
	ManagingCPM               string `json:"managingCPM,omitempty"`
	NumberOfVersionsRetention *int   `json:"numberOfVersionsRetention,omitempty"`
	NumberOfDaysRetention     int    `json:"numberOfDaysRetention,omitempty"`
	AutoPurgeEnabled          bool   `json:"autoPurgeEnabled,omitempty"`
}

// Create creates a new safe in CyberArk.
// This is equivalent to Add-PASSafe in psPAS.
func Create(ctx context.Context, sess *session.Session, opts CreateOptions) (*Safe, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.SafeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	// Validate safe name
	if len(opts.SafeName) > 28 {
		return nil, fmt.Errorf("safe name cannot exceed 28 characters")
	}

	resp, err := sess.Client.Post(ctx, "/Safes", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create safe: %w", err)
	}

	var safe Safe
	if err := json.Unmarshal(resp.Body, &safe); err != nil {
		return nil, fmt.Errorf("failed to parse safe response: %w", err)
	}

	return &safe, nil
}

// UpdateOptions holds options for updating a safe.
type UpdateOptions struct {
	SafeName                  string `json:"safeName,omitempty"`
	Description               string `json:"description,omitempty"`
	Location                  string `json:"location,omitempty"`
	OLACEnabled               *bool  `json:"olacEnabled,omitempty"`
	ManagingCPM               string `json:"managingCPM,omitempty"`
	NumberOfVersionsRetention *int   `json:"numberOfVersionsRetention,omitempty"`
	NumberOfDaysRetention     *int   `json:"numberOfDaysRetention,omitempty"`
	AutoPurgeEnabled          *bool  `json:"autoPurgeEnabled,omitempty"`
}

// Update updates an existing safe.
// This is equivalent to Set-PASSafe in psPAS.
func Update(ctx context.Context, sess *session.Session, safeName string, opts UpdateOptions) (*Safe, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	resp, err := sess.Client.Put(ctx, fmt.Sprintf("/Safes/%s", url.PathEscape(safeName)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update safe: %w", err)
	}

	var safe Safe
	if err := json.Unmarshal(resp.Body, &safe); err != nil {
		return nil, fmt.Errorf("failed to parse safe response: %w", err)
	}

	return &safe, nil
}

// Delete removes a safe from CyberArk.
// This is equivalent to Remove-PASSafe in psPAS.
func Delete(ctx context.Context, sess *session.Session, safeName string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return fmt.Errorf("safeName is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Safes/%s", url.PathEscape(safeName)))
	if err != nil {
		return fmt.Errorf("failed to delete safe: %w", err)
	}

	return nil
}
