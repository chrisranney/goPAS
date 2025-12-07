// Package applications provides tests for application management functionality.
package applications

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
		opts           ListOptions
		serverResponse []Application
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListOptions{},
			serverResponse: []Application{
				{AppID: "App1", Description: "Application 1"},
				{AppID: "App2", Description: "Application 2"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with location",
			opts: ListOptions{Location: "\\Applications"},
			serverResponse: []Application{
				{AppID: "App1", Location: "\\Applications"},
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				response := ApplicationsResponse{Applications: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

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

			if len(result) != len(tt.serverResponse) {
				t.Errorf("List() returned %d applications, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		appID          string
		serverResponse *Application
		serverStatus   int
		wantErr        bool
	}{
		{
			name:  "successful get",
			appID: "App1",
			serverResponse: &Application{
				AppID:       "App1",
				Description: "Test Application",
				Location:    "\\Applications",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:    "empty app ID",
			appID:   "",
			wantErr: true,
		},
		{
			name:         "not found",
			appID:        "nonexistent",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					response := struct {
						Application Application `json:"application"`
					}{Application: *tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := Get(context.Background(), sess, tt.appID)
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

			if result.AppID != tt.serverResponse.AppID {
				t.Errorf("Get().AppID = %v, want %v", result.AppID, tt.serverResponse.AppID)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name         string
		opts         CreateOptions
		serverStatus int
		wantErr      bool
	}{
		{
			name: "successful create",
			opts: CreateOptions{
				AppID:       "NewApp",
				Description: "New Application",
				Location:    "\\Applications",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing app ID",
			opts: CreateOptions{
				Description: "New Application",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := Create(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Create() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Create() unexpected error: %v", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		appID        string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			appID:        "App1",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:    "empty app ID",
			appID:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := Delete(context.Background(), sess, tt.appID)
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

func TestListAuthMethods(t *testing.T) {
	tests := []struct {
		name           string
		appID          string
		serverResponse []AuthMethod
		serverStatus   int
		wantErr        bool
	}{
		{
			name:  "successful list",
			appID: "App1",
			serverResponse: []AuthMethod{
				{AppID: "App1", AuthType: "path", AuthValue: "/app"},
				{AppID: "App1", AuthType: "machineAddress", AuthValue: "192.168.1.1"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:    "empty app ID",
			appID:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				response := struct {
					Authentication []AuthMethod `json:"authentication"`
				}{Authentication: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListAuthMethods(context.Background(), sess, tt.appID)
			if tt.wantErr {
				if err == nil {
					t.Error("ListAuthMethods() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListAuthMethods() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("ListAuthMethods() returned %d methods, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestAddAuthMethod(t *testing.T) {
	tests := []struct {
		name         string
		appID        string
		opts         AddAuthMethodOptions
		serverStatus int
		wantErr      bool
	}{
		{
			name:  "successful add",
			appID: "App1",
			opts: AddAuthMethodOptions{
				AuthType:  "path",
				AuthValue: "/app/path",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name:  "empty app ID",
			appID: "",
			opts: AddAuthMethodOptions{
				AuthType:  "path",
				AuthValue: "/app/path",
			},
			wantErr: true,
		},
		{
			name:  "missing auth type",
			appID: "App1",
			opts: AddAuthMethodOptions{
				AuthValue: "/app/path",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := AddAuthMethod(context.Background(), sess, tt.appID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("AddAuthMethod() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddAuthMethod() unexpected error: %v", err)
			}
		})
	}
}

func TestRemoveAuthMethod(t *testing.T) {
	tests := []struct {
		name         string
		appID        string
		authID       string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			appID:        "App1",
			authID:       "auth-123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:    "empty app ID",
			appID:   "",
			authID:  "auth-123",
			wantErr: true,
		},
		{
			name:    "empty auth ID",
			appID:   "App1",
			authID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := RemoveAuthMethod(context.Background(), sess, tt.appID, tt.authID)
			if tt.wantErr {
				if err == nil {
					t.Error("RemoveAuthMethod() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RemoveAuthMethod() unexpected error: %v", err)
			}
		})
	}
}

func TestApplication_Struct(t *testing.T) {
	app := Application{
		AppID:               "TestApp",
		Description:         "Test Application",
		Location:            "\\Applications",
		AccessPermittedFrom: 0,
		AccessPermittedTo:   24,
		ExpirationDate:      "2025-12-31",
		Disabled:            false,
		BusinessOwnerFName:  "John",
		BusinessOwnerLName:  "Doe",
		BusinessOwnerEmail:  "john.doe@example.com",
		BusinessOwnerPhone:  "555-1234",
	}

	if app.AppID != "TestApp" {
		t.Errorf("AppID = %v, want TestApp", app.AppID)
	}
	if app.Disabled {
		t.Error("Disabled should be false")
	}
}

func TestAuthMethod_Struct(t *testing.T) {
	auth := AuthMethod{
		AppID:                "App1",
		AuthType:             "path",
		AuthValue:            "/app/path",
		Comment:              "Test auth method",
		IsFolder:             false,
		AllowInternalScripts: true,
	}

	if auth.AuthType != "path" {
		t.Errorf("AuthType = %v, want path", auth.AuthType)
	}
	if !auth.AllowInternalScripts {
		t.Error("AllowInternalScripts should be true")
	}
}
