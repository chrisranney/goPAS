// Package accountimport provides CyberArk bulk account import functionality.
// This is equivalent to the account import functions in psPAS including
// Start-PASAccountImportJob, Get-PASAccountImportJob, New-PASAccountObject.
package accountimport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// AccountImportSource represents the source of the account import.
type AccountImportSource string

const (
	// SourceFile indicates importing from a file
	SourceFile AccountImportSource = "File"
	// SourcePendingAccounts indicates importing from pending accounts
	SourcePendingAccounts AccountImportSource = "PendingAccounts"
)

// ImportAccount represents an account to be imported.
type ImportAccount struct {
	UserName                  string                 `json:"userName"`
	Address                   string                 `json:"address"`
	SafeName                  string                 `json:"safeName"`
	PlatformID                types.FlexibleID       `json:"platformId"`
	Secret                    string                 `json:"secret,omitempty"`
	SecretType                string                 `json:"secretType,omitempty"`
	Name                      string                 `json:"name,omitempty"`
	PlatformAccountProperties map[string]interface{} `json:"platformAccountProperties,omitempty"`
	SecretManagement          *SecretManagement      `json:"secretManagement,omitempty"`
	RemoteMachinesAccess      *RemoteMachinesAccess  `json:"remoteMachinesAccess,omitempty"`
}

// SecretManagement holds secret management settings.
type SecretManagement struct {
	AutomaticManagementEnabled bool   `json:"automaticManagementEnabled"`
	ManualManagementReason     string `json:"manualManagementReason,omitempty"`
}

// RemoteMachinesAccess holds remote machines access settings.
type RemoteMachinesAccess struct {
	RemoteMachines                   string `json:"remoteMachines,omitempty"`
	AccessRestrictedToRemoteMachines bool   `json:"accessRestrictedToRemoteMachines"`
}

// StartImportJobOptions holds options for starting an import job.
type StartImportJobOptions struct {
	Source   AccountImportSource `json:"source,omitempty"`
	Accounts []ImportAccount     `json:"accountsList"`
}

// ImportJob represents an account import job.
type ImportJob struct {
	ID         types.FlexibleID `json:"id"`
	Source     string           `json:"source"`
	Status     string           `json:"status"`
	CreatedAt  int64            `json:"createdAt"`
	StartedAt  int64            `json:"startedAt,omitempty"`
	FinishedAt int64            `json:"finishedAt,omitempty"`
}

// ImportJobResult represents the result of an import job.
type ImportJobResult struct {
	ID               types.FlexibleID `json:"id"`
	Source           string           `json:"source"`
	Status           string           `json:"status"`
	TotalAccounts    int              `json:"totalAccounts"`
	SuccessCount     int              `json:"successCount"`
	FailedCount      int              `json:"failedCount"`
	CreatedAt        int64            `json:"createdAt"`
	StartedAt        int64            `json:"startedAt,omitempty"`
	FinishedAt       int64            `json:"finishedAt,omitempty"`
	FailedAccounts   []FailedAccount  `json:"failedAccounts,omitempty"`
	SuccessAccounts  []SuccessAccount `json:"successAccounts,omitempty"`
}

// FailedAccount represents an account that failed to import.
type FailedAccount struct {
	Index      int    `json:"index"`
	AccountName string `json:"accountName"`
	ErrorCode  string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// SuccessAccount represents an account that was successfully imported.
type SuccessAccount struct {
	Index       int              `json:"index"`
	AccountID   types.FlexibleID `json:"accountId"`
	AccountName string           `json:"accountName"`
}

// StartImportJob starts a bulk account import job.
// This is equivalent to Start-PASAccountImportJob in psPAS.
func StartImportJob(ctx context.Context, sess *session.Session, opts StartImportJobOptions) (*ImportJob, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if len(opts.Accounts) == 0 {
		return nil, fmt.Errorf("at least one account is required")
	}

	resp, err := sess.Client.Post(ctx, "/BulkActions/Accounts", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to start import job: %w", err)
	}

	var result ImportJob
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse import job response: %w", err)
	}

	return &result, nil
}

// GetImportJob retrieves the status and results of an import job.
// This is equivalent to Get-PASAccountImportJob in psPAS.
func GetImportJob(ctx context.Context, sess *session.Session, jobID string) (*ImportJobResult, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if jobID == "" {
		return nil, fmt.Errorf("jobID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/BulkActions/Accounts/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get import job: %w", err)
	}

	var result ImportJobResult
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse import job response: %w", err)
	}

	return &result, nil
}

