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
)

// OnboardingRule represents an automatic account onboarding rule.
type OnboardingRule struct {
	RuleID                int    `json:"RuleId,omitempty"`
	RuleName              string `json:"RuleName"`
	RuleDescription       string `json:"RuleDescription,omitempty"`
	TargetPlatformID      string `json:"TargetPlatformId"`
	TargetSafeName        string `json:"TargetSafeName"`
	TargetDeviceType      string `json:"TargetDeviceType,omitempty"`
	IsAdminIDFilter       bool   `json:"IsAdminIDFilter,omitempty"`
	MachineTypeFilter     string `json:"MachineTypeFilter,omitempty"`
	SystemTypeFilter      string `json:"SystemTypeFilter,omitempty"`
	UserNameFilter        string `json:"UserNameFilter,omitempty"`
	UserNameMethod        string `json:"UserNameMethod,omitempty"`
	AddressFilter         string `json:"AddressFilter,omitempty"`
	AddressMethod         string `json:"AddressMethod,omitempty"`
	AccountCategoryFilter string `json:"AccountCategoryFilter,omitempty"`
	RulePrecedence        int    `json:"RulePrecedence,omitempty"`
	ReconcileAccountID    string `json:"ReconcileAccountId,omitempty"`
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
	RuleName              string `json:"RuleName"`
	RuleDescription       string `json:"RuleDescription,omitempty"`
	TargetPlatformID      string `json:"TargetPlatformId"`
	TargetSafeName        string `json:"TargetSafeName"`
	TargetDeviceType      string `json:"TargetDeviceType,omitempty"`
	IsAdminIDFilter       bool   `json:"IsAdminIDFilter,omitempty"`
	MachineTypeFilter     string `json:"MachineTypeFilter,omitempty"`
	SystemTypeFilter      string `json:"SystemTypeFilter,omitempty"`
	UserNameFilter        string `json:"UserNameFilter,omitempty"`
	UserNameMethod        string `json:"UserNameMethod,omitempty"`
	AddressFilter         string `json:"AddressFilter,omitempty"`
	AddressMethod         string `json:"AddressMethod,omitempty"`
	AccountCategoryFilter string `json:"AccountCategoryFilter,omitempty"`
	RulePrecedence        int    `json:"RulePrecedence,omitempty"`
	ReconcileAccountID    string `json:"ReconcileAccountId,omitempty"`
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
	RuleName              string `json:"RuleName,omitempty"`
	RuleDescription       string `json:"RuleDescription,omitempty"`
	TargetPlatformID      string `json:"TargetPlatformId,omitempty"`
	TargetSafeName        string `json:"TargetSafeName,omitempty"`
	TargetDeviceType      string `json:"TargetDeviceType,omitempty"`
	IsAdminIDFilter       *bool  `json:"IsAdminIDFilter,omitempty"`
	MachineTypeFilter     string `json:"MachineTypeFilter,omitempty"`
	SystemTypeFilter      string `json:"SystemTypeFilter,omitempty"`
	UserNameFilter        string `json:"UserNameFilter,omitempty"`
	UserNameMethod        string `json:"UserNameMethod,omitempty"`
	AddressFilter         string `json:"AddressFilter,omitempty"`
	AddressMethod         string `json:"AddressMethod,omitempty"`
	AccountCategoryFilter string `json:"AccountCategoryFilter,omitempty"`
	RulePrecedence        *int   `json:"RulePrecedence,omitempty"`
	ReconcileAccountID    string `json:"ReconcileAccountId,omitempty"`
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
	ID                         string                 `json:"id,omitempty"`
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
