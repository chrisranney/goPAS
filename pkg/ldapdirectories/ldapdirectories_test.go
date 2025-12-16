// Package ldapdirectories provides tests for LDAP directory management functionality.
package ldapdirectories

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
		serverResponse DirectoriesResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			serverResponse: DirectoriesResponse{
				Directories: []Directory{
					{DirectoryID: "dir1", DomainName: "example.com"},
					{DirectoryID: "dir2", DomainName: "test.local"},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty list",
			serverResponse: DirectoriesResponse{
				Directories: []Directory{},
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
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := List(context.Background(), sess)
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

			if len(result) != len(tt.serverResponse.Directories) {
				t.Errorf("List() returned %d directories, want %d", len(result), len(tt.serverResponse.Directories))
			}
		})
	}
}

func TestList_InvalidSession(t *testing.T) {
	_, err := List(context.Background(), nil)
	if err == nil {
		t.Error("List() expected error for nil session, got nil")
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		directoryID    string
		serverResponse Directory
		serverStatus   int
		wantErr        bool
	}{
		{
			name:        "successful get",
			directoryID: "dir123",
			serverResponse: Directory{
				DirectoryID:       "dir123",
				DomainName:        "example.com",
				DomainBaseContext: "DC=example,DC=com",
				SSLConnect:        true,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "empty directory ID",
			directoryID: "",
			wantErr:     true,
		},
		{
			name:         "server error",
			directoryID:  "dir123",
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
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := Get(context.Background(), sess, tt.directoryID)
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

			if result.DirectoryID != tt.serverResponse.DirectoryID {
				t.Errorf("Get() returned DirectoryID %v, want %v", result.DirectoryID, tt.serverResponse.DirectoryID)
			}
		})
	}
}

func TestGet_InvalidSession(t *testing.T) {
	_, err := Get(context.Background(), nil, "dir123")
	if err == nil {
		t.Error("Get() expected error for nil session, got nil")
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		opts           CreateOptions
		serverResponse Directory
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful create",
			opts: CreateOptions{
				DomainName:        "example.com",
				DomainBaseContext: "DC=example,DC=com",
				BindUsername:      "admin",
				BindPassword:      "secret",
				SSLConnect:        true,
			},
			serverResponse: Directory{
				DirectoryID: "dir123",
				DomainName:  "example.com",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "successful create with DC list",
			opts: CreateOptions{
				DomainName: "example.com",
				DCList: []DomainController{
					{Name: "dc1.example.com", Address: "192.168.1.1", Port: 389},
				},
			},
			serverResponse: Directory{
				DirectoryID: "dir124",
				DomainName:  "example.com",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty domain name",
			opts: CreateOptions{
				DomainName: "",
			},
			wantErr: true,
		},
		{
			name: "server error",
			opts: CreateOptions{
				DomainName: "example.com",
			},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
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
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

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

			if result.DomainName != tt.serverResponse.DomainName {
				t.Errorf("Create() returned DomainName %v, want %v", result.DomainName, tt.serverResponse.DomainName)
			}
		})
	}
}

