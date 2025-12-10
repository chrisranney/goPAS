// Package requests provides CyberArk access request functionality.
// This is equivalent to the Requests functions in psPAS including
// Get-PASRequest, New-PASRequest, Approve-PASRequest, Deny-PASRequest, etc.
package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// Request represents an access request.
type Request struct {
	RequestID                types.FlexibleID       `json:"RequestID"`
	SafeName                 string                 `json:"SafeName"`
	RequestorUserName        string                 `json:"RequestorUserName"`
	RequestorReason          string                 `json:"RequestorReason,omitempty"`
	UserReason               string                 `json:"UserReason,omitempty"`
	CreationDate             int64                  `json:"CreationDate"`
	Operation                string                 `json:"Operation"`
	ExpirationDate           int64                  `json:"ExpirationDate,omitempty"`
	OperationType            int                    `json:"OperationType"`
	AccessType               string                 `json:"AccessType,omitempty"`
	ConfirmationsLeft        int                    `json:"ConfirmationsLeft"`
	AccessFrom               int64                  `json:"AccessFrom,omitempty"`
	AccessTo                 int64                  `json:"AccessTo,omitempty"`
	Status                   int                    `json:"Status"`
	StatusTitle              string                 `json:"StatusTitle,omitempty"`
	InvalidRequestReason     string                 `json:"InvalidRequestReason,omitempty"`
	CurrentConfirmationLevel int                    `json:"CurrentConfirmationLevel"`
	RequiredConfirmers       int                    `json:"RequiredConfirmers"`
	ConfirmedByUser          string                 `json:"ConfirmedByUser,omitempty"`
	AdditionalInfo           map[string]interface{} `json:"AdditionalInfo,omitempty"`
	AccountDetails           *AccountDetails        `json:"AccountDetails,omitempty"`
}

// AccountDetails holds account information for a request.
type AccountDetails struct {
	AccountID   types.FlexibleID `json:"AccountID"`
	AccountName string           `json:"AccountName,omitempty"`
	SafeName    string           `json:"SafeName,omitempty"`
	PlatformID  types.FlexibleID `json:"PlatformID,omitempty"`
	Address     string           `json:"Address,omitempty"`
}

// RequestsResponse represents the response from listing requests.
type RequestsResponse struct {
	Requests []Request `json:"Requests"`
	Total    int       `json:"Total"`
}

// ListOptions holds options for listing requests.
type ListOptions struct {
	RequestorUserName string
	SafeName          string
	OnlyWaiting       bool
	Expired           bool
	Offset            int
	Limit             int
}

// ListIncoming retrieves incoming access requests (requests to approve).
// This is equivalent to Get-PASRequest -OnlyWaiting -Incoming in psPAS.
func ListIncoming(ctx context.Context, sess *session.Session, opts ListOptions) (*RequestsResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.OnlyWaiting {
		params.Set("onlyWaiting", "true")
	}
	if opts.Expired {
		params.Set("expired", "true")
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}

	resp, err := sess.Client.Get(ctx, "/IncomingRequests", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list incoming requests: %w", err)
	}

	var result RequestsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse requests response: %w", err)
	}

	return &result, nil
}

// ListMyRequests retrieves the user's own access requests.
// This is equivalent to Get-PASRequest -MyRequests in psPAS.
func ListMyRequests(ctx context.Context, sess *session.Session, opts ListOptions) (*RequestsResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.OnlyWaiting {
		params.Set("onlyWaiting", "true")
	}
	if opts.Expired {
		params.Set("expired", "true")
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}

	resp, err := sess.Client.Get(ctx, "/MyRequests", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list my requests: %w", err)
	}

	var result RequestsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse requests response: %w", err)
	}

	return &result, nil
}

// CreateOptions holds options for creating an access request.
type CreateOptions struct {
	AccountID              string            `json:"AccountId"`
	Reason                 string            `json:"Reason,omitempty"`
	TicketingSystemName    string            `json:"TicketingSystemName,omitempty"`
	TicketID               string            `json:"TicketId,omitempty"`
	MultipleAccessRequired bool              `json:"MultipleAccessRequired,omitempty"`
	FromDate               int64             `json:"FromDate,omitempty"`
	ToDate                 int64             `json:"ToDate,omitempty"`
	AdditionalInfo         map[string]string `json:"AdditionalInfo,omitempty"`
	UseConnect             bool              `json:"UseConnect,omitempty"`
	ConnectionComponent    string            `json:"ConnectionComponent,omitempty"`
	ConnectionParams       map[string]string `json:"ConnectionParams,omitempty"`
}

// Create creates a new access request.
// This is equivalent to New-PASRequest in psPAS.
func Create(ctx context.Context, sess *session.Session, opts CreateOptions) (*Request, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.AccountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Post(ctx, "/MyRequests", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var request Request
	if err := json.Unmarshal(resp.Body, &request); err != nil {
		return nil, fmt.Errorf("failed to parse request response: %w", err)
	}

	return &request, nil
}

// ApproveOptions holds options for approving a request.
type ApproveOptions struct {
	Reason string `json:"Reason,omitempty"`
}

// Approve approves an access request.
// This is equivalent to Approve-PASRequest in psPAS.
func Approve(ctx context.Context, sess *session.Session, requestID string, opts ApproveOptions) (*Request, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if requestID == "" {
		return nil, fmt.Errorf("requestID is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/IncomingRequests/%s/Confirm", url.PathEscape(requestID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to approve request: %w", err)
	}

	var request Request
	if err := json.Unmarshal(resp.Body, &request); err != nil {
		return nil, fmt.Errorf("failed to parse request response: %w", err)
	}

	return &request, nil
}

// DenyOptions holds options for denying a request.
type DenyOptions struct {
	Reason string `json:"Reason,omitempty"`
}

// Deny denies an access request.
// This is equivalent to Deny-PASRequest in psPAS.
func Deny(ctx context.Context, sess *session.Session, requestID string, opts DenyOptions) (*Request, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if requestID == "" {
		return nil, fmt.Errorf("requestID is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/IncomingRequests/%s/Reject", url.PathEscape(requestID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to deny request: %w", err)
	}

	var request Request
	if err := json.Unmarshal(resp.Body, &request); err != nil {
		return nil, fmt.Errorf("failed to parse request response: %w", err)
	}

	return &request, nil
}

// Delete removes an access request.
// This is equivalent to Remove-PASRequest in psPAS.
func Delete(ctx context.Context, sess *session.Session, requestID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if requestID == "" {
		return fmt.Errorf("requestID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/MyRequests/%s", url.PathEscape(requestID)))
	if err != nil {
		return fmt.Errorf("failed to delete request: %w", err)
	}

	return nil
}
