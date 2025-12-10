// Package monitoring provides CyberArk PSM session monitoring functionality.
// This is equivalent to the Monitoring functions in psPAS including
// Get-PASPSMSession, Get-PASPSMRecording, Stop-PASPSMSession, etc.
package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
)

// PSMSession represents a PSM session.
type PSMSession struct {
	SessionID           types.FlexibleID  `json:"SessionID"`
	SessionGuid         types.FlexibleID  `json:"SessionGuid,omitempty"`
	SafeName            string            `json:"SafeName,omitempty"`
	AccountID           types.FlexibleID  `json:"AccountID,omitempty"`
	AccountName         string            `json:"AccountName,omitempty"`
	AccountPlatformID   types.FlexibleID  `json:"AccountPlatformID,omitempty"`
	User                string            `json:"User"`
	RemoteMachine       string            `json:"RemoteMachine,omitempty"`
	Protocol            string            `json:"Protocol,omitempty"`
	Client              string            `json:"Client,omitempty"`
	ClientIP            string            `json:"ClientIP,omitempty"`
	ConnectionComponent string            `json:"ConnectionComponent,omitempty"`
	Start               int64             `json:"Start"`
	End                 int64             `json:"End,omitempty"`
	Duration            int64             `json:"Duration,omitempty"`
	PSMServerID         types.FlexibleID  `json:"PSMServerID,omitempty"`
	FromIP              string            `json:"FromIP,omitempty"`
	RiskScore           float64           `json:"RiskScore,omitempty"`
	IsLive              bool              `json:"IsLive"`
	CanTerminate        bool              `json:"CanTerminate,omitempty"`
	CanMonitor          bool              `json:"CanMonitor,omitempty"`
	CanPlayback         bool              `json:"CanPlayback,omitempty"`
	RecordingFiles      []RecordingFile   `json:"RecordingFiles,omitempty"`
	Properties          map[string]string `json:"Properties,omitempty"`
}

// RecordingFile represents a PSM recording file.
type RecordingFile struct {
	FileName      string `json:"FileName"`
	RecordingType string `json:"RecordingType,omitempty"`
	Format        string `json:"Format,omitempty"`
}

// SessionsResponse represents the response from listing sessions.
type SessionsResponse struct {
	Recordings []PSMSession `json:"Recordings"`
	Total      int          `json:"Total"`
	NextLink   string       `json:"NextLink,omitempty"`
}

// ListOptions holds options for listing sessions.
type ListOptions struct {
	FromTime   int64
	ToTime     int64
	Limit      int
	Offset     int
	Search     string
	Safe       string
	Activities string
}

// ListSessions retrieves PSM sessions.
// This is equivalent to Get-PASPSMSession in psPAS.
func ListSessions(ctx context.Context, sess *session.Session, opts ListOptions) (*SessionsResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.FromTime > 0 {
		params.Set("fromTime", strconv.FormatInt(opts.FromTime, 10))
	}
	if opts.ToTime > 0 {
		params.Set("toTime", strconv.FormatInt(opts.ToTime, 10))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Search != "" {
		params.Set("search", opts.Search)
	}
	if opts.Safe != "" {
		params.Set("safe", opts.Safe)
	}
	if opts.Activities != "" {
		params.Set("activities", opts.Activities)
	}

	resp, err := sess.Client.Get(ctx, "/Recordings", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var result SessionsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse sessions response: %w", err)
	}

	return &result, nil
}

// GetSession retrieves a specific PSM session.
func GetSession(ctx context.Context, sess *session.Session, sessionID string) (*PSMSession, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if sessionID == "" {
		return nil, fmt.Errorf("sessionID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Recordings/%s", url.PathEscape(sessionID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var psmSession PSMSession
	if err := json.Unmarshal(resp.Body, &psmSession); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	return &psmSession, nil
}

// ListLiveSessions retrieves live PSM sessions.
// This is equivalent to Get-PASPSMSession -LiveSession in psPAS.
func ListLiveSessions(ctx context.Context, sess *session.Session, opts ListOptions) (*SessionsResponse, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Search != "" {
		params.Set("search", opts.Search)
	}

	resp, err := sess.Client.Get(ctx, "/LiveSessions", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list live sessions: %w", err)
	}

	var result SessionsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse live sessions response: %w", err)
	}

	return &result, nil
}

// TerminateSession terminates a live PSM session.
// This is equivalent to Stop-PASPSMSession in psPAS.
func TerminateSession(ctx context.Context, sess *session.Session, liveSessionID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if liveSessionID == "" {
		return fmt.Errorf("liveSessionID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/LiveSessions/%s/Terminate", url.PathEscape(liveSessionID)), nil)
	if err != nil {
		return fmt.Errorf("failed to terminate session: %w", err)
	}

	return nil
}

// SuspendSession suspends a live PSM session.
// This is equivalent to Suspend-PASPSMSession in psPAS.
func SuspendSession(ctx context.Context, sess *session.Session, liveSessionID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if liveSessionID == "" {
		return fmt.Errorf("liveSessionID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/LiveSessions/%s/Suspend", url.PathEscape(liveSessionID)), nil)
	if err != nil {
		return fmt.Errorf("failed to suspend session: %w", err)
	}

	return nil
}

// ResumeSession resumes a suspended PSM session.
// This is equivalent to Resume-PASPSMSession in psPAS.
func ResumeSession(ctx context.Context, sess *session.Session, liveSessionID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if liveSessionID == "" {
		return fmt.Errorf("liveSessionID is required")
	}

	_, err := sess.Client.Post(ctx, fmt.Sprintf("/LiveSessions/%s/Resume", url.PathEscape(liveSessionID)), nil)
	if err != nil {
		return fmt.Errorf("failed to resume session: %w", err)
	}

	return nil
}

// GetRecording retrieves the recording file for a session.
// This is equivalent to Get-PASPSMRecording in psPAS.
func GetRecording(ctx context.Context, sess *session.Session, recordingID string) ([]byte, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if recordingID == "" {
		return nil, fmt.Errorf("recordingID is required")
	}

	resp, err := sess.Client.Post(ctx, fmt.Sprintf("/Recordings/%s/Play", url.PathEscape(recordingID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get recording: %w", err)
	}

	return resp.Body, nil
}

// SessionActivity represents an activity in a session.
type SessionActivity struct {
	Time     int64  `json:"Time"`
	Action   string `json:"Action"`
	Details  string `json:"Details,omitempty"`
	Username string `json:"Username,omitempty"`
}

// GetSessionActivities retrieves activities for a session.
// This is equivalent to Get-PASPSMSessionActivity in psPAS.
func GetSessionActivities(ctx context.Context, sess *session.Session, sessionID string) ([]SessionActivity, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if sessionID == "" {
		return nil, fmt.Errorf("sessionID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Recordings/%s/activities", url.PathEscape(sessionID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get session activities: %w", err)
	}

	var result struct {
		Activities []SessionActivity `json:"Activities"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse session activities response: %w", err)
	}

	return result.Activities, nil
}

// GetSessionProperties retrieves properties for a session.
// This is equivalent to Get-PASPSMSessionProperty in psPAS.
func GetSessionProperties(ctx context.Context, sess *session.Session, sessionID string) (map[string]string, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if sessionID == "" {
		return nil, fmt.Errorf("sessionID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Recordings/%s/properties", url.PathEscape(sessionID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get session properties: %w", err)
	}

	var result map[string]string
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse session properties response: %w", err)
	}

	return result, nil
}
