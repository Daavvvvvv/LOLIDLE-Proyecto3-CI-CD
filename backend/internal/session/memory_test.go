package session

import (
	"errors"
	"testing"
	"time"
)

func TestMemoryStore_Create_returnsGameWithUniqueID(t *testing.T) {
	s := NewMemoryStore(time.Minute)
	g1, err := s.Create("ahri")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	g2, _ := s.Create("ahri")
	if g1.ID == "" || g2.ID == "" {
		t.Fatal("expected non-empty IDs")
	}
	if g1.ID == g2.ID {
		t.Error("expected unique IDs")
	}
	if g1.TargetID != "ahri" {
		t.Errorf("TargetID = %s, want ahri", g1.TargetID)
	}
}

func TestMemoryStore_Get_returnsCreatedGame(t *testing.T) {
	s := NewMemoryStore(time.Minute)
	g, _ := s.Create("yasuo")
	got, err := s.Get(g.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.TargetID != "yasuo" {
		t.Errorf("TargetID = %s, want yasuo", got.TargetID)
	}
}

func TestMemoryStore_Get_returnsErrNotFoundForUnknownID(t *testing.T) {
	s := NewMemoryStore(time.Minute)
	_, err := s.Get("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestMemoryStore_Get_returnsErrNotFoundForExpiredGame(t *testing.T) {
	s := NewMemoryStore(10 * time.Millisecond)
	g, _ := s.Create("ahri")
	time.Sleep(20 * time.Millisecond)
	_, err := s.Get(g.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestMemoryStore_Update_persistsChanges(t *testing.T) {
	s := NewMemoryStore(time.Minute)
	g, _ := s.Create("ahri")
	g.Attempts = 3
	g.Won = true
	if err := s.Update(g); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := s.Get(g.ID)
	if got.Attempts != 3 {
		t.Errorf("Attempts = %d, want 3", got.Attempts)
	}
	if !got.Won {
		t.Error("expected Won=true")
	}
}

func TestMemoryStore_Update_returnsErrNotFoundForUnknownGame(t *testing.T) {
	s := NewMemoryStore(time.Minute)
	g := &Game{ID: "nonexistent", TargetID: "ahri"}
	if err := s.Update(g); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
