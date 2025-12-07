// Package authentication provides tests for authentication functionality.
package authentication

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrisranney/gopas/internal/client"
	"github.com/chrisranney/gopas/internal/session"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name           string
		opts           SessionOptions
		serverResponse string
		serverStatus   int
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful login",
			opts: SessionOptions{
				BaseURL: "PLACEHOLDER",
				Credentials: Credentials{
					Username: "admin",
					Password: "password",
				},
			},
			serverResponse: `{"CyberArkLogonResult": "test-token-123"}`,
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name: "login with plain token response",
			opts: SessionOptions{
				BaseURL: "PLACEHOLDER",
				Credentials: Credentials{
					Username: "admin",
					Password: "password",
				},
			},
			serverResponse: `"test-token-456"`,
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name: "missing base URL",
			opts: SessionOptions{
				Credentials: Credentials{
					Username: "admin",
					Password: "password",
				},
			},
			wantErr:     true,
			errContains: "baseURL is required",
		},
		{
			name: "missing username",
			opts: SessionOptions{
				BaseURL: "https://cyberark.example.com",
				Credentials: Credentials{
					Password: "password",
				},
			},
			wantErr:     true,
			errContains: "username is required",
		},
		{
			name: "missing password",
			opts: SessionOptions{
				BaseURL: "https://cyberark.example.com",
				Credentials: Credentials{
					Username: "admin",
				},
			},
			wantErr:     true,
			errContains: "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.serverStatus != 0 {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.serverStatus)
					w.Write([]byte(tt.serverResponse))
				}))
				defer server.Close()
				tt.opts.BaseURL = server.URL
				tt.opts.SkipVersionCheck = true
			}

			sess, err := NewSession(context.Background(), tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("NewSession() expected error, got nil")
				}
				if tt.errContains != "" && err != nil && !containsString(err.Error(), tt.errContains) {
					t.Errorf("NewSession() error = %v, want containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("NewSession() unexpected error: %v", err)
				return
			}
			if sess == nil {
				t.Error("NewSession() returned nil session")
				return
			}
			if !sess.IsValid() {
				t.Error("NewSession() returned invalid session")
			}
		})
	}
}

func TestNewSession_AuthMethods(t *testing.T) {
	tests := []struct {
		name         string
		authMethod   AuthMethod
		expectedPath string
	}{
		{
			name:         "CyberArk auth",
			authMethod:   AuthMethodCyberArk,
			expectedPath: "/Auth/CyberArk/Logon",
		},
		{
			name:         "LDAP auth",
			authMethod:   AuthMethodLDAP,
			expectedPath: "/Auth/LDAP/Logon",
		},
		{
			name:         "RADIUS auth",
			authMethod:   AuthMethodRADIUS,
			expectedPath: "/Auth/RADIUS/Logon",
		},
		{
			name:         "Windows auth",
			authMethod:   AuthMethodWindows,
			expectedPath: "/Auth/Windows/Logon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"CyberArkLogonResult": "test-token"}`))
			}))
			defer server.Close()

			opts := SessionOptions{
				BaseURL: server.URL,
				Credentials: Credentials{
					Username: "admin",
					Password: "password",
				},
				AuthMethod:       tt.authMethod,
				SkipVersionCheck: true,
			}

			_, err := NewSession(context.Background(), opts)
			if err != nil {
				t.Errorf("NewSession() unexpected error: %v", err)
				return
			}

			// Check that the path contains the expected auth endpoint
			expectedSuffix := tt.expectedPath
			if !containsString(capturedPath, expectedSuffix) {
				t.Errorf("NewSession() used path %s, want containing %s", capturedPath, expectedSuffix)
			}
		})
	}
}

func TestCloseSession(t *testing.T) {
	tests := []struct {
		name         string
		sess         *session.Session
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful close",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "already logged out (401)",
			serverStatus: http.StatusUnauthorized,
			wantErr:      false, // Should not error
		},
		{
			name:    "nil session",
			sess:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sess *session.Session

			if tt.sess == nil && tt.name == "nil session" {
				sess = nil
			} else {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.serverStatus)
				}))
				defer server.Close()

				var err error
				sess, err = session.NewSession(server.URL)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				sess.SetAuthenticated("user", "token", "CyberArk")
			}

			err := CloseSession(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("CloseSession() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CloseSession() unexpected error: %v", err)
			}
		})
	}
}

func TestGetServerInfo(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse *ServerInfo
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful get info",
			serverResponse: &ServerInfo{
				ServerID:        "server-123",
				ServerName:      "CyberArkPAS",
				ExternalVersion: "14.0",
				InternalVersion: 14.0,
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			server := httptest.NewServer(handler)
			defer server.Close()

			sess, err := session.NewSession(server.URL)
			if err != nil {
				t.Fatalf("Failed to create session: %v", err)
			}

			result, err := GetServerInfo(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetServerInfo() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetServerInfo() unexpected error: %v", err)
				return
			}

			if result.ServerID != tt.serverResponse.ServerID {
				t.Errorf("GetServerInfo().ServerID = %v, want %v", result.ServerID, tt.serverResponse.ServerID)
			}
		})
	}
}

