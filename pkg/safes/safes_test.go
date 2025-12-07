// Package safes provides tests for safe management functionality.
package safes

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
		serverResponse *SafesResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListOptions{},
			serverResponse: &SafesResponse{
				Value: []Safe{
					{SafeURLId: "TestSafe1", SafeName: "TestSafe1", SafeNumber: 1},
					{SafeURLId: "TestSafe2", SafeName: "TestSafe2", SafeNumber: 2},
				},
				Count: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with search",
			opts: ListOptions{Search: "Test"},
			serverResponse: &SafesResponse{
				Value: []Safe{
					{SafeURLId: "TestSafe1", SafeName: "TestSafe1", SafeNumber: 1},
				},
				Count: 1,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with pagination",
			opts: ListOptions{Offset: 10, Limit: 5},
			serverResponse: &SafesResponse{
				Value:    []Safe{},
				Count:    0,
				NextLink: "https://example.com/api?offset=15",
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

			if result.Count != tt.serverResponse.Count {
				t.Errorf("List().Count = %v, want %v", result.Count, tt.serverResponse.Count)
			}
		})
	}
}

func TestList_InvalidSession(t *testing.T) {
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
			_, err := List(context.Background(), tt.sess, ListOptions{})
			if tt.wantErr && err == nil {
				t.Error("List() expected error, got nil")
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		safeName       string
		serverResponse *Safe
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful get",
			safeName: "TestSafe",
			serverResponse: &Safe{
				SafeURLId:    "TestSafe",
				SafeName:     "TestSafe",
				SafeNumber:   1,
				Description:  "Test safe description",
				ManagingCPM:  "PasswordManager",
				OLACEnabled:  true,
				CreationTime: 1705315800,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "safe not found",
			safeName:     "NonexistentSafe",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
		{
			name:     "empty safe name",
			safeName: "",
			wantErr:  true,
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

			result, err := Get(context.Background(), sess, tt.safeName)
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

			if result.SafeName != tt.serverResponse.SafeName {
				t.Errorf("Get().SafeName = %v, want %v", result.SafeName, tt.serverResponse.SafeName)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		opts           CreateOptions
		serverResponse *Safe
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful create",
			opts: CreateOptions{
				SafeName:    "NewSafe",
				Description: "A new test safe",
				ManagingCPM: "PasswordManager",
			},
			serverResponse: &Safe{
				SafeURLId:   "NewSafe",
				SafeName:    "NewSafe",
				SafeNumber:  10,
				Description: "A new test safe",
				ManagingCPM: "PasswordManager",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing safe name",
			opts: CreateOptions{
				Description: "A test safe",
			},
			wantErr: true,
		},
		{
			name: "safe name too long",
			opts: CreateOptions{
				SafeName: "ThisSafeNameIsWayTooLongToBeValid",
			},
			wantErr: true,
		},
		{
			name: "safe name exactly 28 characters",
			opts: CreateOptions{
				SafeName: "1234567890123456789012345678",
			},
			serverResponse: &Safe{
				SafeURLId: "1234567890123456789012345678",
				SafeName:  "1234567890123456789012345678",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
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

			if result.SafeName != tt.serverResponse.SafeName {
				t.Errorf("Create().SafeName = %v, want %v", result.SafeName, tt.serverResponse.SafeName)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		safeName       string
		opts           UpdateOptions
		serverResponse *Safe
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful update",
			safeName: "TestSafe",
			opts: UpdateOptions{
				Description: "Updated description",
			},
			serverResponse: &Safe{
				SafeURLId:   "TestSafe",
				SafeName:    "TestSafe",
				Description: "Updated description",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty safe name",
			safeName: "",
			opts: UpdateOptions{
				Description: "Updated description",
			},
			wantErr: true,
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

			result, err := Update(context.Background(), sess, tt.safeName, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Update() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Update() unexpected error: %v", err)
				return
			}

			if result.SafeName != tt.serverResponse.SafeName {
				t.Errorf("Update().SafeName = %v, want %v", result.SafeName, tt.serverResponse.SafeName)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		safeName     string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			safeName:     "TestSafe",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "safe not found",
			safeName:     "NonexistentSafe",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
		{
			name:     "empty safe name",
			safeName: "",
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

			err := Delete(context.Background(), sess, tt.safeName)
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

func TestSafe_Struct(t *testing.T) {
	// Test Safe struct fields
	safe := Safe{
		SafeURLId:                 "TestSafe",
		SafeName:                  "TestSafe",
		SafeNumber:                1,
		Description:               "Test Description",
		Location:                  "\\",
		OLACEnabled:               true,
		ManagingCPM:               "PasswordManager",
		NumberOfVersionsRetention: intPtr(10),
		NumberOfDaysRetention:     30,
		AutoPurgeEnabled:          false,
		CreationTime:              1705315800,
	}

	if safe.SafeName != "TestSafe" {
		t.Errorf("SafeName = %v, want TestSafe", safe.SafeName)
	}
	if safe.SafeNumber != 1 {
		t.Errorf("SafeNumber = %v, want 1", safe.SafeNumber)
	}
	if !safe.OLACEnabled {
		t.Error("OLACEnabled should be true")
	}
	if *safe.NumberOfVersionsRetention != 10 {
		t.Errorf("NumberOfVersionsRetention = %v, want 10", *safe.NumberOfVersionsRetention)
	}
}

func TestCreator_Struct(t *testing.T) {
	creator := Creator{
		ID:   "1",
		Name: "Administrator",
	}

	if creator.ID != "1" {
		t.Errorf("Creator.ID = %v, want 1", creator.ID)
	}
	if creator.Name != "Administrator" {
		t.Errorf("Creator.Name = %v, want Administrator", creator.Name)
	}
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}
