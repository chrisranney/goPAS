// Package accounts provides CyberArk account management functionality.
// This is equivalent to the Accounts functions in psPAS including
// Get-PASAccount, Add-PASAccount, Set-PASAccount, Remove-PASAccount, etc.
package accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/chrisranney/gopas/internal/session"
)

// Account represents a CyberArk privileged account.
type Account struct {
	ID                      string                 `json:"id"`
	Name                    string                 `json:"name"`
	Address                 string                 `json:"address"`
	UserName                string                 `json:"userName"`
	PlatformID              string                 `json:"platformId"`
	SafeName                string                 `json:"safeName"`
	SecretType              string                 `json:"secretType"`
	Secret                  string                 `json:"secret,omitempty"`
	PlatformAccountProperties map[string]interface{} `json:"platformAccountProperties,omitempty"`
	SecretManagement        *SecretManagement      `json:"secretManagement,omitempty"`
	RemoteMachinesAccess    *RemoteMachinesAccess  `json:"remoteMachinesAccess,omitempty"`
	CreatedTime             int64                  `json:"createdTime"`
	CategoryModificationTime int64                 `json:"categoryModificationTime,omitempty"`
}

// SecretManagement holds secret management settings for an account.
type SecretManagement struct {
	AutomaticManagementEnabled bool   `json:"automaticManagementEnabled"`
	ManualManagementReason     string `json:"manualManagementReason,omitempty"`
	Status                     string `json:"status,omitempty"`
	LastModifiedTime           int64  `json:"lastModifiedTime,omitempty"`
	LastReconciledTime         int64  `json:"lastReconciledTime,omitempty"`
	LastVerifiedTime           int64  `json:"lastVerifiedTime,omitempty"`
}

// RemoteMachinesAccess holds remote machines access settings.
type RemoteMachinesAccess struct {
	RemoteMachines                   string `json:"remoteMachines,omitempty"`
	AccessRestrictedToRemoteMachines bool   `json:"accessRestrictedToRemoteMachines"`
}

// AccountsResponse represents the response from listing accounts.
type AccountsResponse struct {
	Value    []Account `json:"value"`
	Count    int       `json:"count"`
	NextLink string    `json:"nextLink,omitempty"`
}

// ListOptions holds options for listing accounts.
type ListOptions struct {
	Search       string
	SearchType   string
	Sort         string
	Offset       int
	Limit        int
	Filter       string
	SafeName     string
}

