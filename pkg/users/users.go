// Package users provides CyberArk user management functionality.
// This is equivalent to the User functions in psPAS including
// Get-PASUser, New-PASUser, Set-PASUser, Remove-PASUser, etc.
package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrisranney/gopas/internal/session"
)

// User represents a CyberArk user.
type User struct {
	ID                      int             `json:"id"`
	Username                string          `json:"username"`
	Source                  string          `json:"source,omitempty"`
	UserType                string          `json:"userType,omitempty"`
	ComponentUser           bool            `json:"componentUser"`
	GroupsMembership        []GroupMembership `json:"groupsMembership,omitempty"`
	VaultAuthorization      []string        `json:"vaultAuthorization,omitempty"`
	Location                string          `json:"location,omitempty"`
	PersonalDetails         *PersonalDetails `json:"personalDetails,omitempty"`
	EnableUser              bool            `json:"enableUser"`
	Suspended               bool            `json:"suspended,omitempty"`
	AuthenticationMethod    []string        `json:"authenticationMethod,omitempty"`
	PasswordNeverExpires    bool            `json:"passwordNeverExpires,omitempty"`
	DistinguishedName       string          `json:"distinguishedName,omitempty"`
	Description             string          `json:"description,omitempty"`
	BusinessAddress         *Address        `json:"businessAddress,omitempty"`
	Internet                *Internet       `json:"internet,omitempty"`
	Phones                  *Phones         `json:"phones,omitempty"`
	UnauthorizedInterfaces  []string        `json:"unauthorizedInterfaces,omitempty"`
	ExpiryDate              int64           `json:"expiryDate,omitempty"`
	LastSuccessfulLoginDate int64           `json:"lastSuccessfulLoginDate,omitempty"`
}

