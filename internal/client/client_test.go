// Package client provides tests for the HTTP client.
package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty base URL",
			cfg:     Config{},
			wantErr: true,
			errMsg:  "baseURL is required",
		},
		{
			name: "valid base URL",
			cfg: Config{
				BaseURL: "https://cyberark.example.com",
			},
			wantErr: false,
		},
		{
			name: "base URL with trailing slash",
			cfg: Config{
				BaseURL: "https://cyberark.example.com/",
			},
			wantErr: false,
		},
		{
			name: "custom timeout",
			cfg: Config{
				BaseURL: "https://cyberark.example.com",
				Timeout: 60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "custom HTTP client",
			cfg: Config{
				BaseURL:          "https://cyberark.example.com",
				CustomHTTPClient: &http.Client{Timeout: 10 * time.Second},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
				}
				if err != nil && err.Error() != tt.errMsg {
					t.Errorf("NewClient() error = %v, want %v", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("NewClient() returned nil client")
				return
			}

			// Verify trailing slash is removed
			if client.GetBaseURL() == "https://cyberark.example.com/" {
				t.Error("NewClient() did not remove trailing slash")
			}

			// Verify API URL is constructed correctly
			expectedAPIURL := "https://cyberark.example.com/PasswordVault/API"
			if client.GetAPIURL() != expectedAPIURL {
				t.Errorf("GetAPIURL() = %v, want %v", client.GetAPIURL(), expectedAPIURL)
			}
		})
	}
}

func TestClient_SetAuthToken(t *testing.T) {
	client, err := NewClient(Config{BaseURL: "https://cyberark.example.com"})
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	// Initially empty
	if client.GetAuthToken() != "" {
		t.Error("GetAuthToken() should initially be empty")
	}

	// Set token
	token := "test-auth-token"
	client.SetAuthToken(token)

	if client.GetAuthToken() != token {
		t.Errorf("GetAuthToken() = %v, want %v", client.GetAuthToken(), token)
	}

	// Update token
	newToken := "new-auth-token"
	client.SetAuthToken(newToken)

	if client.GetAuthToken() != newToken {
		t.Errorf("GetAuthToken() = %v, want %v", client.GetAuthToken(), newToken)
	}
}

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		queryParams    url.Values
		headers        map[string]string
		serverResponse string
		serverStatus   int
		wantErr        bool
	}{
		{
			name:           "successful GET request",
			method:         http.MethodGet,
			path:           "/test",
			serverResponse: `{"message": "success"}`,
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "GET request with query params",
			method:         http.MethodGet,
			path:           "/test",
			queryParams:    url.Values{"search": {"test"}, "limit": {"10"}},
			serverResponse: `{"message": "success"}`,
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "successful POST request with body",
			method:         http.MethodPost,
			path:           "/test",
			body:           map[string]string{"key": "value"},
			serverResponse: `{"id": "123"}`,
			serverStatus:   http.StatusCreated,
			wantErr:        false,
		},
		{
			name:           "request with custom headers",
			method:         http.MethodGet,
			path:           "/test",
			headers:        map[string]string{"X-Custom-Header": "custom-value"},
			serverResponse: `{"message": "success"}`,
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "404 error response",
			method:         http.MethodGet,
			path:           "/notfound",
			serverResponse: `{"ErrorCode": "PASWS001", "ErrorMessage": "Not found"}`,
			serverStatus:   http.StatusNotFound,
			wantErr:        true,
		},
		{
			name:           "401 unauthorized",
			method:         http.MethodGet,
			path:           "/unauthorized",
			serverResponse: `{"ErrorCode": "PASWS002", "ErrorMessage": "Unauthorized"}`,
			serverStatus:   http.StatusUnauthorized,
			wantErr:        true,
		},
		{
			name:           "500 server error",
			method:         http.MethodGet,
			path:           "/error",
			serverResponse: `{"ErrorCode": "PASWS003", "ErrorMessage": "Internal Server Error"}`,
			serverStatus:   http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify custom headers
				for key, value := range tt.headers {
					if r.Header.Get(key) != value {
						t.Errorf("Header %s = %v, want %v", key, r.Header.Get(key), value)
					}
				}

				// Verify Content-Type
				if r.Method == http.MethodPost || r.Method == http.MethodPut {
					if r.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Content-Type = %v, want application/json", r.Header.Get("Content-Type"))
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create client pointing to test server
			// We need to strip the /PasswordVault/API suffix that gets added
			client, err := NewClient(Config{BaseURL: server.URL})
			if err != nil {
				t.Fatalf("NewClient() error: %v", err)
			}

			// Override apiURL for testing
			client.apiURL = server.URL

			ctx := context.Background()
			resp, err := client.Do(ctx, Request{
				Method:      tt.method,
				Path:        tt.path,
				Body:        tt.body,
				QueryParams: tt.queryParams,
				Headers:     tt.headers,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Do() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Do() unexpected error: %v", err)
				return
			}

			if resp.StatusCode != tt.serverStatus {
				t.Errorf("Response.StatusCode = %v, want %v", resp.StatusCode, tt.serverStatus)
			}

			if string(resp.Body) != tt.serverResponse {
				t.Errorf("Response.Body = %v, want %v", string(resp.Body), tt.serverResponse)
			}
		})
	}
}

