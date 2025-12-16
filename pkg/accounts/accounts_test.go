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

func TestList_AllOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        ListOptions
		checkParams func(t *testing.T, params map[string]string)
	}{
		{
			name: "with searchType option",
			opts: ListOptions{SearchType: "contains"},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["searchType"] != "contains" {
					t.Errorf("searchType param = %v, want contains", params["searchType"])
				}
			},
		},
		{
			name: "with sort option",
			opts: ListOptions{Sort: "name"},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["sort"] != "name" {
					t.Errorf("sort param = %v, want name", params["sort"])
				}
			},
		},
		{
			name: "with filter option",
			opts: ListOptions{Filter: "safeName eq TestSafe"},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["filter"] != "safeName eq TestSafe" {
					t.Errorf("filter param = %v, want safeName eq TestSafe", params["filter"])
				}
			},
		},
		{
			name: "with safeName option",
			opts: ListOptions{SafeName: "MySafe"},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["filter"] != "safeName eq MySafe" {
					t.Errorf("filter param = %v, want safeName eq MySafe", params["filter"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedParams map[string]string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedParams = make(map[string]string)
				for key, values := range r.URL.Query() {
					if len(values) > 0 {
						capturedParams[key] = values[0]
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&AccountsResponse{Value: []Account{}, Count: 0})
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			_, err := List(context.Background(), sess, tt.opts)
			if err != nil {
				t.Errorf("List() unexpected error: %v", err)
				return
			}

			tt.checkParams(t, capturedParams)
		})
	}
}

func TestGet_InvalidSession(t *testing.T) {
	_, err := Get(context.Background(), nil, "123")
	if err == nil {
		t.Error("Get() expected error for nil session, got nil")
	}
}

func TestCreate_InvalidSession(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateOptions{
		SafeName:   "safe",
		PlatformID: "platform",
		Address:    "server",
		UserName:   "user",
	})
	if err == nil {
		t.Error("Create() expected error for nil session, got nil")
	}
}

func TestUpdate_InvalidSession(t *testing.T) {
	_, err := Update(context.Background(), nil, "123", []PatchOperation{})
	if err == nil {
		t.Error("Update() expected error for nil session, got nil")
	}
}

func TestDelete_InvalidSession(t *testing.T) {
	err := Delete(context.Background(), nil, "123")
	if err == nil {
		t.Error("Delete() expected error for nil session, got nil")
	}
}

func TestGetPassword_InvalidSession(t *testing.T) {
	_, err := GetPassword(context.Background(), nil, "123", "testing")
	if err == nil {
		t.Error("GetPassword() expected error for nil session, got nil")
	}
}

func TestChangeCredentialsImmediately_InvalidSession(t *testing.T) {
	err := ChangeCredentialsImmediately(context.Background(), nil, "123", ChangeCredentialsOptions{})
	if err == nil {
		t.Error("ChangeCredentialsImmediately() expected error for nil session, got nil")
	}
}

func TestVerifyCredentials_InvalidSession(t *testing.T) {
	err := VerifyCredentials(context.Background(), nil, "123")
	if err == nil {
		t.Error("VerifyCredentials() expected error for nil session, got nil")
	}
}

func TestReconcileCredentials_InvalidSession(t *testing.T) {
	err := ReconcileCredentials(context.Background(), nil, "123")
	if err == nil {
		t.Error("ReconcileCredentials() expected error for nil session, got nil")
	}
}

func TestSetNextPassword_InvalidSession(t *testing.T) {
	err := SetNextPassword(context.Background(), nil, "123", "newpass")
	if err == nil {
		t.Error("SetNextPassword() expected error for nil session, got nil")
	}
}

func TestGetActivities_InvalidSession(t *testing.T) {
	_, err := GetActivities(context.Background(), nil, "123")
	if err == nil {
		t.Error("GetActivities() expected error for nil session, got nil")
	}
}

func TestGet_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	_, err := Get(context.Background(), sess, "123")
	if err == nil {
		t.Error("Get() expected error for server error, got nil")
	}
}

func TestCreate_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	_, err := Create(context.Background(), sess, CreateOptions{
		SafeName:   "safe",
		PlatformID: "platform",
		Address:    "server",
		UserName:   "user",
	})
	if err == nil {
		t.Error("Create() expected error for server error, got nil")
	}
}

