// pkg/session/session.go
package session

import (
	"time"
)

// Session represents a user session.  This struct
// holds the data associated with a logged-in user.
type Session struct {
	ID        string    // Unique session ID.
	UserID    string    // ID of the user associated with the session.
	CreatedAt time.Time // Time the session was created.
	ExpiresAt time.Time // Time the session expires.
	Data      map[string]interface{} // Store additional session-specific data.
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
