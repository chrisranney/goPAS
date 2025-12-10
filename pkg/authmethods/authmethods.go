// Package authmethods provides CyberArk authentication method management functionality.
// This is equivalent to the authentication method functions in psPAS including
// Get-PASAuthenticationMethod, Add-PASAuthenticationMethod, Set-PASAuthenticationMethod,
// Remove-PASAuthenticationMethod, etc.
package authmethods

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// AuthenticationMethod represents a vault authentication method.
type AuthenticationMethod struct {
	ID                       types.FlexibleID `json:"id"`
	DisplayName              string           `json:"displayName"`
	Enabled                  bool   `json:"enabled"`
	MobileEnabled            bool   `json:"mobileEnabled"`
	LogoffURL                string `json:"logoffUrl,omitempty"`
	SecondFactorAuth         string `json:"secondFactorAuth,omitempty"`
	SignInLabel              string `json:"signInLabel,omitempty"`
	UsernameFieldLabel       string `json:"usernameFieldLabel,omitempty"`
	PasswordFieldLabel       string `json:"passwordFieldLabel,omitempty"`
}

// AuthenticationMethodsResponse represents the response from listing auth methods.
type AuthenticationMethodsResponse struct {
	Methods []AuthenticationMethod `json:"Methods"`
}

// ListAuthenticationMethods retrieves all configured authentication methods.
// This is equivalent to Get-PASAuthenticationMethod in psPAS.
func ListAuthenticationMethods(ctx context.Context, sess *session.Session) ([]AuthenticationMethod, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/Configuration/AuthenticationMethods", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list authentication methods: %w", err)
	}

	var result AuthenticationMethodsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse authentication methods response: %w", err)
	}

	return result.Methods, nil
}

// GetAuthenticationMethod retrieves a specific authentication method.
func GetAuthenticationMethod(ctx context.Context, sess *session.Session, methodID string) (*AuthenticationMethod, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if methodID == "" {
		return nil, fmt.Errorf("methodID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Configuration/AuthenticationMethods/%s", url.PathEscape(methodID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication method: %w", err)
	}

	var result AuthenticationMethod
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse authentication method response: %w", err)
	}

	return &result, nil
}

// AddAuthenticationMethodOptions holds options for adding an authentication method.
type AddAuthenticationMethodOptions struct {
	ID                       types.FlexibleID `json:"id"`
	DisplayName              string           `json:"displayName"`
	Enabled                  bool   `json:"enabled"`
	MobileEnabled            bool   `json:"mobileEnabled,omitempty"`
	LogoffURL                string `json:"logoffUrl,omitempty"`
	SecondFactorAuth         string `json:"secondFactorAuth,omitempty"`
	SignInLabel              string `json:"signInLabel,omitempty"`
	UsernameFieldLabel       string `json:"usernameFieldLabel,omitempty"`
	PasswordFieldLabel       string `json:"passwordFieldLabel,omitempty"`
}

// AddAuthenticationMethod adds a new authentication method.
// This is equivalent to Add-PASAuthenticationMethod in psPAS.
func AddAuthenticationMethod(ctx context.Context, sess *session.Session, opts AddAuthenticationMethodOptions) (*AuthenticationMethod, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	if opts.DisplayName == "" {
		return nil, fmt.Errorf("displayName is required")
	}

	resp, err := sess.Client.Post(ctx, "/Configuration/AuthenticationMethods", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to add authentication method: %w", err)
	}

	var result AuthenticationMethod
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse authentication method response: %w", err)
	}

	return &result, nil
}

// UpdateAuthenticationMethodOptions holds options for updating an authentication method.
type UpdateAuthenticationMethodOptions struct {
	DisplayName              string `json:"displayName,omitempty"`
	Enabled                  *bool  `json:"enabled,omitempty"`
	MobileEnabled            *bool  `json:"mobileEnabled,omitempty"`
	LogoffURL                string `json:"logoffUrl,omitempty"`
	SecondFactorAuth         string `json:"secondFactorAuth,omitempty"`
	SignInLabel              string `json:"signInLabel,omitempty"`
	UsernameFieldLabel       string `json:"usernameFieldLabel,omitempty"`
	PasswordFieldLabel       string `json:"passwordFieldLabel,omitempty"`
}

