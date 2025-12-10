// Package ccp provides tests for the CyberArk Central Credential Provider functionality.
package ccp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chrisranney/gopas/pkg/types"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ClientConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			cfg: ClientConfig{
				BaseURL: "https://cyberark.example.com",
			},
			wantErr: false,
		},
		{
			name: "valid with all options",
			cfg: ClientConfig{
				BaseURL:       "https://cyberark.example.com",
				SkipTLSVerify: true,
				Timeout:       60 * time.Second,
			},
			wantErr: false,
		},
		{
			name:    "missing base URL",
			cfg:     ClientConfig{},
			wantErr: true,
			errMsg:  "baseURL is required",
		},
		{
			name: "invalid client cert path",
			cfg: ClientConfig{
				BaseURL:    "https://cyberark.example.com",
				ClientCert: "/nonexistent/cert.pem",
				ClientKey:  "/nonexistent/key.pem",
			},
			wantErr: true,
			errMsg:  "failed to load client certificate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Error("NewClient() expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewClient() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestNewClient_WithValidCert(t *testing.T) {
	// Create temporary certificate files for testing
	certPath, keyPath, cleanup := createTestCertificates(t)
	defer cleanup()

	cfg := ClientConfig{
		BaseURL:    "https://cyberark.example.com",
		ClientCert: certPath,
		ClientKey:  keyPath,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient() with valid certs failed: %v", err)
	}
	if client == nil {
		t.Error("NewClient() returned nil client")
	}
}

func TestNewClient_DefaultTimeout(t *testing.T) {
	cfg := ClientConfig{
		BaseURL: "https://cyberark.example.com",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	// Verify the client was created (we can't directly check the timeout,
	// but we verify the client is functional)
	if client == nil {
		t.Error("NewClient() returned nil client")
	}
	if client.baseURL != cfg.BaseURL {
		t.Errorf("client.baseURL = %v, want %v", client.baseURL, cfg.BaseURL)
	}
}

func TestGetCredential(t *testing.T) {
	tests := []struct {
		name           string
		req            CredentialRequest
		serverResponse interface{}
		serverStatus   int
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful credential retrieval",
			req: CredentialRequest{
				AppID: "TestApp",
				Safe:  "TestSafe",
			},
			serverResponse: CredentialResponse{
				Content:    "secretpassword123",
				UserName:   "admin",
				Address:    "server.example.com",
				Safe:       "TestSafe",
				Folder:     "Root",
				Name:       "AdminAccount",
				PolicyID:   "UnixSSH",
				DeviceType: "Operating System",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "with all optional parameters",
			req: CredentialRequest{
				AppID:             "TestApp",
				Safe:              "TestSafe",
				Object:            "MyAccount",
				Folder:            "Root\\SubFolder",
				UserName:          "admin",
				Address:           "server.example.com",
				Query:             "admin",
				QueryFormat:       "Exact",
				Reason:            "Automated login",
				ConnectionTimeout: 30,
			},
			serverResponse: CredentialResponse{
				Content:  "secretpassword123",
				UserName: "admin",
				Safe:     "TestSafe",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "missing AppID",
			req: CredentialRequest{
				Safe: "TestSafe",
			},
			wantErr: true,
			errMsg:  "AppID is required",
		},
		{
			name: "missing Safe",
			req: CredentialRequest{
				AppID: "TestApp",
			},
			wantErr: true,
			errMsg:  "Safe is required",
		},
		{
			name: "credential not found",
			req: CredentialRequest{
				AppID: "TestApp",
				Safe:  "TestSafe",
			},
			serverResponse: ErrorResponse{
				ErrorCode: "APPAP004E",
				ErrorMsg:  "No accounts were found matching the query",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
			errMsg:       "No accounts were found",
		},
		{
			name: "unauthorized AppID",
			req: CredentialRequest{
				AppID: "UnauthorizedApp",
				Safe:  "TestSafe",
			},
			serverResponse: ErrorResponse{
				ErrorCode: "APPAP007E",
				ErrorMsg:  "The Application ID is not authorized to retrieve the requested credential",
			},
			serverStatus: http.StatusForbidden,
			wantErr:      true,
			errMsg:       "not authorized",
		},
		{
			name: "server error",
			req: CredentialRequest{
				AppID: "TestApp",
				Safe:  "TestSafe",
			},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
			errMsg:       "status 500",
		},
		{
			name: "invalid JSON response",
			req: CredentialRequest{
				AppID: "TestApp",
				Safe:  "TestSafe",
			},
			serverResponse: "not valid json",
			serverStatus:   http.StatusOK,
			wantErr:        true,
			errMsg:         "failed to parse response",
		},
		{
			name: "password change in process",
			req: CredentialRequest{
				AppID: "TestApp",
				Safe:  "TestSafe",
			},
			serverResponse: CredentialResponse{
				Content:                 "oldpassword",
				UserName:                "admin",
				PasswordChangeInProcess: types.FlexibleBool(true),
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.serverStatus != 0 {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify request method
					if r.Method != http.MethodGet {
						t.Errorf("Expected GET request, got %s", r.Method)
					}

					// Verify endpoint
					if !strings.Contains(r.URL.Path, "/AIMWebService/api/Accounts") {
						t.Errorf("Unexpected path: %s", r.URL.Path)
					}

					// Verify required query parameters
					query := r.URL.Query()
					if tt.req.AppID != "" && query.Get("AppID") != tt.req.AppID {
						t.Errorf("AppID parameter = %v, want %v", query.Get("AppID"), tt.req.AppID)
					}
					if tt.req.Safe != "" && query.Get("Safe") != tt.req.Safe {
						t.Errorf("Safe parameter = %v, want %v", query.Get("Safe"), tt.req.Safe)
					}

					// Verify optional parameters are passed correctly
					if tt.req.Object != "" && query.Get("Object") != tt.req.Object {
						t.Errorf("Object parameter = %v, want %v", query.Get("Object"), tt.req.Object)
					}
					if tt.req.Folder != "" && query.Get("Folder") != tt.req.Folder {
						t.Errorf("Folder parameter = %v, want %v", query.Get("Folder"), tt.req.Folder)
					}
					if tt.req.UserName != "" && query.Get("UserName") != tt.req.UserName {
						t.Errorf("UserName parameter = %v, want %v", query.Get("UserName"), tt.req.UserName)
					}
					if tt.req.Address != "" && query.Get("Address") != tt.req.Address {
						t.Errorf("Address parameter = %v, want %v", query.Get("Address"), tt.req.Address)
					}
					if tt.req.Query != "" && query.Get("Query") != tt.req.Query {
						t.Errorf("Query parameter = %v, want %v", query.Get("Query"), tt.req.Query)
					}
					if tt.req.QueryFormat != "" && query.Get("QueryFormat") != tt.req.QueryFormat {
						t.Errorf("QueryFormat parameter = %v, want %v", query.Get("QueryFormat"), tt.req.QueryFormat)
					}
					if tt.req.Reason != "" && query.Get("Reason") != tt.req.Reason {
						t.Errorf("Reason parameter = %v, want %v", query.Get("Reason"), tt.req.Reason)
					}
					if tt.req.ConnectionTimeout > 0 {
						expected := fmt.Sprintf("%d", tt.req.ConnectionTimeout)
						if query.Get("ConnectionTimeout") != expected {
							t.Errorf("ConnectionTimeout parameter = %v, want %v", query.Get("ConnectionTimeout"), expected)
						}
					}

					// Verify Accept header
					if r.Header.Get("Accept") != "application/json" {
						t.Errorf("Accept header = %v, want application/json", r.Header.Get("Accept"))
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.serverStatus)

					if tt.serverResponse != nil {
						switch resp := tt.serverResponse.(type) {
						case string:
							w.Write([]byte(resp))
						default:
							json.NewEncoder(w).Encode(resp)
						}
					}
				})
				server = httptest.NewServer(handler)
				defer server.Close()
			}

			var client *Client
			var err error
			if server != nil {
				client, err = NewClient(ClientConfig{BaseURL: server.URL})
			} else {
				client, err = NewClient(ClientConfig{BaseURL: "https://example.com"})
			}
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			result, err := client.GetCredential(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("GetCredential() expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetCredential() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetCredential() unexpected error: %v", err)
				return
			}

			expected := tt.serverResponse.(CredentialResponse)
			if result.Content != expected.Content {
				t.Errorf("GetCredential().Content = %v, want %v", result.Content, expected.Content)
			}
			if result.UserName != expected.UserName {
				t.Errorf("GetCredential().UserName = %v, want %v", result.UserName, expected.UserName)
			}
			if result.Safe != expected.Safe {
				t.Errorf("GetCredential().Safe = %v, want %v", result.Safe, expected.Safe)
			}
			if result.PasswordChangeInProcess != expected.PasswordChangeInProcess {
				t.Errorf("GetCredential().PasswordChangeInProcess = %v, want %v", result.PasswordChangeInProcess, expected.PasswordChangeInProcess)
			}
		})
	}
}

func TestGetCredential_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CredentialResponse{Content: "test"})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = client.GetCredential(ctx, CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})

	if err == nil {
		t.Error("GetCredential() expected context cancellation error, got nil")
	}
}

