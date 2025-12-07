// Package helpers provides tests for utility functions.
package helpers

import (
	"net/url"
	"testing"
	"time"
)

func TestToQueryString(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected url.Values
	}{
		{
			name:     "empty params",
			params:   map[string]interface{}{},
			expected: url.Values{},
		},
		{
			name: "string value",
			params: map[string]interface{}{
				"search": "test",
			},
			expected: url.Values{"search": {"test"}},
		},
		{
			name: "empty string value",
			params: map[string]interface{}{
				"search": "",
			},
			expected: url.Values{},
		},
		{
			name: "int value",
			params: map[string]interface{}{
				"limit": 10,
			},
			expected: url.Values{"limit": {"10"}},
		},
		{
			name: "int64 value",
			params: map[string]interface{}{
				"timestamp": int64(1234567890),
			},
			expected: url.Values{"timestamp": {"1234567890"}},
		},
		{
			name: "bool value",
			params: map[string]interface{}{
				"enabled": true,
			},
			expected: url.Values{"enabled": {"true"}},
		},
		{
			name: "string slice",
			params: map[string]interface{}{
				"tags": []string{"tag1", "tag2"},
			},
			expected: url.Values{"tags": {"tag1", "tag2"}},
		},
		{
			name: "nil value",
			params: map[string]interface{}{
				"key": nil,
			},
			expected: url.Values{},
		},
		{
			name: "mixed values",
			params: map[string]interface{}{
				"search": "test",
				"limit":  10,
				"active": true,
			},
			expected: url.Values{
				"search": {"test"},
				"limit":  {"10"},
				"active": {"true"},
			},
		},
		{
			name: "other type (float)",
			params: map[string]interface{}{
				"value": 3.14,
			},
			expected: url.Values{"value": {"3.14"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToQueryString(tt.params)
			for key, expectedValues := range tt.expected {
				resultValues := result[key]
				if len(resultValues) != len(expectedValues) {
					t.Errorf("ToQueryString()[%s] length = %v, want %v", key, len(resultValues), len(expectedValues))
					continue
				}
				for i, v := range expectedValues {
					if resultValues[i] != v {
						t.Errorf("ToQueryString()[%s][%d] = %v, want %v", key, i, resultValues[i], v)
					}
				}
			}
		})
	}
}

func TestToFilterString(t *testing.T) {
	tests := []struct {
		name     string
		filters  map[string]string
		expected string
	}{
		{
			name:     "empty filters",
			filters:  map[string]string{},
			expected: "",
		},
		{
			name: "single filter",
			filters: map[string]string{
				"safeName": "TestSafe",
			},
			expected: "safeName eq TestSafe",
		},
		// Note: multiple filters will have indeterminate order due to map iteration
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToFilterString(tt.filters)
			if len(tt.filters) <= 1 {
				if result != tt.expected {
					t.Errorf("ToFilterString() = %v, want %v", result, tt.expected)
				}
			} else {
				// For multiple filters, just check it's not empty
				if result == "" {
					t.Error("ToFilterString() should not be empty for non-empty filters")
				}
			}
		})
	}
}

func TestToUnixTime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	result := ToUnixTime(testTime)
	expected := testTime.UnixMilli()

	if result != expected {
		t.Errorf("ToUnixTime() = %v, want %v", result, expected)
	}
}

func TestFromUnixTime(t *testing.T) {
	timestamp := int64(1705315800) // 2024-01-15 10:30:00 UTC
	result := FromUnixTime(timestamp)

	if result.Unix() != timestamp {
		t.Errorf("FromUnixTime() = %v, want Unix = %v", result.Unix(), timestamp)
	}
}

