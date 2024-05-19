package rate_limiter

import "sync"

type InMemoryStorage struct {
	store sync.Map
}

func (s *InMemoryStorage) LoadOrStore(key, value any) (actual any, loaded bool) {
	return s.store.LoadOrStore(key, value)
}

func (s *InMemoryStorage) Store(key, value any) {
	s.store.Store(key, value)
}

func (s *InMemoryStorage) Reset() {
	s.store = sync.Map{}
}

func newStorage() *InMemoryStorage {
	return &InMemoryStorage{store: sync.Map{}}
}
