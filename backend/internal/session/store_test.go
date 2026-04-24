package session

import (
	"testing"
	"time"
)

func TestStore_Create_returnsGameWithUniqueID(t *testing.T) {
	s := NewStore(time.Minute)
	g1 := s.Create("ahri")
	g2 := s.Create("ahri")
	if g1.ID == "" || g2.ID == "" {
		t.Fatal("expected non-empty IDs")
	}
	if g1.ID == g2.ID {
		t.Error("expected unique IDs across Create calls")
	}
	if g1.TargetID != "ahri" {
		t.Errorf("TargetID = %s, want ahri", g1.TargetID)
	}
}

func TestStore_Get_returnsCreatedGame(t *testing.T) {
	s := NewStore(time.Minute)
	g := s.Create("yasuo")

	got, ok := s.Get(g.ID)
	if !ok {
		t.Fatal("expected to find created game")
	}
	if got.TargetID != "yasuo" {
		t.Errorf("TargetID = %s, want yasuo", got.TargetID)
	}
}

func TestStore_Get_returnsFalseForUnknownID(t *testing.T) {
	s := NewStore(time.Minute)
	if _, ok := s.Get("nonexistent"); ok {
		t.Error("expected ok=false for unknown id")
	}
}

func TestStore_Get_returnsFalseForExpiredGame(t *testing.T) {
	s := NewStore(10 * time.Millisecond)
	g := s.Create("ahri")

	time.Sleep(20 * time.Millisecond)

	if _, ok := s.Get(g.ID); ok {
		t.Error("expected expired game to be evicted")
	}
}

func TestStore_recordsAttemptsAndWin(t *testing.T) {
	s := NewStore(time.Minute)
	g := s.Create("ahri")

	g.Attempts++
	g.Attempts++
	g.Won = true

	got, _ := s.Get(g.ID)
	if got.Attempts != 2 {
		t.Errorf("Attempts = %d, want 2", got.Attempts)
	}
	if !got.Won {
		t.Error("expected Won=true")
	}
}
