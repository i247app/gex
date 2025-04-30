// internal/session/session.go
package session

import (
	"encoding/json"
	"time"
)

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

// ToJSON converts the Session to a JSON string.
func (s *Session) ToJSON() (string, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