// PersonalDetails holds personal details for a user.
type PersonalDetails struct {
	FirstName  string `json:"firstName,omitempty"`
	MiddleName string `json:"middleName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	Zip        string `json:"zip,omitempty"`
	Country    string `json:"country,omitempty"`
	Title      string `json:"title,omitempty"`
	Organization string `json:"organization,omitempty"`
	Department string `json:"department,omitempty"`
	Profession string `json:"profession,omitempty"`
}

// Address holds address information.
type Address struct {
	WorkStreet  string `json:"workStreet,omitempty"`
	WorkCity    string `json:"workCity,omitempty"`
	WorkState   string `json:"workState,omitempty"`
	WorkZip     string `json:"workZip,omitempty"`
	WorkCountry string `json:"workCountry,omitempty"`
}

// Internet holds internet contact information.
type Internet struct {
	HomePage       string `json:"homePage,omitempty"`
	HomeEmail      string `json:"homeEmail,omitempty"`
	BusinessEmail  string `json:"businessEmail,omitempty"`
	OtherEmail     string `json:"otherEmail,omitempty"`
}

// Phones holds phone numbers.
type Phones struct {
	HomeNumber     string `json:"homeNumber,omitempty"`
	BusinessNumber string `json:"businessNumber,omitempty"`
	CellularNumber string `json:"cellularNumber,omitempty"`
	FaxNumber      string `json:"faxNumber,omitempty"`
	PagerNumber    string `json:"pagerNumber,omitempty"`
}

// GroupMembership holds group membership information.
type GroupMembership struct {
	GroupID   int    `json:"groupID"`
	GroupName string `json:"groupName"`
	GroupType string `json:"groupType,omitempty"`
}

// UsersResponse represents the response from listing users.
type UsersResponse struct {
	Users    []User `json:"Users"`
	Total    int    `json:"Total"`
	NextLink string `json:"nextLink,omitempty"`
}

// ListOptions holds options for listing users.
type ListOptions struct {
	Search      string
	Sort        string
	Offset      int
	Limit       int
	Filter      string
	UserType    string
	ComponentUser *bool
}

// List retrieves users from CyberArk.
// This is equivalent to Get-PASUser in psPAS.
func List(ctx context.Context, sess *session.Session, opts ListOptions) (*UsersResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.Search != "" {
		params.Set("search", opts.Search)
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}
	if opts.UserType != "" {
		params.Set("userType", opts.UserType)
	}
	if opts.ComponentUser != nil {
		params.Set("componentUser", strconv.FormatBool(*opts.ComponentUser))
	}

	resp, err := sess.Client.Get(ctx, "/Users", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var result UsersResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse users response: %w", err)
	}

	return &result, nil
}

// Get retrieves a specific user by ID.
// This is equivalent to Get-PASUser -id in psPAS.
func Get(ctx context.Context, sess *session.Session, userID int) (*User, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Users/%d", userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user User
	if err := json.Unmarshal(resp.Body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	return &user, nil
}

// CreateOptions holds options for creating a user.
type CreateOptions struct {
	Username               string           `json:"username"`
	InitialPassword        string           `json:"initialPassword,omitempty"`
	UserType               string           `json:"userType,omitempty"`
	UnauthorizedInterfaces []string         `json:"unauthorizedInterfaces,omitempty"`
	EnableUser             *bool            `json:"enableUser,omitempty"`
	AuthenticationMethod   []string         `json:"authenticationMethod,omitempty"`
	PasswordNeverExpires   *bool            `json:"passwordNeverExpires,omitempty"`
	ChangePassOnNextLogon  *bool            `json:"changePassOnNextLogon,omitempty"`
	ExpiryDate             int64            `json:"expiryDate,omitempty"`
	Location               string           `json:"location,omitempty"`
	VaultAuthorization     []string         `json:"vaultAuthorization,omitempty"`
	PersonalDetails        *PersonalDetails `json:"personalDetails,omitempty"`
	Description            string           `json:"description,omitempty"`
	BusinessAddress        *Address         `json:"businessAddress,omitempty"`
	Internet               *Internet        `json:"internet,omitempty"`
	Phones                 *Phones          `json:"phones,omitempty"`
}

// Create creates a new user in CyberArk.
// This is equivalent to New-PASUser in psPAS.
func Create(ctx context.Context, sess *session.Session, opts CreateOptions) (*User, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	resp, err := sess.Client.Post(ctx, "/Users", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	var user User
	if err := json.Unmarshal(resp.Body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	return &user, nil
}

// UpdateOptions holds options for updating a user.
type UpdateOptions struct {
	EnableUser             *bool            `json:"enableUser,omitempty"`
	ChangePassOnNextLogon  *bool            `json:"changePassOnNextLogon,omitempty"`
	Suspended              *bool            `json:"suspended,omitempty"`
	UnauthorizedInterfaces []string         `json:"unauthorizedInterfaces,omitempty"`
	AuthenticationMethod   []string         `json:"authenticationMethod,omitempty"`
	PasswordNeverExpires   *bool            `json:"passwordNeverExpires,omitempty"`
	ExpiryDate             *int64           `json:"expiryDate,omitempty"`
	Location               string           `json:"location,omitempty"`
	VaultAuthorization     []string         `json:"vaultAuthorization,omitempty"`
	PersonalDetails        *PersonalDetails `json:"personalDetails,omitempty"`
	Description            string           `json:"description,omitempty"`
	BusinessAddress        *Address         `json:"businessAddress,omitempty"`
	Internet               *Internet        `json:"internet,omitempty"`
	Phones                 *Phones          `json:"phones,omitempty"`
}

// Update updates an existing user.
// This is equivalent to Set-PASUser in psPAS.
func Update(ctx context.Context, sess *session.Session, userID int, opts UpdateOptions) (*User, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Put(ctx, fmt.Sprintf("/Users/%d", userID), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	var user User
	if err := json.Unmarshal(resp.Body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	return &user, nil
}

// Delete removes a user from CyberArk.
// This is equivalent to Remove-PASUser in psPAS.
func Delete(ctx context.Context, sess *session.Session, userID int) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Users/%d", userID))
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ActivateUser activates a suspended user.
// This is equivalent to Unblock-PASUser in psPAS.
func ActivateUser(ctx context.Context, sess *session.Session, userID int) (*User, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Users/%d/Activate", userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}

	var user User
	if err := json.Unmarshal(resp.Body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	return &user, nil
}

// ResetPassword resets a user's password.
// This is equivalent to Set-PASUserPassword in psPAS.
func ResetPassword(ctx context.Context, sess *session.Session, userID int, newPassword string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if newPassword == "" {
		return fmt.Errorf("newPassword is required")
	}

	body := map[string]string{
		"newPassword": newPassword,
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/Users/%d/ResetPassword", userID), body)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}
