// Package session provides tests for session management.
package session

import (
	"sync"
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name    string
		baseURI string
		wantErr bool
	}{
		{
			name:    "valid base URI",
			baseURI: "https://cyberark.example.com",
			wantErr: false,
		},
		{
			name:    "empty base URI",
			baseURI: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess, err := NewSession(tt.baseURI)
			if tt.wantErr {
				if err == nil {
					t.Error("NewSession() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("NewSession() unexpected error: %v", err)
				return
			}
			if sess == nil {
				t.Error("NewSession() returned nil session")
				return
			}

			// Verify session properties
			if sess.BaseURI != tt.baseURI {
				t.Errorf("BaseURI = %v, want %v", sess.BaseURI, tt.baseURI)
			}
			if sess.APIURI == "" {
				t.Error("APIURI should not be empty")
			}
			if sess.Client == nil {
				t.Error("Client should not be nil")
			}
			if sess.StartTime.IsZero() {
				t.Error("StartTime should be set")
			}
			if sess.IsAuthenticated {
				t.Error("IsAuthenticated should be false initially")
			}
		})
	}
}

func TestSession_SetAuthenticated(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	user := "testuser"
	token := "test-token-123"
	authMethod := "CyberArk"

	sess.SetAuthenticated(user, token, authMethod)

	if sess.User != user {
		t.Errorf("User = %v, want %v", sess.User, user)
	}
	if sess.SessionToken != token {
		t.Errorf("SessionToken = %v, want %v", sess.SessionToken, token)
	}
	if sess.AuthMethod != authMethod {
		t.Errorf("AuthMethod = %v, want %v", sess.AuthMethod, authMethod)
	}
	if !sess.IsAuthenticated {
		t.Error("IsAuthenticated should be true")
	}
	if sess.Client.GetAuthToken() != token {
		t.Errorf("Client.GetAuthToken() = %v, want %v", sess.Client.GetAuthToken(), token)
	}
}

func TestSession_SetVersion(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	version := "14.0"
	sess.SetVersion(version)

	if sess.ExternalVersion != version {
		t.Errorf("ExternalVersion = %v, want %v", sess.ExternalVersion, version)
	}
}

func TestSession_SetPrivilegeCloud(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	// Initially false
	if sess.PrivilegeCloud {
		t.Error("PrivilegeCloud should be false initially")
	}

	sess.SetPrivilegeCloud(true)
	if !sess.PrivilegeCloud {
		t.Error("PrivilegeCloud should be true")
	}

	sess.SetPrivilegeCloud(false)
	if sess.PrivilegeCloud {
		t.Error("PrivilegeCloud should be false")
	}
}

func TestSession_UpdateLastCommand(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	// Initially empty
	if sess.LastCommand != "" {
		t.Error("LastCommand should be empty initially")
	}

	cmd := "Get-PASAccount"
	before := time.Now()
	sess.UpdateLastCommand(cmd)
	after := time.Now()

	if sess.LastCommand != cmd {
		t.Errorf("LastCommand = %v, want %v", sess.LastCommand, cmd)
	}
	if sess.LastCommandTime.Before(before) || sess.LastCommandTime.After(after) {
		t.Error("LastCommandTime should be within expected range")
	}
}

func TestSession_UpdateLastError(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	// Initially nil
	if sess.LastError != nil {
		t.Error("LastError should be nil initially")
	}

	testErr := &testError{msg: "test error"}
	before := time.Now()
	sess.UpdateLastError(testErr)
	after := time.Now()

	if sess.LastError != testErr {
		t.Errorf("LastError = %v, want %v", sess.LastError, testErr)
	}
	if sess.LastErrorTime.Before(before) || sess.LastErrorTime.After(after) {
		t.Error("LastErrorTime should be within expected range")
	}
}

func TestSession_GetElapsedTime(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	// Sleep a bit to ensure elapsed time is > 0
	time.Sleep(10 * time.Millisecond)

	elapsed := sess.GetElapsedTime()
	if elapsed < 10*time.Millisecond {
		t.Errorf("GetElapsedTime() = %v, want >= 10ms", elapsed)
	}
}

