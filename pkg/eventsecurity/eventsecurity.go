// Package eventsecurity provides CyberArk PTA (Privilege Threat Analytics) functionality.
// This is equivalent to the EventSecurity functions in psPAS including
// Get-PASPTAEvent, Set-PASPTARule, Get-PASPTARemediation, etc.
package eventsecurity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrisranney/gopas/internal/session"
)

// PTAEvent represents a PTA security event.
type PTAEvent struct {
	ID                 string                 `json:"id"`
	Type               string                 `json:"type"`
	Score              float64                `json:"score"`
	EventTime          int64                  `json:"eventTime"`
	MachineAddress     string                 `json:"machineAddress,omitempty"`
	UserID             string                 `json:"userId,omitempty"`
	UserName           string                 `json:"userName,omitempty"`
	CloudData          *CloudData             `json:"cloudData,omitempty"`
	Status             string                 `json:"status,omitempty"`
	Details            map[string]interface{} `json:"details,omitempty"`
	AffectedAccounts   []AffectedAccount      `json:"affectedAccounts,omitempty"`
}

// CloudData holds cloud-related event data.
type CloudData struct {
	CloudProvider string `json:"cloudProvider,omitempty"`
	CloudService  string `json:"cloudService,omitempty"`
	Region        string `json:"region,omitempty"`
}

// AffectedAccount represents an account affected by a PTA event.
type AffectedAccount struct {
	AccountID   string `json:"accountId"`
	AccountName string `json:"accountName,omitempty"`
	SafeName    string `json:"safeName,omitempty"`
	PlatformID  string `json:"platformId,omitempty"`
}

// PTAEventsResponse represents the response from listing PTA events.
type PTAEventsResponse struct {
	PTAEvents []PTAEvent `json:"Events"`
	Total     int        `json:"Total"`
	NextLink  string     `json:"NextLink,omitempty"`
}

// ListEventsOptions holds options for listing PTA events.
type ListEventsOptions struct {
	FromDate     int64
	ToDate       int64
	Status       string
	AccountID    string
	Offset       int
	Limit        int
}

// ListEvents retrieves PTA security events.
// This is equivalent to Get-PASPTAEvent in psPAS.
func ListEvents(ctx context.Context, sess *session.Session, opts ListEventsOptions) (*PTAEventsResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.FromDate > 0 {
		params.Set("fromDate", strconv.FormatInt(opts.FromDate, 10))
	}
	if opts.ToDate > 0 {
		params.Set("toDate", strconv.FormatInt(opts.ToDate, 10))
	}
	if opts.Status != "" {
		params.Set("status", opts.Status)
	}
	if opts.AccountID != "" {
		params.Set("accountId", opts.AccountID)
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}

	resp, err := sess.Client.Get(ctx, "/pta/API/Events", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list PTA events: %w", err)
	}

	var result PTAEventsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PTA events response: %w", err)
	}

	return &result, nil
}

// GetEvent retrieves a specific PTA event.
func GetEvent(ctx context.Context, sess *session.Session, eventID string) (*PTAEvent, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if eventID == "" {
		return nil, fmt.Errorf("eventID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/pta/API/Events/%s", url.PathEscape(eventID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PTA event: %w", err)
	}

	var event PTAEvent
	if err := json.Unmarshal(resp.Body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse PTA event response: %w", err)
	}

	return &event, nil
}

// SetEventStatus updates the status of a PTA event.
// This is equivalent to Set-PASPTAEvent in psPAS.
func SetEventStatus(ctx context.Context, sess *session.Session, eventID string, status string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if eventID == "" {
		return fmt.Errorf("eventID is required")
	}

	if status == "" {
		return fmt.Errorf("status is required")
	}

	body := map[string]string{
		"status": status,
	}

	_, err := sess.Client.Patch(ctx, fmt.Sprintf("/pta/API/Events/%s", url.PathEscape(eventID)), body)
	if err != nil {
		return fmt.Errorf("failed to update PTA event status: %w", err)
	}

	return nil
}

// PTARule represents a PTA security rule.
type PTARule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
	Active      bool   `json:"active"`
	Score       int    `json:"score,omitempty"`
}

// ListRules retrieves PTA security rules.
// This is equivalent to Get-PASPTARule in psPAS.
func ListRules(ctx context.Context, sess *session.Session) ([]PTARule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/pta/API/Settings/RiskyActivities", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list PTA rules: %w", err)
	}

	var result []PTARule
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PTA rules response: %w", err)
	}

	return result, nil
}

