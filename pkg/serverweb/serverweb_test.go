// Package serverweb provides tests for server and web service functionality.
package serverweb

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

func TestGetServer(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse *ServerInfo
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful get server",
			serverResponse: &ServerInfo{
				ServerID:        "server-123",
				ServerName:      "CyberArkPAS",
				InternalVersion: 14.0,
				ExternalVersion: "14.0.0",
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

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetServer(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetServer() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetServer() unexpected error: %v", err)
				return
			}

			if result.ServerID != tt.serverResponse.ServerID {
				t.Errorf("GetServer().ServerID = %v, want %v", result.ServerID, tt.serverResponse.ServerID)
			}
		})
	}
}

func TestGetServer_InvalidSession(t *testing.T) {
	_, err := GetServer(context.Background(), nil)
	if err == nil {
		t.Error("GetServer() expected error for nil session, got nil")
	}
}

func TestGetWebServiceStatus(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse *WebServiceStatus
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful get status",
			serverResponse: &WebServiceStatus{
				IsWebServiceEnabled: true,
				WebServiceID:        "ws-123",
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

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetWebServiceStatus(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetWebServiceStatus() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetWebServiceStatus() unexpected error: %v", err)
				return
			}

			if result.IsWebServiceEnabled != tt.serverResponse.IsWebServiceEnabled {
				t.Errorf("GetWebServiceStatus().IsWebServiceEnabled = %v, want %v",
					result.IsWebServiceEnabled, tt.serverResponse.IsWebServiceEnabled)
			}
		})
	}
}

func TestGetWebServiceStatus_InvalidSession(t *testing.T) {
	_, err := GetWebServiceStatus(context.Background(), nil)
	if err == nil {
		t.Error("GetWebServiceStatus() expected error for nil session, got nil")
	}
}

func TestGetWebServiceStatus_InvalidJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	result, err := GetWebServiceStatus(context.Background(), sess)
	if err != nil {
		t.Errorf("GetWebServiceStatus() unexpected error: %v", err)
		return
	}

	// Should return default enabled status when JSON parsing fails
	if !result.IsWebServiceEnabled {
		t.Error("GetWebServiceStatus() should return enabled status when JSON parsing fails")
	}
}

func TestVerifyAPI(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful verify",
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
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := VerifyAPI(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("VerifyAPI() expected error, got nil")
				}
				// Even on error, result should contain status info
				if result == nil {
					t.Error("VerifyAPI() should return status even on error")
				}
				return
			}
			if err != nil {
				t.Errorf("VerifyAPI() unexpected error: %v", err)
				return
			}

			if result.StatusCode != tt.serverStatus {
				t.Errorf("VerifyAPI().StatusCode = %v, want %v", result.StatusCode, tt.serverStatus)
			}
		})
	}
}

func TestVerifyAPI_NilSession(t *testing.T) {
	_, err := VerifyAPI(context.Background(), nil)
	if err == nil {
		t.Error("VerifyAPI() expected error for nil session, got nil")
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

func TestWebServiceStatus_Struct(t *testing.T) {
	status := WebServiceStatus{
		IsWebServiceEnabled: true,
		WebServiceID:        "ws-123",
	}

	if !status.IsWebServiceEnabled {
		t.Error("IsWebServiceEnabled should be true")
	}
	if status.WebServiceID != "ws-123" {
		t.Errorf("WebServiceID = %v, want ws-123", status.WebServiceID)
	}
}

func TestAPIStatus_Struct(t *testing.T) {
	status := APIStatus{
		StatusCode: 200,
		Message:    "OK",
	}

	if status.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", status.StatusCode)
	}
	if status.Message != "OK" {
		t.Errorf("Message = %v, want OK", status.Message)
	}
}
