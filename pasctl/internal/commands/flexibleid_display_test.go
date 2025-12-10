package commands

import (
	"testing"

	"github.com/chrisranney/gopas/pkg/types"
)

// TestFlexibleID_StringForDisplay tests that FlexibleID values are correctly
// converted to strings for table display in various command outputs.
func TestFlexibleID_StringForDisplay(t *testing.T) {
	tests := []struct {
		name string
		id   types.FlexibleID
		want string
	}{
		{
			name: "UUID string",
			id:   types.FlexibleID("550e8400-e29b-41d4-a716-446655440000"),
			want: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name: "numeric ID as string",
			id:   types.FlexibleID("12345"),
			want: "12345",
		},
		{
			name: "alphanumeric ID",
			id:   types.FlexibleID("acc_12_34"),
			want: "acc_12_34",
		},
		{
			name: "empty ID",
			id:   types.FlexibleID(""),
			want: "",
		},
		{
			name: "platform ID format",
			id:   types.FlexibleID("WinServerLocal"),
			want: "WinServerLocal",
		},
		{
			name: "session ID format",
			id:   types.FlexibleID("PSM-session-abc123"),
			want: "PSM-session-abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.String()
			if got != tt.want {
				t.Errorf("FlexibleID.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFlexibleID_TruncateForDisplay tests that FlexibleID values can be
// truncated after conversion to string for table display (used in monitoring.go).
func TestFlexibleID_TruncateForDisplay(t *testing.T) {
	tests := []struct {
		name   string
		id     types.FlexibleID
		maxLen int
		want   string
	}{
		{
			name:   "short ID no truncation",
			id:     types.FlexibleID("abc123"),
			maxLen: 20,
			want:   "abc123",
		},
		{
			name:   "long ID truncated",
			id:     types.FlexibleID("very-long-session-identifier-here"),
			maxLen: 20,
			want:   "very-long-session...",
		},
		{
			name:   "exact length",
			id:     types.FlexibleID("exactly20character!"),
			maxLen: 20,
			want:   "exactly20character!",
		},
		{
			name:   "empty ID",
			id:     types.FlexibleID(""),
			maxLen: 20,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to string then truncate (simulating monitoring.go usage)
			got := truncate(tt.id.String(), tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(FlexibleID.String(), %d) = %v, want %v", tt.maxLen, got, tt.want)
			}
		})
	}
}

// TestFlexibleID_EmptyComparison tests that empty FlexibleID values can be
// detected after string conversion (used in platforms.go).
func TestFlexibleID_EmptyComparison(t *testing.T) {
	tests := []struct {
		name      string
		primary   types.FlexibleID
		fallback  types.FlexibleID
		wantValue string
	}{
		{
			name:      "use primary when not empty",
			primary:   types.FlexibleID("WinServerLocal"),
			fallback:  types.FlexibleID("123"),
			wantValue: "WinServerLocal",
		},
		{
			name:      "use fallback when primary empty",
			primary:   types.FlexibleID(""),
			fallback:  types.FlexibleID("123"),
			wantValue: "123",
		},
		{
			name:      "both empty",
			primary:   types.FlexibleID(""),
			fallback:  types.FlexibleID(""),
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the platforms.go pattern
			id := tt.primary.String()
			if id == "" {
				id = tt.fallback.String()
			}
			if id != tt.wantValue {
				t.Errorf("FlexibleID fallback logic = %v, want %v", id, tt.wantValue)
			}
		})
	}
}

// TestFlexibleID_TableRowValues tests that FlexibleID values are correctly
// converted for use in table rows (simulating accounts.go usage).
func TestFlexibleID_TableRowValues(t *testing.T) {
	// Simulate an account-like structure with FlexibleID fields
	type mockAccount struct {
		ID         types.FlexibleID
		PlatformID types.FlexibleID
		UserName   string
		Address    string
		SafeName   string
	}

	accounts := []mockAccount{
		{
			ID:         types.FlexibleID("12345"),
			PlatformID: types.FlexibleID("WinServerLocal"),
			UserName:   "admin",
			Address:    "server1.example.com",
			SafeName:   "Production",
		},
		{
			ID:         types.FlexibleID("550e8400-e29b-41d4-a716-446655440000"),
			PlatformID: types.FlexibleID("UnixSSH"),
			UserName:   "root",
			Address:    "linux1.example.com",
			SafeName:   "Development",
		},
	}

	for i, acc := range accounts {
		// Test that String() returns expected values for table display
		idStr := acc.ID.String()
		platformStr := acc.PlatformID.String()

		if idStr == "" && string(acc.ID) != "" {
			t.Errorf("Account[%d]: ID.String() returned empty for non-empty ID", i)
		}
		if platformStr == "" && string(acc.PlatformID) != "" {
			t.Errorf("Account[%d]: PlatformID.String() returned empty for non-empty PlatformID", i)
		}

		// Verify the conversion preserves the value
		if idStr != string(acc.ID) {
			t.Errorf("Account[%d]: ID.String() = %v, want %v", i, idStr, string(acc.ID))
		}
		if platformStr != string(acc.PlatformID) {
			t.Errorf("Account[%d]: PlatformID.String() = %v, want %v", i, platformStr, string(acc.PlatformID))
		}
	}
}

// TestFlexibleID_SessionIDDisplay tests FlexibleID conversion for PSM session IDs
// (simulating monitoring.go usage).
func TestFlexibleID_SessionIDDisplay(t *testing.T) {
	type mockSession struct {
		SessionID     types.FlexibleID
		User          string
		RemoteMachine string
	}

	sessions := []mockSession{
		{
			SessionID:     types.FlexibleID("PSM-abc123-def456"),
			User:          "user1",
			RemoteMachine: "server1",
		},
		{
			SessionID:     types.FlexibleID("789012"),
			User:          "user2",
			RemoteMachine: "server2",
		},
	}

	for i, s := range sessions {
		// Simulate the monitoring.go pattern: truncate(s.SessionID.String(), 20)
		sessionIDStr := truncate(s.SessionID.String(), 20)

		if sessionIDStr == "" && string(s.SessionID) != "" {
			t.Errorf("Session[%d]: SessionID display is empty for non-empty ID", i)
		}

		// Verify truncation works correctly
		if len(string(s.SessionID)) <= 20 {
			if sessionIDStr != string(s.SessionID) {
				t.Errorf("Session[%d]: SessionID should not be truncated, got %v want %v", i, sessionIDStr, string(s.SessionID))
			}
		} else {
			if len(sessionIDStr) > 20 {
				t.Errorf("Session[%d]: SessionID should be truncated to 20 chars, got %d", i, len(sessionIDStr))
			}
		}
	}
}
