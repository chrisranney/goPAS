// Package onboardingrules provides CyberArk automatic onboarding rules functionality.
// This is equivalent to the OnboardingRules functions in psPAS including
// Get-PASOnboardingRule, New-PASOnboardingRule, Set-PASOnboardingRule, etc.
package onboardingrules

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// OnboardingRule represents an automatic account onboarding rule.
type OnboardingRule struct {
	RuleID                int              `json:"RuleId,omitempty"`
	RuleName              string           `json:"RuleName"`
	RuleDescription       string           `json:"RuleDescription,omitempty"`
	TargetPlatformID      types.FlexibleID `json:"TargetPlatformId"`
	TargetSafeName        string           `json:"TargetSafeName"`
	TargetDeviceType      string           `json:"TargetDeviceType,omitempty"`
	IsAdminIDFilter       bool             `json:"IsAdminIDFilter,omitempty"`
	MachineTypeFilter     string           `json:"MachineTypeFilter,omitempty"`
	SystemTypeFilter      string           `json:"SystemTypeFilter,omitempty"`
	UserNameFilter        string           `json:"UserNameFilter,omitempty"`
	UserNameMethod        string           `json:"UserNameMethod,omitempty"`
	AddressFilter         string           `json:"AddressFilter,omitempty"`
	AddressMethod         string           `json:"AddressMethod,omitempty"`
	AccountCategoryFilter string           `json:"AccountCategoryFilter,omitempty"`
	RulePrecedence        int              `json:"RulePrecedence,omitempty"`
	ReconcileAccountID    types.FlexibleID `json:"ReconcileAccountId,omitempty"`
}

// OnboardingRulesResponse represents the response from listing onboarding rules.
type OnboardingRulesResponse struct {
	AutomaticOnboardingRules []OnboardingRule `json:"AutomaticOnboardingRules"`
}

// List retrieves automatic onboarding rules.
// This is equivalent to Get-PASOnboardingRule in psPAS.
func List(ctx context.Context, sess *session.Session) ([]OnboardingRule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/AutomaticOnboardingRules", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list onboarding rules: %w", err)
	}

	var result OnboardingRulesResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse onboarding rules response: %w", err)
	}

	return result.AutomaticOnboardingRules, nil
}

// Get retrieves a specific onboarding rule.
func Get(ctx context.Context, sess *session.Session, ruleID int) (*OnboardingRule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/AutomaticOnboardingRules/%d", ruleID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get onboarding rule: %w", err)
	}

	var rule OnboardingRule
	if err := json.Unmarshal(resp.Body, &rule); err != nil {
		return nil, fmt.Errorf("failed to parse onboarding rule response: %w", err)
	}

	return &rule, nil
}

// CreateOptions holds options for creating an onboarding rule.
type CreateOptions struct {
	RuleName              string           `json:"RuleName"`
	RuleDescription       string           `json:"RuleDescription,omitempty"`
	TargetPlatformID      types.FlexibleID `json:"TargetPlatformId"`
	TargetSafeName        string           `json:"TargetSafeName"`
	TargetDeviceType      string           `json:"TargetDeviceType,omitempty"`
	IsAdminIDFilter       bool             `json:"IsAdminIDFilter,omitempty"`
	MachineTypeFilter     string           `json:"MachineTypeFilter,omitempty"`
	SystemTypeFilter      string           `json:"SystemTypeFilter,omitempty"`
	UserNameFilter        string           `json:"UserNameFilter,omitempty"`
	UserNameMethod        string           `json:"UserNameMethod,omitempty"`
	AddressFilter         string           `json:"AddressFilter,omitempty"`
	AddressMethod         string           `json:"AddressMethod,omitempty"`
	AccountCategoryFilter string           `json:"AccountCategoryFilter,omitempty"`
	RulePrecedence        int              `json:"RulePrecedence,omitempty"`
	ReconcileAccountID    types.FlexibleID `json:"ReconcileAccountId,omitempty"`
}

// Create creates a new onboarding rule.
// This is equivalent to New-PASOnboardingRule in psPAS.
func Create(ctx context.Context, sess *session.Session, opts CreateOptions) (*OnboardingRule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.RuleName == "" {
		return nil, fmt.Errorf("ruleName is required")
	}
	if opts.TargetPlatformID == "" {
		return nil, fmt.Errorf("targetPlatformID is required")
	}
	if opts.TargetSafeName == "" {
		return nil, fmt.Errorf("targetSafeName is required")
	}

	resp, err := sess.Client.Post(ctx, "/AutomaticOnboardingRules", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create onboarding rule: %w", err)
	}

	var rule OnboardingRule
	if err := json.Unmarshal(resp.Body, &rule); err != nil {
		return nil, fmt.Errorf("failed to parse onboarding rule response: %w", err)
	}

	return &rule, nil
}

