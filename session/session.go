package session

import "sync"

type SessionRequestContextKey string

// SessionStorer is an interface that defines the methods for a session store.
type SessionStorer interface {
	Put(key string, value any)
	Get(key string) (any, bool)
}

// Container is a container for sessions.
type Container struct {
	sessions      map[string]SessionStorer
	sessionsMutex sync.Mutex
}

func NewContainer() *Container {
	return &Container{sessions: make(map[string]SessionStorer)}
}

// Session is used to get a session from the container.
func (s *Container) Session(sessionKey string) (SessionStorer, bool) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	session, ok := s.sessions[sessionKey]
	if !ok {
		return nil, false
	}
	return session, true
}

// Sessions is used to get all sessions from the container.
func (s *Container) Sessions() *map[string]SessionStorer {
	return &s.sessions
}

// InitSession is used to initialize a session with a given key.
// It accepts a session object to initialize the session with.
func (s *Container) InitSession(sessionKey string, sess SessionStorer) (SessionStorer, bool) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	if _, ok := s.sessions[sessionKey]; ok {
		return nil, false
	}

	s.sessions[sessionKey] = sess
	return sess, true
}

func (s *Container) DeleteSession(sessionKey string) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	delete(s.sessions, sessionKey)
}
