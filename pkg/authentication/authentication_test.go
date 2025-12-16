// Package authentication provides tests for authentication functionality.
package authentication

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestNewSession_AuthenticationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"ErrorCode":"PASWS001E","ErrorMessage":"Authentication failed"}`))
	}))
	defer server.Close()

	opts := SessionOptions{
		BaseURL: server.URL,
		Credentials: Credentials{
			Username: "admin",
			Password: "wrongpassword",
		},
		SkipVersionCheck: true,
	}

	_, err := NewSession(context.Background(), opts)
	if err == nil {
		t.Error("NewSession() expected error for authentication failure, got nil")
	}
}

func TestNewSession_EmptyTokenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"CyberArkLogonResult": ""}`))
	}))
	defer server.Close()

	opts := SessionOptions{
		BaseURL: server.URL,
		Credentials: Credentials{
			Username: "admin",
			Password: "password",
		},
		SkipVersionCheck: true,
	}

	_, err := NewSession(context.Background(), opts)
	if err == nil {
		t.Error("NewSession() expected error for empty token, got nil")
	}
}

func TestNewSession_WithVersionCheck(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if containsString(r.URL.Path, "/Auth/") {
			w.Write([]byte(`{"CyberArkLogonResult": "test-token"}`))
		} else if containsString(r.URL.Path, "/Server") {
			w.Write([]byte(`{"ServerID":"server-1","ExternalVersion":"14.0"}`))
		}
	}))
	defer server.Close()

	opts := SessionOptions{
		BaseURL: server.URL,
		Credentials: Credentials{
			Username: "admin",
			Password: "password",
		},
		SkipVersionCheck: false,
	}

	sess, err := NewSession(context.Background(), opts)
	if err != nil {
		t.Errorf("NewSession() unexpected error: %v", err)
		return
	}

	if sess == nil {
		t.Error("NewSession() returned nil session")
		return
	}

	// Should have made at least 2 requests (login + version check)
	if requestCount < 2 {
		t.Errorf("Expected at least 2 requests, got %d", requestCount)
	}
}

func TestNewSession_WithConcurrentSession(t *testing.T) {
	var capturedBody LoginRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if containsString(r.URL.Path, "/Auth/") {
			json.NewDecoder(r.Body).Decode(&capturedBody)
		}
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
		ConcurrentSession: true,
		SkipVersionCheck:  true,
	}

	_, err := NewSession(context.Background(), opts)
	if err != nil {
		t.Errorf("NewSession() unexpected error: %v", err)
		return
	}

	if !capturedBody.ConcurrentSession {
		t.Error("Expected ConcurrentSession to be true in request body")
	}
}

func TestNewSession_DefaultAuthMethod(t *testing.T) {
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
		// AuthMethod not set - should default to CyberArk
		SkipVersionCheck: true,
	}

	_, err := NewSession(context.Background(), opts)
	if err != nil {
		t.Errorf("NewSession() unexpected error: %v", err)
		return
	}

	if !containsString(capturedPath, "/Auth/CyberArk/Logon") {
		t.Errorf("Default auth method should be CyberArk, got path: %s", capturedPath)
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

func TestCloseSession_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sess, err := session.NewSession(server.URL)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sess.SetAuthenticated("user", "token", "CyberArk")

	err = CloseSession(context.Background(), sess)
	if err == nil {
		t.Error("CloseSession() expected error for server error, got nil")
	}
}

func TestGetComponentsHealth_InvalidSession(t *testing.T) {
	_, err := GetComponentsHealth(context.Background(), nil)
	if err == nil {
		t.Error("GetComponentsHealth() expected error for nil session")
	}
}

func TestGetServerInfo_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	sess, err := session.NewSession(server.URL)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	_, err = GetServerInfo(context.Background(), sess)
	if err == nil {
		t.Error("GetServerInfo() expected error for invalid JSON, got nil")
	}
}

func TestGetComponentsHealth_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	sess, err := session.NewSession(server.URL)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sess.SetAuthenticated("user", "token", "CyberArk")

	_, err = GetComponentsHealth(context.Background(), sess)
	if err == nil {
		t.Error("GetComponentsHealth() expected error for invalid JSON, got nil")
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

// Helper to check if string contains substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGetLoggedOnUser(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse *LoggedOnUser
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful get",
			serverResponse: &LoggedOnUser{
				ID:                 1,
				Username:           "admin",
				Source:             "CyberArk",
				UserType:           "EPVUser",
				ComponentUser:      false,
				VaultAuthorization: []string{"AddUpdateUsers", "ResetUsersPasswords"},
				Location:           "\\",
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
			sess.SetAuthenticated("user", "token", "CyberArk")

			result, err := GetLoggedOnUser(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetLoggedOnUser() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetLoggedOnUser() unexpected error: %v", err)
				return
			}

			if result.Username != tt.serverResponse.Username {
				t.Errorf("GetLoggedOnUser().Username = %v, want %v", result.Username, tt.serverResponse.Username)
			}
		})
	}
}