func TestGetPassword(t *testing.T) {
	expectedPassword := "mysecretpassword123"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CredentialResponse{
			Content:  expectedPassword,
			UserName: "testuser",
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	password, err := client.GetPassword(context.Background(), CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})
	if err != nil {
		t.Errorf("GetPassword() unexpected error: %v", err)
	}
	if password != expectedPassword {
		t.Errorf("GetPassword() = %v, want %v", password, expectedPassword)
	}
}

func TestGetPassword_Error(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			ErrorCode: "APPAP004E",
			ErrorMsg:  "No accounts were found",
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetPassword(context.Background(), CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})
	if err == nil {
		t.Error("GetPassword() expected error, got nil")
	}
}

func TestGetLoginCredentials(t *testing.T) {
	tests := []struct {
		name             string
		serverResponse   CredentialResponse
		serverStatus     int
		expectedUsername string
		expectedPassword string
		wantErr          bool
	}{
		{
			name: "successful login credentials retrieval",
			serverResponse: CredentialResponse{
				Content:  "adminpassword",
				UserName: "administrator",
				Address:  "cyberark.example.com",
			},
			serverStatus:     http.StatusOK,
			expectedUsername: "administrator",
			expectedPassword: "adminpassword",
			wantErr:          false,
		},
		{
			name: "empty username returns empty",
			serverResponse: CredentialResponse{
				Content:  "password",
				UserName: "",
			},
			serverStatus:     http.StatusOK,
			expectedUsername: "",
			expectedPassword: "password",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			})
			server := httptest.NewServer(handler)
			defer server.Close()

			client, err := NewClient(ClientConfig{BaseURL: server.URL})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			username, password, err := client.GetLoginCredentials(context.Background(), CredentialRequest{
				AppID: "TestApp",
				Safe:  "TestSafe",
			})

			if tt.wantErr {
				if err == nil {
					t.Error("GetLoginCredentials() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetLoginCredentials() unexpected error: %v", err)
				return
			}
			if username != tt.expectedUsername {
				t.Errorf("GetLoginCredentials() username = %v, want %v", username, tt.expectedUsername)
			}
			if password != tt.expectedPassword {
				t.Errorf("GetLoginCredentials() password = %v, want %v", password, tt.expectedPassword)
			}
		})
	}
}