// UpdateAuthenticationMethod updates an existing authentication method.
// This is equivalent to Set-PASAuthenticationMethod in psPAS.
func UpdateAuthenticationMethod(ctx context.Context, sess *session.Session, methodID string, opts UpdateAuthenticationMethodOptions) (*AuthenticationMethod, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if methodID == "" {
		return nil, fmt.Errorf("methodID is required")
	}

	resp, err := sess.Client.Put(ctx, fmt.Sprintf("/Configuration/AuthenticationMethods/%s", url.PathEscape(methodID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update authentication method: %w", err)
	}

	var result AuthenticationMethod
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse authentication method response: %w", err)
	}

	return &result, nil
}

// RemoveAuthenticationMethod removes an authentication method.
// This is equivalent to Remove-PASAuthenticationMethod in psPAS.
func RemoveAuthenticationMethod(ctx context.Context, sess *session.Session, methodID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if methodID == "" {
		return fmt.Errorf("methodID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Configuration/AuthenticationMethods/%s", url.PathEscape(methodID)))
	if err != nil {
		return fmt.Errorf("failed to remove authentication method: %w", err)
	}

	return nil
}

// UserAllowedAuthMethod represents an authentication method allowed for a user.
type UserAllowedAuthMethod struct {
	MethodID    types.FlexibleID `json:"authMethodId"`
	MethodName  string           `json:"authMethodDisplayName,omitempty"`
	IsEnabled   bool             `json:"isEnabled"`
}

// ListUserAllowedAuthMethods retrieves allowed authentication methods for a user.
// This is equivalent to getting user's authentication methods in psPAS.
func ListUserAllowedAuthMethods(ctx context.Context, sess *session.Session, userID int) ([]UserAllowedAuthMethod, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Users/%d/AuthenticationMethods", userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list user allowed auth methods: %w", err)
	}

	var result struct {
		AuthenticationMethods []UserAllowedAuthMethod `json:"AuthenticationMethods"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse user auth methods response: %w", err)
	}

	return result.AuthenticationMethods, nil
}

// AddUserAllowedAuthMethod adds an authentication method to user's allowed list.
// This is equivalent to Add-PASUserAllowedAuthenticationMethod in psPAS.
func AddUserAllowedAuthMethod(ctx context.Context, sess *session.Session, userID int, methodID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if methodID == "" {
		return fmt.Errorf("methodID is required")
	}

	body := map[string]string{
		"authMethodId": methodID,
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Users/%d/AuthenticationMethods", userID), body)
	if err != nil {
		return fmt.Errorf("failed to add user allowed auth method: %w", err)
	}

	return nil
}

// RemoveUserAllowedAuthMethod removes an authentication method from user's allowed list.
// This is equivalent to Remove-PASUserAllowedAuthenticationMethod in psPAS.
func RemoveUserAllowedAuthMethod(ctx context.Context, sess *session.Session, userID int, methodID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if methodID == "" {
		return fmt.Errorf("methodID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Users/%d/AuthenticationMethods/%s", userID, url.PathEscape(methodID)))
	if err != nil {
		return fmt.Errorf("failed to remove user allowed auth method: %w", err)
	}

	return nil
}

// AllowedReferrer represents an allowed referrer URL.
type AllowedReferrer struct {
	ReferrerURL     string `json:"referrerURL"`
	RegularExpression bool   `json:"regularExpression"`
}

// ListAllowedReferrers retrieves the list of allowed referrer URLs.
// This is equivalent to Get-PASAllowedReferrer in psPAS.
func ListAllowedReferrers(ctx context.Context, sess *session.Session) ([]AllowedReferrer, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/Configuration/AccessRestriction/AllowedReferrers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list allowed referrers: %w", err)
	}

	var result struct {
		AllowedReferrers []AllowedReferrer `json:"AllowedReferrers"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse allowed referrers response: %w", err)
	}

	return result.AllowedReferrers, nil
}

// AddAllowedReferrer adds an allowed referrer URL.
// This is equivalent to Add-PASAllowedReferrer in psPAS.
func AddAllowedReferrer(ctx context.Context, sess *session.Session, referrerURL string, isRegex bool) (*AllowedReferrer, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if referrerURL == "" {
		return nil, fmt.Errorf("referrerURL is required")
	}

	body := AllowedReferrer{
		ReferrerURL:       referrerURL,
		RegularExpression: isRegex,
	}

	resp, err := sess.Client.Post(ctx, "/Configuration/AccessRestriction/AllowedReferrers", body)
	if err != nil {
		return nil, fmt.Errorf("failed to add allowed referrer: %w", err)
	}

	var result AllowedReferrer
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse allowed referrer response: %w", err)
	}

	return &result, nil
}
