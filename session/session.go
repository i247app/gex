package session

type ISession interface {
	Put(key string, value any)
	Get(key string) (any, bool)
}

type Manager struct {
	sessions map[string]ISession
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]ISession),
	}
}

func (s *Manager) Session(sessionKey string) (ISession, bool) {
	session, ok := s.sessions[sessionKey]
	if !ok {
		return nil, false
	}
	return session, true
}

func (s *Manager) Sessions() *map[string]ISession {
	return &s.sessions
}

// InitSession is used to initialize a session with a given key.
// It accepts a session object to initialize the session with.
func (s *Manager) InitSession(sessionKey string, sess ISession) (ISession, bool) {
	if _, ok := s.sessions[sessionKey]; ok {
		return nil, false
	}

	s.sessions[sessionKey] = sess
	return sess, true
}

func (s *Manager) DeleteSession(sessionKey string) {
	delete(s.sessions, sessionKey)
}