// UpdateOptions holds options for updating an onboarding rule.
type UpdateOptions struct {
	RuleName              string           `json:"RuleName,omitempty"`
	RuleDescription       string           `json:"RuleDescription,omitempty"`
	TargetPlatformID      types.FlexibleID `json:"TargetPlatformId,omitempty"`
	TargetSafeName        string           `json:"TargetSafeName,omitempty"`
	TargetDeviceType      string           `json:"TargetDeviceType,omitempty"`
	IsAdminIDFilter       *bool            `json:"IsAdminIDFilter,omitempty"`
	MachineTypeFilter     string           `json:"MachineTypeFilter,omitempty"`
	SystemTypeFilter      string           `json:"SystemTypeFilter,omitempty"`
	UserNameFilter        string           `json:"UserNameFilter,omitempty"`
	UserNameMethod        string           `json:"UserNameMethod,omitempty"`
	AddressFilter         string           `json:"AddressFilter,omitempty"`
	AddressMethod         string           `json:"AddressMethod,omitempty"`
	AccountCategoryFilter string           `json:"AccountCategoryFilter,omitempty"`
	RulePrecedence        *int             `json:"RulePrecedence,omitempty"`
	ReconcileAccountID    types.FlexibleID `json:"ReconcileAccountId,omitempty"`
}

// Update updates an onboarding rule.
// This is equivalent to Set-PASOnboardingRule in psPAS.
func Update(ctx context.Context, sess *session.Session, ruleID int, opts UpdateOptions) (*OnboardingRule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Put(ctx, fmt.Sprintf("/AutomaticOnboardingRules/%d", ruleID), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update onboarding rule: %w", err)
	}

	var rule OnboardingRule
	if err := json.Unmarshal(resp.Body, &rule); err != nil {
		return nil, fmt.Errorf("failed to parse onboarding rule response: %w", err)
	}

	return &rule, nil
}

// Delete removes an onboarding rule.
// This is equivalent to Remove-PASOnboardingRule in psPAS.
func Delete(ctx context.Context, sess *session.Session, ruleID int) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/AutomaticOnboardingRules/%d", ruleID))
	if err != nil {
		return fmt.Errorf("failed to delete onboarding rule: %w", err)
	}

	return nil
}

// DiscoveredAccount represents a discovered account.
type DiscoveredAccount struct {
	ID                         types.FlexibleID       `json:"id,omitempty"`
	UserName                   string                 `json:"userName"`
	Address                    string                 `json:"address"`
	DiscoveryDateTime          int64                  `json:"discoveryDateTime,omitempty"`
	AccountEnabled             bool                   `json:"accountEnabled,omitempty"`
	OsGroups                   string                 `json:"osGroups,omitempty"`
	PlatformType               string                 `json:"platformType,omitempty"`
	Domain                     string                 `json:"domain,omitempty"`
	LastLogonDateTime          int64                  `json:"lastLogonDateTime,omitempty"`
	LastPasswordSetDateTime    int64                  `json:"lastPasswordSetDateTime,omitempty"`
	PasswordNeverExpires       bool                   `json:"passwordNeverExpires,omitempty"`
	OSVersion                  string                 `json:"osVersion,omitempty"`
	Privileged                 bool                   `json:"privileged,omitempty"`
	UserDisplayName            string                 `json:"userDisplayName,omitempty"`
	Description                string                 `json:"description,omitempty"`
	PasswordExpirationDateTime int64                  `json:"passwordExpirationDateTime,omitempty"`
	OU                         string                 `json:"ou,omitempty"`
	Dependencies               []DiscoveredDependency `json:"dependencies,omitempty"`
}

// DiscoveredDependency represents a dependency of a discovered account.
type DiscoveredDependency struct {
	Name    string `json:"name"`
	Address string `json:"address,omitempty"`
	Type    string `json:"type,omitempty"`
}

// DiscoveredAccountsResponse represents the response from listing discovered accounts.
type DiscoveredAccountsResponse struct {
	Value    []DiscoveredAccount `json:"value"`
	Count    int                 `json:"count"`
	NextLink string              `json:"nextLink,omitempty"`
}

// ListDiscoveredOptions holds options for listing discovered accounts.
type ListDiscoveredOptions struct {
	Search string
	Offset int
	Limit  int
	Filter string
}

// ListDiscoveredAccounts retrieves discovered accounts.
// This is equivalent to Get-PASDiscoveredAccount in psPAS.
func ListDiscoveredAccounts(ctx context.Context, sess *session.Session, opts ListDiscoveredOptions) (*DiscoveredAccountsResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.Search != "" {
		params.Set("search", opts.Search)
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	resp, err := sess.Client.Get(ctx, "/DiscoveredAccounts", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list discovered accounts: %w", err)
	}

	var result DiscoveredAccountsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse discovered accounts response: %w", err)
	}

	return &result, nil
}

