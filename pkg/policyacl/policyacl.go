// Package policyacl provides CyberArk policy ACL management functionality.
// This is equivalent to the PolicyACL functions in psPAS including
// Get-PASPolicyACL, Add-PASPolicyACL, Remove-PASPolicyACL.
package policyacl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
)

// PolicyACL represents a policy ACL entry.
type PolicyACL struct {
	PolicyID           string `json:"PolicyId"`
	UserName           string `json:"UserName"`
	Command            string `json:"Command,omitempty"`
	CommandGroup       bool   `json:"CommandGroup,omitempty"`
	PermissionType     string `json:"PermissionType,omitempty"`
	Restrictions       string `json:"Restrictions,omitempty"`
	IsGroup            bool   `json:"IsGroup,omitempty"`
}

// PolicyACLResponse represents the response from listing policy ACLs.
type PolicyACLResponse struct {
	PolicyACL []PolicyACL `json:"PolicyACL"`
}

// List retrieves policy ACLs.
// This is equivalent to Get-PASPolicyACL in psPAS.
func List(ctx context.Context, sess *session.Session, policyID string) ([]PolicyACL, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if policyID == "" {
		return nil, fmt.Errorf("policyID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Policy/%s/PrivilegedCommands", url.PathEscape(policyID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy ACLs: %w", err)
	}

	var result PolicyACLResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse policy ACL response: %w", err)
	}

	return result.PolicyACL, nil
}

// AddOptions holds options for adding a policy ACL.
type AddOptions struct {
	Command            string `json:"Command"`
	CommandGroup       bool   `json:"CommandGroup,omitempty"`
	PermissionType     string `json:"PermissionType,omitempty"`
	Restrictions       string `json:"Restrictions,omitempty"`
	UserName           string `json:"UserName,omitempty"`
	IsGroup            bool   `json:"IsGroup,omitempty"`
}

// Add adds a policy ACL.
// This is equivalent to Add-PASPolicyACL in psPAS.
func Add(ctx context.Context, sess *session.Session, policyID string, opts AddOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if policyID == "" {
		return fmt.Errorf("policyID is required")
	}
	if opts.Command == "" {
		return fmt.Errorf("command is required")
	}

	body := map[string]interface{}{
		"PolicyACL": opts,
	}

	_, err := sess.Client.Put(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Policy/%s/PrivilegedCommands", url.PathEscape(policyID)), body)
	if err != nil {
		return fmt.Errorf("failed to add policy ACL: %w", err)
	}

	return nil
}

// Remove removes a policy ACL.
// This is equivalent to Remove-PASPolicyACL in psPAS.
func Remove(ctx context.Context, sess *session.Session, policyID string, aclID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if policyID == "" {
		return fmt.Errorf("policyID is required")
	}
	if aclID == "" {
		return fmt.Errorf("aclID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Policy/%s/PrivilegedCommands/%s", url.PathEscape(policyID), url.PathEscape(aclID)))
	if err != nil {
		return fmt.Errorf("failed to remove policy ACL: %w", err)
	}

	return nil
}
