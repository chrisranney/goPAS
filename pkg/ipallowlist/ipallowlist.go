// Package ipallowlist provides CyberArk IP allowlist management functionality.
// This is equivalent to the IPAllowlist functions in psPAS including
// Get-PASIPAllowList, Add-PASIPAllowList, etc.
package ipallowlist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chrisranney/gopas/internal/session"
)

// IPAllowListEntry represents an IP allowlist entry.
type IPAllowListEntry struct {
	IP          string `json:"ip"`
	Description string `json:"description,omitempty"`
}

// IPAllowListResponse represents the response from getting IP allowlist.
type IPAllowListResponse struct {
	IPAllowList []IPAllowListEntry `json:"IPAllowList"`
}

// List retrieves the IP allowlist.
// This is equivalent to Get-PASIPAllowList in psPAS.
func List(ctx context.Context, sess *session.Session) ([]IPAllowListEntry, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/WebServices/PIMServices.svc/IPAllowedList", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP allowlist: %w", err)
	}

	var result IPAllowListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse IP allowlist response: %w", err)
	}

	return result.IPAllowList, nil
}

// AddOptions holds options for adding an IP allowlist entry.
type AddOptions struct {
	IP          string `json:"ip"`
	Description string `json:"description,omitempty"`
}

// Add adds an IP address to the allowlist.
// This is equivalent to Add-PASIPAllowList in psPAS.
func Add(ctx context.Context, sess *session.Session, opts AddOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if opts.IP == "" {
		return fmt.Errorf("IP is required")
	}

	body := map[string]interface{}{
		"IPAllowList": []map[string]string{
			{
				"ip":          opts.IP,
				"description": opts.Description,
			},
		},
	}

	_, err := sess.Client.Put(ctx, "/WebServices/PIMServices.svc/IPAllowedList", body)
	if err != nil {
		return fmt.Errorf("failed to add IP to allowlist: %w", err)
	}

	return nil
}

// Remove removes an IP address from the allowlist.
// This is equivalent to Remove-PASIPAllowList in psPAS.
func Remove(ctx context.Context, sess *session.Session, ip string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if ip == "" {
		return fmt.Errorf("IP is required")
	}

	// Get current list
	current, err := List(ctx, sess)
	if err != nil {
		return fmt.Errorf("failed to get current allowlist: %w", err)
	}

	// Filter out the IP to remove
	var newList []map[string]string
	for _, entry := range current {
		if entry.IP != ip {
			newList = append(newList, map[string]string{
				"ip":          entry.IP,
				"description": entry.Description,
			})
		}
	}

	body := map[string]interface{}{
		"IPAllowList": newList,
	}

	_, err = sess.Client.Put(ctx, "/WebServices/PIMServices.svc/IPAllowedList", body)
	if err != nil {
		return fmt.Errorf("failed to update IP allowlist: %w", err)
	}

	return nil
}
