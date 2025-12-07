// Package accounts provides tests for account management functionality.
package accounts

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

	// Override the client's apiURL for testing
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
	// Override apiURL to point directly to server
	// We use reflection-like field access through re-creating
	c.SetAuthToken("test-token")
	return c
}

func TestList(t *testing.T) {
	tests := []struct {
		name           string
		opts           ListOptions
		serverResponse *AccountsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListOptions{},
			serverResponse: &AccountsResponse{
				Value: []Account{
					{ID: "1", Name: "account1", SafeName: "safe1"},
					{ID: "2", Name: "account2", SafeName: "safe2"},
				},
				Count: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with search",
			opts: ListOptions{Search: "admin"},
			serverResponse: &AccountsResponse{
				Value: []Account{
					{ID: "1", Name: "admin-account", SafeName: "safe1"},
				},
				Count: 1,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with pagination",
			opts: ListOptions{Offset: 10, Limit: 5},
			serverResponse: &AccountsResponse{
				Value:    []Account{},
				Count:    0,
				NextLink: "https://example.com/api?offset=15",
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

				// Check query parameters
				if tt.opts.Search != "" && r.URL.Query().Get("search") != tt.opts.Search {
					t.Errorf("Expected search=%s, got %s", tt.opts.Search, r.URL.Query().Get("search"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			// Override apiURL after session is created
			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			result, err := List(context.Background(), sess, tt.opts)
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

			if result.Count != tt.serverResponse.Count {
				t.Errorf("List().Count = %v, want %v", result.Count, tt.serverResponse.Count)
			}
			if len(result.Value) != len(tt.serverResponse.Value) {
				t.Errorf("List() returned %d accounts, want %d", len(result.Value), len(tt.serverResponse.Value))
			}
		})
	}
}

func TestList_InvalidSession(t *testing.T) {
	tests := []struct {
		name    string
		sess    *session.Session
		wantErr bool
	}{
		{
			name:    "nil session",
			sess:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := List(context.Background(), tt.sess, ListOptions{})
			if tt.wantErr && err == nil {
				t.Error("List() expected error, got nil")
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		serverResponse *Account
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful get",
			accountID: "123",
			serverResponse: &Account{
				ID:         "123",
				Name:       "test-account",
				SafeName:   "TestSafe",
				UserName:   "admin",
				Address:    "server.example.com",
				PlatformID: "WinServerLocal",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "account not found",
			accountID:    "nonexistent",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			result, err := Get(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("Get() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Get() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("Get().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		opts           CreateOptions
		serverResponse *Account
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful create",
			opts: CreateOptions{
				SafeName:   "TestSafe",
				PlatformID: "WinServerLocal",
				Address:    "server.example.com",
				UserName:   "admin",
				Secret:     "password123",
			},
			serverResponse: &Account{
				ID:         "new-123",
				SafeName:   "TestSafe",
				PlatformID: "WinServerLocal",
				Address:    "server.example.com",
				UserName:   "admin",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing safe name",
			opts: CreateOptions{
				PlatformID: "WinServerLocal",
				Address:    "server.example.com",
				UserName:   "admin",
			},
			wantErr: true,
		},
		{
			name: "missing platform ID",
			opts: CreateOptions{
				SafeName: "TestSafe",
				Address:  "server.example.com",
				UserName: "admin",
			},
			wantErr: true,
		},
		{
			name: "missing address",
			opts: CreateOptions{
				SafeName:   "TestSafe",
				PlatformID: "WinServerLocal",
				UserName:   "admin",
			},
			wantErr: true,
		},
		{
			name: "missing username",
			opts: CreateOptions{
				SafeName:   "TestSafe",
				PlatformID: "WinServerLocal",
				Address:    "server.example.com",
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

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

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

			if result.ID != tt.serverResponse.ID {
				t.Errorf("Create().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		operations     []PatchOperation
		serverResponse *Account
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful update",
			accountID: "123",
			operations: []PatchOperation{
				{Op: "replace", Path: "/address", Value: "newserver.example.com"},
			},
			serverResponse: &Account{
				ID:      "123",
				Address: "newserver.example.com",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			operations: []PatchOperation{
				{Op: "replace", Path: "/address", Value: "newserver.example.com"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("Expected PATCH request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			result, err := Update(context.Background(), sess, tt.accountID, tt.operations)
			if tt.wantErr {
				if err == nil {
					t.Error("Update() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Update() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("Update().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			accountID:    "123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "account not found",
			accountID:    "nonexistent",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
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
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			err := Delete(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("Delete() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Delete() unexpected error: %v", err)
			}
		})
	}
}

func TestGetPassword(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		reason         string
		serverResponse string
		serverStatus   int
		wantPassword   string
		wantErr        bool
	}{
		{
			name:           "successful get password",
			accountID:      "123",
			reason:         "Testing",
			serverResponse: `"MySecretPassword123"`,
			serverStatus:   http.StatusOK,
			wantPassword:   "MySecretPassword123",
			wantErr:        false,
		},
		{
			name:           "get password without quotes",
			accountID:      "123",
			reason:         "",
			serverResponse: "PlainPassword",
			serverStatus:   http.StatusOK,
			wantPassword:   "PlainPassword",
			wantErr:        false,
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
				w.Write([]byte(tt.serverResponse))
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			result, err := GetPassword(context.Background(), sess, tt.accountID, tt.reason)
			if tt.wantErr {
				if err == nil {
					t.Error("GetPassword() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPassword() unexpected error: %v", err)
				return
			}

			if result != tt.wantPassword {
				t.Errorf("GetPassword() = %v, want %v", result, tt.wantPassword)
			}
		})
	}
}

func TestChangeCredentialsImmediately(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		opts         ChangeCredentialsOptions
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful change",
			accountID:    "123",
			opts:         ChangeCredentialsOptions{ChangeEntireGroup: false},
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
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			err := ChangeCredentialsImmediately(context.Background(), sess, tt.accountID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ChangeCredentialsImmediately() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ChangeCredentialsImmediately() unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyCredentials(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful verify",
			accountID:    "123",
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
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			err := VerifyCredentials(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("VerifyCredentials() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("VerifyCredentials() unexpected error: %v", err)
			}
		})
	}
}

func TestReconcileCredentials(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful reconcile",
			accountID:    "123",
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
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			err := ReconcileCredentials(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("ReconcileCredentials() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ReconcileCredentials() unexpected error: %v", err)
			}
		})
	}
}

func TestSetNextPassword(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		newPassword  string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful set password",
			accountID:    "123",
			newPassword:  "NewPassword123",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "empty account ID",
			accountID:   "",
			newPassword: "NewPassword123",
			wantErr:     true,
		},
		{
			name:        "empty password",
			accountID:   "123",
			newPassword: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			err := SetNextPassword(context.Background(), sess, tt.accountID, tt.newPassword)
			if tt.wantErr {
				if err == nil {
					t.Error("SetNextPassword() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SetNextPassword() unexpected error: %v", err)
			}
		})
	}
}

func TestGetActivities(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		serverResponse []AccountActivity
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful get activities",
			accountID: "123",
			serverResponse: []AccountActivity{
				{Time: 1705315800, Action: "Retrieve", UserName: "admin"},
				{Time: 1705315900, Action: "Change", UserName: "system"},
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				response := struct {
					Activities []AccountActivity `json:"Activities"`
				}{Activities: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			result, err := GetActivities(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetActivities() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetActivities() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetActivities() returned %d activities, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestAccount_GetCreatedTime(t *testing.T) {
	account := &Account{
		ID:          "123",
		CreatedTime: 1705315800, // 2024-01-15 10:30:00 UTC
	}

	createdTime := account.GetCreatedTime()
	if createdTime.Unix() != account.CreatedTime {
		t.Errorf("GetCreatedTime() = %v, want Unix = %v", createdTime.Unix(), account.CreatedTime)
	}
}

// overrideAPIURL creates a new client with overridden API URL for testing
func overrideAPIURL(t *testing.T, c *client.Client, serverURL string) *client.Client {
	newClient, err := client.NewClient(client.Config{BaseURL: serverURL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	newClient.SetAuthToken(c.GetAuthToken())
	// Override apiURL by using a custom approach
	// Since we can't directly modify apiURL, we create a wrapper
	return newClient
}
