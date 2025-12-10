// Package safemembers provides CyberArk safe member management functionality.
// This is equivalent to the SafeMembers functions in psPAS including
// Get-PASSafeMember, Add-PASSafeMember, Set-PASSafeMember, Remove-PASSafeMember, etc.
package safemembers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// SafeMember represents a safe member.
type SafeMember struct {
	SafeURLID                 types.FlexibleID `json:"safeUrlId,omitempty"`
	SafeName                  string           `json:"safeName,omitempty"`
	SafeNumber                int              `json:"safeNumber,omitempty"`
	MemberID                  types.FlexibleID `json:"memberId,omitempty"`
	MemberName                string       `json:"memberName"`
	MemberType                string       `json:"memberType,omitempty"`
	MembershipExpirationDate  int64        `json:"membershipExpirationDate,omitempty"`
	IsExpiredMembershipEnable bool         `json:"isExpiredMembershipEnable,omitempty"`
	IsPredefinedUser          bool         `json:"isPredefinedUser,omitempty"`
	IsReadOnly                bool         `json:"isReadOnly,omitempty"`
	Permissions               *Permissions `json:"permissions"`
}

// Permissions represents the permissions for a safe member.
type Permissions struct {
	UseAccounts                            bool `json:"useAccounts"`
	RetrieveAccounts                       bool `json:"retrieveAccounts"`
	ListAccounts                           bool `json:"listAccounts"`
	AddAccounts                            bool `json:"addAccounts"`
	UpdateAccountContent                   bool `json:"updateAccountContent"`
	UpdateAccountProperties                bool `json:"updateAccountProperties"`
	InitiateCPMAccountManagementOperations bool `json:"initiateCPMAccountManagementOperations"`
	SpecifyNextAccountContent              bool `json:"specifyNextAccountContent"`
	RenameAccounts                         bool `json:"renameAccounts"`
	DeleteAccounts                         bool `json:"deleteAccounts"`
	UnlockAccounts                         bool `json:"unlockAccounts"`
	ManageSafe                             bool `json:"manageSafe"`
	ManageSafeMembers                      bool `json:"manageSafeMembers"`
	BackupSafe                             bool `json:"backupSafe"`
	ViewAuditLog                           bool `json:"viewAuditLog"`
	ViewSafeMembers                        bool `json:"viewSafeMembers"`
	AccessWithoutConfirmation              bool `json:"accessWithoutConfirmation"`
	CreateFolders                          bool `json:"createFolders"`
	DeleteFolders                          bool `json:"deleteFolders"`
	MoveAccountsAndFolders                 bool `json:"moveAccountsAndFolders"`
	RequestsAuthorizationLevel1            bool `json:"requestsAuthorizationLevel1"`
	RequestsAuthorizationLevel2            bool `json:"requestsAuthorizationLevel2"`
}

// SafeMembersResponse represents the response from listing safe members.
type SafeMembersResponse struct {
	Value    []SafeMember `json:"value"`
	Count    int          `json:"count"`
	NextLink string       `json:"nextLink,omitempty"`
}

// ListOptions holds options for listing safe members.
type ListOptions struct {
	Search string
	Sort   string
	Offset int
	Limit  int
	Filter string
}

// List retrieves safe members.
// This is equivalent to Get-PASSafeMember in psPAS.
func List(ctx context.Context, sess *session.Session, safeName string, opts ListOptions) (*SafeMembersResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return nil, fmt.Errorf("safeName is required")
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
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Safes/%s/Members", url.PathEscape(safeName)), params)
	if err != nil {
		return nil, fmt.Errorf("failed to list safe members: %w", err)
	}

	var result SafeMembersResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse safe members response: %w", err)
	}

	return &result, nil
}

