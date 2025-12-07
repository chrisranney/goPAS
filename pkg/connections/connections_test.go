// Package connections provides tests for PSM connection functionality.
package connections

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

func TestConnect(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		req            ConnectionRequest
		serverResponse *ConnectionResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful connect",
			accountID: "123",
			req: ConnectionRequest{
				Reason:              "Maintenance",
				ConnectionComponent: "PSM-RDP",
			},
			serverResponse: &ConnectionResponse{
				PSMConnectURL: "https://psm.example.com/connect/abc123",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			req:       ConnectionRequest{},
			wantErr:   true,
		},
		{
			name:         "server error",
			accountID:    "123",
			req:          ConnectionRequest{},
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

			result, err := Connect(context.Background(), sess, tt.accountID, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("Connect() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Connect() unexpected error: %v", err)
				return
			}

			if result.PSMConnectURL != tt.serverResponse.PSMConnectURL {
				t.Errorf("Connect().PSMConnectURL = %v, want %v", result.PSMConnectURL, tt.serverResponse.PSMConnectURL)
			}
		})
	}
}

func TestAdHocConnect(t *testing.T) {
	tests := []struct {
		name           string
		req            AdHocConnectRequest
		serverResponse *ConnectionResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful ad-hoc connect",
			req: AdHocConnectRequest{
				UserName:   "admin",
				Secret:     "password",
				Address:    "server.example.com",
				PlatformID: "WinServerLocal",
			},
			serverResponse: &ConnectionResponse{
				PSMConnectURL: "https://psm.example.com/adhoc/xyz789",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "missing username",
			req: AdHocConnectRequest{
				Secret:     "password",
				Address:    "server.example.com",
				PlatformID: "WinServerLocal",
			},
			wantErr: true,
		},
		{
			name: "missing secret",
			req: AdHocConnectRequest{
				UserName:   "admin",
				Address:    "server.example.com",
				PlatformID: "WinServerLocal",
			},
			wantErr: true,
		},
		{
			name: "missing address",
			req: AdHocConnectRequest{
				UserName:   "admin",
				Secret:     "password",
				PlatformID: "WinServerLocal",
			},
			wantErr: true,
		},
		{
			name: "missing platform ID",
			req: AdHocConnectRequest{
				UserName: "admin",
				Secret:   "password",
				Address:  "server.example.com",
			},
			wantErr: true,
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

			result, err := AdHocConnect(context.Background(), sess, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("AdHocConnect() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AdHocConnect() unexpected error: %v", err)
				return
			}

			if result.PSMConnectURL != tt.serverResponse.PSMConnectURL {
				t.Errorf("AdHocConnect().PSMConnectURL = %v, want %v", result.PSMConnectURL, tt.serverResponse.PSMConnectURL)
			}
		})
	}
}

func TestGetConnectionComponents(t *testing.T) {
	tests := []struct {
		name           string
		platformID     string
		serverResponse []ConnectionComponent
		serverStatus   int
		wantErr        bool
	}{
		{
			name:       "successful get",
			platformID: "WinServerLocal",
			serverResponse: []ConnectionComponent{
				{PSMConnectorID: "PSM-RDP", PSMServerID: "PSMServer1"},
				{PSMConnectorID: "PSM-SSH", PSMServerID: "PSMServer1"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "empty platform ID",
			platformID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				response := struct {
					PSMConnectors []ConnectionComponent `json:"PSMConnectors"`
				}{PSMConnectors: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetConnectionComponents(context.Background(), sess, tt.platformID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetConnectionComponents() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetConnectionComponents() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetConnectionComponents() returned %d components, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGetPSMServers(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []PSMServer
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful get",
			serverResponse: []PSMServer{
				{ID: "1", Name: "PSMServer1", Address: "psm1.example.com"},
				{ID: "2", Name: "PSMServer2", Address: "psm2.example.com"},
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
					PSMServers []PSMServer `json:"PSMServers"`
				}{PSMServers: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetPSMServers(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetPSMServers() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPSMServers() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("GetPSMServers() returned %d servers, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestConnectionRequest_Struct(t *testing.T) {
	req := ConnectionRequest{
		Reason:              "Maintenance",
		TicketingSystemName: "ServiceNow",
		TicketID:            "INC12345",
		ConnectionComponent: "PSM-RDP",
		ConnectionParams: map[string]string{
			"AllowMappingLocalDrives": "Yes",
		},
	}

	if req.Reason != "Maintenance" {
		t.Errorf("Reason = %v, want Maintenance", req.Reason)
	}
	if req.ConnectionParams["AllowMappingLocalDrives"] != "Yes" {
		t.Errorf("ConnectionParams[AllowMappingLocalDrives] = %v, want Yes", req.ConnectionParams["AllowMappingLocalDrives"])
	}
}

func TestConnectionResponse_Struct(t *testing.T) {
	resp := ConnectionResponse{
		PSMConnectURL: "https://psm.example.com/connect/abc",
		RDPFile:       "full address:s:server.example.com",
	}

	if resp.PSMConnectURL != "https://psm.example.com/connect/abc" {
		t.Errorf("PSMConnectURL = %v, want https://psm.example.com/connect/abc", resp.PSMConnectURL)
	}
}

func TestPSMPrerequisites_Struct(t *testing.T) {
	prereq := PSMPrerequisites{
		ConnectionComponent: "PSM-RDP",
		ConnectionType:      "RDP",
	}

	if prereq.ConnectionComponent != "PSM-RDP" {
		t.Errorf("ConnectionComponent = %v, want PSM-RDP", prereq.ConnectionComponent)
	}
}

func TestPSMServer_Struct(t *testing.T) {
	server := PSMServer{
		ID:         "1",
		Name:       "PSMServer1",
		Address:    "psm.example.com",
		PSMVersion: "12.6",
	}

	if server.Name != "PSMServer1" {
		t.Errorf("Name = %v, want PSMServer1", server.Name)
	}
	if server.PSMVersion != "12.6" {
		t.Errorf("PSMVersion = %v, want 12.6", server.PSMVersion)
	}
}