// SetRuleOptions holds options for updating a PTA rule.
type SetRuleOptions struct {
	Active bool `json:"active"`
	Score  int  `json:"score,omitempty"`
}

// SetRule updates a PTA security rule.
// This is equivalent to Set-PASPTARule in psPAS.
func SetRule(ctx context.Context, sess *session.Session, ruleID string, opts SetRuleOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if ruleID == "" {
		return fmt.Errorf("ruleID is required")
	}

	_, err := sess.Client.Put(ctx, fmt.Sprintf("/pta/API/Settings/RiskyActivities/%s", url.PathEscape(ruleID)), opts)
	if err != nil {
		return fmt.Errorf("failed to update PTA rule: %w", err)
	}

	return nil
}

// PTARemediation represents a PTA remediation action.
type PTARemediation struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
}

// ListRemediations retrieves PTA remediation options.
// This is equivalent to Get-PASPTARemediation in psPAS.
func ListRemediations(ctx context.Context, sess *session.Session) ([]PTARemediation, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/pta/API/Settings/AutomaticRemediations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list PTA remediations: %w", err)
	}

	var result []PTARemediation
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PTA remediations response: %w", err)
	}

	return result, nil
}

// PrivilegedUser represents a privileged user in PTA.
type PrivilegedUser struct {
	ID       string `json:"id"`
	UserName string `json:"userName"`
	Source   string `json:"source,omitempty"`
}

// GetPrivilegedUsers retrieves PTA privileged users.
// This is equivalent to Get-PASPTAPrivilegedUser in psPAS.
func GetPrivilegedUsers(ctx context.Context, sess *session.Session) ([]PrivilegedUser, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/pta/API/Settings/PrivilegedUsers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PTA privileged users: %w", err)
	}

	var result []PrivilegedUser
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PTA privileged users response: %w", err)
	}

	return result, nil
}

// AddPrivilegedUser adds a privileged user to PTA.
// This is equivalent to Add-PASPTAPrivilegedUser in psPAS.
func AddPrivilegedUser(ctx context.Context, sess *session.Session, userName string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if userName == "" {
		return fmt.Errorf("userName is required")
	}

	body := map[string]string{
		"userName": userName,
	}

	_, err := sess.Client.Post(ctx, "/pta/API/Settings/PrivilegedUsers", body)
	if err != nil {
		return fmt.Errorf("failed to add PTA privileged user: %w", err)
	}

	return nil
}

// RemovePrivilegedUser removes a privileged user from PTA.
// This is equivalent to Remove-PASPTAPrivilegedUser in psPAS.
func RemovePrivilegedUser(ctx context.Context, sess *session.Session, userID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/pta/API/Settings/PrivilegedUsers/%s", url.PathEscape(userID)))
	if err != nil {
		return fmt.Errorf("failed to remove PTA privileged user: %w", err)
	}

	return nil
}

// PrivilegedGroup represents a privileged group in PTA.
type PrivilegedGroup struct {
	ID        string `json:"id"`
	GroupName string `json:"groupName"`
	Source    string `json:"source,omitempty"`
}

// GetPrivilegedGroups retrieves PTA privileged groups.
// This is equivalent to Get-PASPTAPrivilegedGroup in psPAS.
func GetPrivilegedGroups(ctx context.Context, sess *session.Session) ([]PrivilegedGroup, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/pta/API/Settings/PrivilegedGroups", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PTA privileged groups: %w", err)
	}

	var result []PrivilegedGroup
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PTA privileged groups response: %w", err)
	}

	return result, nil
}

// AddPrivilegedGroup adds a privileged group to PTA.
// This is equivalent to Add-PASPTAPrivilegedGroup in psPAS.
func AddPrivilegedGroup(ctx context.Context, sess *session.Session, groupName string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if groupName == "" {
		return fmt.Errorf("groupName is required")
	}

	body := map[string]string{
		"groupName": groupName,
	}

	_, err := sess.Client.Post(ctx, "/pta/API/Settings/PrivilegedGroups", body)
	if err != nil {
		return fmt.Errorf("failed to add PTA privileged group: %w", err)
	}

	return nil
}

// RemovePrivilegedGroup removes a privileged group from PTA.
// This is equivalent to Remove-PASPTAPrivilegedGroup in psPAS.
func RemovePrivilegedGroup(ctx context.Context, sess *session.Session, groupID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if groupID == "" {
		return fmt.Errorf("groupID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/pta/API/Settings/PrivilegedGroups/%s", url.PathEscape(groupID)))
	if err != nil {
		return fmt.Errorf("failed to remove PTA privileged group: %w", err)
	}

	return nil
}
