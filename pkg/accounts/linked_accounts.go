// Package accounts provides linked accounts functionality.
// This is equivalent to Add-PASAccountLinking, Remove-PASAccountLinking in psPAS.
package accounts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chrisranney/gopas/internal/session"
)

// LinkedAccount represents a linked account.
type LinkedAccount struct {
	ID         string `json:"ID"`
	Name       string `json:"Name"`
	SafeName   string `json:"SafeName"`
	ExtraPass1 string `json:"ExtraPass1,omitempty"`
	ExtraPass2 string `json:"ExtraPass2,omitempty"`
	ExtraPass3 string `json:"ExtraPass3,omitempty"`
}

// LinkAccountOptions holds options for linking accounts.
type LinkAccountOptions struct {
	// Safe is the safe name of the linked account
	Safe string `json:"safe"`
	// ExtraPassID is the linked account ID (1, 2, or 3)
	ExtraPassID int `json:"extraPasswordIndex"`
	// Name is the linked account name (only for legacy linking)
	Name string `json:"name,omitempty"`
	// Folder is the linked account folder (only for legacy linking)
	Folder string `json:"folder,omitempty"`
}

// LinkAccount links an account to another account.
// This is equivalent to Add-PASAccountLinking in psPAS.
func LinkAccount(ctx context.Context, sess *session.Session, accountID string, linkedAccountID string, opts LinkAccountOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	if linkedAccountID == "" {
		return fmt.Errorf("linkedAccountID is required")
	}

	body := map[string]interface{}{
		"safe":               opts.Safe,
		"extraPasswordIndex": opts.ExtraPassID,
		"linkedAccountId":    linkedAccountID,
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/LinkAccount", accountID), body)
	if err != nil {
		return fmt.Errorf("failed to link account: %w", err)
	}

	return nil
}

// UnlinkAccount unlinks a linked account.
// This is equivalent to Remove-PASAccountLinking in psPAS.
func UnlinkAccount(ctx context.Context, sess *session.Session, accountID string, extraPassID int) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	if extraPassID < 1 || extraPassID > 3 {
		return fmt.Errorf("extraPassID must be 1, 2, or 3")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Accounts/%s/LinkAccount/%d", accountID, extraPassID))
	if err != nil {
		return fmt.Errorf("failed to unlink account: %w", err)
	}

	return nil
}

// GetLinkedAccounts retrieves the linked accounts for an account.
func GetLinkedAccounts(ctx context.Context, sess *session.Session, accountID string) ([]LinkedAccount, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Accounts/%s/LinkAccount", accountID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked accounts: %w", err)
	}

	var result struct {
		Accounts []LinkedAccount `json:"LinkedAccounts"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse linked accounts response: %w", err)
	}

	return result.Accounts, nil
}