func TestGetLoginCredentials_Error(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			ErrorCode: "APPAP007E",
			ErrorMsg:  "Application not authorized",
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	username, password, err := client.GetLoginCredentials(context.Background(), CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})

	if err == nil {
		t.Error("GetLoginCredentials() expected error, got nil")
	}
	if username != "" {
		t.Errorf("GetLoginCredentials() username should be empty on error, got %v", username)
	}
	if password != "" {
		t.Errorf("GetLoginCredentials() password should be empty on error, got %v", password)
	}
}

func TestCredentialRequest_Struct(t *testing.T) {
	req := CredentialRequest{
		AppID:             "MyApp",
		Safe:              "MySafe",
		Object:            "MyAccount",
		Folder:            "Root\\Folder",
		UserName:          "testuser",
		Address:           "192.168.1.100",
		Query:             "test",
		QueryFormat:       "Exact",
		Reason:            "Automated access",
		ConnectionTimeout: 60,
	}

	if req.AppID != "MyApp" {
		t.Errorf("AppID = %v, want MyApp", req.AppID)
	}
	if req.Safe != "MySafe" {
		t.Errorf("Safe = %v, want MySafe", req.Safe)
	}
	if req.Object != "MyAccount" {
		t.Errorf("Object = %v, want MyAccount", req.Object)
	}
	if req.Folder != "Root\\Folder" {
		t.Errorf("Folder = %v, want Root\\Folder", req.Folder)
	}
	if req.ConnectionTimeout != 60 {
		t.Errorf("ConnectionTimeout = %v, want 60", req.ConnectionTimeout)
	}
}

