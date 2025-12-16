// Package monitoring provides tests for PSM session monitoring functionality.
package monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrisranney/gopas/internal/client"
	"github.com/chrisranney/gopas/internal/session"
)

// createTestSession creates a test session with a mock server
func createTestSession(t *testing.T, handler http.Handler) (*session.Session, *httptest.Server) {
	server := httptest.NewServer(handler)

	sess, err := session.NewSession(server.URL)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	sess.Client = createTestClient(t, server.URL)
	sess.SetAuthenticated("testuser", "test-token", "CyberArk")

	return sess, server
}

// createTestClient creates a test client with mock server URL
func createTestClient(t *testing.T, serverURL string) *client.Client {
	c, err := client.NewClient(client.Config{BaseURL: serverURL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	c.SetAuthToken("test-token")
	return c
}

func TestListSessions(t *testing.T) {
	tests := []struct {
		name           string
		opts           ListOptions
		serverResponse SessionsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list with no options",
			opts: ListOptions{},
			serverResponse: SessionsResponse{
				Recordings: []PSMSession{
					{SessionID: "sess1", User: "admin", RemoteMachine: "server1"},
					{SessionID: "sess2", User: "operator", RemoteMachine: "server2"},
				},
				Total: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "successful list with all options",
			opts: ListOptions{
				FromTime:   1609459200,
				ToTime:     1612137600,
				Limit:      50,
				Offset:     10,
				Search:     "admin",
				Safe:       "PSMRecordings",
				Activities: "Connect",
			},
			serverResponse: SessionsResponse{
				Recordings: []PSMSession{
					{SessionID: "sess1", User: "admin"},
				},
				Total: 1,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			opts:         ListOptions{},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListSessions(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ListSessions() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListSessions() unexpected error: %v", err)
				return
			}

			if len(result.Recordings) != len(tt.serverResponse.Recordings) {
				t.Errorf("ListSessions() returned %d sessions, want %d", len(result.Recordings), len(tt.serverResponse.Recordings))
			}
		})
	}
}

func TestListSessions_InvalidSession(t *testing.T) {
	_, err := ListSessions(context.Background(), nil, ListOptions{})
	if err == nil {
		t.Error("ListSessions() expected error for nil session, got nil")
	}
}

func TestGetSession(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		serverResponse PSMSession
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful get",
			sessionID: "sess123",
			serverResponse: PSMSession{
				SessionID:     "sess123",
				User:          "admin",
				RemoteMachine: "server1.example.com",
				Protocol:      "RDP",
				Start:         1609459200,
				IsLive:        false,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty session ID",
			sessionID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			sessionID:    "sess123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetSession(context.Background(), sess, tt.sessionID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetSession() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetSession() unexpected error: %v", err)
				return
			}

			if result.SessionID != tt.serverResponse.SessionID {
				t.Errorf("GetSession() returned SessionID %v, want %v", result.SessionID, tt.serverResponse.SessionID)
			}
		})
	}
}

func TestGetSession_InvalidSession(t *testing.T) {
	_, err := GetSession(context.Background(), nil, "sess123")
	if err == nil {
		t.Error("GetSession() expected error for nil session, got nil")
	}
}

func TestListLiveSessions(t *testing.T) {
	tests := []struct {
		name           string
		opts           ListOptions
		serverResponse SessionsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListOptions{},
			serverResponse: SessionsResponse{
				Recordings: []PSMSession{
					{SessionID: "live1", User: "admin", IsLive: true, CanTerminate: true},
					{SessionID: "live2", User: "operator", IsLive: true, CanMonitor: true},
				},
				Total: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "successful list with options",
			opts: ListOptions{
				Limit:  10,
				Offset: 5,
				Search: "admin",
			},
			serverResponse: SessionsResponse{
				Recordings: []PSMSession{
					{SessionID: "live1", User: "admin", IsLive: true},
				},
				Total: 1,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			opts:         ListOptions{},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListLiveSessions(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ListLiveSessions() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListLiveSessions() unexpected error: %v", err)
				return
			}

			if len(result.Recordings) != len(tt.serverResponse.Recordings) {
				t.Errorf("ListLiveSessions() returned %d sessions, want %d", len(result.Recordings), len(tt.serverResponse.Recordings))
			}
		})
	}
}

