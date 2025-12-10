// Package safemembers provides tests for safe member management functionality.
package safemembers

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
		safeName       string
		opts           ListOptions
		serverResponse *SafeMembersResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful list",
			safeName: "TestSafe",
			opts:     ListOptions{},
			serverResponse: &SafeMembersResponse{
				Value: []SafeMember{
					{MemberName: "admin", MemberType: "User"},
					{MemberName: "Vault Admins", MemberType: "Group"},
				},
				Count: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty safe name",
			safeName: "",
			opts:     ListOptions{},
			wantErr:  true,
		},
		{
			name:         "server error",
			safeName:     "TestSafe",
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

			result, err := List(context.Background(), sess, tt.safeName, tt.opts)
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

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		safeName       string
		memberName     string
		serverResponse *SafeMember
		serverStatus   int
		wantErr        bool
	}{
		{
			name:       "successful get",
			safeName:   "TestSafe",
			memberName: "admin",
			serverResponse: &SafeMember{
				MemberName: "admin",
				MemberType: "User",
				Permissions: &Permissions{
					UseAccounts:      true,
					RetrieveAccounts: true,
					ListAccounts:     true,
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "empty safe name",
			safeName:   "",
			memberName: "admin",
			wantErr:    true,
		},
		{
			name:       "empty member name",
			safeName:   "TestSafe",
			memberName: "",
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

			result, err := Get(context.Background(), sess, tt.safeName, tt.memberName)
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

			if result.MemberName != tt.serverResponse.MemberName {
				t.Errorf("Get().MemberName = %v, want %v", result.MemberName, tt.serverResponse.MemberName)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name           string
		safeName       string
		opts           AddOptions
		serverResponse *SafeMember
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful add",
			safeName: "TestSafe",
			opts: AddOptions{
				MemberName:  "newuser",
				Permissions: DefaultUserPermissions(),
			},
			serverResponse: &SafeMember{
				MemberName: "newuser",
				MemberType: "User",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name:     "missing safe name",
			safeName: "",
			opts: AddOptions{
				MemberName:  "newuser",
				Permissions: DefaultUserPermissions(),
			},
			wantErr: true,
		},
		{
			name:     "missing member name",
			safeName: "TestSafe",
			opts: AddOptions{
				Permissions: DefaultUserPermissions(),
			},
			wantErr: true,
		},
		{
			name:     "missing permissions",
			safeName: "TestSafe",
			opts: AddOptions{
				MemberName: "newuser",
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

			result, err := Add(context.Background(), sess, tt.safeName, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Add() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Add() unexpected error: %v", err)
				return
			}

			if result.MemberName != tt.serverResponse.MemberName {
				t.Errorf("Add().MemberName = %v, want %v", result.MemberName, tt.serverResponse.MemberName)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		safeName       string
		memberName     string
		opts           UpdateOptions
		serverResponse *SafeMember
		serverStatus   int
		wantErr        bool
	}{
		{
			name:       "successful update",
			safeName:   "TestSafe",
			memberName: "user1",
			opts: UpdateOptions{
				Permissions: DefaultAdminPermissions(),
			},
			serverResponse: &SafeMember{
				MemberName: "user1",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "empty safe name",
			safeName:   "",
			memberName: "user1",
			wantErr:    true,
		},
		{
			name:       "empty member name",
			safeName:   "TestSafe",
			memberName: "",
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

			result, err := Update(context.Background(), sess, tt.safeName, tt.memberName, tt.opts)
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

			if result.MemberName != tt.serverResponse.MemberName {
				t.Errorf("Update().MemberName = %v, want %v", result.MemberName, tt.serverResponse.MemberName)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name         string
		safeName     string
		memberName   string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful remove",
			safeName:     "TestSafe",
			memberName:   "user1",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:       "empty safe name",
			safeName:   "",
			memberName: "user1",
			wantErr:    true,
		},
		{
			name:       "empty member name",
			safeName:   "TestSafe",
			memberName: "",
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

			err := Remove(context.Background(), sess, tt.safeName, tt.memberName)
			if tt.wantErr {
				if err == nil {
					t.Error("Remove() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Remove() unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultUserPermissions(t *testing.T) {
	perms := DefaultUserPermissions()
	if perms == nil {
		t.Fatal("DefaultUserPermissions() returned nil")
	}

	if !perms.UseAccounts {
		t.Error("UseAccounts should be true")
	}
	if !perms.RetrieveAccounts {
		t.Error("RetrieveAccounts should be true")
	}
	if !perms.ListAccounts {
		t.Error("ListAccounts should be true")
	}
	if !perms.ViewSafeMembers {
		t.Error("ViewSafeMembers should be true")
	}
	if perms.ManageSafe {
		t.Error("ManageSafe should be false for user permissions")
	}
}

func TestDefaultAdminPermissions(t *testing.T) {
	perms := DefaultAdminPermissions()
	if perms == nil {
		t.Fatal("DefaultAdminPermissions() returned nil")
	}

	// Check all admin permissions are true
	if !perms.UseAccounts {
		t.Error("UseAccounts should be true")
	}
	if !perms.RetrieveAccounts {
		t.Error("RetrieveAccounts should be true")
	}
	if !perms.ListAccounts {
		t.Error("ListAccounts should be true")
	}
	if !perms.AddAccounts {
		t.Error("AddAccounts should be true")
	}
	if !perms.UpdateAccountContent {
		t.Error("UpdateAccountContent should be true")
	}
	if !perms.DeleteAccounts {
		t.Error("DeleteAccounts should be true")
	}
	if !perms.ManageSafe {
		t.Error("ManageSafe should be true")
	}
	if !perms.ManageSafeMembers {
		t.Error("ManageSafeMembers should be true")
	}
}

func TestPermissions_Struct(t *testing.T) {
	perms := Permissions{
		UseAccounts:                            true,
		RetrieveAccounts:                       true,
		ListAccounts:                           true,
		AddAccounts:                            true,
		UpdateAccountContent:                   true,
		UpdateAccountProperties:                true,
		InitiateCPMAccountManagementOperations: true,
		SpecifyNextAccountContent:              true,
		RenameAccounts:                         true,
		DeleteAccounts:                         true,
		UnlockAccounts:                         true,
		ManageSafe:                             true,
		ManageSafeMembers:                      true,
		BackupSafe:                             true,
		ViewAuditLog:                           true,
		ViewSafeMembers:                        true,
		AccessWithoutConfirmation:              true,
		CreateFolders:                          true,
		DeleteFolders:                          true,
		MoveAccountsAndFolders:                 true,
		RequestsAuthorizationLevel1:            false,
		RequestsAuthorizationLevel2:            false,
	}

	data, err := json.Marshal(perms)
	if err != nil {
		t.Fatalf("Failed to marshal Permissions: %v", err)
	}

	var parsed Permissions
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal Permissions: %v", err)
	}

	if !parsed.UseAccounts {
		t.Error("UseAccounts should be true")
	}
	if !parsed.ManageSafe {
		t.Error("ManageSafe should be true")
	}
}

func TestSafeMember_Struct(t *testing.T) {
	member := SafeMember{
		SafeURLID:                 types.FlexibleID("TestSafe"),
		SafeName:                  "TestSafe",
		SafeNumber:                1,
		MemberID:                  types.FlexibleID("member-123"),
		MemberName:                "admin",
		MemberType:                "User",
		MembershipExpirationDate:  1705315800,
		IsExpiredMembershipEnable: false,
		IsPredefinedUser:          false,
		IsReadOnly:                false,
		Permissions:               DefaultUserPermissions(),
	}

	if member.MemberName != "admin" {
		t.Errorf("MemberName = %v, want admin", member.MemberName)
	}
	if member.MemberType != "User" {
		t.Errorf("MemberType = %v, want User", member.MemberType)
	}
}

func TestFlexibleID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		want     string
		wantErr  bool
	}{
		{
			name:     "unmarshal string UUID",
			jsonData: `"12345678-1234-1234-1234-123456789012"`,
			want:     "12345678-1234-1234-1234-123456789012",
			wantErr:  false,
		},
		{
			name:     "unmarshal string number",
			jsonData: `"12345"`,
			want:     "12345",
			wantErr:  false,
		},
		{
			name:     "unmarshal integer",
			jsonData: `12345`,
			want:     "12345",
			wantErr:  false,
		},
		{
			name:     "unmarshal large integer",
			jsonData: `9876543210`,
			want:     "9876543210",
			wantErr:  false,
		},
		{
			name:     "unmarshal zero",
			jsonData: `0`,
			want:     "0",
			wantErr:  false,
		},
		{
			name:     "unmarshal empty string",
			jsonData: `""`,
			want:     "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f types.FlexibleID
			err := json.Unmarshal([]byte(tt.jsonData), &f)
			if tt.wantErr {
				if err == nil {
					t.Error("UnmarshalJSON() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UnmarshalJSON() unexpected error: %v", err)
				return
			}
			if string(f) != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", string(f), tt.want)
			}
		})
	}
}

func TestFlexibleID_MarshalJSON(t *testing.T) {
	f := types.FlexibleID("12345")
	data, err := json.Marshal(f)
	if err != nil {
		t.Errorf("MarshalJSON() unexpected error: %v", err)
		return
	}
	want := `"12345"`
	if string(data) != want {
		t.Errorf("MarshalJSON() = %v, want %v", string(data), want)
	}
}

func TestFlexibleID_String(t *testing.T) {
	f := types.FlexibleID("test-id")
	if f.String() != "test-id" {
		t.Errorf("String() = %v, want test-id", f.String())
	}
}

func TestSafeMember_UnmarshalWithIntegerMemberID(t *testing.T) {
	// This test verifies that SafeMember correctly handles memberId as an integer
	// which is what the CyberArk API actually returns in some versions
	jsonData := `{
		"safeUrlId": "TestSafe",
		"safeName": "TestSafe",
		"safeNumber": 1,
		"memberId": 12345,
		"memberName": "admin",
		"memberType": "User"
	}`

	var member SafeMember
	err := json.Unmarshal([]byte(jsonData), &member)
	if err != nil {
		t.Fatalf("Failed to unmarshal SafeMember with integer memberId: %v", err)
	}

	if member.MemberID.String() != "12345" {
		t.Errorf("MemberID = %v, want 12345", member.MemberID.String())
	}
	if member.MemberName != "admin" {
		t.Errorf("MemberName = %v, want admin", member.MemberName)
	}
}

func TestSafeMember_UnmarshalWithStringMemberID(t *testing.T) {
	// This test verifies that SafeMember correctly handles memberId as a string UUID
	// which is what the CyberArk documentation shows
	jsonData := `{
		"safeUrlId": "TestSafe",
		"safeName": "TestSafe",
		"safeNumber": 1,
		"memberId": "12345678-1234-1234-1234-123456789012",
		"memberName": "admin",
		"memberType": "User"
	}`

	var member SafeMember
	err := json.Unmarshal([]byte(jsonData), &member)
	if err != nil {
		t.Fatalf("Failed to unmarshal SafeMember with string memberId: %v", err)
	}

	if member.MemberID.String() != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("MemberID = %v, want 12345678-1234-1234-1234-123456789012", member.MemberID.String())
	}
	if member.MemberName != "admin" {
		t.Errorf("MemberName = %v, want admin", member.MemberName)
	}
}

func TestSafeMembersResponse_UnmarshalWithIntegerMemberID(t *testing.T) {
	// This test verifies the full response structure with integer memberId
	jsonData := `{
		"value": [
			{
				"safeUrlId": "TestSafe",
				"safeName": "TestSafe",
				"safeNumber": 1,
				"memberId": 99999,
				"memberName": "admin",
				"memberType": "User"
			}
		],
		"count": 1
	}`

	var response SafeMembersResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal SafeMembersResponse: %v", err)
	}

	if response.Count != 1 {
		t.Errorf("Count = %v, want 1", response.Count)
	}
	if len(response.Value) != 1 {
		t.Fatalf("len(Value) = %v, want 1", len(response.Value))
	}
	if response.Value[0].MemberID.String() != "99999" {
		t.Errorf("MemberID = %v, want 99999", response.Value[0].MemberID.String())
	}
}