func TestCreate_InvalidSession(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateOptions{DomainName: "example.com"})
	if err == nil {
		t.Error("Create() expected error for nil session, got nil")
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		directoryID  string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			directoryID:  "dir123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:        "empty directory ID",
			directoryID: "",
			wantErr:     true,
		},
		{
			name:         "server error",
			directoryID:  "dir123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
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

			err := Delete(context.Background(), sess, tt.directoryID)
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

func TestDelete_InvalidSession(t *testing.T) {
	err := Delete(context.Background(), nil, "dir123")
	if err == nil {
		t.Error("Delete() expected error for nil session, got nil")
	}
}

func TestListMappings(t *testing.T) {
	tests := []struct {
		name           string
		directoryID    string
		serverResponse struct {
			Mappings []DirectoryMapping `json:"Mappings"`
		}
		serverStatus int
		wantErr      bool
	}{
		{
			name:        "successful list",
			directoryID: "dir123",
			serverResponse: struct {
				Mappings []DirectoryMapping `json:"Mappings"`
			}{
				Mappings: []DirectoryMapping{
					{MappingID: "map1", DirectoryMappingName: "AdminMapping", LDAPBranch: "OU=Admins"},
					{MappingID: "map2", DirectoryMappingName: "UserMapping", LDAPBranch: "OU=Users"},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "empty directory ID",
			directoryID: "",
			wantErr:     true,
		},
		{
			name:         "server error",
			directoryID:  "dir123",
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
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListMappings(context.Background(), sess, tt.directoryID)
			if tt.wantErr {
				if err == nil {
					t.Error("ListMappings() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListMappings() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.serverResponse.Mappings) {
				t.Errorf("ListMappings() returned %d mappings, want %d", len(result), len(tt.serverResponse.Mappings))
			}
		})
	}
}

func TestListMappings_InvalidSession(t *testing.T) {
	_, err := ListMappings(context.Background(), nil, "dir123")
	if err == nil {
		t.Error("ListMappings() expected error for nil session, got nil")
	}
}

func TestCreateMapping(t *testing.T) {
	tests := []struct {
		name           string
		directoryID    string
		opts           CreateMappingOptions
		serverResponse DirectoryMapping
		serverStatus   int
		wantErr        bool
	}{
		{
			name:        "successful create",
			directoryID: "dir123",
			opts: CreateMappingOptions{
				DirectoryMappingName: "AdminMapping",
				LDAPBranch:           "OU=Admins,DC=example,DC=com",
				VaultGroups:          []string{"Vault Admins"},
			},
			serverResponse: DirectoryMapping{
				MappingID:            "map123",
				DirectoryMappingName: "AdminMapping",
				LDAPBranch:           "OU=Admins,DC=example,DC=com",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "empty directory ID",
			directoryID: "",
			opts: CreateMappingOptions{
				DirectoryMappingName: "AdminMapping",
			},
			wantErr: true,
		},
		{
			name:        "empty mapping name",
			directoryID: "dir123",
			opts: CreateMappingOptions{
				DirectoryMappingName: "",
			},
			wantErr: true,
		},
		{
			name:        "server error",
			directoryID: "dir123",
			opts: CreateMappingOptions{
				DirectoryMappingName: "AdminMapping",
				LDAPBranch:           "OU=Admins",
			},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
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
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := CreateMapping(context.Background(), sess, tt.directoryID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("CreateMapping() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateMapping() unexpected error: %v", err)
				return
			}

			if result.DirectoryMappingName != tt.serverResponse.DirectoryMappingName {
				t.Errorf("CreateMapping() returned DirectoryMappingName %v, want %v", result.DirectoryMappingName, tt.serverResponse.DirectoryMappingName)
			}
		})
	}
}

func TestCreateMapping_InvalidSession(t *testing.T) {
	_, err := CreateMapping(context.Background(), nil, "dir123", CreateMappingOptions{DirectoryMappingName: "Test"})
	if err == nil {
		t.Error("CreateMapping() expected error for nil session, got nil")
	}
}

func TestDeleteMapping(t *testing.T) {
	tests := []struct {
		name         string
		directoryID  string
		mappingID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			directoryID:  "dir123",
			mappingID:    "map456",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:        "empty directory ID",
			directoryID: "",
			mappingID:   "map456",
			wantErr:     true,
		},
		{
			name:        "empty mapping ID",
			directoryID: "dir123",
			mappingID:   "",
			wantErr:     true,
		},
		{
			name:         "server error",
			directoryID:  "dir123",
			mappingID:    "map456",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
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

			err := DeleteMapping(context.Background(), sess, tt.directoryID, tt.mappingID)
			if tt.wantErr {
				if err == nil {
					t.Error("DeleteMapping() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteMapping() unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteMapping_InvalidSession(t *testing.T) {
	err := DeleteMapping(context.Background(), nil, "dir123", "map456")
	if err == nil {
		t.Error("DeleteMapping() expected error for nil session, got nil")
	}
}

func TestDirectory_Struct(t *testing.T) {
	dir := Directory{
		DirectoryID:        "dir123",
		DomainName:         "example.com",
		DomainBaseContext:  "DC=example,DC=com",
		BindUsername:       "admin",
		SSLConnect:         true,
		VaultUseDomainName: true,
		DCList: []DomainController{
			{Name: "dc1.example.com", Address: "192.168.1.1", Port: 389, SSLConnect: false},
		},
	}

	if dir.DomainName != "example.com" {
		t.Errorf("DomainName = %v, want example.com", dir.DomainName)
	}
	if len(dir.DCList) != 1 {
		t.Errorf("DCList length = %v, want 1", len(dir.DCList))
	}
}

func TestDirectoryMapping_Struct(t *testing.T) {
	mapping := DirectoryMapping{
		MappingID:            "map123",
		DirectoryMappingName: "AdminMapping",
		LDAPBranch:           "OU=Admins,DC=example,DC=com",
		DomainGroups:         []string{"Domain Admins"},
		VaultGroups:          []string{"Vault Admins"},
		Location:             "\\",
		LDAPQuery:            "(objectClass=user)",
	}

	if mapping.DirectoryMappingName != "AdminMapping" {
		t.Errorf("DirectoryMappingName = %v, want AdminMapping", mapping.DirectoryMappingName)
	}
	if len(mapping.DomainGroups) != 1 {
		t.Errorf("DomainGroups length = %v, want 1", len(mapping.DomainGroups))
	}
}
