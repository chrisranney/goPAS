// Package accountgroups provides tests for account group management functionality.
package accountgroups

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

func TestList(t *testing.T) {
	tests := []struct {
		name           string
		safeName       string
		serverResponse []AccountGroup
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful list",
			safeName: "TestSafe",
			serverResponse: []AccountGroup{
				{GroupID: "1", GroupName: "Group1", Safe: "TestSafe"},
				{GroupID: "2", GroupName: "Group2", Safe: "TestSafe"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty safe name",
			safeName: "",
			wantErr:  true,
		},
		{
			name:         "server error",
			safeName:     "TestSafe",
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
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := List(context.Background(), sess, tt.safeName)
			if tt.wantErr {
				if err == nil {
					t.Error("List() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("List() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("List() returned %d groups, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestList_InvalidSession(t *testing.T) {
	_, err := List(context.Background(), nil, "TestSafe")
	if err == nil {
		t.Error("List() expected error for nil session, got nil")
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		opts           CreateOptions
		serverResponse *AccountGroup
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful create",
			opts: CreateOptions{
				GroupName:       "NewGroup",
				GroupPlatformID: "Platform1",
				Safe:            "TestSafe",
			},
			serverResponse: &AccountGroup{
				GroupID:         "new-123",
				GroupName:       "NewGroup",
				GroupPlatformID: "Platform1",
				Safe:            "TestSafe",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing group name",
			opts: CreateOptions{
				GroupPlatformID: "Platform1",
				Safe:            "TestSafe",
			},
			wantErr: true,
		},
		{
			name: "missing platform ID",
			opts: CreateOptions{
				GroupName: "NewGroup",
				Safe:      "TestSafe",
			},
			wantErr: true,
		},
		{
			name: "missing safe",
			opts: CreateOptions{
				GroupName:       "NewGroup",
				GroupPlatformID: "Platform1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := Create(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Create() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Create() unexpected error: %v", err)
				return
			}

			if result.GroupID != tt.serverResponse.GroupID {
				t.Errorf("Create().GroupID = %v, want %v", result.GroupID, tt.serverResponse.GroupID)
			}
		})
	}
}

func TestCreate_InvalidSession(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateOptions{
		GroupName:       "test",
		GroupPlatformID: "platform",
		Safe:            "safe",
	})
	if err == nil {
		t.Error("Create() expected error for nil session, got nil")
	}
}

func TestGetMembers(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		serverResponse []AccountGroupMember
		serverStatus   int
		wantErr        bool
	}{
		{
			name:    "successful get members",
			groupID: "group123",
			serverResponse: []AccountGroupMember{
				{AccountID: "acc1"},
				{AccountID: "acc2"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:    "empty group ID",
			groupID: "",
			wantErr: true,
		},
		{
			name:         "server error",
			groupID:      "group123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				response := struct {
					Members []AccountGroupMember `json:"Members"`
				}{Members: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetMembers(context.Background(), sess, tt.groupID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetMembers() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetMembers() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetMembers() returned %d members, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGetMembers_InvalidSession(t *testing.T) {
	_, err := GetMembers(context.Background(), nil, "group123")
	if err == nil {
		t.Error("GetMembers() expected error for nil session, got nil")
	}
}

func TestAddMember(t *testing.T) {
	tests := []struct {
		name         string
		groupID      string
		accountID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful add",
			groupID:      "group123",
			accountID:    "acc123",
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name:      "empty group ID",
			groupID:   "",
			accountID: "acc123",
			wantErr:   true,
		},
		{
			name:      "empty account ID",
			groupID:   "group123",
			accountID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			groupID:      "group123",
			accountID:    "acc123",
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

			err := AddMember(context.Background(), sess, tt.groupID, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("AddMember() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddMember() unexpected error: %v", err)
			}
		})
	}
}

func TestAddMember_InvalidSession(t *testing.T) {
	err := AddMember(context.Background(), nil, "group123", "acc123")
	if err == nil {
		t.Error("AddMember() expected error for nil session, got nil")
	}
}

func TestRemoveMember(t *testing.T) {
	tests := []struct {
		name         string
		groupID      string
		accountID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			groupID:      "group123",
			accountID:    "acc123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:      "empty group ID",
			groupID:   "",
			accountID: "acc123",
			wantErr:   true,
		},
		{
			name:      "empty account ID",
			groupID:   "group123",
			accountID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			groupID:      "group123",
			accountID:    "acc123",
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

			err := RemoveMember(context.Background(), sess, tt.groupID, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("RemoveMember() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RemoveMember() unexpected error: %v", err)
			}
		})
	}
}

func TestRemoveMember_InvalidSession(t *testing.T) {
	err := RemoveMember(context.Background(), nil, "group123", "acc123")
	if err == nil {
		t.Error("RemoveMember() expected error for nil session, got nil")
	}
}

func TestAccountGroup_Struct(t *testing.T) {
	group := AccountGroup{
		GroupID:         "1",
		GroupName:       "TestGroup",
		GroupPlatformID: "Platform1",
		Safe:            "TestSafe",
		Members: []AccountGroupMember{
			{AccountID: "acc1"},
		},
	}

	if group.GroupID != "1" {
		t.Errorf("GroupID = %v, want 1", group.GroupID)
	}
	if len(group.Members) != 1 {
		t.Errorf("Members length = %v, want 1", len(group.Members))
	}
}
