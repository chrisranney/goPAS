// Package eventsecurity provides tests for PTA event security functionality.
package eventsecurity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrisranney/gopas/internal/client"
	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
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

func TestListEvents(t *testing.T) {
	tests := []struct {
		name           string
		opts           ListEventsOptions
		serverResponse PTAEventsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list with no options",
			opts: ListEventsOptions{},
			serverResponse: PTAEventsResponse{
				PTAEvents: []PTAEvent{
					{ID: "evt1", Type: "SuspiciousActivity", Score: 85.5},
					{ID: "evt2", Type: "UnmanagedPrivilegedAccess", Score: 70.0},
				},
				Total: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "successful list with all options",
			opts: ListEventsOptions{
				FromDate:  1609459200,
				ToDate:    1612137600,
				Status:    "OPEN",
				AccountID: "acc123",
				Offset:    10,
				Limit:     50,
			},
			serverResponse: PTAEventsResponse{
				PTAEvents: []PTAEvent{
					{ID: "evt1", Type: "SuspiciousActivity", Score: 85.5, Status: "OPEN"},
				},
				Total: 1,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			opts:         ListEventsOptions{},
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

			result, err := ListEvents(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ListEvents() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListEvents() unexpected error: %v", err)
				return
			}

			if len(result.PTAEvents) != len(tt.serverResponse.PTAEvents) {
				t.Errorf("ListEvents() returned %d events, want %d", len(result.PTAEvents), len(tt.serverResponse.PTAEvents))
			}
		})
	}
}

func TestListEvents_InvalidSession(t *testing.T) {
	_, err := ListEvents(context.Background(), nil, ListEventsOptions{})
	if err == nil {
		t.Error("ListEvents() expected error for nil session, got nil")
	}
}

