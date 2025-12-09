// Package users provides CyberArk group management functionality.
// This is equivalent to the Group functions in psPAS including
// Get-PASGroup, New-PASGroup, Add-PASGroupMember, Remove-PASGroupMember, etc.
package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrisranney/gopas/internal/session"
)

// Group represents a CyberArk group.
type Group struct {
	ID          int                 `json:"id"`
	GroupName   string              `json:"groupName"`
	Description string              `json:"description,omitempty"`
	Location    string              `json:"location,omitempty"`
	GroupType   string              `json:"groupType,omitempty"`
	Directory   string              `json:"directory,omitempty"`
	DN          string              `json:"dn,omitempty"`
	Members     []GroupMemberDetail `json:"members,omitempty"`
}

// GroupMemberDetail holds detailed information about a group member.
type GroupMemberDetail struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	GroupID    int    `json:"groupID"`
	DomainName string `json:"domainName,omitempty"`
}

// GroupsResponse represents the response from listing groups.
type GroupsResponse struct {
	Value    []Group `json:"value"`
	Count    int     `json:"count"`
	NextLink string  `json:"nextLink,omitempty"`
}

// ListGroupsOptions holds options for listing groups.
type ListGroupsOptions struct {
	Search         string
	Sort           string
	Offset         int
	Limit          int
	Filter         string
	IncludeMembers bool
}

// ListGroups retrieves groups from CyberArk.
// This is equivalent to Get-PASGroup in psPAS.
func ListGroups(ctx context.Context, sess *session.Session, opts ListGroupsOptions) (*GroupsResponse, error) {
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
	if opts.IncludeMembers {
		params.Set("includeMembers", "true")
	}

	resp, err := sess.Client.Get(ctx, "/UserGroups", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	var result GroupsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse groups response: %w", err)
	}

	return &result, nil
}

// GetGroup retrieves a specific group by ID.
func GetGroup(ctx context.Context, sess *session.Session, groupID int) (*Group, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/UserGroups/%d", groupID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	var group Group
	if err := json.Unmarshal(resp.Body, &group); err != nil {
		return nil, fmt.Errorf("failed to parse group response: %w", err)
	}

	return &group, nil
}

// CreateGroupOptions holds options for creating a group.
type CreateGroupOptions struct {
	GroupName   string `json:"groupName"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
}

// CreateGroup creates a new group in CyberArk.
// This is equivalent to New-PASGroup in psPAS.
func CreateGroup(ctx context.Context, sess *session.Session, opts CreateGroupOptions) (*Group, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.GroupName == "" {
		return nil, fmt.Errorf("groupName is required")
	}

	resp, err := sess.Client.Post(ctx, "/UserGroups", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	var group Group
	if err := json.Unmarshal(resp.Body, &group); err != nil {
		return nil, fmt.Errorf("failed to parse group response: %w", err)
	}

	return &group, nil
}

// DeleteGroup removes a group from CyberArk.
// This is equivalent to Remove-PASGroup in psPAS.
func DeleteGroup(ctx context.Context, sess *session.Session, groupID int) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/UserGroups/%d", groupID))
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

// AddGroupMemberOptions holds options for adding a group member.
type AddGroupMemberOptions struct {
	MemberID   int    `json:"memberId,omitempty"`
	MemberName string `json:"memberName,omitempty"`
	DomainName string `json:"domainName,omitempty"`
}

// AddGroupMember adds a user to a group.
// This is equivalent to Add-PASGroupMember in psPAS.
func AddGroupMember(ctx context.Context, sess *session.Session, groupID int, opts AddGroupMemberOptions) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if opts.MemberID == 0 && opts.MemberName == "" {
		return fmt.Errorf("memberId or memberName is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/UserGroups/%d/Members", groupID), opts)
	if err != nil {
		return fmt.Errorf("failed to add group member: %w", err)
	}

	return nil
}

// RemoveGroupMember removes a user from a group.
// This is equivalent to Remove-PASGroupMember in psPAS.
func RemoveGroupMember(ctx context.Context, sess *session.Session, groupID int, memberName string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if memberName == "" {
		return fmt.Errorf("memberName is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/UserGroups/%d/Members/%s", groupID, url.PathEscape(memberName)))
	if err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}

	return nil
}

// ListGroupMembers retrieves the members of a group.
func ListGroupMembers(ctx context.Context, sess *session.Session, groupID int) ([]GroupMemberDetail, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/UserGroups/%d/Members", groupID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %w", err)
	}

	var result struct {
		Members []GroupMemberDetail `json:"members"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse group members response: %w", err)
	}

	return result.Members, nil
}