func TestCredentialResponse_Struct(t *testing.T) {
	resp := CredentialResponse{
		Content:                 "password123",
		UserName:                "admin",
		Address:                 "server.local",
		Safe:                    "TestSafe",
		Folder:                  "Root",
		Name:                    "AdminAccount",
		PolicyID:                "UnixSSH",
		DeviceType:              "Operating System",
		PasswordChangeInProcess: types.FlexibleBool(true),
		CreationMethod:          "AutoDetected",
		Properties: map[string]string{
			"LogonDomain": "CORP",
			"Port":        "22",
		},
	}

	if resp.Content != "password123" {
		t.Errorf("Content = %v, want password123", resp.Content)
	}
	if resp.UserName != "admin" {
		t.Errorf("UserName = %v, want admin", resp.UserName)
	}
	if !resp.PasswordChangeInProcess {
		t.Error("PasswordChangeInProcess should be true")
	}
	if resp.Properties["LogonDomain"] != "CORP" {
		t.Errorf("Properties[LogonDomain] = %v, want CORP", resp.Properties["LogonDomain"])
	}
}

func TestErrorResponse_Struct(t *testing.T) {
	errResp := ErrorResponse{
		ErrorCode: "APPAP004E",
		ErrorMsg:  "No accounts were found matching the query",
	}

	if errResp.ErrorCode != "APPAP004E" {
		t.Errorf("ErrorCode = %v, want APPAP004E", errResp.ErrorCode)
	}
	if errResp.ErrorMsg != "No accounts were found matching the query" {
		t.Errorf("ErrorMsg = %v, want 'No accounts were found matching the query'", errResp.ErrorMsg)
	}
}

func TestClientConfig_Struct(t *testing.T) {
	cfg := ClientConfig{
		BaseURL:       "https://cyberark.example.com",
		SkipTLSVerify: true,
		Timeout:       45 * time.Second,
		ClientCert:    "/path/to/cert.pem",
		ClientKey:     "/path/to/key.pem",
	}

	if cfg.BaseURL != "https://cyberark.example.com" {
		t.Errorf("BaseURL = %v, want https://cyberark.example.com", cfg.BaseURL)
	}
	if !cfg.SkipTLSVerify {
		t.Error("SkipTLSVerify should be true")
	}
	if cfg.Timeout != 45*time.Second {
		t.Errorf("Timeout = %v, want 45s", cfg.Timeout)
	}
}