func TestGetLoggedOnUser_InvalidSession(t *testing.T) {
	_, err := GetLoggedOnUser(context.Background(), nil)
	if err == nil {
		t.Error("GetLoggedOnUser() expected error for nil session")
	}
}

func TestGetUserLoginInfo(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse *UserLoginInfo
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful get",
			serverResponse: &UserLoginInfo{
				Username:           "admin",
				UserType:           "EPVUser",
				InVaultGUID:        "guid-123",
				ComponentUser:      false,
				LoginTime:          1705315800,
				VaultAuthorization: []string{"AddUpdateUsers"},
				Location:           "\\",
				AuthenticationMethods: []string{"CyberArk"},
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
			sess.SetAuthenticated("user", "token", "CyberArk")

			result, err := GetUserLoginInfo(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetUserLoginInfo() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetUserLoginInfo() unexpected error: %v", err)
				return
			}

			if result.Username != tt.serverResponse.Username {
				t.Errorf("GetUserLoginInfo().Username = %v, want %v", result.Username, tt.serverResponse.Username)
			}
		})
	}
}

func TestGetUserLoginInfo_InvalidSession(t *testing.T) {
	_, err := GetUserLoginInfo(context.Background(), nil)
	if err == nil {
		t.Error("GetUserLoginInfo() expected error for nil session")
	}
}

func TestGetSession(t *testing.T) {
	tests := []struct {
		name     string
		sess     *session.Session
		wantNil  bool
	}{
		{
			name:    "nil session",
			sess:    nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSession(tt.sess)
			if tt.wantNil && result != nil {
				t.Error("GetSession() expected nil, got non-nil")
			}
		})
	}
}

func TestGetSession_ValidSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sess, err := session.NewSession(server.URL)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sess.SetAuthenticated("testuser", "test-token", "CyberArk")

	result := GetSession(sess)
	if result == nil {
		t.Error("GetSession() returned nil for valid session")
		return
	}

	if result["User"] != "testuser" {
		t.Errorf("GetSession()[User] = %v, want testuser", result["User"])
	}
	if result["AuthMethod"] != "CyberArk" {
		t.Errorf("GetSession()[AuthMethod] = %v, want CyberArk", result["AuthMethod"])
	}
	if result["IsAuthenticated"] != true {
		t.Errorf("GetSession()[IsAuthenticated] = %v, want true", result["IsAuthenticated"])
	}
}

func TestUseSession(t *testing.T) {
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
			result, err := UseSession(tt.sess)
			if tt.wantErr {
				if err == nil {
					t.Error("UseSession() expected error, got nil")
				}
				return
			}
			if result == nil {
				t.Error("UseSession() returned nil")
			}
		})
	}
}

func TestUseSession_ValidSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sess, err := session.NewSession(server.URL)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sess.SetAuthenticated("testuser", "test-token", "CyberArk")

	result, err := UseSession(sess)
	if err != nil {
		t.Errorf("UseSession() unexpected error: %v", err)
		return
	}
	if result != sess {
		t.Error("UseSession() should return the same session")
	}
}

func TestUseSession_InvalidSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sess, err := session.NewSession(server.URL)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	// Not authenticated - session is not valid

	_, err = UseSession(sess)
	if err == nil {
		t.Error("UseSession() expected error for unauthenticated session")
	}
}

func TestLoggedOnUser_Struct(t *testing.T) {
	user := LoggedOnUser{
		ID:                      1,
		Username:                "admin",
		Source:                  "CyberArk",
		UserType:                "EPVUser",
		ComponentUser:           false,
		VaultAuthorization:      []string{"AddUpdateUsers", "ResetUsersPasswords"},
		Location:                "\\",
		AgentUser:               false,
		Disabled:                false,
		Suspended:               false,
		LastSuccessfulLoginDate: 1705315800,
	}

	if user.Username != "admin" {
		t.Errorf("Username = %v, want admin", user.Username)
	}
	if len(user.VaultAuthorization) != 2 {
		t.Errorf("VaultAuthorization length = %v, want 2", len(user.VaultAuthorization))
	}
}