// ListImportJobs retrieves all account import jobs.
func ListImportJobs(ctx context.Context, sess *session.Session) ([]ImportJob, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/BulkActions/Accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list import jobs: %w", err)
	}

	var result struct {
		Jobs []ImportJob `json:"BulkActions"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse import jobs response: %w", err)
	}

	return result.Jobs, nil
}

// NewAccountObject creates a new account object for bulk import.
// This is equivalent to New-PASAccountObject in psPAS.
// This is a helper function to build ImportAccount objects.
func NewAccountObject(
	userName string,
	address string,
	safeName string,
	platformID string,
	secret string,
) *ImportAccount {
	return &ImportAccount{
		UserName:   userName,
		Address:    address,
		SafeName:   safeName,
		PlatformID: types.FlexibleID(platformID),
		Secret:     secret,
	}
}

// PendingAccount represents a pending account for review.
type PendingAccount struct {
	ID                        types.FlexibleID       `json:"id"`
	UserName                  string                 `json:"userName"`
	Address                   string                 `json:"address"`
	PlatformID                types.FlexibleID       `json:"platformId"`
	SafeName                  string                 `json:"safeName"`
	LastPasswordSetDate       int64                  `json:"lastPasswordSetDate,omitempty"`
	PlatformAccountProperties map[string]interface{} `json:"platformAccountProperties,omitempty"`
	CreationTime              int64                  `json:"creationTime"`
}

// PendingAccountsResponse represents the response from listing pending accounts.
type PendingAccountsResponse struct {
	PendingAccounts []PendingAccount `json:"PendingAccounts"`
	Total           int              `json:"Total"`
}

// AddPendingAccount adds an account to the pending accounts list.
// This is equivalent to Add-PASPendingAccount in psPAS.
func AddPendingAccount(ctx context.Context, sess *session.Session, account ImportAccount) (*PendingAccount, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if account.UserName == "" {
		return nil, fmt.Errorf("userName is required")
	}

	if account.Address == "" {
		return nil, fmt.Errorf("address is required")
	}

	if account.PlatformID == "" {
		return nil, fmt.Errorf("platformID is required")
	}

	resp, err := sess.Client.Post(ctx, "/WebServices/PIMServices.svc/PendingAccounts", account)
	if err != nil {
		return nil, fmt.Errorf("failed to add pending account: %w", err)
	}

	var result PendingAccount
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pending account response: %w", err)
	}

	return &result, nil
}

// PasswordVersionInfo represents a password version for an account.
type PasswordVersionInfo struct {
	Version        int    `json:"version"`
	Status         string `json:"status"`
	LastModified   int64  `json:"lastModified"`
	ModifiedBy     string `json:"modifiedBy"`
}

// GetAccountPasswordVersions retrieves password version history for an account.
// This is equivalent to Get-PASAccountPasswordVersion in psPAS.
func GetAccountPasswordVersions(ctx context.Context, sess *session.Session, accountID string) ([]PasswordVersionInfo, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Accounts/%s/Secret/Versions", url.PathEscape(accountID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get password versions: %w", err)
	}

	var result struct {
		Versions []PasswordVersionInfo `json:"Versions"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse password versions response: %w", err)
	}

	return result.Versions, nil
}

// GeneratePasswordOptions holds options for generating a new password.
type GeneratePasswordOptions struct {
	AccountID string `json:"-"` // Used in URL, not body
}

// GeneratePassword generates a new password for an account.
// This is equivalent to New-PASAccountPassword in psPAS.
func GeneratePassword(ctx context.Context, sess *session.Session, accountID string) (string, error) {
	if sess == nil || !sess.IsValid() {
		return "", fmt.Errorf("valid session is required")
	}

	if accountID == "" {
		return "", fmt.Errorf("accountID is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Accounts/%s/Secret/Generate", url.PathEscape(accountID)), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	// Response is the password as a string
	password := string(resp.Body)
	// Remove surrounding quotes if present
	if len(password) >= 2 && password[0] == '"' && password[len(password)-1] == '"' {
		password = password[1 : len(password)-1]
	}

	return password, nil
}
