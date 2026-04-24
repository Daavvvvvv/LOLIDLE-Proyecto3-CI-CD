package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Game struct {
	ID           string
	TargetID     string
	Attempts     int
	Won          bool
	LastAccessed time.Time
}

type Store struct {
	mu    sync.Mutex
	games map[string]*Game
	ttl   time.Duration
}

func NewStore(ttl time.Duration) *Store {
	return &Store{
		games: make(map[string]*Game),
		ttl:   ttl,
	}
}

func (s *Store) Create(targetID string) *Game {
	s.mu.Lock()
	defer s.mu.Unlock()
	g := &Game{
		ID:           newID(),
		TargetID:     targetID,
		LastAccessed: time.Now(),
	}
	s.games[g.ID] = g
	return g
}

func (s *Store) Get(id string) (*Game, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.games[id]
	if !ok {
		return nil, false
	}
	if time.Since(g.LastAccessed) > s.ttl {
		delete(s.games, id)
		return nil, false
	}
	g.LastAccessed = time.Now()
	return g, true
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