func TestGetServerInfo_NilSession(t *testing.T) {
	_, err := GetServerInfo(context.Background(), nil)
	if err == nil {
		t.Error("GetServerInfo() expected error for nil session")
	}
}

func TestGetComponentsHealth(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []ComponentHealth
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful get health",
			serverResponse: []ComponentHealth{
				{ComponentID: "1", ComponentName: "Vault", IsLoggedOn: true},
				{ComponentID: "2", ComponentName: "CPM", IsLoggedOn: true},
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				response := struct {
					Components []ComponentHealth `json:"Components"`
				}{Components: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			server := httptest.NewServer(handler)
			defer server.Close()

			sess, err := session.NewSession(server.URL)
			if err != nil {
				t.Fatalf("Failed to create session: %v", err)
			}
			sess.SetAuthenticated("user", "token", "CyberArk")

			result, err := GetComponentsHealth(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetComponentsHealth() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetComponentsHealth() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetComponentsHealth() returned %d components, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGetAuthPath(t *testing.T) {
	tests := []struct {
		method   AuthMethod
		expected string
	}{
		{AuthMethodCyberArk, "/Auth/CyberArk/Logon"},
		{AuthMethodLDAP, "/Auth/LDAP/Logon"},
		{AuthMethodRADIUS, "/Auth/RADIUS/Logon"},
		{AuthMethodWindows, "/Auth/Windows/Logon"},
		{AuthMethod("unknown"), "/Auth/CyberArk/Logon"}, // Default
	}

	for _, tt := range tests {
		result := getAuthPath(tt.method)
		if result != tt.expected {
			t.Errorf("getAuthPath(%v) = %v, want %v", tt.method, result, tt.expected)
		}
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"token"`, "token"},
		{"token", "token"},
		{`""`, ""},
		{"", ""},
		{`"`, `"`},
	}

	for _, tt := range tests {
		result := trimQuotes(tt.input)
		if result != tt.expected {
			t.Errorf("trimQuotes(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestAuthMethod_Constants(t *testing.T) {
	if AuthMethodCyberArk != "CyberArk" {
		t.Errorf("AuthMethodCyberArk = %v, want CyberArk", AuthMethodCyberArk)
	}
	if AuthMethodLDAP != "LDAP" {
		t.Errorf("AuthMethodLDAP = %v, want LDAP", AuthMethodLDAP)
	}
	if AuthMethodRADIUS != "RADIUS" {
		t.Errorf("AuthMethodRADIUS = %v, want RADIUS", AuthMethodRADIUS)
	}
	if AuthMethodWindows != "Windows" {
		t.Errorf("AuthMethodWindows = %v, want Windows", AuthMethodWindows)
	}
}

func TestLoginRequest_Struct(t *testing.T) {
	req := LoginRequest{
		Username:          "admin",
		Password:          "password123",
		ConcurrentSession: true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal LoginRequest: %v", err)
	}

	var parsed LoginRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal LoginRequest: %v", err)
	}

	if parsed.Username != req.Username {
		t.Errorf("Username = %v, want %v", parsed.Username, req.Username)
	}
	if parsed.ConcurrentSession != req.ConcurrentSession {
		t.Errorf("ConcurrentSession = %v, want %v", parsed.ConcurrentSession, req.ConcurrentSession)
	}
}

func TestServerInfo_Struct(t *testing.T) {
	info := ServerInfo{
		ServerID:         "server-123",
		ServerName:       "CyberArkPAS",
		ServicesUsed:     "All",
		ApplicationsUsed: "PAS",
		InternalVersion:  14.0,
		ExternalVersion:  "14.0.0",
	}

	if info.ServerID != "server-123" {
		t.Errorf("ServerID = %v, want server-123", info.ServerID)
	}
	if info.ExternalVersion != "14.0.0" {
		t.Errorf("ExternalVersion = %v, want 14.0.0", info.ExternalVersion)
	}
}

func TestComponentHealth_Struct(t *testing.T) {
	health := ComponentHealth{
		ComponentID:          "vault-1",
		ComponentName:        "Vault",
		Description:          "Primary Vault",
		ConnectedComponentID: "dr-vault-1",
		IsLoggedOn:           true,
		LastLogonDate:        1705315800,
	}

	if health.ComponentName != "Vault" {
		t.Errorf("ComponentName = %v, want Vault", health.ComponentName)
	}
	if !health.IsLoggedOn {
		t.Error("IsLoggedOn should be true")
	}
}

// Helper to create test client
func createTestClient(t *testing.T, serverURL string) *client.Client {
	c, err := client.NewClient(client.Config{BaseURL: serverURL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	c.SetAuthToken("test-token")
	return c
}

// Helper to check if string contains substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
