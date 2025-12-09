// Package sshkeys provides tests for SSH key management functionality.
package sshkeys

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

func TestGetUserPublicSSHKeys(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		serverResponse *PublicSSHKeysResponse
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name:   "successful get keys",
			userID: "123",
			serverResponse: &PublicSSHKeysResponse{
				PublicSSHKeys: []PublicSSHKey{
					{KeyID: "key1", PublicSSHKey: "ssh-rsa AAAA..."},
					{KeyID: "key2", PublicSSHKey: "ssh-ed25519 AAAA..."},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
		},
		{
			name:   "empty user ID",
			userID: "",
			wantErr: true,
		},
		{
			name:         "server error",
			userID:       "123",
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

			result, err := GetUserPublicSSHKeys(context.Background(), sess, tt.userID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetUserPublicSSHKeys() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetUserPublicSSHKeys() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("GetUserPublicSSHKeys() returned %d keys, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestAddUserPublicSSHKey(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		publicKey      string
		serverResponse *PublicSSHKey
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful add key",
			userID:    "123",
			publicKey: "ssh-rsa AAAA...",
			serverResponse: &PublicSSHKey{
				KeyID:        "new-key-1",
				PublicSSHKey: "ssh-rsa AAAA...",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name:      "empty user ID",
			userID:    "",
			publicKey: "ssh-rsa AAAA...",
			wantErr:   true,
		},
		{
			name:    "empty public key",
			userID:  "123",
			publicKey: "",
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

			result, err := AddUserPublicSSHKey(context.Background(), sess, tt.userID, tt.publicKey)
			if tt.wantErr {
				if err == nil {
					t.Error("AddUserPublicSSHKey() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddUserPublicSSHKey() unexpected error: %v", err)
				return
			}

			if result.KeyID != tt.serverResponse.KeyID {
				t.Errorf("AddUserPublicSSHKey().KeyID = %v, want %v", result.KeyID, tt.serverResponse.KeyID)
			}
		})
	}
}

func TestRemoveUserPublicSSHKey(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		keyID        string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			userID:       "123",
			keyID:        "key1",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			keyID:   "key1",
			wantErr: true,
		},
		{
			name:    "empty key ID",
			userID:  "123",
			keyID:   "",
			wantErr: true,
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

			err := RemoveUserPublicSSHKey(context.Background(), sess, tt.userID, tt.keyID)
			if tt.wantErr {
				if err == nil {
					t.Error("RemoveUserPublicSSHKey() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RemoveUserPublicSSHKey() unexpected error: %v", err)
			}
		})
	}
}

func TestGetAccountSSHKey(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		opts           GetAccountSSHKeyOptions
		serverResponse *AccountSSHKey
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful get key",
			accountID: "acc-123",
			opts:      GetAccountSSHKeyOptions{Reason: "Testing"},
			serverResponse: &AccountSSHKey{
				PrivateSSHKey: "-----BEGIN RSA PRIVATE KEY-----\n...",
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

			result, err := GetAccountSSHKey(context.Background(), sess, tt.accountID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("GetAccountSSHKey() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetAccountSSHKey() unexpected error: %v", err)
				return
			}

			if result.PrivateSSHKey != tt.serverResponse.PrivateSSHKey {
				t.Errorf("GetAccountSSHKey().PrivateSSHKey mismatch")
			}
		})
	}
}

func TestGeneratePrivateSSHKey(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		opts           GeneratePrivateSSHKeyOptions
		serverResponse *PrivateSSHKey
		serverStatus   int
		wantErr        bool
	}{
		{
			name:   "successful generate",
			userID: "123",
			opts:   GeneratePrivateSSHKeyOptions{Format: "OpenSSH", KeyAlgorithm: "RSA", KeySize: 4096},
			serverResponse: &PrivateSSHKey{
				ID:           "key-123",
				UserID:       "123",
				Format:       "OpenSSH",
				KeyAlgorithm: "RSA",
				KeySize:      4096,
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name:    "empty user ID",
			userID:  "",
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

			result, err := GeneratePrivateSSHKey(context.Background(), sess, tt.userID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("GeneratePrivateSSHKey() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GeneratePrivateSSHKey() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("GeneratePrivateSSHKey().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestRemovePrivateSSHKey(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		keyID        string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			userID:       "123",
			keyID:        "key1",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			keyID:   "key1",
			wantErr: true,
		},
		{
			name:    "empty key ID",
			userID:  "123",
			keyID:   "",
			wantErr: true,
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

			err := RemovePrivateSSHKey(context.Background(), sess, tt.userID, tt.keyID)
			if tt.wantErr {
				if err == nil {
					t.Error("RemovePrivateSSHKey() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RemovePrivateSSHKey() unexpected error: %v", err)
			}
		})
	}
}

func TestClearPrivateSSHKeys(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful clear",
			userID:       "123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
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

			err := ClearPrivateSSHKeys(context.Background(), sess, tt.userID)
			if tt.wantErr {
				if err == nil {
					t.Error("ClearPrivateSSHKeys() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ClearPrivateSSHKeys() unexpected error: %v", err)
			}
		})
	}
}

func TestListMFACachedSSHKeys(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		serverResponse []MFACachedSSHKey
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name:   "successful list",
			userID: "123",
			serverResponse: []MFACachedSSHKey{
				{ID: "cache1", CacheCreationTime: 1705315800, ExpirationTime: 1705402200},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    1,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
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
				response := struct {
					CachedSSHKeys []MFACachedSSHKey `json:"CachedSSHKeys"`
				}{CachedSSHKeys: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListMFACachedSSHKeys(context.Background(), sess, tt.userID)
			if tt.wantErr {
				if err == nil {
					t.Error("ListMFACachedSSHKeys() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListMFACachedSSHKeys() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListMFACachedSSHKeys() returned %d keys, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestGetUserPublicSSHKeys_InvalidSession(t *testing.T) {
	_, err := GetUserPublicSSHKeys(context.Background(), nil, "123")
	if err == nil {
		t.Error("GetUserPublicSSHKeys() with nil session expected error, got nil")
	}
}
