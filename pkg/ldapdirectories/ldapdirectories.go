// Package ldapdirectories provides CyberArk LDAP directory management functionality.
// This is equivalent to the LDAPDirectories functions in psPAS including
// Get-PASDirectory, Add-PASDirectory, Set-PASDirectory, etc.
package ldapdirectories

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
)

// Directory represents an LDAP directory configuration.
type Directory struct {
	DirectoryID         string              `json:"DirectoryID,omitempty"`
	DomainName          string              `json:"DomainName"`
	DomainBaseContext   string              `json:"DomainBaseContext,omitempty"`
	BindUsername        string              `json:"BindUsername,omitempty"`
	BindPassword        string              `json:"BindPassword,omitempty"`
	DCList              []DomainController  `json:"DCList,omitempty"`
	SSLConnect          bool                `json:"SSLConnect,omitempty"`
	VaultUseDomainName  bool                `json:"VaultUseDomainName,omitempty"`
}

// DomainController represents a domain controller.
type DomainController struct {
	Name       string `json:"Name"`
	Address    string `json:"Address,omitempty"`
	Port       int    `json:"Port,omitempty"`
	SSLConnect bool   `json:"SSLConnect,omitempty"`
}

// DirectoriesResponse represents the response from listing directories.
type DirectoriesResponse struct {
	Directories []Directory `json:"Directories"`
}

// List retrieves LDAP directories.
// This is equivalent to Get-PASDirectory in psPAS.
func List(ctx context.Context, sess *session.Session) ([]Directory, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/Configuration/LDAP/Directories", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list directories: %w", err)
	}

	var result DirectoriesResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse directories response: %w", err)
	}

	return result.Directories, nil
}

// Get retrieves a specific LDAP directory.
func Get(ctx context.Context, sess *session.Session, directoryID string) (*Directory, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if directoryID == "" {
		return nil, fmt.Errorf("directoryID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Configuration/LDAP/Directories/%s", url.PathEscape(directoryID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory: %w", err)
	}

	var directory Directory
	if err := json.Unmarshal(resp.Body, &directory); err != nil {
		return nil, fmt.Errorf("failed to parse directory response: %w", err)
	}

	return &directory, nil
}

// CreateOptions holds options for creating a directory.
type CreateOptions struct {
	DomainName          string             `json:"DomainName"`
	DomainBaseContext   string             `json:"DomainBaseContext,omitempty"`
	BindUsername        string             `json:"BindUsername,omitempty"`
	BindPassword        string             `json:"BindPassword,omitempty"`
	DCList              []DomainController `json:"DCList,omitempty"`
	SSLConnect          bool               `json:"SSLConnect,omitempty"`
	VaultUseDomainName  bool               `json:"VaultUseDomainName,omitempty"`
}

// Create creates a new LDAP directory.
// This is equivalent to Add-PASDirectory in psPAS.
func Create(ctx context.Context, sess *session.Session, opts CreateOptions) (*Directory, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.DomainName == "" {
		return nil, fmt.Errorf("domainName is required")
	}

	resp, err := sess.Client.Post(ctx, "/Configuration/LDAP/Directories", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	var directory Directory
	if err := json.Unmarshal(resp.Body, &directory); err != nil {
		return nil, fmt.Errorf("failed to parse directory response: %w", err)
	}

	return &directory, nil
}

// Delete removes an LDAP directory.
// This is equivalent to Remove-PASDirectory in psPAS.
func Delete(ctx context.Context, sess *session.Session, directoryID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if directoryID == "" {
		return fmt.Errorf("directoryID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Configuration/LDAP/Directories/%s", url.PathEscape(directoryID)))
	if err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// DirectoryMapping represents an LDAP directory mapping.
type DirectoryMapping struct {
	MappingID          string   `json:"MappingID,omitempty"`
	DirectoryMappingName string `json:"DirectoryMappingName"`
	LDAPBranch         string   `json:"LDAPBranch"`
	DomainGroups       []string `json:"DomainGroups,omitempty"`
	VaultGroups        []string `json:"VaultGroups,omitempty"`
	Location           string   `json:"Location,omitempty"`
	LDAPQuery          string   `json:"LDAPQuery,omitempty"`
	MappingAuthorizations []string `json:"MappingAuthorizations,omitempty"`
	UserActivityLogPeriod int   `json:"UserActivityLogPeriod,omitempty"`
}

// ListMappings retrieves directory mappings.
// This is equivalent to Get-PASDirectoryMapping in psPAS.
func ListMappings(ctx context.Context, sess *session.Session, directoryID string) ([]DirectoryMapping, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if directoryID == "" {
		return nil, fmt.Errorf("directoryID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Configuration/LDAP/Directories/%s/Mappings", url.PathEscape(directoryID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list mappings: %w", err)
	}

	var result struct {
		Mappings []DirectoryMapping `json:"Mappings"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse mappings response: %w", err)
	}

	return result.Mappings, nil
}

// CreateMappingOptions holds options for creating a directory mapping.
type CreateMappingOptions struct {
	DirectoryMappingName string   `json:"DirectoryMappingName"`
	LDAPBranch           string   `json:"LDAPBranch"`
	DomainGroups         []string `json:"DomainGroups,omitempty"`
	VaultGroups          []string `json:"VaultGroups,omitempty"`
	Location             string   `json:"Location,omitempty"`
	LDAPQuery            string   `json:"LDAPQuery,omitempty"`
	MappingAuthorizations []string `json:"MappingAuthorizations,omitempty"`
	UserActivityLogPeriod int     `json:"UserActivityLogPeriod,omitempty"`
}

// CreateMapping creates a new directory mapping.
// This is equivalent to New-PASDirectoryMapping in psPAS.
func CreateMapping(ctx context.Context, sess *session.Session, directoryID string, opts CreateMappingOptions) (*DirectoryMapping, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if directoryID == "" {
		return nil, fmt.Errorf("directoryID is required")
	}

	if opts.DirectoryMappingName == "" {
		return nil, fmt.Errorf("directoryMappingName is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Configuration/LDAP/Directories/%s/Mappings", url.PathEscape(directoryID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create mapping: %w", err)
	}

	var mapping DirectoryMapping
	if err := json.Unmarshal(resp.Body, &mapping); err != nil {
		return nil, fmt.Errorf("failed to parse mapping response: %w", err)
	}

	return &mapping, nil
}

// DeleteMapping removes a directory mapping.
// This is equivalent to Remove-PASDirectoryMapping in psPAS.
func DeleteMapping(ctx context.Context, sess *session.Session, directoryID string, mappingID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if directoryID == "" {
		return fmt.Errorf("directoryID is required")
	}

	if mappingID == "" {
		return fmt.Errorf("mappingID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Configuration/LDAP/Directories/%s/Mappings/%s", url.PathEscape(directoryID), url.PathEscape(mappingID)))
	if err != nil {
		return fmt.Errorf("failed to delete mapping: %w", err)
	}

	return nil
}
