package champions

import (
	"testing"
)

func TestStore_All_returnsAllChampions(t *testing.T) {
	s, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	all := s.All()
	if len(all) < 20 {
		t.Errorf("expected at least 20 champions, got %d", len(all))
	}
}

func TestStore_ByID_returnsChampionWhenFound(t *testing.T) {
	s, _ := NewStore()
	c, ok := s.ByID("ahri")
	if !ok {
		t.Fatal("expected to find ahri")
	}
	if c.Name != "Ahri" {
		t.Errorf("expected name Ahri, got %s", c.Name)
	}
}

func TestStore_ByID_returnsFalseWhenNotFound(t *testing.T) {
	s, _ := NewStore()
	if _, ok := s.ByID("nonexistent"); ok {
		t.Error("expected ok=false for unknown id")
	}
}

func TestStore_Random_returnsChampionFromList(t *testing.T) {
	s, _ := NewStore()
	c := s.Random()
	if c.ID == "" {
		t.Error("expected non-empty ID")
	}
	if _, ok := s.ByID(c.ID); !ok {
		t.Error("expected random champion to exist in store")
	}
}

func TestStore_ImageKey_isPopulated(t *testing.T) {
	s, _ := NewStore()
	ahri, ok := s.ByID("ahri")
	if !ok {
		t.Fatal("expected ahri")
	}
	if ahri.ImageKey != "Ahri" {
		t.Errorf("ImageKey = %q, want %q", ahri.ImageKey, "Ahri")
	}
	leeSin, _ := s.ByID("lee-sin")
	if leeSin.ImageKey != "LeeSin" {
		t.Errorf("LeeSin ImageKey = %q, want %q", leeSin.ImageKey, "LeeSin")
	}
}
