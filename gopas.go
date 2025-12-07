// Package gopas provides a Go SDK for the CyberArk Privileged Access Security (PAS) REST API.
//
// goPAS is a port of the popular psPAS PowerShell module, providing the same functionality
// in a native Go implementation. It supports all major CyberArk operations including
// authentication, account management, safe management, user management, and more.
//
// # Quick Start
//
// Create a new session and authenticate:
//
//	ctx := context.Background()
//	sess, err := gopas.NewSession(ctx, gopas.SessionOptions{
//		BaseURL: "https://cyberark.example.com",
//		Credentials: gopas.Credentials{
//			Username: "admin",
//			Password: "secretpassword",
//		},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer gopas.CloseSession(ctx, sess)
//
// List accounts:
//
//	accounts, err := gopas.ListAccounts(ctx, sess, gopas.ListAccountsOptions{
//		SafeName: "MySafe",
//	})
//
// # Authentication
//
// goPAS supports multiple authentication methods:
//   - CyberArk native authentication
//   - LDAP authentication
//   - RADIUS authentication
//   - SAML authentication
//   - Windows integrated authentication
//
// # Package Structure
//
// The SDK is organized into functional packages:
//   - authentication: Session management and authentication
//   - accounts: Account CRUD operations and password management
//   - safes: Safe management
//   - safemembers: Safe member and permissions management
//   - users: User and group management
//   - platforms: Platform management
//   - requests: Access request workflows
//   - applications: Application management
//   - monitoring: PSM session monitoring
//   - connections: PSM connection management
//   - eventsecurity: PTA (Privilege Threat Analytics)
//   - systemhealth: Component health monitoring
//   - ldapdirectories: LDAP directory configuration
//   - onboardingrules: Automatic account onboarding
//   - accountgroups: Account group management
//
// # Version Compatibility
//
// goPAS supports CyberArk versions up to v14.0, matching the functionality
// of psPAS 6.4.x.
package gopas

import (
	"context"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/accounts"
	"github.com/chrisranney/gopas/pkg/authentication"
	"github.com/chrisranney/gopas/pkg/safes"
)

// Version is the current version of the goPAS SDK.
const Version = "1.0.0"

// Re-export common types for convenience

// Session represents an authenticated CyberArk session.
type Session = session.Session

// Credentials holds authentication credentials.
type Credentials = authentication.Credentials

// SessionOptions holds options for creating a session.
type SessionOptions = authentication.SessionOptions

// AuthMethod represents an authentication method.
type AuthMethod = authentication.AuthMethod

// Authentication method constants
const (
	AuthMethodCyberArk = authentication.AuthMethodCyberArk
	AuthMethodLDAP     = authentication.AuthMethodLDAP
	AuthMethodRADIUS   = authentication.AuthMethodRADIUS
	AuthMethodWindows  = authentication.AuthMethodWindows
)

// Account represents a CyberArk privileged account.
type Account = accounts.Account

// Safe represents a CyberArk safe.
type Safe = safes.Safe

// NewSession creates a new authenticated session with CyberArk.
// This is the main entry point for using the SDK.
//
// Example:
//
//	sess, err := gopas.NewSession(ctx, gopas.SessionOptions{
//		BaseURL: "https://cyberark.example.com",
//		Credentials: gopas.Credentials{
//			Username: "admin",
//			Password: "password",
//		},
//		AuthMethod: gopas.AuthMethodCyberArk,
//	})
func NewSession(ctx context.Context, opts SessionOptions) (*Session, error) {
	return authentication.NewSession(ctx, opts)
}

// CloseSession closes an authenticated session.
// Always call this when done with a session.
func CloseSession(ctx context.Context, sess *Session) error {
	return authentication.CloseSession(ctx, sess)
}

// ListAccountsOptions holds options for listing accounts.
type ListAccountsOptions = accounts.ListOptions

// ListAccounts retrieves accounts from CyberArk.
func ListAccounts(ctx context.Context, sess *Session, opts ListAccountsOptions) (*accounts.AccountsResponse, error) {
	return accounts.List(ctx, sess, opts)
}

// GetAccount retrieves a specific account by ID.
func GetAccount(ctx context.Context, sess *Session, accountID string) (*Account, error) {
	return accounts.Get(ctx, sess, accountID)
}

// CreateAccountOptions holds options for creating an account.
type CreateAccountOptions = accounts.CreateOptions

// CreateAccount creates a new account in CyberArk.
func CreateAccount(ctx context.Context, sess *Session, opts CreateAccountOptions) (*Account, error) {
	return accounts.Create(ctx, sess, opts)
}

// DeleteAccount removes an account from CyberArk.
func DeleteAccount(ctx context.Context, sess *Session, accountID string) error {
	return accounts.Delete(ctx, sess, accountID)
}

// GetAccountPassword retrieves the password for an account.
func GetAccountPassword(ctx context.Context, sess *Session, accountID string, reason string) (string, error) {
	return accounts.GetPassword(ctx, sess, accountID, reason)
}

// ListSafesOptions holds options for listing safes.
type ListSafesOptions = safes.ListOptions

// ListSafes retrieves safes from CyberArk.
func ListSafes(ctx context.Context, sess *Session, opts ListSafesOptions) (*safes.SafesResponse, error) {
	return safes.List(ctx, sess, opts)
}

// GetSafe retrieves a specific safe by name.
func GetSafe(ctx context.Context, sess *Session, safeName string) (*Safe, error) {
	return safes.Get(ctx, sess, safeName)
}

// CreateSafeOptions holds options for creating a safe.
type CreateSafeOptions = safes.CreateOptions

// CreateSafe creates a new safe in CyberArk.
func CreateSafe(ctx context.Context, sess *Session, opts CreateSafeOptions) (*Safe, error) {
	return safes.Create(ctx, sess, opts)
}

// DeleteSafe removes a safe from CyberArk.
func DeleteSafe(ctx context.Context, sess *Session, safeName string) error {
	return safes.Delete(ctx, sess, safeName)
}

// GetServerInfo retrieves CyberArk server information.
func GetServerInfo(ctx context.Context, sess *Session) (*authentication.ServerInfo, error) {
	return authentication.GetServerInfo(ctx, sess)
}