func TestGetEvent(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		serverResponse PTAEvent
		serverStatus   int
		wantErr        bool
	}{
		{
			name:    "successful get",
			eventID: "evt123",
			serverResponse: PTAEvent{
				ID:             "evt123",
				Type:           "SuspiciousActivity",
				Score:          85.5,
				MachineAddress: "192.168.1.100",
				UserName:       "admin",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:    "empty event ID",
			eventID: "",
			wantErr: true,
		},
		{
			name:         "server error",
			eventID:      "evt123",
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

			result, err := GetEvent(context.Background(), sess, tt.eventID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetEvent() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetEvent() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("GetEvent() returned ID %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestGetEvent_InvalidSession(t *testing.T) {
	_, err := GetEvent(context.Background(), nil, "evt123")
	if err == nil {
		t.Error("GetEvent() expected error for nil session, got nil")
	}
}

func TestSetEventStatus(t *testing.T) {
	tests := []struct {
		name         string
		eventID      string
		status       string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful update",
			eventID:      "evt123",
			status:       "CLOSED",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:    "empty event ID",
			eventID: "",
			status:  "CLOSED",
			wantErr: true,
		},
		{
			name:    "empty status",
			eventID: "evt123",
			status:  "",
			wantErr: true,
		},
		{
			name:         "server error",
			eventID:      "evt123",
			status:       "CLOSED",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("Expected PATCH request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := SetEventStatus(context.Background(), sess, tt.eventID, tt.status)
			if tt.wantErr {
				if err == nil {
					t.Error("SetEventStatus() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SetEventStatus() unexpected error: %v", err)
			}
		})
	}
}

func TestSetEventStatus_InvalidSession(t *testing.T) {
	err := SetEventStatus(context.Background(), nil, "evt123", "CLOSED")
	if err == nil {
		t.Error("SetEventStatus() expected error for nil session, got nil")
	}
}

func TestListRules(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []PTARule
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			serverResponse: []PTARule{
				{ID: "rule1", Name: "Suspicious Activity", Type: "Risky", Active: true},
				{ID: "rule2", Name: "Unmanaged Access", Type: "Alert", Active: false},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
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

			result, err := ListRules(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("ListRules() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListRules() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("ListRules() returned %d rules, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestListRules_InvalidSession(t *testing.T) {
	_, err := ListRules(context.Background(), nil)
	if err == nil {
		t.Error("ListRules() expected error for nil session, got nil")
	}
}

func TestSetRule(t *testing.T) {
	tests := []struct {
		name         string
		ruleID       string
		opts         SetRuleOptions
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful update",
			ruleID:       "rule123",
			opts:         SetRuleOptions{Active: true, Score: 85},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:    "empty rule ID",
			ruleID:  "",
			opts:    SetRuleOptions{Active: true},
			wantErr: true,
		},
		{
			name:         "server error",
			ruleID:       "rule123",
			opts:         SetRuleOptions{Active: true},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := SetRule(context.Background(), sess, tt.ruleID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("SetRule() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SetRule() unexpected error: %v", err)
			}
		})
	}
}

func TestSetRule_InvalidSession(t *testing.T) {
	err := SetRule(context.Background(), nil, "rule123", SetRuleOptions{Active: true})
	if err == nil {
		t.Error("SetRule() expected error for nil session, got nil")
	}
}

func TestListRemediations(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []PTARemediation
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			serverResponse: []PTARemediation{
				{ID: "rem1", Type: "Password", Name: "Password Rotation"},
				{ID: "rem2", Type: "Session", Name: "Session Termination"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
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

			result, err := ListRemediations(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("ListRemediations() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListRemediations() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("ListRemediations() returned %d remediations, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestListRemediations_InvalidSession(t *testing.T) {
	_, err := ListRemediations(context.Background(), nil)
	if err == nil {
		t.Error("ListRemediations() expected error for nil session, got nil")
	}
}

func TestGetPrivilegedUsers(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []PrivilegedUser
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			serverResponse: []PrivilegedUser{
				{ID: "user1", UserName: "admin", Source: "Vault"},
				{ID: "user2", UserName: "operator", Source: "LDAP"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
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

			result, err := GetPrivilegedUsers(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetPrivilegedUsers() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPrivilegedUsers() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetPrivilegedUsers() returned %d users, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGetPrivilegedUsers_InvalidSession(t *testing.T) {
	_, err := GetPrivilegedUsers(context.Background(), nil)
	if err == nil {
		t.Error("GetPrivilegedUsers() expected error for nil session, got nil")
	}
}

func TestAddPrivilegedUser(t *testing.T) {
	tests := []struct {
		name         string
		userName     string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful add",
			userName:     "newadmin",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty username",
			userName: "",
			wantErr:  true,
		},
		{
			name:         "server error",
			userName:     "newadmin",
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
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := AddPrivilegedUser(context.Background(), sess, tt.userName)
			if tt.wantErr {
				if err == nil {
					t.Error("AddPrivilegedUser() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddPrivilegedUser() unexpected error: %v", err)
			}
		})
	}
}

func TestAddPrivilegedUser_InvalidSession(t *testing.T) {
	err := AddPrivilegedUser(context.Background(), nil, "admin")
	if err == nil {
		t.Error("AddPrivilegedUser() expected error for nil session, got nil")
	}
}

func TestRemovePrivilegedUser(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			userID:       "user123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
		},
		{
			name:         "server error",
			userID:       "user123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := RemovePrivilegedUser(context.Background(), sess, tt.userID)
			if tt.wantErr {
				if err == nil {
					t.Error("RemovePrivilegedUser() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RemovePrivilegedUser() unexpected error: %v", err)
			}
		})
	}
}

func TestRemovePrivilegedUser_InvalidSession(t *testing.T) {
	err := RemovePrivilegedUser(context.Background(), nil, "user123")
	if err == nil {
		t.Error("RemovePrivilegedUser() expected error for nil session, got nil")
	}
}

func TestGetPrivilegedGroups(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []PrivilegedGroup
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			serverResponse: []PrivilegedGroup{
				{ID: "grp1", GroupName: "Administrators", Source: "Vault"},
				{ID: "grp2", GroupName: "Operators", Source: "LDAP"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
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

			result, err := GetPrivilegedGroups(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetPrivilegedGroups() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPrivilegedGroups() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetPrivilegedGroups() returned %d groups, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGetPrivilegedGroups_InvalidSession(t *testing.T) {
	_, err := GetPrivilegedGroups(context.Background(), nil)
	if err == nil {
		t.Error("GetPrivilegedGroups() expected error for nil session, got nil")
	}
}

func TestAddPrivilegedGroup(t *testing.T) {
	tests := []struct {
		name         string
		groupName    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful add",
			groupName:    "NewAdmins",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty group name",
			groupName: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			groupName:    "NewAdmins",
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
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := AddPrivilegedGroup(context.Background(), sess, tt.groupName)
			if tt.wantErr {
				if err == nil {
					t.Error("AddPrivilegedGroup() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddPrivilegedGroup() unexpected error: %v", err)
			}
		})
	}
}

func TestAddPrivilegedGroup_InvalidSession(t *testing.T) {
	err := AddPrivilegedGroup(context.Background(), nil, "Admins")
	if err == nil {
		t.Error("AddPrivilegedGroup() expected error for nil session, got nil")
	}
}

func TestRemovePrivilegedGroup(t *testing.T) {
	tests := []struct {
		name         string
		groupID      string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			groupID:      "grp123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:    "empty group ID",
			groupID: "",
			wantErr: true,
		},
		{
			name:         "server error",
			groupID:      "grp123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := RemovePrivilegedGroup(context.Background(), sess, tt.groupID)
			if tt.wantErr {
				if err == nil {
					t.Error("RemovePrivilegedGroup() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RemovePrivilegedGroup() unexpected error: %v", err)
			}
		})
	}
}

func TestRemovePrivilegedGroup_InvalidSession(t *testing.T) {
	err := RemovePrivilegedGroup(context.Background(), nil, "grp123")
	if err == nil {
		t.Error("RemovePrivilegedGroup() expected error for nil session, got nil")
	}
}

func TestPTAEvent_Struct(t *testing.T) {
	event := PTAEvent{
		ID:             "evt123",
		Type:           "SuspiciousActivity",
		Score:          85.5,
		EventTime:      1609459200,
		MachineAddress: "192.168.1.100",
		UserID:         "user456",
		UserName:       "admin",
		Status:         "OPEN",
		CloudData: &CloudData{
			CloudProvider: "AWS",
			CloudService:  "EC2",
			Region:        "us-east-1",
		},
		AffectedAccounts: []AffectedAccount{
			{AccountID: "acc1", AccountName: "root", SafeName: "Unix"},
		},
	}

	if event.ID != "evt123" {
		t.Errorf("ID = %v, want evt123", event.ID)
	}
	if event.Score != 85.5 {
		t.Errorf("Score = %v, want 85.5", event.Score)
	}
	if event.CloudData.CloudProvider != "AWS" {
		t.Errorf("CloudData.CloudProvider = %v, want AWS", event.CloudData.CloudProvider)
	}
}

func TestPTARule_Struct(t *testing.T) {
	rule := PTARule{
		ID:          types.FlexibleID("rule123"),
		Name:        "Test Rule",
		Description: "A test rule",
		Type:        "Risky",
		Active:      true,
		Score:       75,
	}

	if rule.Name != "Test Rule" {
		t.Errorf("Name = %v, want Test Rule", rule.Name)
	}
	if rule.Active != true {
		t.Errorf("Active = %v, want true", rule.Active)
	}
}