func TestClient_DoWithAuthToken(t *testing.T) {
	token := "test-auth-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != token {
			t.Errorf("Authorization header = %v, want %v", authHeader, token)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	client.apiURL = server.URL
	client.SetAuthToken(token)

	ctx := context.Background()
	_, err = client.Get(ctx, "/test", nil)
	if err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Query().Get("search") != "test" {
			t.Errorf("Query param search = %v, want test", r.URL.Query().Get("search"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": "value"}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})
	client.apiURL = server.URL

	ctx := context.Background()
	params := url.Values{"search": {"test"}}
	resp, err := client.Get(ctx, "/test", params)
	if err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Get() StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %v, want POST", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]string
		json.Unmarshal(body, &data)

		if data["key"] != "value" {
			t.Errorf("Body key = %v, want value", data["key"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})
	client.apiURL = server.URL

	ctx := context.Background()
	resp, err := client.Post(ctx, "/test", map[string]string{"key": "value"})
	if err != nil {
		t.Errorf("Post() unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Post() StatusCode = %v, want %v", resp.StatusCode, http.StatusCreated)
	}
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Method = %v, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"updated": true}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})
	client.apiURL = server.URL

	ctx := context.Background()
	resp, err := client.Put(ctx, "/test", map[string]string{"key": "newvalue"})
	if err != nil {
		t.Errorf("Put() unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Put() StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Method = %v, want PATCH", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"patched": true}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})
	client.apiURL = server.URL

	ctx := context.Background()
	resp, err := client.Patch(ctx, "/test", map[string]string{"field": "value"})
	if err != nil {
		t.Errorf("Patch() unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Patch() StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})
	client.apiURL = server.URL

	ctx := context.Background()
	resp, err := client.Delete(ctx, "/test/123")
	if err != nil {
		t.Errorf("Delete() unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Delete() StatusCode = %v, want %v", resp.StatusCode, http.StatusNoContent)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})
	client.apiURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "/test", nil)
	if err == nil {
		t.Error("Get() expected error for cancelled context")
	}
}

func TestClient_InvalidBodyMarshal(t *testing.T) {
	client, _ := NewClient(Config{BaseURL: "https://cyberark.example.com"})

	ctx := context.Background()
	// Create a body that cannot be marshaled (channel type)
	invalidBody := make(chan int)

	_, err := client.Post(ctx, "/test", invalidBody)
	if err == nil {
		t.Error("Post() expected error for invalid body marshal")
	}
}
