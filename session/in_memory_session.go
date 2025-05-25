package session

type InMemorySession struct {
	Data map[string]any
}

func NewInMemorySession() *InMemorySession {
	return &InMemorySession{Data: make(map[string]any)}
}

func (s *InMemorySession) Put(key string, value any) {
	s.Data[key] = value
}

func (s *InMemorySession) Get(key string) (any, bool) {
	value, ok := s.Data[key]
	return value, ok
}
