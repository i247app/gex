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
		fmt.Println("ERROR: key 'user_id' NOT IN SESSION STORE")
		return 0, false
	}

	userID, ok := result.(int64)
	if !ok {
		fmt.Println("ERROR: 'user_id' is in session store but cant be converted to int64")
		return 0, false
	}

	return userID, true
}

/**
 * Interface for locating the session token in the request
 */
type TokenLocator interface {
	Locate(r *http.Request) (string, error)
}

type Manager struct {
	tokenLocator TokenLocator
	sessionCache map[string]*sync.Map
}

func NewManager(tokenLocator TokenLocator) *Manager {
	return &Manager{
		tokenLocator: tokenLocator,
		sessionCache: make(map[string]*sync.Map),
	}
}

func (s *Manager) GetSessionFromRequest(r *http.Request) (*Session, error) {
	return s.getSessionFromRequest(r)
}

func (s *Manager) CreateSession(key string, data *map[string]any) {
	sess := sync.Map{}
	for k, v := range *data {
		sess.Store(k, v)
	}

	s.sessionCache[key] = &sess
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

func (s *Manager) getSessionFromRequest(r *http.Request) (*Session, error) {
	// Get JWT
	sessToken, err := s.tokenLocator.Locate(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth header: %v", err)
	}

	// Check cache
	bucket, ok := s.sessionCache[sessToken]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	// fmt.Printf("sessToken %v\n", sessToken)
	// fmt.Printf("sessToken bucket %v\n", bucket)

	return &Session{bucket: bucket}, nil
}