func TestUpdate_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	_, err := Update(context.Background(), sess, "123", []PatchOperation{
		{Op: "replace", Path: "/name", Value: "newname"},
	})
	if err == nil {
		t.Error("Update() expected error for server error, got nil")
	}
}

func TestGetPassword_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	_, err := GetPassword(context.Background(), sess, "123", "testing")
	if err == nil {
		t.Error("GetPassword() expected error for server error, got nil")
	}
}

func TestChangeCredentialsImmediately_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	err := ChangeCredentialsImmediately(context.Background(), sess, "123", ChangeCredentialsOptions{})
	if err == nil {
		t.Error("ChangeCredentialsImmediately() expected error for server error, got nil")
	}
}

func TestVerifyCredentials_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	err := VerifyCredentials(context.Background(), sess, "123")
	if err == nil {
		t.Error("VerifyCredentials() expected error for server error, got nil")
	}
}

func TestReconcileCredentials_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	err := ReconcileCredentials(context.Background(), sess, "123")
	if err == nil {
		t.Error("ReconcileCredentials() expected error for server error, got nil")
	}
}

func TestSetNextPassword_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	err := SetNextPassword(context.Background(), sess, "123", "newpass")
	if err == nil {
		t.Error("SetNextPassword() expected error for server error, got nil")
	}
}

func TestGetActivities_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	_, err := GetActivities(context.Background(), sess, "123")
	if err == nil {
		t.Error("GetActivities() expected error for server error, got nil")
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

// Tests for linked_accounts.go

func TestLinkAccount(t *testing.T) {
	tests := []struct {
		name            string
		accountID       string
		linkedAccountID string
		opts            LinkAccountOptions
		serverStatus    int
		wantErr         bool
	}{
		{
			name:            "successful link",
			accountID:       "123",
			linkedAccountID: "456",
			opts:            LinkAccountOptions{Safe: "TestSafe", ExtraPassID: 1},
			serverStatus:    http.StatusOK,
			wantErr:         false,
		},
		{
			name:            "empty account ID",
			accountID:       "",
			linkedAccountID: "456",
			opts:            LinkAccountOptions{Safe: "TestSafe", ExtraPassID: 1},
			wantErr:         true,
		},
		{
			name:            "empty linked account ID",
			accountID:       "123",
			linkedAccountID: "",
			opts:            LinkAccountOptions{Safe: "TestSafe", ExtraPassID: 1},
			wantErr:         true,
		},
		{
			name:            "server error",
			accountID:       "123",
			linkedAccountID: "456",
			opts:            LinkAccountOptions{Safe: "TestSafe", ExtraPassID: 1},
			serverStatus:    http.StatusInternalServerError,
			wantErr:         true,
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

			err := LinkAccount(context.Background(), sess, tt.accountID, tt.linkedAccountID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("LinkAccount() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("LinkAccount() unexpected error: %v", err)
			}
		})
	}
}

func TestLinkAccount_InvalidSession(t *testing.T) {
	err := LinkAccount(context.Background(), nil, "123", "456", LinkAccountOptions{})
	if err == nil {
		t.Error("LinkAccount() expected error for nil session")
	}
}

func TestUnlinkAccount(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		extraPassID  int
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful unlink",
			accountID:    "123",
			extraPassID:  1,
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "unlink extraPassID 2",
			accountID:    "123",
			extraPassID:  2,
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "unlink extraPassID 3",
			accountID:    "123",
			extraPassID:  3,
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:        "empty account ID",
			accountID:   "",
			extraPassID: 1,
			wantErr:     true,
		},
		{
			name:        "invalid extraPassID 0",
			accountID:   "123",
			extraPassID: 0,
			wantErr:     true,
		},
		{
			name:        "invalid extraPassID 4",
			accountID:   "123",
			extraPassID: 4,
			wantErr:     true,
		},
		{
			name:         "server error",
			accountID:    "123",
			extraPassID:  1,
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
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

			err := UnlinkAccount(context.Background(), sess, tt.accountID, tt.extraPassID)
			if tt.wantErr {
				if err == nil {
					t.Error("UnlinkAccount() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UnlinkAccount() unexpected error: %v", err)
			}
		})
	}
}

func TestUnlinkAccount_InvalidSession(t *testing.T) {
	err := UnlinkAccount(context.Background(), nil, "123", 1)
	if err == nil {
		t.Error("UnlinkAccount() expected error for nil session")
	}
}

