// Package ipallowlist provides tests for IP allowlist management functionality.
package ipallowlist

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
		serverResponse IPAllowListResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			serverResponse: IPAllowListResponse{
				IPAllowList: []IPAllowListEntry{
					{IP: "192.168.1.1", Description: "Office"},
					{IP: "10.0.0.0/8", Description: "Internal network"},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty list",
			serverResponse: IPAllowListResponse{
				IPAllowList: []IPAllowListEntry{},
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

			result, err := List(context.Background(), sess)
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

			if len(result) != len(tt.serverResponse.IPAllowList) {
				t.Errorf("List() returned %d entries, want %d", len(result), len(tt.serverResponse.IPAllowList))
			}
		})
	}
}

func TestList_InvalidSession(t *testing.T) {
	_, err := List(context.Background(), nil)
	if err == nil {
		t.Error("List() expected error for nil session, got nil")
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name         string
		opts         AddOptions
		serverStatus int
		wantErr      bool
	}{
		{
			name: "successful add",
			opts: AddOptions{
				IP:          "192.168.1.100",
				Description: "New server",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "successful add without description",
			opts: AddOptions{
				IP: "10.0.0.1",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty IP",
			opts: AddOptions{
				IP: "",
			},
			wantErr: true,
		},
		{
			name: "server error",
			opts: AddOptions{
				IP: "192.168.1.100",
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

			err := Add(context.Background(), sess, tt.opts)
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
	err := Add(context.Background(), nil, AddOptions{IP: "192.168.1.1"})
	if err == nil {
		t.Error("Add() expected error for nil session, got nil")
	}
}

func TestRemove(t *testing.T) {
	requestCount := 0
	tests := []struct {
		name              string
		ip                string
		currentList       []IPAllowListEntry
		serverStatus      int
		wantErr           bool
		skipListOperation bool
	}{
		{
			name: "successful remove",
			ip:   "192.168.1.100",
			currentList: []IPAllowListEntry{
				{IP: "192.168.1.100", Description: "To remove"},
				{IP: "10.0.0.1", Description: "Keep"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:    "empty IP",
			ip:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount = 0
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(IPAllowListResponse{IPAllowList: tt.currentList})
				} else if r.Method == http.MethodPut {
					w.WriteHeader(tt.serverStatus)
				} else {
					t.Errorf("Unexpected method: %s", r.Method)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := Remove(context.Background(), sess, tt.ip)
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
	err := Remove(context.Background(), nil, "192.168.1.1")
	if err == nil {
		t.Error("Remove() expected error for nil session, got nil")
	}
}

func TestRemove_ListError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	err := Remove(context.Background(), sess, "192.168.1.1")
	if err == nil {
		t.Error("Remove() expected error when List fails, got nil")
	}
}

func TestRemove_UpdateError(t *testing.T) {
	requestCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if requestCount == 1 {
			// First request is List (GET)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(IPAllowListResponse{
				IPAllowList: []IPAllowListEntry{
					{IP: "192.168.1.100", Description: "Test"},
				},
			})
		} else {
			// Second request is Put (update)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	err := Remove(context.Background(), sess, "192.168.1.100")
	if err == nil {
		t.Error("Remove() expected error when update fails, got nil")
	}
}

func TestIPAllowListEntry_Struct(t *testing.T) {
	entry := IPAllowListEntry{
		IP:          "192.168.1.1",
		Description: "Test server",
	}

	if entry.IP != "192.168.1.1" {
		t.Errorf("IP = %v, want 192.168.1.1", entry.IP)
	}
	if entry.Description != "Test server" {
		t.Errorf("Description = %v, want Test server", entry.Description)
	}
}
