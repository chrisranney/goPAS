// Package accountacl provides tests for account ACL management functionality.
package accountacl

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
		accountID      string
		safeName       string
		folderName     string
		serverResponse []AccountACL
		serverStatus   int
		wantErr        bool
	}{
		{
			name:       "successful list",
			accountID:  "acc123",
			safeName:   "TestSafe",
			folderName: "Root",
			serverResponse: []AccountACL{
				{VaultUserName: "admin", Command: "ls"},
				{VaultUserName: "user1", Command: "cat"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "successful list with default folder",
			accountID:  "acc123",
			safeName:   "TestSafe",
			folderName: "",
			serverResponse: []AccountACL{
				{VaultUserName: "admin", Command: "ls"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			safeName:  "TestSafe",
			wantErr:   true,
		},
		{
			name:      "empty safe name",
			accountID: "acc123",
			safeName:  "",
			wantErr:   true,
		},
		{
			name:         "server error",
			accountID:    "acc123",
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
				response := AccountACLResponse{ListAccountPrivilegedCommandsResult: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := List(context.Background(), sess, tt.accountID, tt.safeName, tt.folderName)
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
				t.Errorf("List() returned %d ACLs, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestList_InvalidSession(t *testing.T) {
	_, err := List(context.Background(), nil, "acc123", "TestSafe", "")
	if err == nil {
		t.Error("List() expected error for nil session, got nil")
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		safeName     string
		folderName   string
		opts         AddOptions
		serverStatus int
		wantErr      bool
	}{
		{
			name:       "successful add",
			accountID:  "acc123",
			safeName:   "TestSafe",
			folderName: "Root",
			opts: AddOptions{
				Command:        "ls -la",
				PermissionType: "Allow",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "successful add with default folder",
			accountID:  "acc123",
			safeName:   "TestSafe",
			folderName: "",
			opts: AddOptions{
				Command: "ls",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			safeName:  "TestSafe",
			opts: AddOptions{
				Command: "ls",
			},
			wantErr: true,
		},
		{
			name:      "empty safe name",
			accountID: "acc123",
			safeName:  "",
			opts: AddOptions{
				Command: "ls",
			},
			wantErr: true,
		},
		{
			name:      "empty command",
			accountID: "acc123",
			safeName:  "TestSafe",
			opts:      AddOptions{},
			wantErr:   true,
		},
		{
			name:       "server error",
			accountID:  "acc123",
			safeName:   "TestSafe",
			folderName: "Root",
			opts: AddOptions{
				Command: "ls",
			},
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

			err := Add(context.Background(), sess, tt.accountID, tt.safeName, tt.folderName, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Add() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Add() unexpected error: %v", err)
			}
		})
	}
}

func TestAdd_InvalidSession(t *testing.T) {
	err := Add(context.Background(), nil, "acc123", "TestSafe", "", AddOptions{Command: "ls"})
	if err == nil {
		t.Error("Add() expected error for nil session, got nil")
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		safeName     string
		folderName   string
		aclID        string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			accountID:    "acc123",
			safeName:     "TestSafe",
			folderName:   "Root",
			aclID:        "acl456",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "successful remove with default folder",
			accountID:    "acc123",
			safeName:     "TestSafe",
			folderName:   "",
			aclID:        "acl456",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			safeName:  "TestSafe",
			aclID:     "acl456",
			wantErr:   true,
		},
		{
			name:      "empty safe name",
			accountID: "acc123",
			safeName:  "",
			aclID:     "acl456",
			wantErr:   true,
		},
		{
			name:      "empty ACL ID",
			accountID: "acc123",
			safeName:  "TestSafe",
			aclID:     "",
			wantErr:   true,
		},
		{
			name:         "server error",
			accountID:    "acc123",
			safeName:     "TestSafe",
			folderName:   "Root",
			aclID:        "acl456",
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

			err := Remove(context.Background(), sess, tt.accountID, tt.safeName, tt.folderName, tt.aclID)
			if tt.wantErr {
				if err == nil {
					t.Error("Remove() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Remove() unexpected error: %v", err)
			}
		})
	}
}

func TestRemove_InvalidSession(t *testing.T) {
	err := Remove(context.Background(), nil, "acc123", "TestSafe", "", "acl456")
	if err == nil {
		t.Error("Remove() expected error for nil session, got nil")
	}
}

func TestAccountACL_Struct(t *testing.T) {
	acl := AccountACL{
		VaultUserName:  "admin",
		SafeName:       "TestSafe",
		FolderName:     "Root",
		ObjectName:     "account1",
		Command:        "ls",
		CommandGroup:   false,
		PermissionType: "Allow",
		Restrictions:   "",
		IsGroup:        false,
		UserName:       "admin",
	}

	if acl.VaultUserName != "admin" {
		t.Errorf("VaultUserName = %v, want admin", acl.VaultUserName)
	}
	if acl.SafeName != "TestSafe" {
		t.Errorf("SafeName = %v, want TestSafe", acl.SafeName)
	}
}
