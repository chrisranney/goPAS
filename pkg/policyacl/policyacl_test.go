// Package policyacl provides tests for policy ACL management functionality.
package policyacl

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
		policyID       string
		serverResponse []PolicyACL
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful list",
			policyID: "policy123",
			serverResponse: []PolicyACL{
				{PolicyID: "policy123", UserName: "admin", Command: "ls"},
				{PolicyID: "policy123", UserName: "user1", Command: "cat"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty policy ID",
			policyID: "",
			wantErr:  true,
		},
		{
			name:         "server error",
			policyID:     "policy123",
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
				response := PolicyACLResponse{PolicyACL: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := List(context.Background(), sess, tt.policyID)
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
	_, err := List(context.Background(), nil, "policy123")
	if err == nil {
		t.Error("List() expected error for nil session, got nil")
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name         string
		policyID     string
		opts         AddOptions
		serverStatus int
		wantErr      bool
	}{
		{
			name:     "successful add",
			policyID: "policy123",
			opts: AddOptions{
				Command:        "ls -la",
				PermissionType: "Allow",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty policy ID",
			policyID: "",
			opts: AddOptions{
				Command: "ls",
			},
			wantErr: true,
		},
		{
			name:     "empty command",
			policyID: "policy123",
			opts:     AddOptions{},
			wantErr:  true,
		},
		{
			name:     "server error",
			policyID: "policy123",
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

			err := Add(context.Background(), sess, tt.policyID, tt.opts)
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
	err := Add(context.Background(), nil, "policy123", AddOptions{Command: "ls"})
	if err == nil {
		t.Error("Add() expected error for nil session, got nil")
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name         string
		policyID     string
		aclID        string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			policyID:     "policy123",
			aclID:        "acl456",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:     "empty policy ID",
			policyID: "",
			aclID:    "acl456",
			wantErr:  true,
		},
		{
			name:     "empty ACL ID",
			policyID: "policy123",
			aclID:    "",
			wantErr:  true,
		},
		{
			name:         "server error",
			policyID:     "policy123",
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

			err := Remove(context.Background(), sess, tt.policyID, tt.aclID)
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
	err := Remove(context.Background(), nil, "policy123", "acl456")
	if err == nil {
		t.Error("Remove() expected error for nil session, got nil")
	}
}

func TestPolicyACL_Struct(t *testing.T) {
	acl := PolicyACL{
		PolicyID:       "policy123",
		UserName:       "admin",
		Command:        "ls",
		CommandGroup:   false,
		PermissionType: "Allow",
		Restrictions:   "",
		IsGroup:        false,
	}

	if acl.PolicyID != "policy123" {
		t.Errorf("PolicyID = %v, want policy123", acl.PolicyID)
	}
	if acl.UserName != "admin" {
		t.Errorf("UserName = %v, want admin", acl.UserName)
	}
}