// List retrieves accounts from CyberArk.
// This is equivalent to Get-PASAccount in psPAS.
func List(ctx context.Context, sess *session.Session, opts ListOptions) (*AccountsResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.Search != "" {
		params.Set("search", opts.Search)
	}
	if opts.SearchType != "" {
		params.Set("searchType", opts.SearchType)
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
	if opts.SafeName != "" {
		params.Set("filter", fmt.Sprintf("safeName eq %s", opts.SafeName))
	}

	resp, err := sess.Client.Get(ctx, "/Accounts", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	var result AccountsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse accounts response: %w", err)
	}

	return &result, nil
}

// Get retrieves a specific account by ID.
// This is equivalent to Get-PASAccount -id in psPAS.
func Get(ctx context.Context, sess *session.Session, accountID string) (*Account, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Accounts/%s", accountID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var account Account
	if err := json.Unmarshal(resp.Body, &account); err != nil {
		return nil, fmt.Errorf("failed to parse account response: %w", err)
	}

	return &account, nil
}

// CreateOptions holds options for creating an account.
type CreateOptions struct {
	Name                    string                 `json:"name,omitempty"`
	Address                 string                 `json:"address"`
	UserName                string                 `json:"userName"`
	PlatformID              string                 `json:"platformId"`
	SafeName                string                 `json:"safeName"`
	SecretType              string                 `json:"secretType,omitempty"`
	Secret                  string                 `json:"secret,omitempty"`
	PlatformAccountProperties map[string]interface{} `json:"platformAccountProperties,omitempty"`
	SecretManagement        *SecretManagement      `json:"secretManagement,omitempty"`
	RemoteMachinesAccess    *RemoteMachinesAccess  `json:"remoteMachinesAccess,omitempty"`
}

// Create creates a new account in CyberArk.
// This is equivalent to Add-PASAccount in psPAS.
func Create(ctx context.Context, sess *session.Session, opts CreateOptions) (*Account, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.SafeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}
	if opts.PlatformID == "" {
		return nil, fmt.Errorf("platformID is required")
	}
	if opts.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if opts.UserName == "" {
		return nil, fmt.Errorf("userName is required")
	}

	resp, err := sess.Client.Post(ctx, "/Accounts", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	var account Account
	if err := json.Unmarshal(resp.Body, &account); err != nil {
		return nil, fmt.Errorf("failed to parse account response: %w", err)
	}

	return &account, nil
}

// UpdateOptions holds options for updating an account.
type UpdateOptions struct {
	Name                    string                 `json:"name,omitempty"`
	Address                 string                 `json:"address,omitempty"`
	UserName                string                 `json:"userName,omitempty"`
	PlatformID              string                 `json:"platformId,omitempty"`
	PlatformAccountProperties map[string]interface{} `json:"platformAccountProperties,omitempty"`
	SecretManagement        *SecretManagement      `json:"secretManagement,omitempty"`
	RemoteMachinesAccess    *RemoteMachinesAccess  `json:"remoteMachinesAccess,omitempty"`
}

// PatchOperation represents a JSON Patch operation.
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// Update updates an existing account.
// This is equivalent to Set-PASAccount in psPAS.
func Update(ctx context.Context, sess *session.Session, accountID string, operations []PatchOperation) (*Account, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Patch(ctx, fmt.Sprintf("/Accounts/%s", accountID), operations)
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	var account Account
	if err := json.Unmarshal(resp.Body, &account); err != nil {
		return nil, fmt.Errorf("failed to parse account response: %w", err)
	}

	return &account, nil
}

// Delete removes an account from CyberArk.
// This is equivalent to Remove-PASAccount in psPAS.
func Delete(ctx context.Context, sess *session.Session, accountID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Accounts/%s", accountID))
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// GetPassword retrieves the password for an account.
// This is equivalent to Get-PASAccountPassword in psPAS.
func GetPassword(ctx context.Context, sess *session.Session, accountID string, reason string) (string, error) {
	if sess == nil || !sess.IsValid() {
		return "", fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return "", fmt.Errorf("accountID is required")
	}

	body := map[string]string{}
	if reason != "" {
		body["reason"] = reason
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/Password/Retrieve", accountID), body)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve password: %w", err)
	}

	// Response is the password as a string
	password := string(resp.Body)
	// Remove surrounding quotes if present
	if len(password) >= 2 && password[0] == '"' && password[len(password)-1] == '"' {
		password = password[1 : len(password)-1]
	}

	return password, nil
}

// ChangeCredentialsOptions holds options for changing credentials.
type ChangeCredentialsOptions struct {
	ChangeEntireGroup bool `json:"ChangeEntireGroup,omitempty"`
}

// ChangeCredentialsImmediately initiates an immediate password change.
// This is equivalent to Invoke-PASCPMOperation -ChangeImmediately in psPAS.
func ChangeCredentialsImmediately(ctx context.Context, sess *session.Session, accountID string, opts ChangeCredentialsOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/Change", accountID), opts)
	if err != nil {
		return fmt.Errorf("failed to change credentials: %w", err)
	}

	return nil
}

// VerifyCredentials initiates a credentials verification.
// This is equivalent to Invoke-PASCPMOperation -VerifyTask in psPAS.
func VerifyCredentials(ctx context.Context, sess *session.Session, accountID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/Verify", accountID), nil)
	if err != nil {
		return fmt.Errorf("failed to verify credentials: %w", err)
	}

	return nil
}

// ReconcileCredentials initiates a credentials reconciliation.
// This is equivalent to Invoke-PASCPMOperation -ReconcileTask in psPAS.
func ReconcileCredentials(ctx context.Context, sess *session.Session, accountID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/Reconcile", accountID), nil)
	if err != nil {
		return fmt.Errorf("failed to reconcile credentials: %w", err)
	}

	return nil
}

// SetNextPassword sets the next password value for an account.
// This is equivalent to Set-PASAccountPassword in psPAS.
func SetNextPassword(ctx context.Context, sess *session.Session, accountID string, newPassword string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	if newPassword == "" {
		return fmt.Errorf("newPassword is required")
	}

	body := map[string]string{
		"NewCredentials": newPassword,
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/SetNextPassword", accountID), body)
	if err != nil {
		return fmt.Errorf("failed to set next password: %w", err)
	}

	return nil
}

// AccountActivity represents account activity information.
type AccountActivity struct {
	Time       int64  `json:"Time"`
	Action     string `json:"Action"`
	ClientID   string `json:"ClientID"`
	ActionID   string `json:"ActionID"`
	Alert      bool   `json:"Alert"`
	Reason     string `json:"Reason"`
	UserName   string `json:"UserName"`
}

// GetActivities retrieves the activity log for an account.
// This is equivalent to Get-PASAccountActivity in psPAS.
func GetActivities(ctx context.Context, sess *session.Session, accountID string) ([]AccountActivity, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Accounts/%s/Activities", accountID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get account activities: %w", err)
	}

	var result struct {
		Activities []AccountActivity `json:"Activities"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse activities response: %w", err)
	}

	return result.Activities, nil
}

// GetCreatedTime returns the account's creation time as time.Time.
func (a *Account) GetCreatedTime() time.Time {
	return time.Unix(a.CreatedTime, 0)
}
