// Package systemhealth provides tests for system health monitoring functionality.
package systemhealth

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

func TestListComponentSummary(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []ComponentSummary
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			serverResponse: []ComponentSummary{
				{ComponentID: "1", ComponentName: "Vault", ComponentType: "Vault", IsLoggedOn: true},
				{ComponentID: "2", ComponentName: "CPM", ComponentType: "CPM", IsLoggedOn: true},
				{ComponentID: "3", ComponentName: "PVWA", ComponentType: "PVWA", IsLoggedOn: true},
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
				response := ComponentSummaryResponse{Components: tt.serverResponse}
				json.NewEncoder(w).Encode(response)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListComponentSummary(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("ListComponentSummary() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListComponentSummary() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse) {
				t.Errorf("ListComponentSummary() returned %d components, want %d", len(result), len(tt.serverResponse))
			}
		})
	}
}

func TestGetComponentDetail(t *testing.T) {
	tests := []struct {
		name           string
		componentID    string
		serverResponse *ComponentDetail
		serverStatus   int
		wantErr        bool
	}{
		{
			name:        "successful get",
			componentID: "vault-1",
			serverResponse: &ComponentDetail{
				ComponentID:      "vault-1",
				ComponentName:    "Vault",
				ComponentType:    "Vault",
				IsLoggedOn:       true,
				ComponentVersion: "14.0",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "empty component ID",
			componentID: "",
			wantErr:     true,
		},
		{
			name:         "not found",
			componentID:  "nonexistent",
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

			result, err := GetComponentDetail(context.Background(), sess, tt.componentID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetComponentDetail() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetComponentDetail() unexpected error: %v", err)
				return
			}

			if result.ComponentID != tt.serverResponse.ComponentID {
				t.Errorf("GetComponentDetail().ComponentID = %v, want %v", result.ComponentID, tt.serverResponse.ComponentID)
			}
		})
	}
}

func TestGetVaultHealth(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse *VaultHealth
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "healthy vault",
			serverResponse: &VaultHealth{
				IsHealthy:     true,
				HealthDetails: "All systems operational",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "unhealthy vault",
			serverResponse: &VaultHealth{
				IsHealthy:     false,
				HealthDetails: "CPM connectivity issue",
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

			result, err := GetVaultHealth(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetVaultHealth() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetVaultHealth() unexpected error: %v", err)
				return
			}

			if result.IsHealthy != tt.serverResponse.IsHealthy {
				t.Errorf("GetVaultHealth().IsHealthy = %v, want %v", result.IsHealthy, tt.serverResponse.IsHealthy)
			}
		})
	}
}

func TestComponentSummary_Struct(t *testing.T) {
	summary := ComponentSummary{
		ComponentID:            "vault-1",
		ComponentName:          "Primary Vault",
		ComponentType:          "Vault",
		Description:            "Production vault",
		ConnectedComponentID:   "dr-vault-1",
		ConnectedComponentName: "DR Vault",
		IsLoggedOn:             true,
		LastLogonDate:          1705315800,
	}

	if summary.ComponentName != "Primary Vault" {
		t.Errorf("ComponentName = %v, want Primary Vault", summary.ComponentName)
	}
	if !summary.IsLoggedOn {
		t.Error("IsLoggedOn should be true")
	}
}

func TestComponentDetail_Struct(t *testing.T) {
	detail := ComponentDetail{
		ComponentID:            "vault-1",
		ComponentName:          "Primary Vault",
		ComponentType:          "Vault",
		Description:            "Production vault",
		ConnectedComponentID:   "dr-vault-1",
		ConnectedComponentName: "DR Vault",
		IsLoggedOn:             true,
		LastLogonDate:          1705315800,
		ComponentVersion:       "14.0.0",
		ComponentSpecificData: map[string]interface{}{
			"LicenseCapacity": 1000,
			"UsedCapacity":    500,
		},
	}

	if detail.ComponentVersion != "14.0.0" {
		t.Errorf("ComponentVersion = %v, want 14.0.0", detail.ComponentVersion)
	}
	if detail.ComponentSpecificData["LicenseCapacity"] != 1000 {
		t.Errorf("ComponentSpecificData[LicenseCapacity] = %v, want 1000", detail.ComponentSpecificData["LicenseCapacity"])
	}
}

func TestVaultHealth_Struct(t *testing.T) {
	health := VaultHealth{
		IsHealthy:     true,
		HealthDetails: "All systems operational",
	}

	if !health.IsHealthy {
		t.Error("IsHealthy should be true")
	}
	if health.HealthDetails != "All systems operational" {
		t.Errorf("HealthDetails = %v, want All systems operational", health.HealthDetails)
	}
}
