// Package requests provides tests for access request functionality.
package requests

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

func TestListIncoming(t *testing.T) {
	tests := []struct {
		name           string
		opts           ListOptions
		serverResponse *RequestsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListOptions{},
			serverResponse: &RequestsResponse{
				Requests: []Request{
					{RequestID: "1", SafeName: "Safe1", RequestorUserName: "user1"},
					{RequestID: "2", SafeName: "Safe2", RequestorUserName: "user2"},
				},
				Total: 2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "list only waiting",
			opts: ListOptions{OnlyWaiting: true},
			serverResponse: &RequestsResponse{
				Requests: []Request{
					{RequestID: "1", SafeName: "Safe1", Status: 1},
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

			result, err := ListIncoming(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ListIncoming() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListIncoming() unexpected error: %v", err)
				return
			}

			if result.Total != tt.serverResponse.Total {
				t.Errorf("ListIncoming().Total = %v, want %v", result.Total, tt.serverResponse.Total)
			}
		})
	}
}

func TestListMyRequests(t *testing.T) {
	tests := []struct {
		name           string
		opts           ListOptions
		serverResponse *RequestsResponse
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful list",
			opts: ListOptions{},
			serverResponse: &RequestsResponse{
				Requests: []Request{
					{RequestID: "1", SafeName: "Safe1"},
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

			result, err := ListMyRequests(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ListMyRequests() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListMyRequests() unexpected error: %v", err)
				return
			}

			if result.Total != tt.serverResponse.Total {
				t.Errorf("ListMyRequests().Total = %v, want %v", result.Total, tt.serverResponse.Total)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		opts           CreateOptions
		serverResponse *Request
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful create",
			opts: CreateOptions{
				AccountID: "123",
				Reason:    "Maintenance",
			},
			serverResponse: &Request{
				RequestID: "new-123",
				SafeName:  "Safe1",
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing account ID",
			opts: CreateOptions{
				Reason: "Maintenance",
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

			if result.RequestID != tt.serverResponse.RequestID {
				t.Errorf("Create().RequestID = %v, want %v", result.RequestID, tt.serverResponse.RequestID)
			}
		})
	}
}

func TestApprove(t *testing.T) {
	tests := []struct {
		name           string
		requestID      string
		opts           ApproveOptions
		serverResponse *Request
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful approve",
			requestID: "123",
			opts: ApproveOptions{
				Reason: "Approved",
			},
			serverResponse: &Request{
				RequestID: "123",
				Status:    2,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty request ID",
			requestID: "",
			opts:      ApproveOptions{},
			wantErr:   true,
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

			result, err := Approve(context.Background(), sess, tt.requestID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Approve() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Approve() unexpected error: %v", err)
				return
			}

			if result.RequestID != tt.serverResponse.RequestID {
				t.Errorf("Approve().RequestID = %v, want %v", result.RequestID, tt.serverResponse.RequestID)
			}
		})
	}
}

func TestDeny(t *testing.T) {
	tests := []struct {
		name           string
		requestID      string
		opts           DenyOptions
		serverResponse *Request
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "successful deny",
			requestID: "123",
			opts: DenyOptions{
				Reason: "Not approved",
			},
			serverResponse: &Request{
				RequestID: "123",
				Status:    3,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:      "empty request ID",
			requestID: "",
			opts:      DenyOptions{},
			wantErr:   true,
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

			result, err := Deny(context.Background(), sess, tt.requestID, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Deny() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Deny() unexpected error: %v", err)
				return
			}

			if result.RequestID != tt.serverResponse.RequestID {
				t.Errorf("Deny().RequestID = %v, want %v", result.RequestID, tt.serverResponse.RequestID)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		requestID    string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			requestID:    "123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:      "empty request ID",
			requestID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := Delete(context.Background(), sess, tt.requestID)
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

func TestRequest_Struct(t *testing.T) {
	req := Request{
		RequestID:                "123",
		SafeName:                 "Safe1",
		RequestorUserName:        "user1",
		RequestorReason:          "Maintenance",
		CreationDate:             1705315800,
		Operation:                "Retrieve",
		OperationType:            1,
		ConfirmationsLeft:        1,
		Status:                   1,
		StatusTitle:              "Pending",
		CurrentConfirmationLevel: 1,
		RequiredConfirmers:       1,
	}

	if req.RequestID != "123" {
		t.Errorf("RequestID = %v, want 123", req.RequestID)
	}
	if req.Status != 1 {
		t.Errorf("Status = %v, want 1", req.Status)
	}
}

func TestAccountDetails_Struct(t *testing.T) {
	details := AccountDetails{
		AccountID:   "acc-123",
		AccountName: "admin@server",
		SafeName:    "Safe1",
		PlatformID:  "WinServerLocal",
		Address:     "server.example.com",
	}

	if details.AccountID != "acc-123" {
		t.Errorf("AccountID = %v, want acc-123", details.AccountID)
	}
}

// Tests for nil session and additional edge cases

func TestListIncoming_InvalidSession(t *testing.T) {
	_, err := ListIncoming(context.Background(), nil, ListOptions{})
	if err == nil {
		t.Error("ListIncoming() expected error for nil session")
	}
}

func TestListIncoming_AllOptions(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		q := r.URL.Query()
		if q.Get("onlyWaiting") != "true" {
			t.Errorf("Expected onlyWaiting=true, got %s", q.Get("onlyWaiting"))
		}
		if q.Get("expired") != "true" {
			t.Errorf("Expected expired=true, got %s", q.Get("expired"))
		}
		if q.Get("offset") != "10" {
			t.Errorf("Expected offset=10, got %s", q.Get("offset"))
		}
		if q.Get("limit") != "50" {
			t.Errorf("Expected limit=50, got %s", q.Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RequestsResponse{Requests: []Request{}, Total: 0})
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	opts := ListOptions{
		OnlyWaiting: true,
		Expired:     true,
		Offset:      10,
		Limit:       50,
	}
	_, err := ListIncoming(context.Background(), sess, opts)
	if err != nil {
		t.Errorf("ListIncoming() unexpected error: %v", err)
	}
}

func TestListMyRequests_InvalidSession(t *testing.T) {
	_, err := ListMyRequests(context.Background(), nil, ListOptions{})
	if err == nil {
		t.Error("ListMyRequests() expected error for nil session")
	}
}

func TestListMyRequests_AllOptions(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("onlyWaiting") != "true" {
			t.Errorf("Expected onlyWaiting=true, got %s", q.Get("onlyWaiting"))
		}
		if q.Get("expired") != "true" {
			t.Errorf("Expected expired=true, got %s", q.Get("expired"))
		}
		if q.Get("offset") != "5" {
			t.Errorf("Expected offset=5, got %s", q.Get("offset"))
		}
		if q.Get("limit") != "25" {
			t.Errorf("Expected limit=25, got %s", q.Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RequestsResponse{Requests: []Request{}, Total: 0})
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	opts := ListOptions{
		OnlyWaiting: true,
		Expired:     true,
		Offset:      5,
		Limit:       25,
	}
	_, err := ListMyRequests(context.Background(), sess, opts)
	if err != nil {
		t.Errorf("ListMyRequests() unexpected error: %v", err)
	}
}

func TestCreate_InvalidSession(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateOptions{AccountID: "123"})
	if err == nil {
		t.Error("Create() expected error for nil session")
	}
}

func TestCreate_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := Create(context.Background(), sess, CreateOptions{AccountID: "123"})
	if err == nil {
		t.Error("Create() expected error for server error")
	}
}

func TestApprove_InvalidSession(t *testing.T) {
	_, err := Approve(context.Background(), nil, "123", ApproveOptions{})
	if err == nil {
		t.Error("Approve() expected error for nil session")
	}
}

func TestApprove_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := Approve(context.Background(), sess, "123", ApproveOptions{})
	if err == nil {
		t.Error("Approve() expected error for server error")
	}
}

func TestDeny_InvalidSession(t *testing.T) {
	_, err := Deny(context.Background(), nil, "123", DenyOptions{})
	if err == nil {
		t.Error("Deny() expected error for nil session")
	}
}

func TestDeny_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	_, err := Deny(context.Background(), sess, "123", DenyOptions{})
	if err == nil {
		t.Error("Deny() expected error for server error")
	}
}

func TestDelete_InvalidSession(t *testing.T) {
	err := Delete(context.Background(), nil, "123")
	if err == nil {
		t.Error("Delete() expected error for nil session")
	}
}

func TestDelete_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	sess, server := createTestSession(t, handler)
	defer server.Close()

	err := Delete(context.Background(), sess, "123")
	if err == nil {
		t.Error("Delete() expected error for server error")
	}
}

func TestCreateOptions_Struct(t *testing.T) {
	opts := CreateOptions{
		AccountID:              "acc-123",
		Reason:                 "Maintenance",
		TicketingSystemName:    "ServiceNow",
		TicketID:               "INC001",
		MultipleAccessRequired: true,
		FromDate:               1705315800,
		ToDate:                 1705920600,
		AdditionalInfo:         map[string]string{"key": "value"},
		UseConnect:             true,
		ConnectionComponent:    "PSM-RDP",
		ConnectionParams:       map[string]string{"param": "value"},
	}

	if opts.AccountID != "acc-123" {
		t.Errorf("AccountID = %v, want acc-123", opts.AccountID)
	}
	if opts.TicketingSystemName != "ServiceNow" {
		t.Errorf("TicketingSystemName = %v, want ServiceNow", opts.TicketingSystemName)
	}
}

func TestListOptions_Struct(t *testing.T) {
	opts := ListOptions{
		RequestorUserName: "admin",
		SafeName:          "Safe1",
		OnlyWaiting:       true,
		Expired:           false,
		Offset:            10,
		Limit:             50,
	}

	if opts.RequestorUserName != "admin" {
		t.Errorf("RequestorUserName = %v, want admin", opts.RequestorUserName)
	}
	if opts.Limit != 50 {
		t.Errorf("Limit = %v, want 50", opts.Limit)
	}
}
