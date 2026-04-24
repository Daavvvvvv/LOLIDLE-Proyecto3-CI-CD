package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type MemoryStore struct {
	mu    sync.Mutex
	games map[string]*Game
	ttl   time.Duration
}

func NewMemoryStore(ttl time.Duration) *MemoryStore {
	return &MemoryStore{
		games: make(map[string]*Game),
		ttl:   ttl,
	}
}

func (s *MemoryStore) Create(targetID string) (*Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g := &Game{
		ID:           newID(),
		TargetID:     targetID,
		LastAccessed: time.Now(),
	}
	s.games[g.ID] = g
	return g, nil
}

func (s *MemoryStore) Get(id string) (*Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.games[id]
	if !ok {
		return nil, ErrNotFound
	}
	if time.Since(g.LastAccessed) > s.ttl {
		delete(s.games, id)
		return nil, ErrNotFound
	}
	g.LastAccessed = time.Now()
	return g, nil
}

func (s *MemoryStore) Update(g *Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[g.ID]; !ok {
		return ErrNotFound
	}
	s.games[g.ID] = g
	return nil
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
