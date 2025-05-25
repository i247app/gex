package session

import (
	"fmt"
	"net/http"
	"sync"
)

type ISession interface {
	Put(key string, value any)
	Get(key string) (any, bool)
}

type Session struct {
	bucket *sync.Map
}

func (s *Session) Put(key string, value any) {
	s.bucket.Store(key, value)
}

func (s *Session) Get(key string) (any, bool) {
	return s.bucket.Load(key)
}

func (s *Session) UserID() (int64, bool) {
	result, ok := s.Get("user_id")
	if !ok {
		fmt.Println("ERROR: key 'user_id' not in session store")
		return 0, false
	}

	userID, ok := result.(int64)
	if !ok {
		fmt.Printf("ERROR: 'user_id' is in session store but expected to be an int64, is a %T\n", result)
		return 0, false
	}

	return userID, true
}

// AuthTokenLocator is an interface for locating the session token in the request
// This allows for different implementations, such as JWT, API Key, etc.
type AuthTokenLocator interface {
	Locate(r *http.Request) (string, error)
}

type Manager struct {
	tokenLocator AuthTokenLocator
	sessionCache map[string]*sync.Map
}

func NewManager(tokenLocator AuthTokenLocator) *Manager {
	return &Manager{
		tokenLocator: tokenLocator,
		sessionCache: make(map[string]*sync.Map),
	}
}

func (s *Manager) GetSession(sessionKey string) (*Session, bool) {
	bucket, ok := s.sessionCache[sessionKey]
	if !ok {
		return nil, false
	}
	return &Session{bucket: bucket}, true
}

func (s *Manager) GetSessionFromRequest(r *http.Request) (*Session, error) {
	// Get JWT
	sessToken, err := s.tokenLocator.Locate(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth header: %v", err)
	}

	// Check cache
	sess, ok := s.GetSession(sessToken)
	if !ok {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}

	return sess, nil
}

func (s *Manager) CreateSession(sessionKey string) *Session {
	return s.CreateSessionFromMap(sessionKey, &map[string]any{})
}

func (s *Manager) CreateSessionFromMap(sessionKey string, data *map[string]any) *Session {
	sess := Session{bucket: new(sync.Map)}
	for k, v := range *data {
		sess.Put(k, v)
	}

	// Store bucket in sessionCache (Where session data is actually kept)
	s.sessionCache[sessionKey] = sess.bucket

	return &sess
}

func (s *Manager) DeleteSession(sessionKey string) {
	delete(s.sessionCache, sessionKey)
}

func (s *Manager) Dump() *map[string]map[string]any {
	result := make(map[string]map[string]any)

	for k, v := range s.sessionCache {
		result[k] = make(map[string]any)

		v.Range(func(key any, value any) bool {
			result[k][key.(string)] = value
			return true
		})
	}

	return &result
}
