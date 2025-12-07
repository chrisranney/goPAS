// Package platforms provides tests for platform management functionality.
package platforms

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
		serverResponse *PlatformsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListOptions{},
			serverResponse: &PlatformsResponse{
				Platforms: []Platform{
					{ID: "1", Name: "WinServerLocal", Active: true},
					{ID: "2", Name: "UnixSSH", Active: true},
				},
				Total: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with search",
			opts: ListOptions{Search: "Win"},
			serverResponse: &PlatformsResponse{
				Platforms: []Platform{
					{ID: "1", Name: "WinServerLocal", Active: true},
				},
				Total: 1,
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
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
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

			if result.Total != tt.serverResponse.Total {
				t.Errorf("List().Total = %v, want %v", result.Total, tt.serverResponse.Total)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		platformID     string
		serverResponse *Platform
		serverStatus   int
		wantErr        bool
	}{
		{
			name:       "successful get",
			platformID: "WinServerLocal",
			serverResponse: &Platform{
				ID:     "1",
				Name:   "WinServerLocal",
				Active: true,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "empty platform ID",
			platformID: "",
			wantErr:    true,
		},
		{
			name:         "not found",
			platformID:   "nonexistent",
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

			result, err := Get(context.Background(), sess, tt.platformID)
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

			if result.Name != tt.serverResponse.Name {
				t.Errorf("Get().Name = %v, want %v", result.Name, tt.serverResponse.Name)
			}
		})
	}
}

func TestActivate(t *testing.T) {
	tests := []struct {
		name         string
		platformID   string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful activate",
			platformID:   "WinServerLocal",
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
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := Activate(context.Background(), sess, tt.platformID)
			if tt.wantErr {
				if err == nil {
					t.Error("Activate() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Activate() unexpected error: %v", err)
			}
		})
	}
}

func TestDeactivate(t *testing.T) {
	tests := []struct {
		name         string
		platformID   string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful deactivate",
			platformID:   "WinServerLocal",
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
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := Deactivate(context.Background(), sess, tt.platformID)
			if tt.wantErr {
				if err == nil {
					t.Error("Deactivate() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Deactivate() unexpected error: %v", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		platformID   string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			platformID:   "WinServerLocal",
			serverStatus: http.StatusNoContent,
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
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := Delete(context.Background(), sess, tt.platformID)
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

func TestDuplicate(t *testing.T) {
	tests := []struct {
		name           string
		platformID     string
		opts           DuplicateOptions
		serverResponse *Platform
		serverStatus   int
		wantErr        bool
	}{
		{
			name:       "successful duplicate",
			platformID: "WinServerLocal",
			opts: DuplicateOptions{
				Name:        "WinServerLocal_Copy",
				Description: "Copy of WinServerLocal",
			},
			serverResponse: &Platform{
				ID:   "3",
				Name: "WinServerLocal_Copy",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name:       "empty platform ID",
			platformID: "",
			opts:       DuplicateOptions{Name: "Copy"},
			wantErr:    true,
		},
		{
			name:       "missing name",
			platformID: "WinServerLocal",
			opts:       DuplicateOptions{},
			wantErr:    true,
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

			result, err := Duplicate(context.Background(), sess, tt.platformID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Duplicate() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Duplicate() unexpected error: %v", err)
				return
			}

			if result.Name != tt.serverResponse.Name {
				t.Errorf("Duplicate().Name = %v, want %v", result.Name, tt.serverResponse.Name)
			}
		})
	}
}

func TestExportPlatform(t *testing.T) {
	tests := []struct {
		name           string
		platformID     string
		serverResponse []byte
		serverStatus   int
		wantErr        bool
	}{
		{
			name:           "successful export",
			platformID:     "WinServerLocal",
			serverResponse: []byte("ZIP_FILE_CONTENTS"),
			serverStatus:   http.StatusOK,
			wantErr:        false,
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
				w.WriteHeader(tt.serverStatus)
				w.Write(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ExportPlatform(context.Background(), sess, tt.platformID)
			if tt.wantErr {
				if err == nil {
					t.Error("ExportPlatform() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ExportPlatform() unexpected error: %v", err)
				return
			}

			if string(result) != string(tt.serverResponse) {
				t.Errorf("ExportPlatform() = %v, want %v", string(result), string(tt.serverResponse))
			}
		})
	}
}

func TestImportPlatform(t *testing.T) {
	tests := []struct {
		name         string
		platformZip  []byte
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful import",
			platformZip:  []byte("ZIP_FILE_CONTENTS"),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "empty platform zip",
			platformZip: []byte{},
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

			err := ImportPlatform(context.Background(), sess, tt.platformZip)
			if tt.wantErr {
				if err == nil {
					t.Error("ImportPlatform() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ImportPlatform() unexpected error: %v", err)
			}
		})
	}
}

func TestPlatform_Struct(t *testing.T) {
	platform := Platform{
		ID:           "1",
		PlatformID:   "WinServerLocal",
		Name:         "Windows Server Local",
		Active:       true,
		Description:  "Windows local accounts",
		SystemType:   "Windows",
		PlatformType: "Regular",
		AllowedSafes: ".*",
	}

	if platform.Name != "Windows Server Local" {
		t.Errorf("Name = %v, want Windows Server Local", platform.Name)
	}
	if !platform.Active {
		t.Error("Active should be true")
	}
}

func TestCredentialsPolicy_Struct(t *testing.T) {
	policy := CredentialsPolicy{
		Verification: &VerificationPolicy{
			PerformAutomatic:          true,
			RequirePasswordEveryXDays: 30,
			AutoOnAdd:                 true,
			AllowManual:               true,
		},
		Change: &ChangePolicy{
			PerformAutomatic:          true,
			RequirePasswordEveryXDays: 90,
			AllowManual:               true,
		},
		Reconcile: &ReconcilePolicy{
			AutomaticReconcileWhenUnsynced: true,
			AllowManual:                    true,
		},
	}

	if !policy.Verification.PerformAutomatic {
		t.Error("Verification.PerformAutomatic should be true")
	}
	if policy.Change.RequirePasswordEveryXDays != 90 {
		t.Errorf("Change.RequirePasswordEveryXDays = %v, want 90", policy.Change.RequirePasswordEveryXDays)
	}
}