func TestGetLinkedAccounts(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		serverResponse []LinkedAccount
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful get linked accounts",
			accountID: "123",
			serverResponse: []LinkedAccount{
				{ID: "456", Name: "linked-account-1", SafeName: "TestSafe"},
				{ID: "789", Name: "linked-account-2", SafeName: "TestSafe"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:           "empty linked accounts",
			accountID:      "123",
			serverResponse: []LinkedAccount{},
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			accountID:    "123",
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
					LinkedAccounts []LinkedAccount `json:"LinkedAccounts"`
				}{LinkedAccounts: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()
			sess.Client = overrideAPIURL(t, sess.Client, server.URL)

			result, err := GetLinkedAccounts(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetLinkedAccounts() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetLinkedAccounts() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetLinkedAccounts() returned %d accounts, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGetLinkedAccounts_InvalidSession(t *testing.T) {
	_, err := GetLinkedAccounts(context.Background(), nil, "123")
	if err == nil {
		t.Error("GetLinkedAccounts() expected error for nil session")
	}
}

func TestDelete_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()
	sess.Client = overrideAPIURL(t, sess.Client, server.URL)

	err := Delete(context.Background(), sess, "123")
	if err == nil {
		t.Error("Delete() expected error for server error, got nil")
	}
}

func TestLinkedAccount_Struct(t *testing.T) {
	la := LinkedAccount{
		ID:         "123",
		Name:       "linked-account",
		SafeName:   "TestSafe",
		ExtraPass1: "extrapass1",
		ExtraPass2: "extrapass2",
		ExtraPass3: "extrapass3",
	}

	if la.ID.String() != "123" {
		t.Errorf("LinkedAccount.ID = %v, want 123", la.ID)
	}
	if la.Name != "linked-account" {
		t.Errorf("LinkedAccount.Name = %v, want linked-account", la.Name)
	}
}

func TestLinkAccountOptions_Struct(t *testing.T) {
	opts := LinkAccountOptions{
		Safe:        "TestSafe",
		ExtraPassID: 1,
		Name:        "linked-name",
		Folder:      "Root",
	}

	if opts.Safe != "TestSafe" {
		t.Errorf("LinkAccountOptions.Safe = %v, want TestSafe", opts.Safe)
	}
	if opts.ExtraPassID != 1 {
		t.Errorf("LinkAccountOptions.ExtraPassID = %v, want 1", opts.ExtraPassID)
	}
}

func TestListOptions_Struct(t *testing.T) {
	opts := ListOptions{
		Search:     "admin",
		SearchType: "contains",
		Sort:       "name",
		Offset:     10,
		Limit:      50,
		Filter:     "safeName eq Test",
		SafeName:   "TestSafe",
	}

	if opts.Search != "admin" {
		t.Errorf("ListOptions.Search = %v, want admin", opts.Search)
	}
	if opts.Limit != 50 {
		t.Errorf("ListOptions.Limit = %v, want 50", opts.Limit)
	}
}

func TestCreateOptions_Struct(t *testing.T) {
	opts := CreateOptions{
		Name:       "test-account",
		Address:    "server.example.com",
		UserName:   "admin",
		PlatformID: "WinServerLocal",
		SafeName:   "TestSafe",
		SecretType: "password",
		Secret:     "mysecret",
		PlatformAccountProperties: map[string]interface{}{
			"LogonDomain": "DOMAIN",
		},
		SecretManagement: &SecretManagement{
			AutomaticManagementEnabled: true,
		},
		RemoteMachinesAccess: &RemoteMachinesAccess{
			RemoteMachines:                   "server1,server2",
			AccessRestrictedToRemoteMachines: true,
		},
	}

	if opts.Name != "test-account" {
		t.Errorf("CreateOptions.Name = %v, want test-account", opts.Name)
	}
	if opts.SecretManagement == nil {
		t.Error("CreateOptions.SecretManagement should not be nil")
	}
}

func TestUpdateOptions_Struct(t *testing.T) {
	opts := UpdateOptions{
		Name:       "updated-account",
		Address:    "newserver.example.com",
		UserName:   "newadmin",
		PlatformID: "UnixSSH",
		PlatformAccountProperties: map[string]interface{}{
			"Port": "22",
		},
		SecretManagement: &SecretManagement{
			AutomaticManagementEnabled: false,
			ManualManagementReason:     "test reason",
		},
		RemoteMachinesAccess: &RemoteMachinesAccess{
			RemoteMachines: "server3",
		},
	}

	if opts.Name != "updated-account" {
		t.Errorf("UpdateOptions.Name = %v, want updated-account", opts.Name)
	}
}

func TestPatchOperation_Struct(t *testing.T) {
	ops := []PatchOperation{
		{Op: "replace", Path: "/name", Value: "newname"},
		{Op: "add", Path: "/property", Value: "value"},
		{Op: "remove", Path: "/oldprop"},
	}

	if ops[0].Op != "replace" {
		t.Errorf("PatchOperation.Op = %v, want replace", ops[0].Op)
	}
	if ops[2].Value != nil {
		t.Errorf("PatchOperation.Value for remove should be nil")
	}
}

func TestSecretManagement_Struct(t *testing.T) {
	sm := SecretManagement{
		AutomaticManagementEnabled: true,
		ManualManagementReason:     "test",
		Status:                     "verified",
		LastModifiedTime:           1705315800,
		LastReconciledTime:         1705315900,
		LastVerifiedTime:           1705316000,
	}

	if !sm.AutomaticManagementEnabled {
		t.Error("SecretManagement.AutomaticManagementEnabled should be true")
	}
	if sm.Status != "verified" {
		t.Errorf("SecretManagement.Status = %v, want verified", sm.Status)
	}
}

func TestRemoteMachinesAccess_Struct(t *testing.T) {
	rma := RemoteMachinesAccess{
		RemoteMachines:                   "server1,server2",
		AccessRestrictedToRemoteMachines: true,
	}

	if rma.RemoteMachines != "server1,server2" {
		t.Errorf("RemoteMachinesAccess.RemoteMachines = %v, want server1,server2", rma.RemoteMachines)
	}
	if !rma.AccessRestrictedToRemoteMachines {
		t.Error("RemoteMachinesAccess.AccessRestrictedToRemoteMachines should be true")
	}
}

func TestAccountActivity_Struct(t *testing.T) {
	activity := AccountActivity{
		Time:     1705315800,
		Action:   "Retrieve",
		ClientID: "client-123",
		ActionID: "action-456",
		Alert:    true,
		Reason:   "testing",
		UserName: "admin",
	}

	if activity.Action != "Retrieve" {
		t.Errorf("AccountActivity.Action = %v, want Retrieve", activity.Action)
	}
	if !activity.Alert {
		t.Error("AccountActivity.Alert should be true")
	}
}

func TestChangeCredentialsOptions_Struct(t *testing.T) {
	opts := ChangeCredentialsOptions{
		ChangeEntireGroup: true,
	}

	if !opts.ChangeEntireGroup {
		t.Error("ChangeCredentialsOptions.ChangeEntireGroup should be true")
	}
}

func TestAccount_Struct(t *testing.T) {
	account := Account{
		ID:                       "123",
		Name:                     "test-account",
		Address:                  "server.example.com",
		UserName:                 "admin",
		PlatformID:               "WinServerLocal",
		SafeName:                 "TestSafe",
		SecretType:               "password",
		Secret:                   "secret123",
		CreatedTime:              1705315800,
		CategoryModificationTime: 1705315900,
		PlatformAccountProperties: map[string]interface{}{
			"LogonDomain": "DOMAIN",
		},
		SecretManagement: &SecretManagement{
			AutomaticManagementEnabled: true,
		},
		RemoteMachinesAccess: &RemoteMachinesAccess{
			RemoteMachines: "server1",
		},
	}

	if account.Name != "test-account" {
		t.Errorf("Account.Name = %v, want test-account", account.Name)
	}
	if account.SecretManagement == nil {
		t.Error("Account.SecretManagement should not be nil")
	}
}

func TestAccountsResponse_Struct(t *testing.T) {
	resp := AccountsResponse{
		Value: []Account{
			{ID: "1", Name: "account1"},
			{ID: "2", Name: "account2"},
		},
		Count:    2,
		NextLink: "https://example.com/next",
	}

	if resp.Count != 2 {
		t.Errorf("AccountsResponse.Count = %v, want 2", resp.Count)
	}
	if len(resp.Value) != 2 {
		t.Errorf("AccountsResponse.Value length = %v, want 2", len(resp.Value))
	}
}
