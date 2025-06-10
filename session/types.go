package session

import "fmt"

var log = fmt.Println

type SessionFactory func() SessionStorer

// SessionResult contains the session and metadata about the session retrieval
type SessionResult struct {
	Session        SessionStorer
	DidAutoRefresh bool
	AuthToken      string
}
