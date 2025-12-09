// Package users provides tests for user management functionality.
package users

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
		serverResponse *UsersResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListOptions{},
			serverResponse: &UsersResponse{
				Users: []User{
					{ID: 1, Username: "admin", UserType: "EPVUser"},
					{ID: 2, Username: "user1", UserType: "EPVUser"},
				},
				Total: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with search",
			opts: ListOptions{Search: "admin"},
			serverResponse: &UsersResponse{
				Users: []User{
					{ID: 1, Username: "admin", UserType: "EPVUser"},
				},
				Total: 1,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list with user type filter",
			opts: ListOptions{UserType: "EPVUser"},
			serverResponse: &UsersResponse{
				Users: []User{},
				Total: 0,
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

			if result.Total != tt.serverResponse.Total {
				t.Errorf("List().Total = %v, want %v", result.Total, tt.serverResponse.Total)
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

func TestList_AllOptions(t *testing.T) {
	componentUser := true
	tests := []struct {
		name        string
		opts        ListOptions
		checkParams func(t *testing.T, params map[string]string)
	}{
		{
			name: "with sort option",
			opts: ListOptions{Sort: "username"},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["sort"] != "username" {
					t.Errorf("sort param = %v, want username", params["sort"])
				}
			},
		},
		{
			name: "with offset option",
			opts: ListOptions{Offset: 10},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["offset"] != "10" {
					t.Errorf("offset param = %v, want 10", params["offset"])
				}
			},
		},
		{
			name: "with limit option",
			opts: ListOptions{Limit: 25},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["limit"] != "25" {
					t.Errorf("limit param = %v, want 25", params["limit"])
				}
			},
		},
		{
			name: "with filter option",
			opts: ListOptions{Filter: "userType eq EPVUser"},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["filter"] != "userType eq EPVUser" {
					t.Errorf("filter param = %v, want userType eq EPVUser", params["filter"])
				}
			},
		},
		{
			name: "with componentUser option",
			opts: ListOptions{ComponentUser: &componentUser},
			checkParams: func(t *testing.T, params map[string]string) {
				if params["componentUser"] != "true" {
					t.Errorf("componentUser param = %v, want true", params["componentUser"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedParams map[string]string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedParams = make(map[string]string)
				for key, values := range r.URL.Query() {
					if len(values) > 0 {
						capturedParams[key] = values[0]
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&UsersResponse{Users: []User{}, Total: 0})
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			_, err := List(context.Background(), sess, tt.opts)
			if err != nil {
				t.Errorf("List() unexpected error: %v", err)
				return
			}

			tt.checkParams(t, capturedParams)
		})
	}
}

func TestGet_InvalidSession(t *testing.T) {
	_, err := Get(context.Background(), nil, 1)
	if err == nil {
		t.Error("Get() expected error for nil session, got nil")
	}
}

func TestCreate_InvalidSession(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateOptions{Username: "test"})
	if err == nil {
		t.Error("Create() expected error for nil session, got nil")
	}
}

func TestUpdate_InvalidSession(t *testing.T) {
	_, err := Update(context.Background(), nil, 1, UpdateOptions{})
	if err == nil {
		t.Error("Update() expected error for nil session, got nil")
	}
}

func TestDelete_InvalidSession(t *testing.T) {
	err := Delete(context.Background(), nil, 1)
	if err == nil {
		t.Error("Delete() expected error for nil session, got nil")
	}
}

func TestActivateUser_InvalidSession(t *testing.T) {
	_, err := ActivateUser(context.Background(), nil, 1)
	if err == nil {
		t.Error("ActivateUser() expected error for nil session, got nil")
	}
}

func TestResetPassword_InvalidSession(t *testing.T) {
	err := ResetPassword(context.Background(), nil, 1, "newpass")
	if err == nil {
		t.Error("ResetPassword() expected error for nil session, got nil")
	}
}

func TestGet_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := Get(context.Background(), sess, 1)
	if err == nil {
		t.Error("Get() expected error for server error, got nil")
	}
}

func TestCreate_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := Create(context.Background(), sess, CreateOptions{
		Username:        "newuser",
		InitialPassword: "pass123",
	})
	if err == nil {
		t.Error("Create() expected error for server error, got nil")
	}
}

func TestUpdate_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := Update(context.Background(), sess, 1, UpdateOptions{Description: "test"})
	if err == nil {
		t.Error("Update() expected error for server error, got nil")
	}
}

