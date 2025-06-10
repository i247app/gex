package session

import "sync"

type InMemorySession struct {
	Data      map[string]any
	dataMutex sync.Mutex
}

func NewInMemorySession() *InMemorySession {
	return &InMemorySession{Data: make(map[string]any)}
}

func (s *InMemorySession) Put(key string, value any) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	s.Data[key] = value
}

func (s *InMemorySession) Get(key string) (any, bool) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	value, ok := s.Data[key]
	return value, ok
}
