// internal/session/session_manager.go
package session

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// ChangeType represents the type of session change.
type ChangeType string

const (
	ChangeTypeCreate ChangeType = "create"
	ChangeTypeUpdate ChangeType = "update"
	ChangeTypeDelete ChangeType = "delete"
)

// SessionChange represents a change to a session.
type SessionChange struct {
	Timestamp    time.Time  `json:"timestamp"`
	Type         ChangeType `json:"type"`
	SessionToken string     `json:"session_token,omitempty"` // Use sessionToken instead of Session
	Session      *Session   `json:"session,omitempty"`
}

// Session represents a user session.
type Session struct {
	Token      string    `json:"token"`
	IsSecure   bool      `json:"is_secure"`
	UserID     string    `json:"user_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	ModifyDate time.Time `json:"modify_date"` // Track modification time
}

// NewSession creates a new session.
func NewSession(token string, isSecure bool, userID string, expiresAt time.Time) *Session {
	return &Session{
		Token:      token,
		IsSecure:   isSecure,
		UserID:     userID,
		ExpiresAt:  expiresAt,
		ModifyDate: time.Now(), // Initialize ModifyDate
	}
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

// SessionManager manages user sessions.
type SessionManager struct {
	sessions      sync.Map // Use sync.Map for concurrent access
	sessionTimeout time.Duration
	jwtSecret      string
	changeLog     []SessionChange // Change log
	changeLogLock sync.RWMutex    // Protect changeLog
	nodeCommunicator NodeCommunicator
	nodeID           string
	lastSyncTime    sync.Map //map[string]time.Time // Keep track of last sync time for each node.  Use sync.Map
}

// NodeCommunicator defines the interface for node-to-node communication.
type NodeCommunicator interface {
	SendChanges(nodeID string, changes []SessionChange) error
	BroadcastChanges(changes []SessionChange) error
	GetNodeID() string
	GetOtherNodeIDs() []string
}

// NewSessionManager creates a new session manager.
func NewSessionManager(sessionTimeout time.Duration, jwtSecret string, communicator NodeCommunicator, nodeID string) *SessionManager {
	sm := &SessionManager{
		sessions:      sync.Map{},
		sessionTimeout: sessionTimeout,
		jwtSecret:      jwtSecret,
		changeLog:     []SessionChange{},
		nodeCommunicator: communicator,
		nodeID:           nodeID,
		lastSyncTime:    sync.Map{},
	}
	go sm.cleanupExpiredSessions()
	return sm
}

// CreateSession creates a new session and returns the token.
func (sm *SessionManager) CreateSession(userID string, isSecure bool) (*Session, string, error) {
	now := time.Now()
	expiresAt := now.Add(sm.sessionTimeout)
	claims := map[string]interface{}{
		"user_id":   userID,
		"is_secure": isSecure,
		"exp":       expiresAt.Unix(),
		"iat":       now.Unix(),
		"nbf":       now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenString, err := token.SignedString([]byte(sm.jwtSecret))
	if err != nil {
		return nil, "", err
	}

	session := NewSession(tokenString, isSecure, userID, expiresAt)
	sm.SetSession(session)
	sm.addChangeLogEntry(SessionChange{
		Timestamp:    now,
		Type:         ChangeTypeCreate,
		SessionToken: session.Token, // Store the token.
		Session:      session,
	})
	//sm.broadcastChanges() // Remove this line for pull model
	return session, tokenString, nil
}

// SetSession stores a session.
func (sm *SessionManager) SetSession(session *Session) {
	sm.sessions.Store(session.Token, session)
	sm.addChangeLogEntry(SessionChange{
		Timestamp:    time.Now(),
		Type:         ChangeTypeUpdate,
		SessionToken: session.Token, // Store the token
		Session:      session,
	})
	//sm.broadcastChanges()  // Remove this line for pull model
}

// GetSession retrieves a session by its token.
func (sm *SessionManager) GetSession(token string) *Session {
	sess, ok := sm.sessions.Load(token)
	if ok {
		s, ok := sess.(*Session)
		if !ok {
			log.Printf("unexpected session type")
			return nil
		}
		if !s.IsExpired() {
			return s
		}
		sm.DeleteSession(token)
		return nil
	}
	return nil
}

// DeleteSession deletes a session by its token.
func (sm *SessionManager) DeleteSession(token string) {
	sm.sessions.Delete(token)
	sm.addChangeLogEntry(SessionChange{
		Timestamp:    time.Now(),
		Type:         ChangeTypeDelete,
		SessionToken: token, // Store the token
	})
	//sm.broadcastChanges() // Remove this line for pull model
}

// cleanupExpiredSessions periodically removes expired sessions.
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.sessions.Range(func(key, value interface{}) bool {
			session, ok := value.(*Session)
			if !ok {
				log.Printf("unexpected session value type: %T, expected *Session", value)
				return true
			}
			if session.IsExpired() {
				sm.sessions.Delete(key)
				log.Printf("Expired session deleted: %s", session.Token)
				sm.addChangeLogEntry(SessionChange{
					Timestamp:    time.Now(),
					Type:         ChangeTypeDelete,
					SessionToken: session.Token,
				})
				//sm.broadcastChanges() // Remove this line for pull model
			}
			return true
		})
	}
}

func (sm *SessionManager) addChangeLogEntry(change SessionChange) {
	sm.changeLogLock.Lock()
	defer sm.changeLogLock.Unlock()
	sm.changeLog = append(sm.changeLog, change)
}

// GetChangesSince returns all session changes since the given timestamp.
// TODO - implement poll or trigger to pull
func (sm *SessionManager) GetChangesSince(since time.Time, nodeID string) []SessionChange {
	sm.changeLogLock.RLock()
	defer sm.changeLogLock.RUnlock()
	var changes []SessionChange
	for _, change := range sm.changeLog {
		if change.Timestamp.After(since) {
			changes = append(changes, change)
		}
	}
	sm.lastSyncTime.Store(nodeID, time.Now()) //update the last sync time
	return changes
}

// ApplyChanges applies a batch of session changes to the SessionManager.
func (sm *SessionManager) ApplyChanges(changes []SessionChange) {
	for _, change := range changes {
		switch change.Type {
		case ChangeTypeCreate:
			if change.Session != nil {
				sm.sessions.Store(change.Session.Token, change.Session)
			}
		case ChangeTypeUpdate:
			if change.Session != nil {
				if currentSession, ok := sm.sessions.Load(change.Session.Token); ok {
					// Only update if the session exists and the ModifyDate is newer
					if currentSession.(*Session).ModifyDate.Before(change.Session.ModifyDate) {
						sm.sessions.Store(change.Session.Token, change.Session)
					}
				}

			}
		case ChangeTypeDelete:
			sm.sessions.Delete(change.SessionToken) // Use SessionToken
		}
	}
}

// SessionsToJSON converts all sessions to a JSON string.
func (sm *SessionManager) SessionsToJSON() (string, error) {
	sessionMap := make(map[string]*Session)
	sm.sessions.Range(func(key, value interface{}) bool {
		token, ok := key.(string)
		if !ok {
			log.Printf("unexpected session key type: %T, expected string", key)
			return true
		}
		session, ok := value.(*Session)
		if !ok {
			log.Printf("unexpected session value type: %T, expected *Session", value)
			return true
		}
		sessionMap[token] = session
		return true
	})
	bytes, err := json.Marshal(sessionMap)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// SessionsFromJSON loads sessions from a JSON string.
func (sm *SessionManager) SessionsFromJSON(jsonStr string) error {
	var sessionMap map[string]*Session
	err := json.Unmarshal([]byte(jsonStr), &sessionMap)
	if err != nil {
		return err
	}
	sm.sessions = sync.Map{}
	for token, session := range sessionMap {
		sm.sessions.Store(token, session)
	}
	return nil
}

// broadcastChanges broadcasts the changes
// This is now for a different use case, like force syncing.
func (sm *SessionManager) broadcastChanges() {
	sm.changeLogLock.RLock()
	defer sm.changeLogLock.RUnlock()
	if sm.nodeCommunicator != nil {
		sm.nodeCommunicator.BroadcastChanges(sm.changeLog)
	}
}

// HandleChanges is called by the NodeCommunicator when it receives changes from another node.
func (sm *SessionManager) HandleChanges(changes []SessionChange) {
	sm.ApplyChanges(changes)
}

// GetLastSyncTime returns the last sync time for a given node.
func (sm *SessionManager) GetLastSyncTime(nodeID string) time.Time {
	val, ok := sm.lastSyncTime.Load(nodeID)
	if !ok {
		return time.Time{} // Return zero time if not found
	}
	lastSyncTime, ok := val.(time.Time)
	if !ok{
		return time.Time{}
	}
	return lastSyncTime
}

// ForceSync triggers a broadcast of all changes to all connected nodes.
// This is the new function for the use case where you want to force a sync.
func (sm *SessionManager) ForceSync() {
	sm.changeLogLock.RLock()
	defer sm.changeLogLock.RUnlock()
	if sm.nodeCommunicator != nil {
		sm.nodeCommunicator.BroadcastChanges(sm.changeLog)
	}
}
