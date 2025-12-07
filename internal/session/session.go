// Package session manages the CyberArk session state.
// This is equivalent to the $psPASSession script-scoped variable in psPAS.
package session

import (
	"sync"
	"time"

	"github.com/chrisranney/gopas/internal/client"
)

// Session represents an authenticated session with CyberArk.
type Session struct {
	mu sync.RWMutex

	// Client is the HTTP client used for API requests
	Client *client.Client

	// BaseURI is the base URL of the CyberArk server
	BaseURI string

	// APIURI is the computed API URL
	APIURI string

	// User is the authenticated username
	User string

	// ExternalVersion is the CyberArk version
	ExternalVersion string

	// StartTime is when the session was created
	StartTime time.Time

	// LastCommand is the last executed command
	LastCommand string

	// LastCommandTime is when the last command was executed
	LastCommandTime time.Time

	// LastError holds the last error that occurred
	LastError error

	// LastErrorTime is when the last error occurred
	LastErrorTime time.Time

	// IsAuthenticated indicates if the session is authenticated
	IsAuthenticated bool

	// AuthMethod is the authentication method used
	AuthMethod string

	// SessionToken is the authentication token
	SessionToken string

	// PrivilegeCloud indicates if connected to Privilege Cloud (ISPSS)
	PrivilegeCloud bool
}

// NewSession creates a new unauthenticated session.
func NewSession(baseURI string) (*Session, error) {
	cfg := client.Config{
		BaseURL: baseURI,
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Session{
		Client:    c,
		BaseURI:   baseURI,
		APIURI:    c.GetAPIURL(),
		StartTime: time.Now(),
	}, nil
}

// SetAuthenticated marks the session as authenticated.
func (s *Session) SetAuthenticated(user, token, authMethod string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.User = user
	s.SessionToken = token
	s.AuthMethod = authMethod
	s.IsAuthenticated = true
	s.Client.SetAuthToken(token)
}

// SetVersion sets the CyberArk version for the session.
func (s *Session) SetVersion(version string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ExternalVersion = version
}

// SetPrivilegeCloud marks the session as a Privilege Cloud connection.
func (s *Session) SetPrivilegeCloud(isCloud bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PrivilegeCloud = isCloud
}

// UpdateLastCommand updates the last command tracking.
func (s *Session) UpdateLastCommand(cmd string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastCommand = cmd
	s.LastCommandTime = time.Now()
}

// UpdateLastError updates the last error tracking.
func (s *Session) UpdateLastError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastError = err
	s.LastErrorTime = time.Now()
}

// GetElapsedTime returns the duration since the session started.
func (s *Session) GetElapsedTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.StartTime)
}

// Close closes the session (does not log out from CyberArk).
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsAuthenticated = false
	s.SessionToken = ""
}

// Clone creates a copy of the session.
// This is equivalent to Get-SessionClone in psPAS.
func (s *Session) Clone() *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &Session{
		Client:          s.Client,
		BaseURI:         s.BaseURI,
		APIURI:          s.APIURI,
		User:            s.User,
		ExternalVersion: s.ExternalVersion,
		StartTime:       s.StartTime,
		IsAuthenticated: s.IsAuthenticated,
		AuthMethod:      s.AuthMethod,
		SessionToken:    s.SessionToken,
		PrivilegeCloud:  s.PrivilegeCloud,
	}
}

// IsValid returns true if the session is valid and authenticated.
func (s *Session) IsValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IsAuthenticated && s.SessionToken != ""
}
