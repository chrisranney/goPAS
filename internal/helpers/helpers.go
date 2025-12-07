// Package helpers provides utility functions for the goPAS module.
// These correspond to the private helper functions in psPAS.
package helpers

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ToQueryString converts a map of parameters to a URL query string.
// This is equivalent to ConvertTo-QueryString in psPAS.
func ToQueryString(params map[string]interface{}) url.Values {
	values := url.Values{}
	for key, value := range params {
		if value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			if v != "" {
				values.Set(key, v)
			}
		case int:
			values.Set(key, strconv.Itoa(v))
		case int64:
			values.Set(key, strconv.FormatInt(v, 10))
		case bool:
			values.Set(key, strconv.FormatBool(v))
		case []string:
			for _, s := range v {
				values.Add(key, s)
			}
		default:
			values.Set(key, fmt.Sprintf("%v", v))
		}
	}
	return values
}

// ToFilterString converts filter parameters to a CyberArk filter string.
// This is equivalent to ConvertTo-FilterString in psPAS.
func ToFilterString(filters map[string]string) string {
	if len(filters) == 0 {
		return ""
	}

	var parts []string
	for key, value := range filters {
		parts = append(parts, fmt.Sprintf("%s eq %s", key, value))
	}
	return strings.Join(parts, " AND ")
}

// ToUnixTime converts a time.Time to Unix timestamp in milliseconds.
// This is equivalent to ConvertTo-UnixTime in psPAS.
func ToUnixTime(t time.Time) int64 {
	return t.UnixMilli()
}

// FromUnixTime converts Unix timestamp in seconds to time.Time.
func FromUnixTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// FromUnixTimeMillis converts Unix timestamp in milliseconds to time.Time.
func FromUnixTimeMillis(timestamp int64) time.Time {
	return time.UnixMilli(timestamp)
}

// EscapeString escapes special characters in a string for API requests.
// This is equivalent to Get-EscapedString in psPAS.
func EscapeString(s string) string {
	// Escape special characters
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// HideSecretValue masks a secret value for logging.
// This is equivalent to Hide-SecretValue in psPAS.
func HideSecretValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}

// ValidateSafeName validates a safe name according to CyberArk rules.
func ValidateSafeName(name string) error {
	if name == "" {
		return fmt.Errorf("safe name cannot be empty")
	}
	if len(name) > 28 {
		return fmt.Errorf("safe name cannot exceed 28 characters")
	}
	// Check for invalid characters
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|]`)
	if invalidChars.MatchString(name) {
		return fmt.Errorf("safe name contains invalid characters")
	}
	return nil
}

// ValidateAccountName validates an account name.
func ValidateAccountName(name string) error {
	if name == "" {
		return fmt.Errorf("account name cannot be empty")
	}
	return nil
}

// BuildSearchQuery builds a search query string for the CyberArk API.
func BuildSearchQuery(keywords []string) string {
	return strings.Join(keywords, " ")
}

// ParseNextLink extracts the offset from a next link URL.
// This is equivalent to Get-NextLink in psPAS.
func ParseNextLink(nextLink string) (int, error) {
	if nextLink == "" {
		return 0, fmt.Errorf("empty next link")
	}

	u, err := url.Parse(nextLink)
	if err != nil {
		return 0, err
	}

	offset := u.Query().Get("offset")
	if offset == "" {
		return 0, fmt.Errorf("offset not found in next link")
	}

	return strconv.Atoi(offset)
}

// PtrString returns a pointer to a string.
func PtrString(s string) *string {
	return &s
}

// PtrInt returns a pointer to an int.
func PtrInt(i int) *int {
	return &i
}

// PtrBool returns a pointer to a bool.
func PtrBool(b bool) *bool {
	return &b
}

// DerefString returns the value of a string pointer or empty string if nil.
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DerefInt returns the value of an int pointer or 0 if nil.
func DerefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// DerefBool returns the value of a bool pointer or false if nil.
func DerefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