func TestGetCredential_URLConstruction(t *testing.T) {
	var capturedURL string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CredentialResponse{
			Content:  "password",
			UserName: "user",
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetCredential(context.Background(), CredentialRequest{
		AppID:    "TestApp",
		Safe:     "TestSafe",
		Object:   "TestObject",
		UserName: "testuser",
	})
	if err != nil {
		t.Fatalf("GetCredential() failed: %v", err)
	}

	// Verify URL contains expected path
	if !strings.Contains(capturedURL, "/AIMWebService/api/Accounts") {
		t.Errorf("URL should contain CCP endpoint path, got %v", capturedURL)
	}

	// Verify query parameters are present
	if !strings.Contains(capturedURL, "AppID=TestApp") {
		t.Errorf("URL should contain AppID parameter, got %v", capturedURL)
	}
	if !strings.Contains(capturedURL, "Safe=TestSafe") {
		t.Errorf("URL should contain Safe parameter, got %v", capturedURL)
	}
	if !strings.Contains(capturedURL, "Object=TestObject") {
		t.Errorf("URL should contain Object parameter, got %v", capturedURL)
	}
	if !strings.Contains(capturedURL, "UserName=testuser") {
		t.Errorf("URL should contain UserName parameter, got %v", capturedURL)
	}
}

func TestGetCredential_SpecialCharactersInParameters(t *testing.T) {
	var capturedQuery string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CredentialResponse{
			Content:  "password",
			UserName: "user",
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetCredential(context.Background(), CredentialRequest{
		AppID:  "Test App",     // Space in value
		Safe:   "Test/Safe",    // Slash in value
		Folder: "Root\\Folder", // Backslash in value
	})
	if err != nil {
		t.Fatalf("GetCredential() failed: %v", err)
	}

	// Verify parameters are URL-encoded
	if !strings.Contains(capturedQuery, "AppID=Test+App") && !strings.Contains(capturedQuery, "AppID=Test%20App") {
		t.Errorf("AppID should be URL-encoded, got query: %v", capturedQuery)
	}
}

func TestClient_TLSConfiguration(t *testing.T) {
	// Create a test HTTPS server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CredentialResponse{
			Content:  "password",
			UserName: "user",
		})
	})
	server := httptest.NewTLSServer(handler)
	defer server.Close()

	// Test with SkipTLSVerify = true (should succeed with self-signed cert)
	client, err := NewClient(ClientConfig{
		BaseURL:       server.URL,
		SkipTLSVerify: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetCredential(context.Background(), CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})
	if err != nil {
		t.Errorf("GetCredential() with SkipTLSVerify=true failed: %v", err)
	}

	// Test with SkipTLSVerify = false (should fail with self-signed cert)
	clientStrict, err := NewClient(ClientConfig{
		BaseURL:       server.URL,
		SkipTLSVerify: false,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = clientStrict.GetCredential(context.Background(), CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})
	if err == nil {
		t.Error("GetCredential() with SkipTLSVerify=false should fail with self-signed cert")
	}
}

func TestGetCredential_EmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return empty JSON object
		w.Write([]byte("{}"))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	result, err := client.GetCredential(context.Background(), CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})
	if err != nil {
		t.Errorf("GetCredential() unexpected error: %v", err)
		return
	}

	// Empty response should still parse without error
	if result.Content != "" {
		t.Errorf("Expected empty Content, got %v", result.Content)
	}
}

func TestGetCredential_MalformedErrorResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetCredential(context.Background(), CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})
	if err == nil {
		t.Error("GetCredential() expected error for bad request")
	}
	if !strings.Contains(err.Error(), "status 400") {
		t.Errorf("Error should contain status code, got: %v", err)
	}
}

