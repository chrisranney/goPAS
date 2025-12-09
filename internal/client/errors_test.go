// Package client provides tests for error handling.
package client

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *APIError
		expected string
	}{
		{
			name: "error with error code",
			apiErr: &APIError{
				StatusCode: 404,
				ErrorCode:  "PASWS001",
				ErrorMsg:   "Account not found",
			},
			expected: "CyberArk API error [404] PASWS001: Account not found",
		},
		{
			name: "error without error code",
			apiErr: &APIError{
				StatusCode: 500,
				ErrorMsg:   "Internal Server Error",
			},
			expected: "CyberArk API error [500]: Internal Server Error",
		},
		{
			name: "error with empty error code",
			apiErr: &APIError{
				StatusCode: 401,
				ErrorCode:  "",
				ErrorMsg:   "Unauthorized",
			},
			expected: "CyberArk API error [401]: Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiErr.Error()
			if result != tt.expected {
				t.Errorf("Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAPIError_StatusChecks(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		isNotFound bool
		isUnauth   bool
		isForbid   bool
		isConflict bool
		isBadReq   bool
	}{
		{
			name:       "404 Not Found",
			statusCode: 404,
			isNotFound: true,
		},
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			isUnauth:   true,
		},
		{
			name:       "403 Forbidden",
			statusCode: 403,
			isForbid:   true,
		},
		{
			name:       "409 Conflict",
			statusCode: 409,
			isConflict: true,
		},
		{
			name:       "400 Bad Request",
			statusCode: 400,
			isBadReq:   true,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
		},
		{
			name:       "200 OK",
			statusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &APIError{StatusCode: tt.statusCode}

			if apiErr.IsNotFound() != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", apiErr.IsNotFound(), tt.isNotFound)
			}
			if apiErr.IsUnauthorized() != tt.isUnauth {
				t.Errorf("IsUnauthorized() = %v, want %v", apiErr.IsUnauthorized(), tt.isUnauth)
			}
			if apiErr.IsForbidden() != tt.isForbid {
				t.Errorf("IsForbidden() = %v, want %v", apiErr.IsForbidden(), tt.isForbid)
			}
			if apiErr.IsConflict() != tt.isConflict {
				t.Errorf("IsConflict() = %v, want %v", apiErr.IsConflict(), tt.isConflict)
			}
			if apiErr.IsBadRequest() != tt.isBadReq {
				t.Errorf("IsBadRequest() = %v, want %v", apiErr.IsBadRequest(), tt.isBadReq)
			}
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name           string
		response       *Response
		expectedCode   string
		expectedMsg    string
		expectedStatus int
	}{
		{
			name: "standard error format",
			response: &Response{
				StatusCode: 404,
				Body:       []byte(`{"ErrorCode": "PASWS001", "ErrorMessage": "Account not found"}`),
			},
			expectedCode:   "PASWS001",
			expectedMsg:    "Account not found",
			expectedStatus: 404,
		},
		{
			name: "error with details",
			response: &Response{
				StatusCode: 400,
				Body:       []byte(`{"ErrorCode": "PASWS002", "ErrorMessage": "Invalid request", "Details": "Missing required field"}`),
			},
			expectedCode:   "PASWS002",
			expectedMsg:    "Invalid request",
			expectedStatus: 400,
		},
		{
			name: "plain text error body",
			response: &Response{
				StatusCode: 500,
				Body:       []byte(`Internal Server Error occurred`),
			},
			expectedCode:   "",
			expectedMsg:    "Internal Server Error occurred",
			expectedStatus: 500,
		},
		{
			name: "empty body - 400",
			response: &Response{
				StatusCode: 400,
				Body:       []byte{},
			},
			expectedMsg:    "Bad Request",
			expectedStatus: 400,
		},
		{
			name: "empty body - 401",
			response: &Response{
				StatusCode: 401,
				Body:       []byte{},
			},
			expectedMsg:    "Unauthorized",
			expectedStatus: 401,
		},
		{
			name: "empty body - 403",
			response: &Response{
				StatusCode: 403,
				Body:       []byte{},
			},
			expectedMsg:    "Forbidden",
			expectedStatus: 403,
		},
		{
			name: "empty body - 404",
			response: &Response{
				StatusCode: 404,
				Body:       []byte{},
			},
			expectedMsg:    "Not Found",
			expectedStatus: 404,
		},
		{
			name: "empty body - 409",
			response: &Response{
				StatusCode: 409,
				Body:       []byte{},
			},
			expectedMsg:    "Conflict",
			expectedStatus: 409,
		},
		{
			name: "empty body - 500",
			response: &Response{
				StatusCode: 500,
				Body:       []byte{},
			},
			expectedMsg:    "Internal Server Error",
			expectedStatus: 500,
		},
		{
			name: "empty body - unknown status",
			response: &Response{
				StatusCode: 502,
				Body:       []byte{},
			},
			expectedMsg:    "HTTP Error 502",
			expectedStatus: 502,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(tt.response)
			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("parseAPIError() did not return *APIError")
			}

			if apiErr.StatusCode != tt.expectedStatus {
				t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, tt.expectedStatus)
			}
			if apiErr.ErrorCode != tt.expectedCode {
				t.Errorf("ErrorCode = %v, want %v", apiErr.ErrorCode, tt.expectedCode)
			}
			if apiErr.ErrorMsg != tt.expectedMsg {
				t.Errorf("ErrorMsg = %v, want %v", apiErr.ErrorMsg, tt.expectedMsg)
			}
		})
	}
}

func TestIsAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError type",
			err:      &APIError{StatusCode: 404, ErrorMsg: "Not found"},
			expected: true,
		},
		{
			name:     "other error type",
			err:      &testError{msg: "test error"},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAPIError(tt.err)
			if result != tt.expected {
				t.Errorf("IsAPIError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAsAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantOk   bool
		wantCode int
	}{
		{
			name:     "APIError type",
			err:      &APIError{StatusCode: 404, ErrorMsg: "Not found"},
			wantOk:   true,
			wantCode: 404,
		},
		{
			name:   "other error type",
			err:    &testError{msg: "test error"},
			wantOk: false,
		},
		{
			name:   "nil error",
			err:    nil,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr, ok := AsAPIError(tt.err)
			if ok != tt.wantOk {
				t.Errorf("AsAPIError() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && apiErr.StatusCode != tt.wantCode {
				t.Errorf("AsAPIError() StatusCode = %v, want %v", apiErr.StatusCode, tt.wantCode)
			}
		})
	}
}

// testError is a helper error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
