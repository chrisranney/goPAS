// Package helpers provides version checking utilities.
// This is equivalent to Assert-VersionRequirement and version comparison functions in psPAS.
package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a version string into a Version struct.
func ParseVersion(v string) (*Version, error) {
	// Remove leading 'v' if present
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")

	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	version := &Version{}

	if len(parts) >= 1 {
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid major version: %s", parts[0])
		}
		version.Major = major
	}

	if len(parts) >= 2 {
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
		version.Minor = minor
	}

	if len(parts) >= 3 {
		patch, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
		version.Patch = patch
	}

	return version, nil
}

// String returns the version as a string.
func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares two versions.
// Returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	return 0
}

// GreaterThanOrEqual returns true if v >= other.
func (v *Version) GreaterThanOrEqual(other *Version) bool {
	return v.Compare(other) >= 0
}

// LessThan returns true if v < other.
func (v *Version) LessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// VersionRequirement represents a CyberArk version requirement.
type VersionRequirement struct {
	MinVersion *Version
	MaxVersion *Version
}

// NewVersionRequirement creates a new version requirement.
func NewVersionRequirement(minVersion, maxVersion string) (*VersionRequirement, error) {
	req := &VersionRequirement{}

	if minVersion != "" {
		v, err := ParseVersion(minVersion)
		if err != nil {
			return nil, err
		}
		req.MinVersion = v
	}

	if maxVersion != "" {
		v, err := ParseVersion(maxVersion)
		if err != nil {
			return nil, err
		}
		req.MaxVersion = v
	}

	return req, nil
}

// IsSatisfied returns true if the given version satisfies the requirement.
func (r *VersionRequirement) IsSatisfied(version *Version) bool {
	if r.MinVersion != nil && version.LessThan(r.MinVersion) {
		return false
	}
	if r.MaxVersion != nil && version.Compare(r.MaxVersion) > 0 {
		return false
	}
	return true
}

// AssertVersionRequirement checks if the current version meets the requirement.
// This is equivalent to Assert-VersionRequirement in psPAS.
func AssertVersionRequirement(currentVersion string, minVersion string, maxVersion string, privilegeCloudRequired bool, selfHostedRequired bool, isPrivilegeCloud bool) error {
	if privilegeCloudRequired && !isPrivilegeCloud {
		return fmt.Errorf("this operation requires Privilege Cloud")
	}

	if selfHostedRequired && isPrivilegeCloud {
		return fmt.Errorf("this operation requires Self-Hosted (not supported in Privilege Cloud)")
	}

	current, err := ParseVersion(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to parse current version: %w", err)
	}

	req, err := NewVersionRequirement(minVersion, maxVersion)
	if err != nil {
		return fmt.Errorf("failed to create version requirement: %w", err)
	}

	if !req.IsSatisfied(current) {
		if minVersion != "" && maxVersion != "" {
			return fmt.Errorf("this operation requires CyberArk version between %s and %s (current: %s)", minVersion, maxVersion, currentVersion)
		}
		if minVersion != "" {
			return fmt.Errorf("this operation requires CyberArk version %s or higher (current: %s)", minVersion, currentVersion)
		}
		if maxVersion != "" {
			return fmt.Errorf("this operation requires CyberArk version %s or lower (current: %s)", maxVersion, currentVersion)
		}
	}

	return nil
}
