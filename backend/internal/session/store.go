package session

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("game not found or expired")

type Game struct {
	ID           string
	TargetID     string
	Attempts     int
	Won          bool
	LastAccessed time.Time
}

type Store interface {
	Create(targetID string) (*Game, error)
	Get(id string) (*Game, error)
	Update(g *Game) error
}
