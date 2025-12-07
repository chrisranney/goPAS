// Package accountacl provides CyberArk account ACL management functionality.
// This is equivalent to the AccountACL functions in psPAS including
// Get-PASAccountACL, Add-PASAccountACL, Remove-PASAccountACL.
package accountacl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
)

// AccountACL represents an account ACL entry.
type AccountACL struct {
	VaultUserName  string `json:"VaultUserName"`
	SafeName       string `json:"SafeName,omitempty"`
	FolderName     string `json:"FolderName,omitempty"`
	ObjectName     string `json:"ObjectName,omitempty"`
	Command        string `json:"Command,omitempty"`
	CommandGroup   bool   `json:"CommandGroup,omitempty"`
	PermissionType string `json:"PermissionType,omitempty"`
	Restrictions   string `json:"Restrictions,omitempty"`
	IsGroup        bool   `json:"IsGroup,omitempty"`
	UserName       string `json:"UserName,omitempty"`
}

// AccountACLResponse represents the response from listing account ACLs.
type AccountACLResponse struct {
	ListAccountPrivilegedCommandsResult []AccountACL `json:"ListAccountPrivilegedCommandsResult"`
}

// List retrieves account ACLs.
// This is equivalent to Get-PASAccountACL in psPAS.
func List(ctx context.Context, sess *session.Session, accountID string, safeName string, folderName string) ([]AccountACL, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}
	if safeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	folder := folderName
	if folder == "" {
		folder = "Root"
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Account/%s|%s|%s/PrivilegedCommands",
		url.PathEscape(safeName), url.PathEscape(folder), url.PathEscape(accountID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get account ACLs: %w", err)
	}

	var result AccountACLResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse account ACL response: %w", err)
	}

	return result.ListAccountPrivilegedCommandsResult, nil
}

// AddOptions holds options for adding an account ACL.
type AddOptions struct {
	Command        string `json:"Command"`
	CommandGroup   bool   `json:"CommandGroup,omitempty"`
	PermissionType string `json:"PermissionType,omitempty"`
	Restrictions   string `json:"Restrictions,omitempty"`
	UserName       string `json:"UserName,omitempty"`
}

// Add adds an account ACL.
// This is equivalent to Add-PASAccountACL in psPAS.
func Add(ctx context.Context, sess *session.Session, accountID string, safeName string, folderName string, opts AddOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}
	if safeName == "" {
		return fmt.Errorf("safeName is required")
	}
	if opts.Command == "" {
		return fmt.Errorf("command is required")
	}

	folder := folderName
	if folder == "" {
		folder = "Root"
	}

	body := map[string]interface{}{
		"PrivilegedCommand": opts,
	}

	_, err := sess.Client.Put(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Account/%s|%s|%s/PrivilegedCommands",
		url.PathEscape(safeName), url.PathEscape(folder), url.PathEscape(accountID)), body)
	if err != nil {
		return fmt.Errorf("failed to add account ACL: %w", err)
	}

	return nil
}

// Remove removes an account ACL.
// This is equivalent to Remove-PASAccountACL in psPAS.
func Remove(ctx context.Context, sess *session.Session, accountID string, safeName string, folderName string, aclID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}
	if safeName == "" {
		return fmt.Errorf("safeName is required")
	}
	if aclID == "" {
		return fmt.Errorf("aclID is required")
	}

	folder := folderName
	if folder == "" {
		folder = "Root"
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Account/%s|%s|%s/PrivilegedCommands/%s",
		url.PathEscape(safeName), url.PathEscape(folder), url.PathEscape(accountID), url.PathEscape(aclID)))
	if err != nil {
		return fmt.Errorf("failed to remove account ACL: %w", err)
	}

	return nil
}