// TestGetCredential_PasswordChangeInProcessAsString tests that PasswordChangeInProcess
// can be parsed when the API returns it as a string ("true"/"false") instead of a boolean.
// This is a known behavior of some CyberArk CCP API versions.
func TestGetCredential_PasswordChangeInProcessAsString(t *testing.T) {
	tests := []struct {
		name         string
		jsonResponse string
		wantValue    bool
	}{
		{
			name:         "string true",
			jsonResponse: `{"Content": "password", "UserName": "admin", "PasswordChangeInProcess": "true"}`,
			wantValue:    true,
		},
		{
			name:         "string false",
			jsonResponse: `{"Content": "password", "UserName": "admin", "PasswordChangeInProcess": "false"}`,
			wantValue:    false,
		},
		{
			name:         "boolean true",
			jsonResponse: `{"Content": "password", "UserName": "admin", "PasswordChangeInProcess": true}`,
			wantValue:    true,
		},
		{
			name:         "boolean false",
			jsonResponse: `{"Content": "password", "UserName": "admin", "PasswordChangeInProcess": false}`,
			wantValue:    false,
		},
		{
			name:         "string True titlecase",
			jsonResponse: `{"Content": "password", "UserName": "admin", "PasswordChangeInProcess": "True"}`,
			wantValue:    true,
		},
		{
			name:         "string False titlecase",
			jsonResponse: `{"Content": "password", "UserName": "admin", "PasswordChangeInProcess": "False"}`,
			wantValue:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.jsonResponse))
			})
			server := httptest.NewServer(handler)
			defer server.Close()

			client, err := NewClient(ClientConfig{BaseURL: server.URL})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			result, err := client.GetCredential(context.Background(), CredentialRequest{
				AppID: "TestApp",
				Safe:  "TestSafe",
			})
			if err != nil {
				t.Fatalf("GetCredential() unexpected error: %v", err)
			}

			if result.PasswordChangeInProcess.Bool() != tt.wantValue {
				t.Errorf("PasswordChangeInProcess = %v, want %v", result.PasswordChangeInProcess, tt.wantValue)
			}
		})
	}
}

// createTestCertificates creates temporary test certificate files and returns paths
func createTestCertificates(t *testing.T) (certPath, keyPath string, cleanup func()) {
	t.Helper()

	// Generate a self-signed certificate
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Write certificate
	certPath = filepath.Join(tmpDir, "cert.pem")
	certFile, err := os.Create(certPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create cert file: %v", err)
	}
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certFile.Close()

	// Write private key
	keyPath = filepath.Join(tmpDir, "key.pem")
	keyFile, err := os.Create(keyPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create key file: %v", err)
	}
	pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyFile.Close()

	cleanup = func() {
		os.RemoveAll(tmpDir)
	}

	return certPath, keyPath, cleanup
}

// Benchmark tests
func BenchmarkGetCredential(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CredentialResponse{
			Content:  "password123",
			UserName: "admin",
		})
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client, _ := NewClient(ClientConfig{BaseURL: server.URL})
	ctx := context.Background()
	req := CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.GetCredential(ctx, req)
	}
}

func BenchmarkNewClient(b *testing.B) {
	cfg := ClientConfig{
		BaseURL: "https://cyberark.example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewClient(cfg)
	}
}

// Test helper to verify client state is properly isolated
func TestClient_Isolation(t *testing.T) {
	client1, err := NewClient(ClientConfig{
		BaseURL: "https://server1.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create client1: %v", err)
	}

	client2, err := NewClient(ClientConfig{
		BaseURL: "https://server2.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create client2: %v", err)
	}

	// Verify clients are independent
	if client1.baseURL == client2.baseURL {
		t.Error("Clients should have different base URLs")
	}
	if client1.httpClient == client2.httpClient {
		t.Error("Clients should have different HTTP clients")
	}
}

// Test that client can be used with custom TLS config
func TestNewClient_MutualTLS(t *testing.T) {
	// Create server that requires client cert
	certPath, keyPath, cleanup := createTestCertificates(t)
	defer cleanup()

	// Create a TLS server that accepts any client cert
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CredentialResponse{
			Content:  "password",
			UserName: "user",
		})
	})

	server := httptest.NewUnstartedServer(handler)
	server.TLS = &tls.Config{
		ClientAuth: tls.RequestClientCert, // Don't require, just accept
	}
	server.StartTLS()
	defer server.Close()

	// Create client with mutual TLS
	client, err := NewClient(ClientConfig{
		BaseURL:       server.URL,
		SkipTLSVerify: true, // Skip verify since server has self-signed cert
		ClientCert:    certPath,
		ClientKey:     keyPath,
	})
	if err != nil {
		t.Fatalf("Failed to create client with mTLS: %v", err)
	}

	_, err = client.GetCredential(context.Background(), CredentialRequest{
		AppID: "TestApp",
		Safe:  "TestSafe",
	})
	if err != nil {
		t.Errorf("GetCredential() with mTLS failed: %v", err)
	}
}
