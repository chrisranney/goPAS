// Package accountimport provides tests for bulk account import functionality.
package accountimport

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

func TestStartImportJob(t *testing.T) {
	tests := []struct {
		name           string
		opts           StartImportJobOptions
		serverResponse *ImportJob
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful start",
			opts: StartImportJobOptions{
				Accounts: []ImportAccount{
					{UserName: "admin", Address: "server1", SafeName: "Safe1", PlatformID: "WinServer"},
					{UserName: "root", Address: "server2", SafeName: "Safe1", PlatformID: "UnixSSH"},
				},
			},
			serverResponse: &ImportJob{
				ID:     "job-123",
				Source: "File",
				Status: "Pending",
			},
			serverStatus: http.StatusAccepted,
			wantErr:      false,
		},
		{
			name: "empty accounts",
			opts: StartImportJobOptions{
				Accounts: []ImportAccount{},
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

			result, err := StartImportJob(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("StartImportJob() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("StartImportJob() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("StartImportJob().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestGetImportJob(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		serverResponse *ImportJobResult
		serverStatus   int
		wantErr        bool
	}{
		{
			name:  "successful get",
			jobID: "job-123",
			serverResponse: &ImportJobResult{
				ID:            "job-123",
				Status:        "Completed",
				TotalAccounts: 10,
				SuccessCount:  9,
				FailedCount:   1,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:    "empty job ID",
			jobID:   "",
			wantErr: true,
		},
		{
			name:         "job not found",
			jobID:        "nonexistent",
			serverStatus: http.StatusNotFound,
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

			result, err := GetImportJob(context.Background(), sess, tt.jobID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetImportJob() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetImportJob() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("GetImportJob().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
			if result.SuccessCount != tt.serverResponse.SuccessCount {
				t.Errorf("GetImportJob().SuccessCount = %v, want %v", result.SuccessCount, tt.serverResponse.SuccessCount)
			}
		})
	}
}

func TestListImportJobs(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []ImportJob
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful list",
			serverResponse: []ImportJob{
				{ID: "job-1", Status: "Completed"},
				{ID: "job-2", Status: "Running"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
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
					response := struct {
						BulkActions []ImportJob `json:"BulkActions"`
					}{BulkActions: tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListImportJobs(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("ListImportJobs() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListImportJobs() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListImportJobs() returned %d jobs, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestNewAccountObject(t *testing.T) {
	account := NewAccountObject("admin", "server.example.com", "TestSafe", "WinServer", "password123")

	if account.UserName != "admin" {
		t.Errorf("NewAccountObject().UserName = %v, want admin", account.UserName)
	}
	if account.Address != "server.example.com" {
		t.Errorf("NewAccountObject().Address = %v, want server.example.com", account.Address)
	}
	if account.SafeName != "TestSafe" {
		t.Errorf("NewAccountObject().SafeName = %v, want TestSafe", account.SafeName)
	}
	if account.PlatformID != "WinServer" {
		t.Errorf("NewAccountObject().PlatformID = %v, want WinServer", account.PlatformID)
	}
	if account.Secret != "password123" {
		t.Errorf("NewAccountObject().Secret = %v, want password123", account.Secret)
	}
}

func TestAddPendingAccount(t *testing.T) {
	tests := []struct {
		name           string
		account        ImportAccount
		serverResponse *PendingAccount
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful add",
			account: ImportAccount{
				UserName:   "admin",
				Address:    "server.example.com",
				PlatformID: "WinServer",
			},
			serverResponse: &PendingAccount{
				ID:         "pending-123",
				UserName:   "admin",
				Address:    "server.example.com",
				PlatformID: "WinServer",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing userName",
			account: ImportAccount{
				Address:    "server.example.com",
				PlatformID: "WinServer",
			},
			wantErr: true,
		},
		{
			name: "missing address",
			account: ImportAccount{
				UserName:   "admin",
				PlatformID: "WinServer",
			},
			wantErr: true,
		},
		{
			name: "missing platformID",
			account: ImportAccount{
				UserName: "admin",
				Address:  "server.example.com",
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

			result, err := AddPendingAccount(context.Background(), sess, tt.account)
			if tt.wantErr {
				if err == nil {
					t.Error("AddPendingAccount() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddPendingAccount() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("AddPendingAccount().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestGetAccountPasswordVersions(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		serverResponse []PasswordVersionInfo
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name:      "successful get",
			accountID: "acc-123",
			serverResponse: []PasswordVersionInfo{
				{Version: 1, Status: "Previous"},
				{Version: 2, Status: "Current"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					response := struct {
						Versions []PasswordVersionInfo `json:"Versions"`
					}{Versions: tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetAccountPasswordVersions(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetAccountPasswordVersions() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetAccountPasswordVersions() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("GetAccountPasswordVersions() returned %d versions, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		serverResponse string
		serverStatus   int
		wantPassword   string
		wantErr        bool
	}{
		{
			name:           "successful generate",
			accountID:      "acc-123",
			serverResponse: `"GeneratedPassword123!"`,
			serverStatus:   http.StatusOK,
			wantPassword:   "GeneratedPassword123!",
			wantErr:        false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
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
				w.Write([]byte(tt.serverResponse))
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GeneratePassword(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("GeneratePassword() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GeneratePassword() unexpected error: %v", err)
				return
			}

			if result != tt.wantPassword {
				t.Errorf("GeneratePassword() = %v, want %v", result, tt.wantPassword)
			}
		})
	}
}

func TestStartImportJob_InvalidSession(t *testing.T) {
	_, err := StartImportJob(context.Background(), nil, StartImportJobOptions{
		Accounts: []ImportAccount{{UserName: "test"}},
	})
	if err == nil {
		t.Error("StartImportJob() with nil session expected error, got nil")
	}
}

func TestGetImportJob_InvalidSession(t *testing.T) {
	_, err := GetImportJob(context.Background(), nil, "job-123")
	if err == nil {
		t.Error("GetImportJob() with nil session expected error, got nil")
	}
}

func TestListImportJobs_InvalidSession(t *testing.T) {
	_, err := ListImportJobs(context.Background(), nil)
	if err == nil {
		t.Error("ListImportJobs() with nil session expected error, got nil")
	}
}

func TestAddPendingAccount_InvalidSession(t *testing.T) {
	_, err := AddPendingAccount(context.Background(), nil, ImportAccount{
		UserName:   "admin",
		Address:    "server",
		PlatformID: "WinServer",
	})
	if err == nil {
		t.Error("AddPendingAccount() with nil session expected error, got nil")
	}
}

func TestGetAccountPasswordVersions_InvalidSession(t *testing.T) {
	_, err := GetAccountPasswordVersions(context.Background(), nil, "acc-123")
	if err == nil {
		t.Error("GetAccountPasswordVersions() with nil session expected error, got nil")
	}
}

func TestGeneratePassword_InvalidSession(t *testing.T) {
	_, err := GeneratePassword(context.Background(), nil, "acc-123")
	if err == nil {
		t.Error("GeneratePassword() with nil session expected error, got nil")
	}
}

func TestStartImportJob_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := StartImportJob(context.Background(), sess, StartImportJobOptions{
		Accounts: []ImportAccount{{UserName: "test", Address: "server", PlatformID: "WinServer", SafeName: "Safe"}},
	})
	if err == nil {
		t.Error("StartImportJob() expected error for server error")
	}
}

func TestGetImportJob_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := GetImportJob(context.Background(), sess, "job-123")
	if err == nil {
		t.Error("GetImportJob() expected error for server error")
	}
}

func TestAddPendingAccount_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := AddPendingAccount(context.Background(), sess, ImportAccount{
		UserName:   "admin",
		Address:    "server",
		PlatformID: "WinServer",
	})
	if err == nil {
		t.Error("AddPendingAccount() expected error for server error")
	}
}

func TestGetAccountPasswordVersions_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := GetAccountPasswordVersions(context.Background(), sess, "acc-123")
	if err == nil {
		t.Error("GetAccountPasswordVersions() expected error for server error")
	}
}

func TestGeneratePassword_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := GeneratePassword(context.Background(), sess, "acc-123")
	if err == nil {
		t.Error("GeneratePassword() expected error for server error")
	}
}

func TestImportAccount_Struct(t *testing.T) {
	account := ImportAccount{
		UserName:   "admin",
		Address:    "server.example.com",
		SafeName:   "TestSafe",
		PlatformID: "WinServer",
		Secret:     "password123",
		PlatformAccountProperties: map[string]interface{}{
			"LogonDomain": "DOMAIN",
		},
	}

	if account.UserName != "admin" {
		t.Errorf("UserName = %v, want admin", account.UserName)
	}
	if account.PlatformAccountProperties["LogonDomain"] != "DOMAIN" {
		t.Errorf("PlatformAccountProperties[LogonDomain] = %v, want DOMAIN", account.PlatformAccountProperties["LogonDomain"])
	}
}

func TestImportJob_Struct(t *testing.T) {
	job := ImportJob{
		ID:     "job-123",
		Source: "File",
		Status: "Completed",
	}

	if job.ID != "job-123" {
		t.Errorf("ID = %v, want job-123", job.ID)
	}
	if job.Status != "Completed" {
		t.Errorf("Status = %v, want Completed", job.Status)
	}
}

func TestImportJobResult_Struct(t *testing.T) {
	result := ImportJobResult{
		ID:            "job-123",
		Status:        "Completed",
		TotalAccounts: 10,
		SuccessCount:  9,
		FailedCount:   1,
	}

	if result.TotalAccounts != 10 {
		t.Errorf("TotalAccounts = %v, want 10", result.TotalAccounts)
	}
	if result.SuccessCount != 9 {
		t.Errorf("SuccessCount = %v, want 9", result.SuccessCount)
	}
}

func TestStartImportJobOptions_Struct(t *testing.T) {
	opts := StartImportJobOptions{
		Accounts: []ImportAccount{
			{UserName: "admin", Address: "server1"},
			{UserName: "root", Address: "server2"},
		},
	}

	if len(opts.Accounts) != 2 {
		t.Errorf("Accounts length = %v, want 2", len(opts.Accounts))
	}
}

func TestPendingAccount_Struct(t *testing.T) {
	account := PendingAccount{
		ID:         "pending-123",
		UserName:   "admin",
		Address:    "server",
		PlatformID: "WinServer",
	}

	if account.ID != "pending-123" {
		t.Errorf("ID = %v, want pending-123", account.ID)
	}
}

func TestPasswordVersionInfo_Struct(t *testing.T) {
	info := PasswordVersionInfo{
		Version: 2,
		Status:  "Current",
	}

	if info.Version != 2 {
		t.Errorf("Version = %v, want 2", info.Version)
	}
	if info.Status != "Current" {
		t.Errorf("Status = %v, want Current", info.Status)
	}
}
