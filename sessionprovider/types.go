package sessionprovider

import (
	"errors"

	"github.com/i247app/gex/session"
)

var (
	ErrMalformedJwt = errors.New("invalid or malformed JWT")
)

type SessionFactory func() session.SessionStorer

// SessionResult contains the session and metadata about the session retrieval
type SessionResult struct {
	Session        session.SessionStorer
	DidAutoRefresh bool
	AuthToken      string
}