func TestDelete_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	err := Delete(context.Background(), sess, 1)
	if err == nil {
		t.Error("Delete() expected error for server error, got nil")
	}
}

func TestActivateUser_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := ActivateUser(context.Background(), sess, 1)
	if err == nil {
		t.Error("ActivateUser() expected error for server error, got nil")
	}
}

func TestResetPassword_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	err := ResetPassword(context.Background(), sess, 1, "newpass")
	if err == nil {
		t.Error("ResetPassword() expected error for server error, got nil")
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		serverResponse *User
		serverStatus   int
		wantErr        bool
	}{
		{
			name:   "successful get",
			userID: 1,
			serverResponse: &User{
				ID:         1,
				Username:   "admin",
				UserType:   "EPVUser",
				EnableUser: true,
				Source:     "CyberArk",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "user not found",
			userID:       9999,
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

			result, err := Get(context.Background(), sess, tt.userID)
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

			if result.ID != tt.serverResponse.ID {
				t.Errorf("Get().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
			if result.Username != tt.serverResponse.Username {
				t.Errorf("Get().Username = %v, want %v", result.Username, tt.serverResponse.Username)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		opts           CreateOptions
		serverResponse *User
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful create",
			opts: CreateOptions{
				Username:        "newuser",
				InitialPassword: "Password123!",
				UserType:        "EPVUser",
			},
			serverResponse: &User{
				ID:         10,
				Username:   "newuser",
				UserType:   "EPVUser",
				EnableUser: true,
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing username",
			opts: CreateOptions{
				InitialPassword: "Password123!",
			},
			wantErr: true,
		},
		{
			name: "create with personal details",
			opts: CreateOptions{
				Username: "newuser",
				PersonalDetails: &PersonalDetails{
					FirstName: "John",
					LastName:  "Doe",
				},
			},
			serverResponse: &User{
				ID:       11,
				Username: "newuser",
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

			if result.ID != tt.serverResponse.ID {
				t.Errorf("Create().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		opts           UpdateOptions
		serverResponse *User
		serverStatus   int
		wantErr        bool
	}{
		{
			name:   "successful update",
			userID: 1,
			opts: UpdateOptions{
				Description: "Updated user",
			},
			serverResponse: &User{
				ID:          1,
				Username:    "admin",
				Description: "Updated user",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:   "update with enable user",
			userID: 1,
			opts: UpdateOptions{
				EnableUser: boolPtr(true),
			},
			serverResponse: &User{
				ID:         1,
				EnableUser: true,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
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

			result, err := Update(context.Background(), sess, tt.userID, tt.opts)
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

			if result.ID != tt.serverResponse.ID {
				t.Errorf("Update().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			userID:       1,
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "user not found",
			userID:       9999,
			serverStatus: http.StatusNotFound,
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

			err := Delete(context.Background(), sess, tt.userID)
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

func TestActivateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		serverResponse *User
		serverStatus   int
		wantErr        bool
	}{
		{
			name:   "successful activate",
			userID: 1,
			serverResponse: &User{
				ID:        1,
				Username:  "admin",
				Suspended: false,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "user not found",
			userID:       9999,
			serverStatus: http.StatusNotFound,
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
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ActivateUser(context.Background(), sess, tt.userID)
			if tt.wantErr {
				if err == nil {
					t.Error("ActivateUser() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ActivateUser() unexpected error: %v", err)
				return
			}

			if result.Suspended {
				t.Error("ActivateUser() should return unsuspended user")
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		newPassword  string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful reset",
			userID:       1,
			newPassword:  "NewPassword123!",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:        "empty password",
			userID:      1,
			newPassword: "",
			wantErr:     true,
		},
		{
			name:         "user not found",
			userID:       9999,
			newPassword:  "NewPassword123!",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := ResetPassword(context.Background(), sess, tt.userID, tt.newPassword)
			if tt.wantErr {
				if err == nil {
					t.Error("ResetPassword() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ResetPassword() unexpected error: %v", err)
			}
		})
	}
}

func TestUser_Structs(t *testing.T) {
	// Test User struct
	user := User{
		ID:       1,
		Username: "admin",
		Source:   "CyberArk",
		UserType: "EPVUser",
		PersonalDetails: &PersonalDetails{
			FirstName: "John",
			LastName:  "Doe",
		},
		GroupsMembership: []GroupMembership{
			{GroupID: 1, GroupName: "Vault Admins"},
		},
	}

	if user.ID != 1 {
		t.Errorf("User.ID = %v, want 1", user.ID)
	}
	if user.PersonalDetails.FirstName != "John" {
		t.Errorf("User.PersonalDetails.FirstName = %v, want John", user.PersonalDetails.FirstName)
	}
	if len(user.GroupsMembership) != 1 {
		t.Errorf("User.GroupsMembership length = %v, want 1", len(user.GroupsMembership))
	}
}

func TestPersonalDetails_Struct(t *testing.T) {
	pd := PersonalDetails{
		FirstName:    "John",
		MiddleName:   "Q",
		LastName:     "Doe",
		Street:       "123 Main St",
		City:         "Boston",
		State:        "MA",
		Zip:          "02101",
		Country:      "USA",
		Title:        "Mr",
		Organization: "Acme Corp",
		Department:   "IT",
		Profession:   "Engineer",
	}

	if pd.FirstName != "John" {
		t.Errorf("FirstName = %v, want John", pd.FirstName)
	}
	if pd.Organization != "Acme Corp" {
		t.Errorf("Organization = %v, want Acme Corp", pd.Organization)
	}
}

func TestAddress_Struct(t *testing.T) {
	addr := Address{
		WorkStreet:  "123 Business Ave",
		WorkCity:    "Boston",
		WorkState:   "MA",
		WorkZip:     "02101",
		WorkCountry: "USA",
	}

	if addr.WorkCity != "Boston" {
		t.Errorf("WorkCity = %v, want Boston", addr.WorkCity)
	}
}

func TestInternet_Struct(t *testing.T) {
	internet := Internet{
		HomePage:      "https://example.com",
		HomeEmail:     "home@example.com",
		BusinessEmail: "work@example.com",
		OtherEmail:    "other@example.com",
	}

	if internet.BusinessEmail != "work@example.com" {
		t.Errorf("BusinessEmail = %v, want work@example.com", internet.BusinessEmail)
	}
}

func TestPhones_Struct(t *testing.T) {
	phones := Phones{
		HomeNumber:     "555-1234",
		BusinessNumber: "555-5678",
		CellularNumber: "555-9012",
		FaxNumber:      "555-3456",
		PagerNumber:    "555-7890",
	}

	if phones.CellularNumber != "555-9012" {
		t.Errorf("CellularNumber = %v, want 555-9012", phones.CellularNumber)
	}
}

func TestGroupMembership_Struct(t *testing.T) {
	gm := GroupMembership{
		GroupID:   1,
		GroupName: "Vault Admins",
		GroupType: "Vault",
	}

	if gm.GroupID != 1 {
		t.Errorf("GroupID = %v, want 1", gm.GroupID)
	}
	if gm.GroupName != "Vault Admins" {
		t.Errorf("GroupName = %v, want Vault Admins", gm.GroupName)
	}
}

// boolPtr returns a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