func TestListLiveSessions_InvalidSession(t *testing.T) {
	_, err := ListLiveSessions(context.Background(), nil, ListOptions{})
	if err == nil {
		t.Error("ListLiveSessions() expected error for nil session, got nil")
	}
}

func TestTerminateSession(t *testing.T) {
	tests := []struct {
		name          string
		liveSessionID string
		serverStatus  int
		wantErr       bool
	}{
		{
			name:          "successful terminate",
			liveSessionID: "live123",
			serverStatus:  http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "empty session ID",
			liveSessionID: "",
			wantErr:       true,
		},
		{
			name:          "server error",
			liveSessionID: "live123",
			serverStatus:  http.StatusInternalServerError,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := TerminateSession(context.Background(), sess, tt.liveSessionID)
			if tt.wantErr {
				if err == nil {
					t.Error("TerminateSession() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("TerminateSession() unexpected error: %v", err)
			}
		})
	}
}

func TestTerminateSession_InvalidSession(t *testing.T) {
	err := TerminateSession(context.Background(), nil, "live123")
	if err == nil {
		t.Error("TerminateSession() expected error for nil session, got nil")
	}
}

func TestSuspendSession(t *testing.T) {
	tests := []struct {
		name          string
		liveSessionID string
		serverStatus  int
		wantErr       bool
	}{
		{
			name:          "successful suspend",
			liveSessionID: "live123",
			serverStatus:  http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "empty session ID",
			liveSessionID: "",
			wantErr:       true,
		},
		{
			name:          "server error",
			liveSessionID: "live123",
			serverStatus:  http.StatusInternalServerError,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := SuspendSession(context.Background(), sess, tt.liveSessionID)
			if tt.wantErr {
				if err == nil {
					t.Error("SuspendSession() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SuspendSession() unexpected error: %v", err)
			}
		})
	}
}

func TestSuspendSession_InvalidSession(t *testing.T) {
	err := SuspendSession(context.Background(), nil, "live123")
	if err == nil {
		t.Error("SuspendSession() expected error for nil session, got nil")
	}
}

func TestResumeSession(t *testing.T) {
	tests := []struct {
		name          string
		liveSessionID string
		serverStatus  int
		wantErr       bool
	}{
		{
			name:          "successful resume",
			liveSessionID: "live123",
			serverStatus:  http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "empty session ID",
			liveSessionID: "",
			wantErr:       true,
		},
		{
			name:          "server error",
			liveSessionID: "live123",
			serverStatus:  http.StatusInternalServerError,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := ResumeSession(context.Background(), sess, tt.liveSessionID)
			if tt.wantErr {
				if err == nil {
					t.Error("ResumeSession() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ResumeSession() unexpected error: %v", err)
			}
		})
	}
}

func TestResumeSession_InvalidSession(t *testing.T) {
	err := ResumeSession(context.Background(), nil, "live123")
	if err == nil {
		t.Error("ResumeSession() expected error for nil session, got nil")
	}
}

func TestGetRecording(t *testing.T) {
	tests := []struct {
		name           string
		recordingID    string
		serverResponse []byte
		serverStatus   int
		wantErr        bool
	}{
		{
			name:           "successful get",
			recordingID:    "rec123",
			serverResponse: []byte("recording data"),
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name:        "empty recording ID",
			recordingID: "",
			wantErr:     true,
		},
		{
			name:         "server error",
			recordingID:  "rec123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
				w.Write(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetRecording(context.Background(), sess, tt.recordingID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetRecording() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetRecording() unexpected error: %v", err)
				return
			}

			if string(result) != string(tt.serverResponse) {
				t.Errorf("GetRecording() returned %v, want %v", string(result), string(tt.serverResponse))
			}
		})
	}
}

func TestGetRecording_InvalidSession(t *testing.T) {
	_, err := GetRecording(context.Background(), nil, "rec123")
	if err == nil {
		t.Error("GetRecording() expected error for nil session, got nil")
	}
}

func TestGetSessionActivities(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		serverResponse struct {
			Activities []SessionActivity `json:"Activities"`
		}
		serverStatus int
		wantErr      bool
	}{
		{
			name:      "successful get",
			sessionID: "sess123",
			serverResponse: struct {
				Activities []SessionActivity `json:"Activities"`
			}{
				Activities: []SessionActivity{
					{Time: 1609459200, Action: "Connect", Username: "admin"},
					{Time: 1609459260, Action: "Command", Details: "ls -la"},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty session ID",
			sessionID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			sessionID:    "sess123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetSessionActivities(context.Background(), sess, tt.sessionID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetSessionActivities() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetSessionActivities() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse.Activities) {
				t.Errorf("GetSessionActivities() returned %d activities, want %d", len(result), len(tt.serverResponse.Activities))
			}
		})
	}
}

