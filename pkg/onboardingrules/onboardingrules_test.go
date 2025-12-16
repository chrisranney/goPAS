// Package onboardingrules provides tests for automatic onboarding rules functionality.
package onboardingrules

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrisranney/gopas/internal/client"
	"github.com/chrisranney/gopas/internal/session"
	"github.com/chrisranney/gopas/pkg/types"
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
		serverResponse OnboardingRulesResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			serverResponse: OnboardingRulesResponse{
				AutomaticOnboardingRules: []OnboardingRule{
					{RuleID: 1, RuleName: "Linux Admin", TargetPlatformID: "UnixSSH", TargetSafeName: "LinuxSafe"},
					{RuleID: 2, RuleName: "Windows Admin", TargetPlatformID: "WinDomain", TargetSafeName: "WindowsSafe"},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty list",
			serverResponse: OnboardingRulesResponse{
				AutomaticOnboardingRules: []OnboardingRule{},
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

			if len(result) != len(tt.serverResponse.AutomaticOnboardingRules) {
				t.Errorf("List() returned %d rules, want %d", len(result), len(tt.serverResponse.AutomaticOnboardingRules))
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
		ruleID         int
		serverResponse OnboardingRule
		serverStatus   int
		wantErr        bool
	}{
		{
			name:   "successful get",
			ruleID: 123,
			serverResponse: OnboardingRule{
				RuleID:           123,
				RuleName:         "Linux Admin",
				RuleDescription:  "Onboard Linux admin accounts",
				TargetPlatformID: "UnixSSH",
				TargetSafeName:   "LinuxSafe",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			ruleID:       123,
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

			result, err := Get(context.Background(), sess, tt.ruleID)
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

			if result.RuleID != tt.serverResponse.RuleID {
				t.Errorf("Get() returned RuleID %v, want %v", result.RuleID, tt.serverResponse.RuleID)
			}
		})
	}
}

func TestGet_InvalidSession(t *testing.T) {
	_, err := Get(context.Background(), nil, 123)
	if err == nil {
		t.Error("Get() expected error for nil session, got nil")
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		opts           CreateOptions
		serverResponse OnboardingRule
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful create",
			opts: CreateOptions{
				RuleName:         "Linux Admin",
				RuleDescription:  "Onboard Linux admin accounts",
				TargetPlatformID: "UnixSSH",
				TargetSafeName:   "LinuxSafe",
				UserNameFilter:   "admin*",
			},
			serverResponse: OnboardingRule{
				RuleID:           1,
				RuleName:         "Linux Admin",
				TargetPlatformID: "UnixSSH",
				TargetSafeName:   "LinuxSafe",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty rule name",
			opts: CreateOptions{
				RuleName:         "",
				TargetPlatformID: "UnixSSH",
				TargetSafeName:   "LinuxSafe",
			},
			wantErr: true,
		},
		{
			name: "empty platform ID",
			opts: CreateOptions{
				RuleName:         "Linux Admin",
				TargetPlatformID: "",
				TargetSafeName:   "LinuxSafe",
			},
			wantErr: true,
		},
		{
			name: "empty safe name",
			opts: CreateOptions{
				RuleName:         "Linux Admin",
				TargetPlatformID: "UnixSSH",
				TargetSafeName:   "",
			},
			wantErr: true,
		},
		{
			name: "server error",
			opts: CreateOptions{
				RuleName:         "Linux Admin",
				TargetPlatformID: "UnixSSH",
				TargetSafeName:   "LinuxSafe",
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

			if result.RuleName != tt.serverResponse.RuleName {
				t.Errorf("Create() returned RuleName %v, want %v", result.RuleName, tt.serverResponse.RuleName)
			}
		})
	}
}

func TestCreate_InvalidSession(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateOptions{RuleName: "Test", TargetPlatformID: "UnixSSH", TargetSafeName: "Safe"})
	if err == nil {
		t.Error("Create() expected error for nil session, got nil")
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         int
		opts           UpdateOptions
		serverResponse OnboardingRule
		serverStatus   int
		wantErr        bool
	}{
		{
			name:   "successful update",
			ruleID: 123,
			opts: UpdateOptions{
				RuleName:        "Updated Linux Admin",
				RuleDescription: "Updated description",
			},
			serverResponse: OnboardingRule{
				RuleID:   123,
				RuleName: "Updated Linux Admin",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			ruleID:       123,
			opts:         UpdateOptions{RuleName: "Test"},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
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
				json.NewEncoder(w).Encode(tt.serverResponse)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := Update(context.Background(), sess, tt.ruleID, tt.opts)
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

			if result.RuleName != tt.serverResponse.RuleName {
				t.Errorf("Update() returned RuleName %v, want %v", result.RuleName, tt.serverResponse.RuleName)
			}
		})
	}
}

func TestUpdate_InvalidSession(t *testing.T) {
	_, err := Update(context.Background(), nil, 123, UpdateOptions{RuleName: "Test"})
	if err == nil {
		t.Error("Update() expected error for nil session, got nil")
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		ruleID       int
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			ruleID:       123,
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "server error",
			ruleID:       123,
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

			err := Delete(context.Background(), sess, tt.ruleID)
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
	err := Delete(context.Background(), nil, 123)
	if err == nil {
		t.Error("Delete() expected error for nil session, got nil")
	}
}

func TestListDiscoveredAccounts(t *testing.T) {
	tests := []struct {
		name           string
		opts           ListDiscoveredOptions
		serverResponse DiscoveredAccountsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListDiscoveredOptions{},
			serverResponse: DiscoveredAccountsResponse{
				Value: []DiscoveredAccount{
					{ID: "disc1", UserName: "admin", Address: "server1.example.com"},
					{ID: "disc2", UserName: "root", Address: "server2.example.com"},
				},
				Count: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with options",
			opts: ListDiscoveredOptions{
				Search: "admin",
				Filter: "platformType eq Windows",
			},
			serverResponse: DiscoveredAccountsResponse{
				Value: []DiscoveredAccount{
					{ID: "disc1", UserName: "admin", Address: "server1.example.com"},
				},
				Count: 1,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			opts:         ListDiscoveredOptions{},
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

			result, err := ListDiscoveredAccounts(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ListDiscoveredAccounts() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListDiscoveredAccounts() unexpected error: %v", err)
				return
			}

			if len(result.Value) != len(tt.serverResponse.Value) {
				t.Errorf("ListDiscoveredAccounts() returned %d accounts, want %d", len(result.Value), len(tt.serverResponse.Value))
			}
		})
	}
}

func TestListDiscoveredAccounts_InvalidSession(t *testing.T) {
	_, err := ListDiscoveredAccounts(context.Background(), nil, ListDiscoveredOptions{})
	if err == nil {
		t.Error("ListDiscoveredAccounts() expected error for nil session, got nil")
	}
}

func TestGetDiscoveredAccount(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		serverResponse DiscoveredAccount
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful get",
			accountID: "disc123",
			serverResponse: DiscoveredAccount{
				ID:           "disc123",
				UserName:     "admin",
				Address:      "server1.example.com",
				PlatformType: "Windows",
				Privileged:   true,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			accountID:    "disc123",
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

			result, err := GetDiscoveredAccount(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetDiscoveredAccount() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetDiscoveredAccount() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("GetDiscoveredAccount() returned ID %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestGetDiscoveredAccount_InvalidSession(t *testing.T) {
	_, err := GetDiscoveredAccount(context.Background(), nil, "disc123")
	if err == nil {
		t.Error("GetDiscoveredAccount() expected error for nil session, got nil")
	}
}

func TestAddDiscoveredAccount(t *testing.T) {
	tests := []struct {
		name           string
		opts           AddDiscoveredAccountOptions
		serverResponse DiscoveredAccount
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful add",
			opts: AddDiscoveredAccountOptions{
				UserName:     "admin",
				Address:      "server1.example.com",
				PlatformType: "Windows",
			},
			serverResponse: DiscoveredAccount{
				ID:       "disc123",
				UserName: "admin",
				Address:  "server1.example.com",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty username",
			opts: AddDiscoveredAccountOptions{
				UserName: "",
				Address:  "server1.example.com",
			},
			wantErr: true,
		},
		{
			name: "empty address",
			opts: AddDiscoveredAccountOptions{
				UserName: "admin",
				Address:  "",
			},
			wantErr: true,
		},
		{
			name: "server error",
			opts: AddDiscoveredAccountOptions{
				UserName: "admin",
				Address:  "server1.example.com",
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

			result, err := AddDiscoveredAccount(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("AddDiscoveredAccount() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("AddDiscoveredAccount() unexpected error: %v", err)
				return
			}

			if result.UserName != tt.serverResponse.UserName {
				t.Errorf("AddDiscoveredAccount() returned UserName %v, want %v", result.UserName, tt.serverResponse.UserName)
			}
		})
	}
}

func TestAddDiscoveredAccount_InvalidSession(t *testing.T) {
	_, err := AddDiscoveredAccount(context.Background(), nil, AddDiscoveredAccountOptions{UserName: "admin", Address: "server1"})
	if err == nil {
		t.Error("AddDiscoveredAccount() expected error for nil session, got nil")
	}
}

func TestDeleteDiscoveredAccount(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			accountID:    "disc123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
		},
		{
			name:         "server error",
			accountID:    "disc123",
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

			err := DeleteDiscoveredAccount(context.Background(), sess, tt.accountID)
			if tt.wantErr {
				if err == nil {
					t.Error("DeleteDiscoveredAccount() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteDiscoveredAccount() unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteDiscoveredAccount_InvalidSession(t *testing.T) {
	err := DeleteDiscoveredAccount(context.Background(), nil, "disc123")
	if err == nil {
		t.Error("DeleteDiscoveredAccount() expected error for nil session, got nil")
	}
}

func TestClearDiscoveredAccounts(t *testing.T) {
	tests := []struct {
		name         string
		opts         ClearDiscoveredAccountsOptions
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful clear",
			opts:         ClearDiscoveredAccountsOptions{},
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "clear with discovery source",
			opts:         ClearDiscoveredAccountsOptions{DiscoverySource: "Windows Discovery"},
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "server error",
			opts:         ClearDiscoveredAccountsOptions{},
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

			err := ClearDiscoveredAccounts(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ClearDiscoveredAccounts() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ClearDiscoveredAccounts() unexpected error: %v", err)
			}
		})
	}
}

func TestClearDiscoveredAccounts_InvalidSession(t *testing.T) {
	err := ClearDiscoveredAccounts(context.Background(), nil, ClearDiscoveredAccountsOptions{})
	if err == nil {
		t.Error("ClearDiscoveredAccounts() expected error for nil session, got nil")
	}
}

func TestPublishDiscoveredAccount(t *testing.T) {
	tests := []struct {
		name           string
		opts           PublishDiscoveredAccountOptions
		serverResponse PublishedAccount
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful publish",
			opts: PublishDiscoveredAccountOptions{
				AccountID:  "disc123",
				SafeName:   "TargetSafe",
				PlatformID: "UnixSSH",
			},
			serverResponse: PublishedAccount{
				ID:         "acc123",
				SafeName:   "TargetSafe",
				PlatformID: "UnixSSH",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty account ID",
			opts: PublishDiscoveredAccountOptions{
				AccountID: "",
				SafeName:  "TargetSafe",
			},
			wantErr: true,
		},
		{
			name: "empty safe name",
			opts: PublishDiscoveredAccountOptions{
				AccountID: "disc123",
				SafeName:  "",
			},
			wantErr: true,
		},
		{
			name: "server error",
			opts: PublishDiscoveredAccountOptions{
				AccountID: "disc123",
				SafeName:  "TargetSafe",
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

			result, err := PublishDiscoveredAccount(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("PublishDiscoveredAccount() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("PublishDiscoveredAccount() unexpected error: %v", err)
				return
			}

			if result.SafeName != tt.serverResponse.SafeName {
				t.Errorf("PublishDiscoveredAccount() returned SafeName %v, want %v", result.SafeName, tt.serverResponse.SafeName)
			}
		})
	}
}

func TestPublishDiscoveredAccount_InvalidSession(t *testing.T) {
	_, err := PublishDiscoveredAccount(context.Background(), nil, PublishDiscoveredAccountOptions{AccountID: "disc123", SafeName: "Safe"})
	if err == nil {
		t.Error("PublishDiscoveredAccount() expected error for nil session, got nil")
	}
}

func TestOnboardingRule_Struct(t *testing.T) {
	rule := OnboardingRule{
		RuleID:                1,
		RuleName:              "Linux Admin",
		RuleDescription:       "Onboard Linux admin accounts",
		TargetPlatformID:      "UnixSSH",
		TargetSafeName:        "LinuxSafe",
		TargetDeviceType:      "Server",
		IsAdminIDFilter:       true,
		MachineTypeFilter:     "Server",
		SystemTypeFilter:      "Linux",
		UserNameFilter:        "admin*",
		UserNameMethod:        "Begins",
		AddressFilter:         "*.example.com",
		AddressMethod:         "Ends",
		AccountCategoryFilter: "Privileged",
		RulePrecedence:        1,
	}

	if rule.RuleName != "Linux Admin" {
		t.Errorf("RuleName = %v, want Linux Admin", rule.RuleName)
	}
	if rule.IsAdminIDFilter != true {
		t.Errorf("IsAdminIDFilter = %v, want true", rule.IsAdminIDFilter)
	}
}

func TestDiscoveredAccount_Struct(t *testing.T) {
	account := DiscoveredAccount{
		ID:                   types.FlexibleID("disc123"),
		UserName:             "admin",
		Address:              "server1.example.com",
		DiscoveryDateTime:    1609459200,
		AccountEnabled:       true,
		OsGroups:             "Administrators",
		PlatformType:         "Windows",
		Domain:               "EXAMPLE",
		Privileged:           true,
		UserDisplayName:      "Administrator",
		Description:          "Built-in administrator account",
		PasswordNeverExpires: true,
		Dependencies: []DiscoveredDependency{
			{Name: "Service1", Address: "server1.example.com", Type: "Service"},
		},
	}

	if account.UserName != "admin" {
		t.Errorf("UserName = %v, want admin", account.UserName)
	}
	if len(account.Dependencies) != 1 {
		t.Errorf("Dependencies length = %v, want 1", len(account.Dependencies))
	}
}
