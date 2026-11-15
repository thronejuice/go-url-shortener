package link

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu    sync.RWMutex
	links map[string]Link
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{links: make(map[string]Link)}
}

func (s *MemoryStore) Create(_ context.Context, l Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.links[l.Code]; ok {
		return ErrConflict
	}
	s.links[l.Code] = l
	return nil
}

func (s *MemoryStore) Get(_ context.Context, code string) (Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return l, nil
}
