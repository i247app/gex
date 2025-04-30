// pkg/session/session_manager.go
package session

import (
	"sync"
	"time"
)

// SessionManager manages user sessions.  This struct
// provides methods for creating, retrieving, and
// destroying sessions.
type SessionManager struct {
	mu      sync.RWMutex
	sessions map[string]*Session
	// Configuration options (e.g., session timeout) can be added here.
	sessionDuration time.Duration
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager(sessionDuration time.Duration) *SessionManager {
	return &SessionManager{
		sessions:        make(map[string]*Session),
		sessionDuration: sessionDuration,
	}
}

// CreateSession creates a new session for the given user ID.
// It generates a unique session ID.
func (m *SessionManager) CreateSession(userID string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := generateSessionID() //  Replace with a proper ID generator (e.g., UUID)
	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(m.sessionDuration),
		Data:      make(map[string]interface{}),
	}
	m.sessions[sessionID] = session
	return session
}

// GetSession retrieves a session by its ID.  If the session
// is expired or doesn't exist, it returns nil.
func (m *SessionManager) GetSession(sessionID string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil // Session not found.
	}
	if session.IsExpired() {
		m.deleteSession(sessionID) // Clean up expired session.
		return nil                 // Return nil to indicate expiration.
	}
	return session
}

// DeleteSession deletes a session by its ID.
func (m *SessionManager) DeleteSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// generateSessionID generates a unique session ID.
//
// This is a placeholder.  In a real application, you should use
// a proper UUID generator (e.g., from the "github.com/google/uuid" package).
func generateSessionID() string {
	//  Replace this with a call to a UUID generator.
	//  Example (using github.com/google/uuid):
	//  uuid := uuid.New()
	//  return uuid.String()
	return "session-" + time.Now().Format("20060102150405") // Insecure, for demonstration only.
}

// CleanupExpiredSessions removes expired sessions.  This method
// should be called periodically (e.g., in a background goroutine)
// to prevent the session store from growing indefinitely.
func (m *SessionManager) CleanupExpiredSessions() {
	m.mu.Lock() // Get a write lock because we're modifying the map.
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if session.ExpiresAt.Before(now) {
			delete(m.sessions, id) // Remove expired session.
			log.Printf("Session %s expired and was removed", id)
		}
	}
}

// StartCleanup runs a goroutine to clean up expired sessions periodically.
func (m *SessionManager) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop() // Ensure the ticker is stopped when the function exits.
		for range ticker.C {
			m.CleanupExpiredSessions()
		}
	}()
}
