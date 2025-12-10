// Package sshkeys provides CyberArk SSH key management functionality.
// This is equivalent to the SSH key functions in psPAS including
// Add-PASPublicSSHKey, Get-PASPublicSSHKey, Remove-PASPublicSSHKey,
// Get-PASAccountSSHKey, New-PASPrivateSSHKey, Remove-PASPrivateSSHKey, etc.
package sshkeys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// PublicSSHKey represents a user's public SSH key.
type PublicSSHKey struct {
	KeyID        types.FlexibleID `json:"KeyID"`
	PublicSSHKey string           `json:"PublicSSHKey"`
}

// PublicSSHKeysResponse represents the response from listing public SSH keys.
type PublicSSHKeysResponse struct {
	PublicSSHKeys []PublicSSHKey `json:"PublicSSHKeys"`
}

// GetUserPublicSSHKeys retrieves public SSH keys for a user.
// This is equivalent to Get-PASPublicSSHKey in psPAS.
func GetUserPublicSSHKeys(ctx context.Context, sess *session.Session, userID string) ([]PublicSSHKey, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Users/%s/Secret/SSHKeys", url.PathEscape(userID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user SSH keys: %w", err)
	}

	var result PublicSSHKeysResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse SSH keys response: %w", err)
	}

	return result.PublicSSHKeys, nil
}

// AddUserPublicSSHKey adds a public SSH key to a user.
// This is equivalent to Add-PASPublicSSHKey in psPAS.
func AddUserPublicSSHKey(ctx context.Context, sess *session.Session, userID string, publicKey string) (*PublicSSHKey, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	if publicKey == "" {
		return nil, fmt.Errorf("publicKey is required")
	}

	body := map[string]string{
		"PublicSSHKey": publicKey,
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Users/%s/Secret/SSHKeys", url.PathEscape(userID)), body)
	if err != nil {
		return nil, fmt.Errorf("failed to add user SSH key: %w", err)
	}

	var result PublicSSHKey
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse SSH key response: %w", err)
	}

	return &result, nil
}

// RemoveUserPublicSSHKey removes a public SSH key from a user.
// This is equivalent to Remove-PASPublicSSHKey in psPAS.
func RemoveUserPublicSSHKey(ctx context.Context, sess *session.Session, userID string, keyID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	if keyID == "" {
		return fmt.Errorf("keyID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Users/%s/Secret/SSHKeys/%s", url.PathEscape(userID), url.PathEscape(keyID)))
	if err != nil {
		return fmt.Errorf("failed to remove user SSH key: %w", err)
	}

	return nil
}

// AccountSSHKey represents an account's SSH private key.
type AccountSSHKey struct {
	PrivateSSHKey string `json:"PrivateSSHKey"`
	Passphrase    string `json:"Passphrase,omitempty"`
}

// GetAccountSSHKeyOptions holds options for retrieving account SSH keys.
type GetAccountSSHKeyOptions struct {
	Reason               string `json:"reason,omitempty"`
	TicketingSystemName  string `json:"TicketingSystemName,omitempty"`
	TicketID             string `json:"TicketId,omitempty"`
	Version              int    `json:"Version,omitempty"`
	ActionType           string `json:"ActionType,omitempty"`
	IsUse                bool   `json:"isUse,omitempty"`
	Machine              string `json:"Machine,omitempty"`
}

// GetAccountSSHKey retrieves the SSH private key from an account.
// This is equivalent to Get-PASAccountSSHKey in psPAS.
func GetAccountSSHKey(ctx context.Context, sess *session.Session, accountID string, opts GetAccountSSHKeyOptions) (*AccountSSHKey, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/Secret/Retrieve", url.PathEscape(accountID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get account SSH key: %w", err)
	}

	var result AccountSSHKey
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse SSH key response: %w", err)
	}

	return &result, nil
}

// PrivateSSHKey represents a private SSH key managed by CyberArk.
type PrivateSSHKey struct {
	ID             types.FlexibleID `json:"id"`
	UserID         types.FlexibleID `json:"userId"`
	Format         string           `json:"format,omitempty"`
	KeyAlgorithm   string           `json:"keyAlgorithm,omitempty"`
	KeySize        int              `json:"keySize,omitempty"`
	CreationTime   int64            `json:"creationTime,omitempty"`
	ExpirationTime int64            `json:"expirationTime,omitempty"`
}

// GeneratePrivateSSHKeyOptions holds options for generating a private SSH key.
type GeneratePrivateSSHKeyOptions struct {
	Format       string `json:"format,omitempty"`       // OpenSSH or PEM
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"` // RSA, DSA, ECDSA, ED25519
	KeySize      int    `json:"keySize,omitempty"`      // Key size in bits
}

// GeneratePrivateSSHKey generates a new private SSH key for a user.
// This is equivalent to New-PASPrivateSSHKey in psPAS.
func GeneratePrivateSSHKey(ctx context.Context, sess *session.Session, userID string, opts GeneratePrivateSSHKeyOptions) (*PrivateSSHKey, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Users/%s/Secret/SSHKeys/Cache", url.PathEscape(userID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private SSH key: %w", err)
	}

	var result PrivateSSHKey
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse private SSH key response: %w", err)
	}

	return &result, nil
}

// RemovePrivateSSHKey removes a specific private SSH key for a user.
// This is equivalent to Remove-PASPrivateSSHKey in psPAS.
func RemovePrivateSSHKey(ctx context.Context, sess *session.Session, userID string, keyID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	if keyID == "" {
		return fmt.Errorf("keyID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Users/%s/Secret/SSHKeys/Cache/%s", url.PathEscape(userID), url.PathEscape(keyID)))
	if err != nil {
		return fmt.Errorf("failed to remove private SSH key: %w", err)
	}

	return nil
}

// ClearPrivateSSHKeys removes all cached private SSH keys for a user.
// This is equivalent to Clear-PASPrivateSSHKey in psPAS.
func ClearPrivateSSHKeys(ctx context.Context, sess *session.Session, userID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Users/%s/Secret/SSHKeys/ClearCache", url.PathEscape(userID)))
	if err != nil {
		return fmt.Errorf("failed to clear private SSH keys: %w", err)
	}

	return nil
}

// MFACachedSSHKey represents an MFA-cached SSH key.
type MFACachedSSHKey struct {
	ID                types.FlexibleID `json:"id"`
	CacheCreationTime int64            `json:"cacheCreationTime"`
	ExpirationTime    int64            `json:"expirationTime"`
}

// ListMFACachedSSHKeys retrieves MFA-cached SSH keys for a user.
func ListMFACachedSSHKeys(ctx context.Context, sess *session.Session, userID string) ([]MFACachedSSHKey, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Users/%s/Secret/SSHKeys/Cache", url.PathEscape(userID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list MFA cached SSH keys: %w", err)
	}

	var result struct {
		CachedSSHKeys []MFACachedSSHKey `json:"CachedSSHKeys"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse cached SSH keys response: %w", err)
	}

	return result.CachedSSHKeys, nil
}