// GetDiscoveredAccount retrieves a specific discovered account by ID.
func GetDiscoveredAccount(ctx context.Context, sess *session.Session, accountID string) (*DiscoveredAccount, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/DiscoveredAccounts/%s", url.PathEscape(accountID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get discovered account: %w", err)
	}

	var account DiscoveredAccount
	if err := json.Unmarshal(resp.Body, &account); err != nil {
		return nil, fmt.Errorf("failed to parse discovered account response: %w", err)
	}

	return &account, nil
}

// AddDiscoveredAccountOptions holds options for adding a discovered account.
type AddDiscoveredAccountOptions struct {
	UserName                   string                 `json:"userName"`
	Address                    string                 `json:"address"`
	DiscoveryDateTime          int64                  `json:"discoveryDate,omitempty"`
	AccountEnabled             *bool                  `json:"accountEnabled,omitempty"`
	OsGroups                   string                 `json:"osGroups,omitempty"`
	PlatformType               string                 `json:"platformType,omitempty"`
	Domain                     string                 `json:"domain,omitempty"`
	LastLogonDateTime          int64                  `json:"lastLogonDateTime,omitempty"`
	LastPasswordSetDateTime    int64                  `json:"lastPasswordSetDateTime,omitempty"`
	PasswordNeverExpires       *bool                  `json:"passwordNeverExpires,omitempty"`
	OSVersion                  string                 `json:"osVersion,omitempty"`
	Privileged                 *bool                  `json:"privileged,omitempty"`
	UserDisplayName            string                 `json:"userDisplayName,omitempty"`
	Description                string                 `json:"description,omitempty"`
	PasswordExpirationDateTime int64                  `json:"passwordExpirationDateTime,omitempty"`
	OU                         string                 `json:"organizationalUnit,omitempty"`
	Dependencies               []DiscoveredDependency `json:"Dependencies,omitempty"`
}

// AddDiscoveredAccount adds an account to the discovered accounts list.
// This is equivalent to Add-PASDiscoveredAccount in psPAS.
func AddDiscoveredAccount(ctx context.Context, sess *session.Session, opts AddDiscoveredAccountOptions) (*DiscoveredAccount, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.UserName == "" {
		return nil, fmt.Errorf("userName is required")
	}

	if opts.Address == "" {
		return nil, fmt.Errorf("address is required")
	}

	resp, err := sess.Client.Post(ctx, "/DiscoveredAccounts", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to add discovered account: %w", err)
	}

	var account DiscoveredAccount
	if err := json.Unmarshal(resp.Body, &account); err != nil {
		return nil, fmt.Errorf("failed to parse discovered account response: %w", err)
	}

	return &account, nil
}

// DeleteDiscoveredAccount removes a discovered account from the list.
func DeleteDiscoveredAccount(ctx context.Context, sess *session.Session, accountID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/DiscoveredAccounts/%s", url.PathEscape(accountID)))
	if err != nil {
		return fmt.Errorf("failed to delete discovered account: %w", err)
	}

	return nil
}

// ClearDiscoveredAccountsOptions holds options for clearing discovered accounts.
type ClearDiscoveredAccountsOptions struct {
	DiscoverySource string `json:"discoverySource,omitempty"` // Discovery source to filter by
}

// ClearDiscoveredAccounts clears all discovered accounts from the list.
// This is equivalent to Clear-PASDiscoveredAccountList in psPAS.
func ClearDiscoveredAccounts(ctx context.Context, sess *session.Session, opts ClearDiscoveredAccountsOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	_, err := sess.Client.Delete(ctx, "/DiscoveredAccounts")
	if err != nil {
		return fmt.Errorf("failed to clear discovered accounts: %w", err)
	}

	return nil
}

// PublishDiscoveredAccountOptions holds options for publishing (onboarding) a discovered account.
type PublishDiscoveredAccountOptions struct {
	AccountID        string           `json:"-"` // Used in URL
	SafeName         string           `json:"safeName"`
	PlatformID       types.FlexibleID `json:"platformId,omitempty"`
	Secret           string `json:"secret,omitempty"`
	SecretType       string `json:"secretType,omitempty"`
	AutomaticManagement *bool  `json:"automaticManagement,omitempty"`
	ManualManagementReason string `json:"manualManagementReason,omitempty"`
}

// PublishedAccount represents the result of publishing a discovered account.
type PublishedAccount struct {
	ID         types.FlexibleID `json:"id"`
	Name       string           `json:"name,omitempty"`
	SafeName   string           `json:"safeName"`
	PlatformID types.FlexibleID `json:"platformId"`
}

// PublishDiscoveredAccount onboards a discovered account to CyberArk vault.
// This is equivalent to Publish-PASDiscoveredAccount in psPAS.
func PublishDiscoveredAccount(ctx context.Context, sess *session.Session, opts PublishDiscoveredAccountOptions) (*PublishedAccount, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.AccountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	if opts.SafeName == "" {
		return nil, fmt.Errorf("safeName is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/DiscoveredAccounts/%s/Onboard", url.PathEscape(opts.AccountID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to publish discovered account: %w", err)
	}

	var result PublishedAccount
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse published account response: %w", err)
	}

	return &result, nil
}
