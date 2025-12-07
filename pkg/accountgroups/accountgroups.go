// Package accountgroups provides CyberArk account group management functionality.
// This is equivalent to the AccountGroups functions in psPAS including
// Get-PASAccountGroup, Add-PASAccountGroup, Add-PASAccountGroupMember, etc.
package accountgroups

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
)

// AccountGroup represents an account group.
type AccountGroup struct {
	GroupID       string              `json:"GroupID"`
	GroupName     string              `json:"GroupName"`
	GroupPlatformID string            `json:"GroupPlatformID,omitempty"`
	Safe          string              `json:"Safe"`
	Members       []AccountGroupMember `json:"Members,omitempty"`
}

// AccountGroupMember represents a member of an account group.
type AccountGroupMember struct {
	AccountID string `json:"AccountID"`
}

// List retrieves account groups from a safe.
// This is equivalent to Get-PASAccountGroup in psPAS.
func List(ctx context.Context, sess *session.Session, safeName string) ([]AccountGroup, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	params := url.Values{}
	params.Set("Safe", safeName)

	resp, err := sess.Client.Get(ctx, "/AccountGroups", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list account groups: %w", err)
	}

	var result []AccountGroup
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse account groups response: %w", err)
	}

	return result, nil
}

// CreateOptions holds options for creating an account group.
type CreateOptions struct {
	GroupName       string `json:"GroupName"`
	GroupPlatformID string `json:"GroupPlatformID"`
	Safe            string `json:"Safe"`
}

// Create creates a new account group.
// This is equivalent to Add-PASAccountGroup in psPAS.
func Create(ctx context.Context, sess *session.Session, opts CreateOptions) (*AccountGroup, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.GroupName == "" {
		return nil, fmt.Errorf("groupName is required")
	}
	if opts.GroupPlatformID == "" {
		return nil, fmt.Errorf("groupPlatformID is required")
	}
	if opts.Safe == "" {
		return nil, fmt.Errorf("safe is required")
	}

	resp, err := sess.Client.Post(ctx, "/AccountGroups", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create account group: %w", err)
	}

	var group AccountGroup
	if err := json.Unmarshal(resp.Body, &group); err != nil {
		return nil, fmt.Errorf("failed to parse account group response: %w", err)
	}

	return &group, nil
}

// GetMembers retrieves members of an account group.
// This is equivalent to Get-PASAccountGroupMember in psPAS.
func GetMembers(ctx context.Context, sess *session.Session, groupID string) ([]AccountGroupMember, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if groupID == "" {
		return nil, fmt.Errorf("groupID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/AccountGroups/%s/Members", url.PathEscape(groupID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get account group members: %w", err)
	}

	var result struct {
		Members []AccountGroupMember `json:"Members"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse account group members response: %w", err)
	}

	return result.Members, nil
}

// AddMember adds an account to an account group.
// This is equivalent to Add-PASAccountGroupMember in psPAS.
func AddMember(ctx context.Context, sess *session.Session, groupID string, accountID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if groupID == "" {
		return fmt.Errorf("groupID is required")
	}
	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	body := map[string]string{
		"AccountID": accountID,
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/AccountGroups/%s/Members", url.PathEscape(groupID)), body)
	if err != nil {
		return fmt.Errorf("failed to add account group member: %w", err)
	}

	return nil
}

// RemoveMember removes an account from an account group.
// This is equivalent to Remove-PASAccountGroupMember in psPAS.
func RemoveMember(ctx context.Context, sess *session.Session, groupID string, accountID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if groupID == "" {
		return fmt.Errorf("groupID is required")
	}
	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/AccountGroups/%s/Members/%s", url.PathEscape(groupID), url.PathEscape(accountID)))
	if err != nil {
		return fmt.Errorf("failed to remove account group member: %w", err)
	}

	return nil
}
