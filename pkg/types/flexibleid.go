// Package types provides shared types used across the gopas library.
package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexibleID is a type that can unmarshal from either a JSON string or number.
// CyberArk API documentation often shows IDs as UUID strings, but some API versions
// return them as integers. This type handles both cases transparently.
type FlexibleID string

// UnmarshalJSON implements json.Unmarshaler for FlexibleID.
// It accepts both JSON strings and numbers, converting numbers to strings.
func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleID(s)
		return nil
	}

	// Try to unmarshal as number (integer)
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleID(strconv.FormatInt(n, 10))
		return nil
	}

	// Try to unmarshal as float (for large numbers that might come as floats)
	var fn float64
	if err := json.Unmarshal(data, &fn); err == nil {
		*f = FlexibleID(strconv.FormatFloat(fn, 'f', -1, 64))
		return nil
	}

	return fmt.Errorf("FlexibleID: cannot unmarshal %s", string(data))
}

// MarshalJSON implements json.Marshaler for FlexibleID.
// It always marshals as a string.
func (f FlexibleID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

// String returns the string representation of the FlexibleID.
func (f FlexibleID) String() string {
	return string(f)
}
