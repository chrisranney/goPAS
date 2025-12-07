// Package client provides error types for CyberArk API errors.
package client

import (
	"encoding/json"
	"fmt"
)

// APIError represents a CyberArk API error response.
type APIError struct {
	StatusCode int    `json:"-"`
	ErrorCode  string `json:"ErrorCode"`
	ErrorMsg   string `json:"ErrorMessage"`
	Details    string `json:"Details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("CyberArk API error [%d] %s: %s", e.StatusCode, e.ErrorCode, e.ErrorMsg)
	}
	return fmt.Sprintf("CyberArk API error [%d]: %s", e.StatusCode, e.ErrorMsg)
}

// IsNotFound returns true if the error is a 404 Not Found error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsForbidden returns true if the error is a 403 Forbidden error.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403
}

// IsConflict returns true if the error is a 409 Conflict error.
func (e *APIError) IsConflict() bool {
	return e.StatusCode == 409
}

// IsBadRequest returns true if the error is a 400 Bad Request error.
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == 400
}

// parseAPIError parses the API error response.
func parseAPIError(resp *Response) error {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
	}

	// Try to parse the error response body
	if len(resp.Body) > 0 {
		// Try standard error format first
		if err := json.Unmarshal(resp.Body, apiErr); err != nil {
			// If parsing fails, use the raw body as the error message
			apiErr.ErrorMsg = string(resp.Body)
		}
	}

	// Set default message if empty
	if apiErr.ErrorMsg == "" {
		switch resp.StatusCode {
		case 400:
			apiErr.ErrorMsg = "Bad Request"
		case 401:
			apiErr.ErrorMsg = "Unauthorized"
		case 403:
			apiErr.ErrorMsg = "Forbidden"
		case 404:
			apiErr.ErrorMsg = "Not Found"
		case 409:
			apiErr.ErrorMsg = "Conflict"
		case 500:
			apiErr.ErrorMsg = "Internal Server Error"
		default:
			apiErr.ErrorMsg = fmt.Sprintf("HTTP Error %d", resp.StatusCode)
		}
	}

	return apiErr
}

// IsAPIError returns true if the error is an APIError.
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// AsAPIError attempts to convert an error to an APIError.
func AsAPIError(err error) (*APIError, bool) {
	apiErr, ok := err.(*APIError)
	return apiErr, ok
}