func TestUserLoginInfo_Struct(t *testing.T) {
	info := UserLoginInfo{
		Username:              "admin",
		UserType:              "EPVUser",
		InVaultGUID:           "guid-123",
		ComponentUser:         false,
		LoginTime:             1705315800,
		LastFailedLogin:       1705315700,
		VaultAuthorization:    []string{"AddUpdateUsers"},
		Location:              "\\",
		AuthenticationMethods: []string{"CyberArk", "LDAP"},
	}

	if info.Username != "admin" {
		t.Errorf("Username = %v, want admin", info.Username)
	}
	if len(info.AuthenticationMethods) != 2 {
		t.Errorf("AuthenticationMethods length = %v, want 2", len(info.AuthenticationMethods))
	}
}

// Tests for saml.go

func TestNewSAMLSession(t *testing.T) {
	tests := []struct {
		name           string
		opts           SAMLSessionOptions
		serverResponse string
		serverStatus   int
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful SAML login",
			opts: SAMLSessionOptions{
				BaseURL:      "PLACEHOLDER",
				SAMLResponse: "base64-encoded-saml-response",
			},
			serverResponse: `{"CyberArkLogonResult": "saml-token-123"}`,
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name: "SAML login with plain token response",
			opts: SAMLSessionOptions{
				BaseURL:      "PLACEHOLDER",
				SAMLResponse: "base64-saml-response",
			},
			serverResponse: `"saml-token-456"`,
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name: "missing base URL",
			opts: SAMLSessionOptions{
				SAMLResponse: "base64-saml-response",
			},
			wantErr:     true,
			errContains: "baseURL is required",
		},
		{
			name: "missing SAML response without integrated auth",
			opts: SAMLSessionOptions{
				BaseURL: "https://cyberark.example.com",
			},
			wantErr:     true,
			errContains: "SAMLResponse is required",
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
			}

			sess, err := NewSAMLSession(context.Background(), tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("NewSAMLSession() expected error, got nil")
				}
				if tt.errContains != "" && err != nil && !containsString(err.Error(), tt.errContains) {
					t.Errorf("NewSAMLSession() error = %v, want containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("NewSAMLSession() unexpected error: %v", err)
				return
			}
			if sess == nil {
				t.Error("NewSAMLSession() returned nil session")
				return
			}
			if !sess.IsValid() {
				t.Error("NewSAMLSession() returned invalid session")
			}
		})
	}
}

func TestNewSAMLSession_AuthenticationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"ErrorCode":"PASWS001E","ErrorMessage":"SAML Authentication failed"}`))
	}))
	defer server.Close()

	opts := SAMLSessionOptions{
		BaseURL:      server.URL,
		SAMLResponse: "invalid-saml-response",
	}

	_, err := NewSAMLSession(context.Background(), opts)
	if err == nil {
		t.Error("NewSAMLSession() expected error for authentication failure, got nil")
	}
}

func TestNewSAMLSession_EmptyTokenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"CyberArkLogonResult": ""}`))
	}))
	defer server.Close()

	opts := SAMLSessionOptions{
		BaseURL:      server.URL,
		SAMLResponse: "base64-saml-response",
	}

	_, err := NewSAMLSession(context.Background(), opts)
	if err == nil {
		t.Error("NewSAMLSession() expected error for empty token, got nil")
	}
}

func TestSAMLSessionOptions_Struct(t *testing.T) {
	opts := SAMLSessionOptions{
		BaseURL:           "https://cyberark.example.com",
		SAMLResponse:      "base64-encoded-response",
		UseIntegratedAuth: false,
		IDPLoginURL:       "https://idp.example.com/saml",
	}

	if opts.BaseURL != "https://cyberark.example.com" {
		t.Errorf("BaseURL = %v, want https://cyberark.example.com", opts.BaseURL)
	}
}

func TestSAMLLoginRequest_Struct(t *testing.T) {
	req := SAMLLoginRequest{
		SAMLResponse:      "base64-saml",
		ConcurrentSession: true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal SAMLLoginRequest: %v", err)
	}

	var parsed SAMLLoginRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal SAMLLoginRequest: %v", err)
	}

	if parsed.SAMLResponse != req.SAMLResponse {
		t.Errorf("SAMLResponse = %v, want %v", parsed.SAMLResponse, req.SAMLResponse)
	}
}
