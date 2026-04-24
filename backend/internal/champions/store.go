package champions

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand/v2"
)

//go:embed champions.json
var rawData []byte

type Champion struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Gender      string   `json:"gender"`
	Positions   []string `json:"positions"`
	Species     string   `json:"species"`
	Resource    string   `json:"resource"`
	RangeType   string   `json:"rangeType"`
	Regions     []string `json:"regions"`
	ReleaseYear int      `json:"releaseYear"`
}

type Store struct {
	list []Champion
	byID map[string]Champion
}

func NewStore() (*Store, error) {
	var list []Champion
	if err := json.Unmarshal(rawData, &list); err != nil {
		return nil, fmt.Errorf("unmarshal champions.json: %w", err)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("champions.json is empty")
	}
	byID := make(map[string]Champion, len(list))
	for _, c := range list {
		byID[c.ID] = c
	}
	return &Store{list: list, byID: byID}, nil
}

func (s *Store) All() []Champion {
	return s.list
}

func (s *Store) ByID(id string) (Champion, bool) {
	c, ok := s.byID[id]
	return c, ok
}

func (s *Store) Random() Champion {
	return s.list[rand.IntN(len(s.list))]
}
