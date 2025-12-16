// Package authmethods provides tests for authentication method management functionality.
package authmethods

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

func TestListAuthenticationMethods(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []AuthenticationMethod
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful list",
			serverResponse: []AuthenticationMethod{
				{ID: "cyberark", DisplayName: "CyberArk", Enabled: true},
				{ID: "ldap", DisplayName: "LDAP", Enabled: true},
				{ID: "radius", DisplayName: "RADIUS", Enabled: false},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    3,
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
					response := AuthenticationMethodsResponse{Methods: tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListAuthenticationMethods(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("ListAuthenticationMethods() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListAuthenticationMethods() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListAuthenticationMethods() returned %d methods, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestGetAuthenticationMethod(t *testing.T) {
	tests := []struct {
		name           string
		methodID       string
		serverResponse *AuthenticationMethod
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful get",
			methodID: "cyberark",
			serverResponse: &AuthenticationMethod{
				ID:          "cyberark",
				DisplayName: "CyberArk",
				Enabled:     true,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty method ID",
			methodID: "",
			wantErr:  true,
		},
		{
			name:         "not found",
			methodID:     "nonexistent",
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
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetAuthenticationMethod(context.Background(), sess, tt.methodID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetAuthenticationMethod() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetAuthenticationMethod() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("GetAuthenticationMethod().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestAddAuthenticationMethod(t *testing.T) {
	tests := []struct {
		name           string
		opts           AddAuthenticationMethodOptions
		serverResponse *AuthenticationMethod
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful add",
			opts: AddAuthenticationMethodOptions{
				ID:          "custom",
				DisplayName: "Custom Auth",
				Enabled:     true,
			},
			serverResponse: &AuthenticationMethod{
				ID:          "custom",
				DisplayName: "Custom Auth",
				Enabled:     true,
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing ID",
			opts: AddAuthenticationMethodOptions{
				DisplayName: "Custom Auth",
			},
			wantErr: true,
		},
		{
			name: "missing display name",
			opts: AddAuthenticationMethodOptions{
				ID: "custom",
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

			result, err := AddAuthenticationMethod(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("AddAuthenticationMethod() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddAuthenticationMethod() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("AddAuthenticationMethod().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestUpdateAuthenticationMethod(t *testing.T) {
	tests := []struct {
		name           string
		methodID       string
		opts           UpdateAuthenticationMethodOptions
		serverResponse *AuthenticationMethod
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful update",
			methodID: "custom",
			opts: UpdateAuthenticationMethodOptions{
				DisplayName: "Updated Custom Auth",
			},
			serverResponse: &AuthenticationMethod{
				ID:          "custom",
				DisplayName: "Updated Custom Auth",
				Enabled:     true,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty method ID",
			methodID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := UpdateAuthenticationMethod(context.Background(), sess, tt.methodID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("UpdateAuthenticationMethod() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateAuthenticationMethod() unexpected error: %v", err)
				return
			}

			if result.DisplayName != tt.serverResponse.DisplayName {
				t.Errorf("UpdateAuthenticationMethod().DisplayName = %v, want %v", result.DisplayName, tt.serverResponse.DisplayName)
			}
		})
	}
}

func TestRemoveAuthenticationMethod(t *testing.T) {
	tests := []struct {
		name         string
		methodID     string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			methodID:     "custom",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:     "empty method ID",
			methodID: "",
			wantErr:  true,
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

			err := RemoveAuthenticationMethod(context.Background(), sess, tt.methodID)
			if tt.wantErr {
				if err == nil {
					t.Error("RemoveAuthenticationMethod() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RemoveAuthenticationMethod() unexpected error: %v", err)
			}
		})
	}
}

func TestListUserAllowedAuthMethods(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		serverResponse []UserAllowedAuthMethod
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name:   "successful list",
			userID: 123,
			serverResponse: []UserAllowedAuthMethod{
				{MethodID: "cyberark", IsEnabled: true},
				{MethodID: "ldap", IsEnabled: true},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					response := struct {
						AuthenticationMethods []UserAllowedAuthMethod `json:"AuthenticationMethods"`
					}{AuthenticationMethods: tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListUserAllowedAuthMethods(context.Background(), sess, tt.userID)
			if tt.wantErr {
				if err == nil {
					t.Error("ListUserAllowedAuthMethods() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListUserAllowedAuthMethods() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListUserAllowedAuthMethods() returned %d methods, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestAddUserAllowedAuthMethod(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		methodID     string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful add",
			userID:       123,
			methodID:     "radius",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty method ID",
			userID:   123,
			methodID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := AddUserAllowedAuthMethod(context.Background(), sess, tt.userID, tt.methodID)
			if tt.wantErr {
				if err == nil {
					t.Error("AddUserAllowedAuthMethod() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddUserAllowedAuthMethod() unexpected error: %v", err)
			}
		})
	}
}

func TestRemoveUserAllowedAuthMethod(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		methodID     string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			userID:       123,
			methodID:     "radius",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:     "empty method ID",
			userID:   123,
			methodID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := RemoveUserAllowedAuthMethod(context.Background(), sess, tt.userID, tt.methodID)
			if tt.wantErr {
				if err == nil {
					t.Error("RemoveUserAllowedAuthMethod() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("RemoveUserAllowedAuthMethod() unexpected error: %v", err)
			}
		})
	}
}

func TestListAllowedReferrers(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []AllowedReferrer
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful list",
			serverResponse: []AllowedReferrer{
				{ReferrerURL: "https://example.com", RegularExpression: false},
				{ReferrerURL: "https://*.internal.com", RegularExpression: true},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					response := struct {
						AllowedReferrers []AllowedReferrer `json:"AllowedReferrers"`
					}{AllowedReferrers: tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListAllowedReferrers(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("ListAllowedReferrers() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListAllowedReferrers() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListAllowedReferrers() returned %d referrers, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestAddAllowedReferrer(t *testing.T) {
	tests := []struct {
		name           string
		referrerURL    string
		isRegex        bool
		serverResponse *AllowedReferrer
		serverStatus   int
		wantErr        bool
	}{
		{
			name:        "successful add",
			referrerURL: "https://example.com",
			isRegex:     false,
			serverResponse: &AllowedReferrer{
				ReferrerURL:       "https://example.com",
				RegularExpression: false,
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name:        "empty URL",
			referrerURL: "",
			wantErr:     true,
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

			result, err := AddAllowedReferrer(context.Background(), sess, tt.referrerURL, tt.isRegex)
			if tt.wantErr {
				if err == nil {
					t.Error("AddAllowedReferrer() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddAllowedReferrer() unexpected error: %v", err)
				return
			}

			if result.ReferrerURL != tt.serverResponse.ReferrerURL {
				t.Errorf("AddAllowedReferrer().ReferrerURL = %v, want %v", result.ReferrerURL, tt.serverResponse.ReferrerURL)
			}
		})
	}
}

func TestListAuthenticationMethods_InvalidSession(t *testing.T) {
	_, err := ListAuthenticationMethods(context.Background(), nil)
	if err == nil {
		t.Error("ListAuthenticationMethods() with nil session expected error, got nil")
	}
}

func TestGetAuthenticationMethod_InvalidSession(t *testing.T) {
	_, err := GetAuthenticationMethod(context.Background(), nil, "cyberark")
	if err == nil {
		t.Error("GetAuthenticationMethod() expected error for nil session")
	}
}

func TestGetAuthenticationMethod_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := GetAuthenticationMethod(context.Background(), sess, "cyberark")
	if err == nil {
		t.Error("GetAuthenticationMethod() expected error for server error")
	}
}

func TestAddAuthenticationMethod_InvalidSession(t *testing.T) {
	_, err := AddAuthenticationMethod(context.Background(), nil, AddAuthenticationMethodOptions{ID: "test", DisplayName: "Test"})
	if err == nil {
		t.Error("AddAuthenticationMethod() expected error for nil session")
	}
}

func TestAddAuthenticationMethod_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := AddAuthenticationMethod(context.Background(), sess, AddAuthenticationMethodOptions{ID: "test", DisplayName: "Test"})
	if err == nil {
		t.Error("AddAuthenticationMethod() expected error for server error")
	}
}

func TestUpdateAuthenticationMethod_InvalidSession(t *testing.T) {
	_, err := UpdateAuthenticationMethod(context.Background(), nil, "cyberark", UpdateAuthenticationMethodOptions{})
	if err == nil {
		t.Error("UpdateAuthenticationMethod() expected error for nil session")
	}
}

func TestUpdateAuthenticationMethod_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := UpdateAuthenticationMethod(context.Background(), sess, "cyberark", UpdateAuthenticationMethodOptions{})
	if err == nil {
		t.Error("UpdateAuthenticationMethod() expected error for server error")
	}
}

func TestRemoveAuthenticationMethod_InvalidSession(t *testing.T) {
	err := RemoveAuthenticationMethod(context.Background(), nil, "cyberark")
	if err == nil {
		t.Error("RemoveAuthenticationMethod() expected error for nil session")
	}
}

func TestRemoveAuthenticationMethod_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	err := RemoveAuthenticationMethod(context.Background(), sess, "cyberark")
	if err == nil {
		t.Error("RemoveAuthenticationMethod() expected error for server error")
	}
}

func TestListUserAllowedAuthMethods_InvalidSession(t *testing.T) {
	_, err := ListUserAllowedAuthMethods(context.Background(), nil, 123)
	if err == nil {
		t.Error("ListUserAllowedAuthMethods() expected error for nil session")
	}
}

func TestListUserAllowedAuthMethods_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := ListUserAllowedAuthMethods(context.Background(), sess, 123)
	if err == nil {
		t.Error("ListUserAllowedAuthMethods() expected error for server error")
	}
}

func TestAddUserAllowedAuthMethod_InvalidSession(t *testing.T) {
	err := AddUserAllowedAuthMethod(context.Background(), nil, 123, "cyberark")
	if err == nil {
		t.Error("AddUserAllowedAuthMethod() expected error for nil session")
	}
}

func TestAddUserAllowedAuthMethod_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	err := AddUserAllowedAuthMethod(context.Background(), sess, 123, "cyberark")
	if err == nil {
		t.Error("AddUserAllowedAuthMethod() expected error for server error")
	}
}

func TestRemoveUserAllowedAuthMethod_InvalidSession(t *testing.T) {
	err := RemoveUserAllowedAuthMethod(context.Background(), nil, 123, "cyberark")
	if err == nil {
		t.Error("RemoveUserAllowedAuthMethod() expected error for nil session")
	}
}

func TestRemoveUserAllowedAuthMethod_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	err := RemoveUserAllowedAuthMethod(context.Background(), sess, 123, "cyberark")
	if err == nil {
		t.Error("RemoveUserAllowedAuthMethod() expected error for server error")
	}
}

func TestListAllowedReferrers_InvalidSession(t *testing.T) {
	_, err := ListAllowedReferrers(context.Background(), nil)
	if err == nil {
		t.Error("ListAllowedReferrers() expected error for nil session")
	}
}

func TestListAllowedReferrers_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := ListAllowedReferrers(context.Background(), sess)
	if err == nil {
		t.Error("ListAllowedReferrers() expected error for server error")
	}
}

func TestAddAllowedReferrer_InvalidSession(t *testing.T) {
	_, err := AddAllowedReferrer(context.Background(), nil, "https://example.com", false)
	if err == nil {
		t.Error("AddAllowedReferrer() expected error for nil session")
	}
}

func TestAddAllowedReferrer_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := AddAllowedReferrer(context.Background(), sess, "https://example.com", false)
	if err == nil {
		t.Error("AddAllowedReferrer() expected error for server error")
	}
}

func TestAuthenticationMethod_Struct(t *testing.T) {
	method := AuthenticationMethod{
		ID:          "cyberark",
		DisplayName: "CyberArk Authentication",
		Enabled:     true,
		SignInLabel: "Sign In",
	}

	if method.ID != "cyberark" {
		t.Errorf("ID = %v, want cyberark", method.ID)
	}
	if !method.Enabled {
		t.Error("Enabled should be true")
	}
}