func TestFromUnixTimeMillis(t *testing.T) {
	timestampMillis := int64(1705315800000) // 2024-01-15 10:30:00.000 UTC
	result := FromUnixTimeMillis(timestampMillis)

	if result.UnixMilli() != timestampMillis {
		t.Errorf("FromUnixTimeMillis() = %v, want UnixMilli = %v", result.UnixMilli(), timestampMillis)
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "simple string",
			expected: "simple string",
		},
		{
			name:     "backslash",
			input:    "path\\to\\file",
			expected: "path\\\\to\\\\file",
		},
		{
			name:     "double quotes",
			input:    `say "hello"`,
			expected: `say \"hello\"`,
		},
		{
			name:     "both backslash and quotes",
			input:    `path\\to\\"file"`,
			expected: `path\\\\to\\\\\"file\"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeString(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHideSecretValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short value (4 chars)",
			input:    "1234",
			expected: "****",
		},
		{
			name:     "short value (3 chars)",
			input:    "abc",
			expected: "****",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "****",
		},
		{
			name:     "normal password",
			input:    "MySecretPassword123",
			expected: "My****23",
		},
		{
			name:     "5 character value",
			input:    "abcde",
			expected: "ab****de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HideSecretValue(tt.input)
			if result != tt.expected {
				t.Errorf("HideSecretValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateSafeName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid safe name",
			input:   "MySafe",
			wantErr: false,
		},
		{
			name:    "empty safe name",
			input:   "",
			wantErr: true,
			errMsg:  "safe name cannot be empty",
		},
		{
			name:    "safe name too long",
			input:   "ThisSafeNameIsWayTooLongToBeValid",
			wantErr: true,
			errMsg:  "safe name cannot exceed 28 characters",
		},
		{
			name:    "safe name with backslash",
			input:   "My\\Safe",
			wantErr: true,
			errMsg:  "safe name contains invalid characters",
		},
		{
			name:    "safe name with colon",
			input:   "My:Safe",
			wantErr: true,
			errMsg:  "safe name contains invalid characters",
		},
		{
			name:    "safe name with asterisk",
			input:   "My*Safe",
			wantErr: true,
			errMsg:  "safe name contains invalid characters",
		},
		{
			name:    "safe name with question mark",
			input:   "My?Safe",
			wantErr: true,
			errMsg:  "safe name contains invalid characters",
		},
		{
			name:    "safe name with quotes",
			input:   `My"Safe"`,
			wantErr: true,
			errMsg:  "safe name contains invalid characters",
		},
		{
			name:    "safe name with angle brackets",
			input:   "My<Safe>",
			wantErr: true,
			errMsg:  "safe name contains invalid characters",
		},
		{
			name:    "safe name with pipe",
			input:   "My|Safe",
			wantErr: true,
			errMsg:  "safe name contains invalid characters",
		},
		{
			name:    "safe name exactly 28 characters",
			input:   "1234567890123456789012345678",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSafeName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("ValidateSafeName() expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("ValidateSafeName() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateSafeName() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAccountName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid account name",
			input:   "MyAccount",
			wantErr: false,
		},
		{
			name:    "empty account name",
			input:   "",
			wantErr: true,
			errMsg:  "account name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAccountName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("ValidateAccountName() expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("ValidateAccountName() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateAccountName() unexpected error: %v", err)
			}
		})
	}
}

func TestBuildSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		keywords []string
		expected string
	}{
		{
			name:     "empty keywords",
			keywords: []string{},
			expected: "",
		},
		{
			name:     "single keyword",
			keywords: []string{"admin"},
			expected: "admin",
		},
		{
			name:     "multiple keywords",
			keywords: []string{"admin", "production", "server"},
			expected: "admin production server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSearchQuery(tt.keywords)
			if result != tt.expected {
				t.Errorf("BuildSearchQuery() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseNextLink(t *testing.T) {
	tests := []struct {
		name     string
		nextLink string
		expected int
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid next link",
			nextLink: "https://cyberark.example.com/PasswordVault/API/Accounts?offset=100",
			expected: 100,
			wantErr:  false,
		},
		{
			name:     "empty next link",
			nextLink: "",
			wantErr:  true,
			errMsg:   "empty next link",
		},
		{
			name:     "no offset param",
			nextLink: "https://cyberark.example.com/PasswordVault/API/Accounts?limit=50",
			wantErr:  true,
			errMsg:   "offset not found in next link",
		},
		{
			name:     "invalid offset value",
			nextLink: "https://cyberark.example.com/PasswordVault/API/Accounts?offset=invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNextLink(tt.nextLink)
			if tt.wantErr {
				if err == nil {
					t.Error("ParseNextLink() expected error, got nil")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ParseNextLink() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseNextLink() unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseNextLink() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPtrString(t *testing.T) {
	input := "test"
	result := PtrString(input)
	if result == nil {
		t.Error("PtrString() returned nil")
		return
	}
	if *result != input {
		t.Errorf("PtrString() = %v, want %v", *result, input)
	}
}

func TestPtrInt(t *testing.T) {
	input := 42
	result := PtrInt(input)
	if result == nil {
		t.Error("PtrInt() returned nil")
		return
	}
	if *result != input {
		t.Errorf("PtrInt() = %v, want %v", *result, input)
	}
}

func TestPtrBool(t *testing.T) {
	for _, input := range []bool{true, false} {
		result := PtrBool(input)
		if result == nil {
			t.Error("PtrBool() returned nil")
			continue
		}
		if *result != input {
			t.Errorf("PtrBool() = %v, want %v", *result, input)
		}
	}
}

func TestDerefString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "non-nil pointer",
			input:    PtrString("test"),
			expected: "test",
		},
		{
			name:     "nil pointer",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DerefString(tt.input)
			if result != tt.expected {
				t.Errorf("DerefString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDerefInt(t *testing.T) {
	tests := []struct {
		name     string
		input    *int
		expected int
	}{
		{
			name:     "non-nil pointer",
			input:    PtrInt(42),
			expected: 42,
		},
		{
			name:     "nil pointer",
			input:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DerefInt(tt.input)
			if result != tt.expected {
				t.Errorf("DerefInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDerefBool(t *testing.T) {
	tests := []struct {
		name     string
		input    *bool
		expected bool
	}{
		{
			name:     "non-nil true pointer",
			input:    PtrBool(true),
			expected: true,
		},
		{
			name:     "non-nil false pointer",
			input:    PtrBool(false),
			expected: false,
		},
		{
			name:     "nil pointer",
			input:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DerefBool(tt.input)
			if result != tt.expected {
				t.Errorf("DerefBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}
