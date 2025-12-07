// Package helpers provides tests for version utilities.
package helpers

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMajor   int
		wantMinor   int
		wantPatch   int
		wantErr     bool
	}{
		{
			name:      "full version",
			input:     "14.0.0",
			wantMajor: 14,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:      "version with v prefix",
			input:     "v14.0.0",
			wantMajor: 14,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:      "version with V prefix",
			input:     "V14.0.0",
			wantMajor: 14,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:      "major.minor only",
			input:     "14.0",
			wantMajor: 14,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:      "major only",
			input:     "14",
			wantMajor: 14,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:    "empty version",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid major version",
			input:   "abc.0.0",
			wantErr: true,
		},
		{
			name:    "invalid minor version",
			input:   "14.abc.0",
			wantErr: true,
		},
		{
			name:    "invalid patch version",
			input:   "14.0.abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("ParseVersion() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ParseVersion() unexpected error: %v", err)
				return
			}
			if v.Major != tt.wantMajor {
				t.Errorf("ParseVersion().Major = %v, want %v", v.Major, tt.wantMajor)
			}
			if v.Minor != tt.wantMinor {
				t.Errorf("ParseVersion().Minor = %v, want %v", v.Minor, tt.wantMinor)
			}
			if v.Patch != tt.wantPatch {
				t.Errorf("ParseVersion().Patch = %v, want %v", v.Patch, tt.wantPatch)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{
			name:     "full version",
			version:  &Version{Major: 14, Minor: 0, Patch: 0},
			expected: "14.0.0",
		},
		{
			name:     "version with non-zero components",
			version:  &Version{Major: 12, Minor: 6, Patch: 3},
			expected: "12.6.3",
		},
		{
			name:     "zero version",
			version:  &Version{Major: 0, Minor: 0, Patch: 0},
			expected: "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("Version.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		name     string
		v        *Version
		other    *Version
		expected int
	}{
		{
			name:     "equal versions",
			v:        &Version{Major: 14, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: 0,
		},
		{
			name:     "v greater major",
			v:        &Version{Major: 15, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: 1,
		},
		{
			name:     "v lesser major",
			v:        &Version{Major: 13, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: -1,
		},
		{
			name:     "v greater minor",
			v:        &Version{Major: 14, Minor: 1, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: 1,
		},
		{
			name:     "v lesser minor",
			v:        &Version{Major: 14, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 1, Patch: 0},
			expected: -1,
		},
		{
			name:     "v greater patch",
			v:        &Version{Major: 14, Minor: 0, Patch: 1},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: 1,
		},
		{
			name:     "v lesser patch",
			v:        &Version{Major: 14, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 1},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Compare(tt.other)
			if result != tt.expected {
				t.Errorf("Version.Compare() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVersion_GreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		name     string
		v        *Version
		other    *Version
		expected bool
	}{
		{
			name:     "equal versions",
			v:        &Version{Major: 14, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: true,
		},
		{
			name:     "v greater",
			v:        &Version{Major: 15, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: true,
		},
		{
			name:     "v lesser",
			v:        &Version{Major: 13, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.GreaterThanOrEqual(tt.other)
			if result != tt.expected {
				t.Errorf("Version.GreaterThanOrEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVersion_LessThan(t *testing.T) {
	tests := []struct {
		name     string
		v        *Version
		other    *Version
		expected bool
	}{
		{
			name:     "equal versions",
			v:        &Version{Major: 14, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: false,
		},
		{
			name:     "v greater",
			v:        &Version{Major: 15, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: false,
		},
		{
			name:     "v lesser",
			v:        &Version{Major: 13, Minor: 0, Patch: 0},
			other:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.LessThan(tt.other)
			if result != tt.expected {
				t.Errorf("Version.LessThan() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewVersionRequirement(t *testing.T) {
	tests := []struct {
		name       string
		minVersion string
		maxVersion string
		wantErr    bool
	}{
		{
			name:       "valid min and max",
			minVersion: "12.0.0",
			maxVersion: "14.0.0",
			wantErr:    false,
		},
		{
			name:       "valid min only",
			minVersion: "12.0.0",
			maxVersion: "",
			wantErr:    false,
		},
		{
			name:       "valid max only",
			minVersion: "",
			maxVersion: "14.0.0",
			wantErr:    false,
		},
		{
			name:       "both empty",
			minVersion: "",
			maxVersion: "",
			wantErr:    false,
		},
		{
			name:       "invalid min version",
			minVersion: "invalid",
			maxVersion: "",
			wantErr:    true,
		},
		{
			name:       "invalid max version",
			minVersion: "",
			maxVersion: "invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewVersionRequirement(tt.minVersion, tt.maxVersion)
			if tt.wantErr {
				if err == nil {
					t.Error("NewVersionRequirement() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("NewVersionRequirement() unexpected error: %v", err)
				return
			}
			if req == nil {
				t.Error("NewVersionRequirement() returned nil")
			}
		})
	}
}

func TestVersionRequirement_IsSatisfied(t *testing.T) {
	tests := []struct {
		name       string
		minVersion string
		maxVersion string
		version    *Version
		expected   bool
	}{
		{
			name:       "within range",
			minVersion: "12.0.0",
			maxVersion: "14.0.0",
			version:    &Version{Major: 13, Minor: 0, Patch: 0},
			expected:   true,
		},
		{
			name:       "at min boundary",
			minVersion: "12.0.0",
			maxVersion: "14.0.0",
			version:    &Version{Major: 12, Minor: 0, Patch: 0},
			expected:   true,
		},
		{
			name:       "at max boundary",
			minVersion: "12.0.0",
			maxVersion: "14.0.0",
			version:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected:   true,
		},
		{
			name:       "below min",
			minVersion: "12.0.0",
			maxVersion: "14.0.0",
			version:    &Version{Major: 11, Minor: 0, Patch: 0},
			expected:   false,
		},
		{
			name:       "above max",
			minVersion: "12.0.0",
			maxVersion: "14.0.0",
			version:    &Version{Major: 15, Minor: 0, Patch: 0},
			expected:   false,
		},
		{
			name:       "no min constraint",
			minVersion: "",
			maxVersion: "14.0.0",
			version:    &Version{Major: 10, Minor: 0, Patch: 0},
			expected:   true,
		},
		{
			name:       "no max constraint",
			minVersion: "12.0.0",
			maxVersion: "",
			version:    &Version{Major: 20, Minor: 0, Patch: 0},
			expected:   true,
		},
		{
			name:       "no constraints",
			minVersion: "",
			maxVersion: "",
			version:    &Version{Major: 14, Minor: 0, Patch: 0},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewVersionRequirement(tt.minVersion, tt.maxVersion)
			if err != nil {
				t.Fatalf("NewVersionRequirement() error: %v", err)
			}
			result := req.IsSatisfied(tt.version)
			if result != tt.expected {
				t.Errorf("IsSatisfied() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAssertVersionRequirement(t *testing.T) {
	tests := []struct {
		name                   string
		currentVersion         string
		minVersion             string
		maxVersion             string
		privilegeCloudRequired bool
		selfHostedRequired     bool
		isPrivilegeCloud       bool
		wantErr                bool
		errContains            string
	}{
		{
			name:           "version satisfied",
			currentVersion: "14.0.0",
			minVersion:     "12.0.0",
			maxVersion:     "",
			wantErr:        false,
		},
		{
			name:           "version below min",
			currentVersion: "11.0.0",
			minVersion:     "12.0.0",
			maxVersion:     "",
			wantErr:        true,
			errContains:    "requires CyberArk version 12.0.0 or higher",
		},
		{
			name:           "version above max",
			currentVersion: "15.0.0",
			minVersion:     "",
			maxVersion:     "14.0.0",
			wantErr:        true,
			errContains:    "requires CyberArk version 14.0.0 or lower",
		},
		{
			name:           "version outside range",
			currentVersion: "11.0.0",
			minVersion:     "12.0.0",
			maxVersion:     "14.0.0",
			wantErr:        true,
			errContains:    "requires CyberArk version between 12.0.0 and 14.0.0",
		},
		{
			name:                   "requires privilege cloud",
			currentVersion:         "14.0.0",
			minVersion:             "",
			maxVersion:             "",
			privilegeCloudRequired: true,
			isPrivilegeCloud:       false,
			wantErr:                true,
			errContains:            "requires Privilege Cloud",
		},
		{
			name:               "requires self-hosted",
			currentVersion:     "14.0.0",
			minVersion:         "",
			maxVersion:         "",
			selfHostedRequired: true,
			isPrivilegeCloud:   true,
			wantErr:            true,
			errContains:        "requires Self-Hosted",
		},
		{
			name:                   "privilege cloud satisfied",
			currentVersion:         "14.0.0",
			minVersion:             "",
			maxVersion:             "",
			privilegeCloudRequired: true,
			isPrivilegeCloud:       true,
			wantErr:                false,
		},
		{
			name:               "self-hosted satisfied",
			currentVersion:     "14.0.0",
			minVersion:         "",
			maxVersion:         "",
			selfHostedRequired: true,
			isPrivilegeCloud:   false,
			wantErr:            false,
		},
		{
			name:           "invalid current version",
			currentVersion: "invalid",
			minVersion:     "12.0.0",
			maxVersion:     "",
			wantErr:        true,
			errContains:    "failed to parse current version",
		},
		{
			name:           "invalid min version",
			currentVersion: "14.0.0",
			minVersion:     "invalid",
			maxVersion:     "",
			wantErr:        true,
			errContains:    "failed to create version requirement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertVersionRequirement(
				tt.currentVersion,
				tt.minVersion,
				tt.maxVersion,
				tt.privilegeCloudRequired,
				tt.selfHostedRequired,
				tt.isPrivilegeCloud,
			)
			if tt.wantErr {
				if err == nil {
					t.Error("AssertVersionRequirement() expected error, got nil")
				} else if tt.errContains != "" && !containsSubstring(err.Error(), tt.errContains) {
					t.Errorf("AssertVersionRequirement() error = %v, want containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("AssertVersionRequirement() unexpected error: %v", err)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && contains(s, substr)))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
