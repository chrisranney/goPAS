// Package applications provides CyberArk application management functionality.
// This is equivalent to the Applications functions in psPAS including
// Get-PASApplication, Add-PASApplication, Remove-PASApplication, etc.
package applications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
)

// Application represents a CyberArk application.
type Application struct {
	AppID               string `json:"AppID"`
	Description         string `json:"Description,omitempty"`
	Location            string `json:"Location,omitempty"`
	AccessPermittedFrom int    `json:"AccessPermittedFrom,omitempty"`
	AccessPermittedTo   int    `json:"AccessPermittedTo,omitempty"`
	ExpirationDate      string `json:"ExpirationDate,omitempty"`
	Disabled            bool   `json:"Disabled"`
	BusinessOwnerFName  string `json:"BusinessOwnerFName,omitempty"`
	BusinessOwnerLName  string `json:"BusinessOwnerLName,omitempty"`
	BusinessOwnerEmail  string `json:"BusinessOwnerEmail,omitempty"`
	BusinessOwnerPhone  string `json:"BusinessOwnerPhone,omitempty"`
}

// ApplicationsResponse represents the response from listing applications.
type ApplicationsResponse struct {
	Applications []Application `json:"application"`
}

// ListOptions holds options for listing applications.
type ListOptions struct {
	Location     string
	SubLocations bool
}

// List retrieves applications from CyberArk.
// This is equivalent to Get-PASApplication in psPAS.
func List(ctx context.Context, sess *session.Session, opts ListOptions) ([]Application, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.Location != "" {
		params.Set("location", opts.Location)
	}
	if opts.SubLocations {
		params.Set("includeSublocations", "true")
	}

	resp, err := sess.Client.Get(ctx, "/WebServices/PIMServices.svc/Applications", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	var result ApplicationsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse applications response: %w", err)
	}

	return result.Applications, nil
}

// Get retrieves a specific application.
func Get(ctx context.Context, sess *session.Session, appID string) (*Application, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if appID == "" {
		return nil, fmt.Errorf("appID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Applications/%s", url.PathEscape(appID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	var result struct {
		Application Application `json:"application"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse application response: %w", err)
	}

	return &result.Application, nil
}

// CreateOptions holds options for creating an application.
type CreateOptions struct {
	AppID               string `json:"AppID"`
	Description         string `json:"Description,omitempty"`
	Location            string `json:"Location,omitempty"`
	AccessPermittedFrom int    `json:"AccessPermittedFrom,omitempty"`
	AccessPermittedTo   int    `json:"AccessPermittedTo,omitempty"`
	ExpirationDate      string `json:"ExpirationDate,omitempty"`
	Disabled            bool   `json:"Disabled,omitempty"`
	BusinessOwnerFName  string `json:"BusinessOwnerFName,omitempty"`
	BusinessOwnerLName  string `json:"BusinessOwnerLName,omitempty"`
	BusinessOwnerEmail  string `json:"BusinessOwnerEmail,omitempty"`
	BusinessOwnerPhone  string `json:"BusinessOwnerPhone,omitempty"`
}

// Create creates a new application in CyberArk.
// This is equivalent to Add-PASApplication in psPAS.
func Create(ctx context.Context, sess *session.Session, opts CreateOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if opts.AppID == "" {
		return fmt.Errorf("appID is required")
	}

	body := map[string]interface{}{
		"application": opts,
	}

	_, err := sess.Client.Post(ctx, "/WebServices/PIMServices.svc/Applications", body)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	return nil
}

// Delete removes an application from CyberArk.
// This is equivalent to Remove-PASApplication in psPAS.
func Delete(ctx context.Context, sess *session.Session, appID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if appID == "" {
		return fmt.Errorf("appID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Applications/%s", url.PathEscape(appID)))
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	return nil
}

// AuthMethod represents an application authentication method.
type AuthMethod struct {
	AppID                string `json:"AppID"`
	AuthType             string `json:"AuthType"`
	AuthValue            string `json:"AuthValue"`
	Comment              string `json:"Comment,omitempty"`
	IsFolder             bool   `json:"IsFolder,omitempty"`
	AllowInternalScripts bool   `json:"AllowInternalScripts,omitempty"`
}

// ListAuthMethods retrieves authentication methods for an application.
// This is equivalent to Get-PASApplicationAuthenticationMethod in psPAS.
func ListAuthMethods(ctx context.Context, sess *session.Session, appID string) ([]AuthMethod, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if appID == "" {
		return nil, fmt.Errorf("appID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Applications/%s/Authentications", url.PathEscape(appID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list auth methods: %w", err)
	}

	var result struct {
		Authentication []AuthMethod `json:"authentication"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse auth methods response: %w", err)
	}

	return result.Authentication, nil
}

// AddAuthMethodOptions holds options for adding an authentication method.
type AddAuthMethodOptions struct {
	AuthType             string `json:"AuthType"`
	AuthValue            string `json:"AuthValue"`
	Comment              string `json:"Comment,omitempty"`
	IsFolder             bool   `json:"IsFolder,omitempty"`
	AllowInternalScripts bool   `json:"AllowInternalScripts,omitempty"`
}

// AddAuthMethod adds an authentication method to an application.
// This is equivalent to Add-PASApplicationAuthenticationMethod in psPAS.
func AddAuthMethod(ctx context.Context, sess *session.Session, appID string, opts AddAuthMethodOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if appID == "" {
		return fmt.Errorf("appID is required")
	}

	if opts.AuthType == "" {
		return fmt.Errorf("authType is required")
	}

	body := map[string]interface{}{
		"authentication": opts,
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Applications/%s/Authentications", url.PathEscape(appID)), body)
	if err != nil {
		return fmt.Errorf("failed to add auth method: %w", err)
	}

	return nil
}

// RemoveAuthMethod removes an authentication method from an application.
// This is equivalent to Remove-PASApplicationAuthenticationMethod in psPAS.
func RemoveAuthMethod(ctx context.Context, sess *session.Session, appID string, authID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if appID == "" {
		return fmt.Errorf("appID is required")
	}

	if authID == "" {
		return fmt.Errorf("authID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/WebServices/PIMServices.svc/Applications/%s/Authentications/%s", url.PathEscape(appID), url.PathEscape(authID)))
	if err != nil {
		return fmt.Errorf("failed to remove auth method: %w", err)
	}

	return nil
}