// Get retrieves a specific safe member.
func Get(ctx context.Context, sess *session.Session, safeName string, memberName string) (*SafeMember, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	if memberName == "" {
		return nil, fmt.Errorf("memberName is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Safes/%s/Members/%s", url.PathEscape(safeName), url.PathEscape(memberName)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get safe member: %w", err)
	}

	var member SafeMember
	if err := json.Unmarshal(resp.Body, &member); err != nil {
		return nil, fmt.Errorf("failed to parse safe member response: %w", err)
	}

	return &member, nil
}

// AddOptions holds options for adding a safe member.
type AddOptions struct {
	MemberName               string       `json:"memberName"`
	SearchIn                 string       `json:"searchIn,omitempty"`
	MembershipExpirationDate int64        `json:"membershipExpirationDate,omitempty"`
	Permissions              *Permissions `json:"permissions"`
}

// Add adds a member to a safe.
// This is equivalent to Add-PASSafeMember in psPAS.
func Add(ctx context.Context, sess *session.Session, safeName string, opts AddOptions) (*SafeMember, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	if opts.MemberName == "" {
		return nil, fmt.Errorf("memberName is required")
	}

	if opts.Permissions == nil {
		return nil, fmt.Errorf("permissions are required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Safes/%s/Members", url.PathEscape(safeName)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to add safe member: %w", err)
	}

	var member SafeMember
	if err := json.Unmarshal(resp.Body, &member); err != nil {
		return nil, fmt.Errorf("failed to parse safe member response: %w", err)
	}

	return &member, nil
}

// UpdateOptions holds options for updating a safe member.
type UpdateOptions struct {
	MembershipExpirationDate int64        `json:"membershipExpirationDate,omitempty"`
	Permissions              *Permissions `json:"permissions"`
}

// Update updates a safe member.
// This is equivalent to Set-PASSafeMember in psPAS.
func Update(ctx context.Context, sess *session.Session, safeName string, memberName string, opts UpdateOptions) (*SafeMember, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	if memberName == "" {
		return nil, fmt.Errorf("memberName is required")
	}

	resp, err := sess.Client.Put(ctx, fmt.Sprintf("/Safes/%s/Members/%s", url.PathEscape(safeName), url.PathEscape(memberName)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update safe member: %w", err)
	}

	var member SafeMember
	if err := json.Unmarshal(resp.Body, &member); err != nil {
		return nil, fmt.Errorf("failed to parse safe member response: %w", err)
	}

	return &member, nil
}

// Remove removes a member from a safe.
// This is equivalent to Remove-PASSafeMember in psPAS.
func Remove(ctx context.Context, sess *session.Session, safeName string, memberName string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if safeName == "" {
		return fmt.Errorf("safeName is required")
	}

	if memberName == "" {
		return fmt.Errorf("memberName is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Safes/%s/Members/%s", url.PathEscape(safeName), url.PathEscape(memberName)))
	if err != nil {
		return fmt.Errorf("failed to remove safe member: %w", err)
	}

	return nil
}

// DefaultUserPermissions returns the default permissions for a regular user.
func DefaultUserPermissions() *Permissions {
	return &Permissions{
		UseAccounts:      true,
		RetrieveAccounts: true,
		ListAccounts:     true,
		ViewSafeMembers:  true,
	}
}

// DefaultAdminPermissions returns the default permissions for an admin.
func DefaultAdminPermissions() *Permissions {
	return &Permissions{
		UseAccounts:                            true,
		RetrieveAccounts:                       true,
		ListAccounts:                           true,
		AddAccounts:                            true,
		UpdateAccountContent:                   true,
		UpdateAccountProperties:                true,
		InitiateCPMAccountManagementOperations: true,
		SpecifyNextAccountContent:              true,
		RenameAccounts:                         true,
		DeleteAccounts:                         true,
		UnlockAccounts:                         true,
		ManageSafe:                             true,
		ManageSafeMembers:                      true,
		BackupSafe:                             true,
		ViewAuditLog:                           true,
		ViewSafeMembers:                        true,
		AccessWithoutConfirmation:              true,
		CreateFolders:                          true,
		DeleteFolders:                          true,
		MoveAccountsAndFolders:                 true,
	}
}