func TestGetSessionActivities_InvalidSession(t *testing.T) {
	_, err := GetSessionActivities(context.Background(), nil, "sess123")
	if err == nil {
		t.Error("GetSessionActivities() expected error for nil session, got nil")
	}
}

func TestGetSessionProperties(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		serverResponse map[string]string
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful get",
			sessionID: "sess123",
			serverResponse: map[string]string{
				"User":          "admin",
				"RemoteMachine": "server1.example.com",
				"Protocol":      "RDP",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty session ID",
			sessionID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			sessionID:    "sess123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetSessionProperties(context.Background(), sess, tt.sessionID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetSessionProperties() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetSessionProperties() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetSessionProperties() returned %d properties, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGetSessionProperties_InvalidSession(t *testing.T) {
	_, err := GetSessionProperties(context.Background(), nil, "sess123")
	if err == nil {
		t.Error("GetSessionProperties() expected error for nil session, got nil")
	}
}

func TestPSMSession_Struct(t *testing.T) {
	session := PSMSession{
		SessionID:           "sess123",
		SessionGuid:         "guid-123",
		SafeName:            "PSMRecordings",
		AccountID:           "acc456",
		AccountName:         "root",
		User:                "admin",
		RemoteMachine:       "server1.example.com",
		Protocol:            "RDP",
		Client:              "PSMGW",
		ClientIP:            "192.168.1.100",
		ConnectionComponent: "PSM-RDP",
		Start:               1609459200,
		End:                 1609462800,
		Duration:            3600,
		FromIP:              "10.0.0.1",
		RiskScore:           25.5,
		IsLive:              false,
		CanTerminate:        true,
		CanMonitor:          true,
		CanPlayback:         true,
		RecordingFiles: []RecordingFile{
			{FileName: "recording.avi", RecordingType: "Video", Format: "AVI"},
		},
		Properties: map[string]string{"custom": "value"},
	}

	if session.SessionID != "sess123" {
		t.Errorf("SessionID = %v, want sess123", session.SessionID)
	}
	if session.Protocol != "RDP" {
		t.Errorf("Protocol = %v, want RDP", session.Protocol)
	}
	if len(session.RecordingFiles) != 1 {
		t.Errorf("RecordingFiles length = %v, want 1", len(session.RecordingFiles))
	}
}

func TestSessionActivity_Struct(t *testing.T) {
	activity := SessionActivity{
		Time:     1609459200,
		Action:   "Command",
		Details:  "ls -la",
		Username: "admin",
	}

	if activity.Action != "Command" {
		t.Errorf("Action = %v, want Command", activity.Action)
	}
	if activity.Details != "ls -la" {
		t.Errorf("Details = %v, want ls -la", activity.Details)
	}
}