func TestSession_Close(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	// Set authenticated state
	sess.SetAuthenticated("user", "token", "CyberArk")
	if !sess.IsAuthenticated {
		t.Error("Session should be authenticated before close")
	}

	sess.Close()

	if sess.IsAuthenticated {
		t.Error("IsAuthenticated should be false after close")
	}
	if sess.SessionToken != "" {
		t.Error("SessionToken should be empty after close")
	}
}

func TestSession_Clone(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	// Set up session state
	sess.SetAuthenticated("testuser", "test-token", "CyberArk")
	sess.SetVersion("14.0")
	sess.SetPrivilegeCloud(true)

	// Clone the session
	clone := sess.Clone()

	// Verify clone has same values
	if clone.BaseURI != sess.BaseURI {
		t.Errorf("Clone.BaseURI = %v, want %v", clone.BaseURI, sess.BaseURI)
	}
	if clone.APIURI != sess.APIURI {
		t.Errorf("Clone.APIURI = %v, want %v", clone.APIURI, sess.APIURI)
	}
	if clone.User != sess.User {
		t.Errorf("Clone.User = %v, want %v", clone.User, sess.User)
	}
	if clone.SessionToken != sess.SessionToken {
		t.Errorf("Clone.SessionToken = %v, want %v", clone.SessionToken, sess.SessionToken)
	}
	if clone.AuthMethod != sess.AuthMethod {
		t.Errorf("Clone.AuthMethod = %v, want %v", clone.AuthMethod, sess.AuthMethod)
	}
	if clone.IsAuthenticated != sess.IsAuthenticated {
		t.Errorf("Clone.IsAuthenticated = %v, want %v", clone.IsAuthenticated, sess.IsAuthenticated)
	}
	if clone.ExternalVersion != sess.ExternalVersion {
		t.Errorf("Clone.ExternalVersion = %v, want %v", clone.ExternalVersion, sess.ExternalVersion)
	}
	if clone.PrivilegeCloud != sess.PrivilegeCloud {
		t.Errorf("Clone.PrivilegeCloud = %v, want %v", clone.PrivilegeCloud, sess.PrivilegeCloud)
	}
	if clone.Client != sess.Client {
		t.Error("Clone.Client should be the same client instance")
	}

	// Verify clone is independent (LastCommand not copied)
	if clone.LastCommand != "" {
		t.Error("Clone.LastCommand should be empty (not copied)")
	}
}

func TestSession_IsValid(t *testing.T) {
	tests := []struct {
		name           string
		isAuthenticated bool
		sessionToken   string
		expected       bool
	}{
		{
			name:           "valid session",
			isAuthenticated: true,
			sessionToken:   "valid-token",
			expected:       true,
		},
		{
			name:           "not authenticated",
			isAuthenticated: false,
			sessionToken:   "token",
			expected:       false,
		},
		{
			name:           "empty token",
			isAuthenticated: true,
			sessionToken:   "",
			expected:       false,
		},
		{
			name:           "not authenticated and empty token",
			isAuthenticated: false,
			sessionToken:   "",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess, err := NewSession("https://cyberark.example.com")
			if err != nil {
				t.Fatalf("NewSession() error: %v", err)
			}

			sess.IsAuthenticated = tt.isAuthenticated
			sess.SessionToken = tt.sessionToken

			result := sess.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSession_ThreadSafety(t *testing.T) {
	sess, err := NewSession("https://cyberark.example.com")
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	wg.Add(iterations * 5)

	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			sess.SetAuthenticated("user", "token", "method")
		}(i)

		go func(i int) {
			defer wg.Done()
			sess.SetVersion("14.0")
		}(i)

		go func(i int) {
			defer wg.Done()
			sess.SetPrivilegeCloud(true)
		}(i)

		go func(i int) {
			defer wg.Done()
			sess.UpdateLastCommand("cmd")
		}(i)

		go func(i int) {
			defer wg.Done()
			sess.UpdateLastError(&testError{msg: "error"})
		}(i)
	}

	// Concurrent reads
	wg.Add(iterations * 3)

	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = sess.GetElapsedTime()
		}()

		go func() {
			defer wg.Done()
			_ = sess.IsValid()
		}()

		go func() {
			defer wg.Done()
			_ = sess.Clone()
		}()
	}

	wg.Wait()
	// If we get here without panic/race, test passes
}

// testError is a helper error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
