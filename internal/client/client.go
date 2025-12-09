// Package client provides the HTTP client for CyberArk REST API communication.
// This is equivalent to the Invoke-PASRestMethod private function in psPAS.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents an HTTP client for CyberArk API communication.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	apiURL      string
	authToken   string
	contentType string
	timeout     time.Duration
}

// Config holds the client configuration options.
type Config struct {
	BaseURL          string
	Timeout          time.Duration
	SkipTLSVerify    bool
	CustomHTTPClient *http.Client
}

// NewClient creates a new HTTP client for CyberArk API communication.
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	// Ensure baseURL doesn't have trailing slash
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	httpClient := cfg.CustomHTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &Client{
		httpClient:  httpClient,
		baseURL:     cfg.BaseURL,
		apiURL:      cfg.BaseURL + "/PasswordVault/API",
		contentType: "application/json",
		timeout:     timeout,
	}, nil
}

// SetAuthToken sets the authentication token for subsequent requests.
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// GetAuthToken returns the current authentication token.
func (c *Client) GetAuthToken() string {
	return c.authToken
}

// GetBaseURL returns the base URL.
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetAPIURL returns the API URL.
func (c *Client) GetAPIURL() string {
	return c.apiURL
}

// Request represents an API request.
type Request struct {
	Method      string
	Path        string
	Body        interface{}
	QueryParams url.Values
	Headers     map[string]string
}

// Response represents an API response.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Do executes an HTTP request to the CyberArk API.
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	// Build the full URL
	fullURL := c.apiURL + req.Path
	if len(req.QueryParams) > 0 {
		fullURL += "?" + req.QueryParams.Encode()
	}

	// Serialize body if present
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	httpReq.Header.Set("Content-Type", c.contentType)
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", c.authToken)
	}

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute the request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Body:       respBody,
		Headers:    httpResp.Header,
	}

	// Check for error responses
	if httpResp.StatusCode >= 400 {
		return resp, parseAPIError(resp)
	}

	return resp, nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, queryParams url.Values) (*Response, error) {
	return c.Do(ctx, Request{
		Method:      http.MethodGet,
		Path:        path,
		QueryParams: queryParams,
	})
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, Request{
		Method: http.MethodPut,
		Path:   path,
		Body:   body,
	})
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, Request{
		Method: http.MethodPatch,
		Path:   path,
		Body:   body,
	})
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, Request{
		Method: http.MethodDelete,
		Path:   path,
	})
}
