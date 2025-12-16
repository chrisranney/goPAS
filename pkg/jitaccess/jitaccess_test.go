// Package jitaccess provides tests for JIT access functionality.
package jitaccess

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

func TestRequestJITAccess(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		opts           JITAccessRequest
		serverResponse *JITAccess
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful request",
			accountID: "acc-123",
			opts:      JITAccessRequest{Reason: "Testing JIT access"},
			serverResponse: &JITAccess{
				AccountID:      "acc-123",
				Status:         "Granted",
				RequestTime:    1705315800,
				ExpirationTime: 1705319400,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "with ticketing info",
			accountID: "acc-123",
			opts: JITAccessRequest{
				Reason:              "Emergency access",
				TicketingSystemName: "ServiceNow",
				TicketID:            "INC0012345",
			},
			serverResponse: &JITAccess{
				AccountID: "acc-123",
				Status:    "Granted",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			accountID:    "acc-123",
			serverStatus: http.StatusForbidden,
			wantErr:      true,
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

			result, err := RequestJITAccess(context.Background(), sess, tt.accountID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("RequestJITAccess() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RequestJITAccess() unexpected error: %v", err)
				return
			}

			if result.AccountID != tt.serverResponse.AccountID {
				t.Errorf("RequestJITAccess().AccountID = %v, want %v", result.AccountID, tt.serverResponse.AccountID)
			}
		})
	}
}

func TestRevokeJITAccess(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful revoke",
			accountID:    "acc-123",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
		},
		{
			name:         "not found",
			accountID:    "nonexistent",
			serverStatus: http.StatusNotFound,
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

			err := RevokeJITAccess(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("RevokeJITAccess() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RevokeJITAccess() unexpected error: %v", err)
			}
		})
	}
}

func TestGetJITAccessStatus(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		serverResponse *JITAccessStatus
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful get with active access",
			accountID: "acc-123",
			serverResponse: &JITAccessStatus{
				IsJITEnabled:      true,
				HasActiveAccess:   true,
				ExpirationTime:    1705319400,
				RemainingDuration: 3600,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "no active access",
			accountID: "acc-456",
			serverResponse: &JITAccessStatus{
				IsJITEnabled:    true,
				HasActiveAccess: false,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
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

			result, err := GetJITAccessStatus(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetJITAccessStatus() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetJITAccessStatus() unexpected error: %v", err)
				return
			}

			if result.IsJITEnabled != tt.serverResponse.IsJITEnabled {
				t.Errorf("GetJITAccessStatus().IsJITEnabled = %v, want %v", result.IsJITEnabled, tt.serverResponse.IsJITEnabled)
			}
			if result.HasActiveAccess != tt.serverResponse.HasActiveAccess {
				t.Errorf("GetJITAccessStatus().HasActiveAccess = %v, want %v", result.HasActiveAccess, tt.serverResponse.HasActiveAccess)
			}
		})
	}
}

func TestListEPVUserAccess(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		serverResponse []EPVUserAccess
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name:      "successful list",
			accountID: "acc-123",
			serverResponse: []EPVUserAccess{
				{UserID: 1, Username: "admin", AccessType: "Full"},
				{UserID: 2, Username: "user1", AccessType: "ReadOnly"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
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
					response := struct {
						AccessGrants []EPVUserAccess `json:"AccessGrants"`
					}{AccessGrants: tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListEPVUserAccess(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("ListEPVUserAccess() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListEPVUserAccess() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListEPVUserAccess() returned %d grants, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestRequestJITAccess_InvalidSession(t *testing.T) {
	_, err := RequestJITAccess(context.Background(), nil, "acc-123", JITAccessRequest{})
	if err == nil {
		t.Error("RequestJITAccess() with nil session expected error, got nil")
	}
}

func TestRevokeJITAccess_InvalidSession(t *testing.T) {
	err := RevokeJITAccess(context.Background(), nil, "acc-123")
	if err == nil {
		t.Error("RevokeJITAccess() with nil session expected error, got nil")
	}
}

func TestGetJITAccessStatus_InvalidSession(t *testing.T) {
	_, err := GetJITAccessStatus(context.Background(), nil, "acc-123")
	if err == nil {
		t.Error("GetJITAccessStatus() with nil session expected error, got nil")
	}
}

func TestListEPVUserAccess_InvalidSession(t *testing.T) {
	_, err := ListEPVUserAccess(context.Background(), nil, "acc-123")
	if err == nil {
		t.Error("ListEPVUserAccess() with nil session expected error, got nil")
	}
}

func TestGetJITAccessStatus_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := GetJITAccessStatus(context.Background(), sess, "acc-123")
	if err == nil {
		t.Error("GetJITAccessStatus() expected error for server error")
	}
}

func TestListEPVUserAccess_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := ListEPVUserAccess(context.Background(), sess, "acc-123")
	if err == nil {
		t.Error("ListEPVUserAccess() expected error for server error")
	}
}

func TestJITAccessRequest_Struct(t *testing.T) {
	req := JITAccessRequest{
		Reason:              "Emergency access",
		TicketingSystemName: "ServiceNow",
		TicketID:            "INC0012345",
	}

	if req.Reason != "Emergency access" {
		t.Errorf("Reason = %v, want Emergency access", req.Reason)
	}
	if req.TicketID != "INC0012345" {
		t.Errorf("TicketID = %v, want INC0012345", req.TicketID)
	}
}

func TestJITAccess_Struct(t *testing.T) {
	access := JITAccess{
		AccountID:      "acc-123",
		Status:         "Granted",
		RequestTime:    1705315800,
		ExpirationTime: 1705319400,
	}

	if access.AccountID != "acc-123" {
		t.Errorf("AccountID = %v, want acc-123", access.AccountID)
	}
	if access.Status != "Granted" {
		t.Errorf("Status = %v, want Granted", access.Status)
	}
}

func TestJITAccessStatus_Struct(t *testing.T) {
	status := JITAccessStatus{
		IsJITEnabled:      true,
		HasActiveAccess:   true,
		ExpirationTime:    1705319400,
		RemainingDuration: 3600,
	}

	if !status.IsJITEnabled {
		t.Error("IsJITEnabled should be true")
	}
	if !status.HasActiveAccess {
		t.Error("HasActiveAccess should be true")
	}
	if status.RemainingDuration != 3600 {
		t.Errorf("RemainingDuration = %v, want 3600", status.RemainingDuration)
	}
}

func TestEPVUserAccess_Struct(t *testing.T) {
	access := EPVUserAccess{
		UserID:     1,
		Username:   "admin",
		AccessType: "Full",
	}

	if access.UserID != 1 {
		t.Errorf("UserID = %v, want 1", access.UserID)
	}
	if access.Username != "admin" {
		t.Errorf("Username = %v, want admin", access.Username)
	}
}
