// Package jitaccess provides CyberArk Just-In-Time (JIT) access functionality.
// This is equivalent to the JIT access functions in psPAS including
// Request-PASJustInTimeAccess, Revoke-PASJustInTimeAccess.
package jitaccess

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// JITAccessRequest represents a JIT access request.
type JITAccessRequest struct {
	AccountID string `json:"-"` // Used in URL
	Reason    string `json:"reason,omitempty"`
	TicketingSystemName string `json:"ticketingSystemName,omitempty"`
	TicketID  types.FlexibleID `json:"ticketId,omitempty"`
}

// JITAccess represents an active JIT access session.
type JITAccess struct {
	AccountID      types.FlexibleID `json:"accountId"`
	Status         string           `json:"status"`
	RequestTime    int64            `json:"requestTime"`
	ExpirationTime int64            `json:"expirationTime,omitempty"`
	Reason         string           `json:"reason,omitempty"`
}

// RequestJITAccess requests Just-In-Time access to an account.
// This is equivalent to Request-PASJustInTimeAccess in psPAS.
func RequestJITAccess(ctx context.Context, sess *session.Session, accountID string, opts JITAccessRequest) (*JITAccess, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	body := map[string]interface{}{}
	if opts.Reason != "" {
		body["reason"] = opts.Reason
	}
	if opts.TicketingSystemName != "" {
		body["ticketingSystemName"] = opts.TicketingSystemName
	}
	if opts.TicketID != "" {
		body["ticketId"] = opts.TicketID
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/grantAdministrativeAccess", url.PathEscape(accountID)), body)
	if err != nil {
		return nil, fmt.Errorf("failed to request JIT access: %w", err)
	}

	var result JITAccess
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		// JIT access might return empty response on success
		result = JITAccess{
			AccountID: types.FlexibleID(accountID),
			Status:    "Granted",
		}
	}

	return &result, nil
}

// RevokeJITAccess revokes Just-In-Time access to an account.
// This is equivalent to Revoke-PASJustInTimeAccess in psPAS.
func RevokeJITAccess(ctx context.Context, sess *session.Session, accountID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/revokeAdministrativeAccess", url.PathEscape(accountID)), nil)
	if err != nil {
		return fmt.Errorf("failed to revoke JIT access: %w", err)
	}

	return nil
}

// JITAccessStatus represents the JIT access status for an account.
type JITAccessStatus struct {
	AccountID         string `json:"accountId"`
	IsJITEnabled      bool   `json:"isJITEnabled"`
	HasActiveAccess   bool   `json:"hasActiveAccess"`
	ExpirationTime    int64  `json:"expirationTime,omitempty"`
	RemainingDuration int    `json:"remainingDuration,omitempty"` // In seconds
}

// GetJITAccessStatus retrieves the JIT access status for an account.
func GetJITAccessStatus(ctx context.Context, sess *session.Session, accountID string) (*JITAccessStatus, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Accounts/%s/grantAdministrativeAccess", url.PathEscape(accountID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get JIT access status: %w", err)
	}

	var result JITAccessStatus
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JIT access status: %w", err)
	}
	result.AccountID = accountID

	return &result, nil
}

// EPVUserAccess represents an EPV user's access grant.
type EPVUserAccess struct {
	UserID         int    `json:"userId"`
	Username       string `json:"username"`
	AccessType     string `json:"accessType"`
	GrantedTime    int64  `json:"grantedTime"`
	ExpirationTime int64  `json:"expirationTime,omitempty"`
}

// ListEPVUserAccess lists all EPV user access grants for an account.
func ListEPVUserAccess(ctx context.Context, sess *session.Session, accountID string) ([]EPVUserAccess, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Accounts/%s/AdministrativeAccess", url.PathEscape(accountID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list EPV user access: %w", err)
	}

	var result struct {
		AccessGrants []EPVUserAccess `json:"AccessGrants"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse EPV user access: %w", err)
	}

	return result.AccessGrants, nil
}
