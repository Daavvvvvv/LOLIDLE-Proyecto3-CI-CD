package lore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeCache struct {
	store map[string]string
}

func newFakeCache() *fakeCache { return &fakeCache{store: map[string]string{}} }
func (c *fakeCache) Get(_ context.Context, id string) (string, bool, error) {
	v, ok := c.store[id]
	return v, ok, nil
}
func (c *fakeCache) Put(_ context.Context, id, lore string) error {
	c.store[id] = lore
	return nil
}

func TestService_Generate_returnsCachedValueOnHit(t *testing.T) {
	cache := newFakeCache()
	cache.store["ahri"] = "cached lore"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Gemini should not be called on cache hit")
	}))
	defer server.Close()

	s := New(server.URL, "test-key", cache)
	out, err := s.Generate(context.Background(), "ahri", "Ahri")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out != "cached lore" {
		t.Errorf("got %q, want %q", out, "cached lore")
	}
}

func TestService_Generate_callsGeminiOnMissAndCaches(t *testing.T) {
	cache := newFakeCache()
	called := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{
					"parts": []map[string]any{{"text": "generated lore for Ahri"}},
				},
			}},
		})
	}))
	defer server.Close()

	s := New(server.URL, "test-key", cache)
	out, err := s.Generate(context.Background(), "ahri", "Ahri")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if called != 1 {
		t.Errorf("Gemini called %d times, want 1", called)
	}
	if out != "generated lore for Ahri" {
		t.Errorf("got %q, want generated lore", out)
	}
	if cache.store["ahri"] != "generated lore for Ahri" {
		t.Errorf("cache not populated, got %q", cache.store["ahri"])
	}
}

func TestService_Generate_returnsEmptyOnGeminiError(t *testing.T) {
	cache := newFakeCache()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	s := New(server.URL, "test-key", cache)
	out, err := s.Generate(context.Background(), "ahri", "Ahri")
	if err != nil {
		t.Fatalf("Generate should not return error on Gemini failure: %v", err)
	}
	if out != "" {
		t.Errorf("got %q, want empty string on error", out)
	}
}

func TestService_Generate_returnsEmptyWhenAPIKeyMissing(t *testing.T) {
	s := New("http://unused", "", newFakeCache())
	out, err := s.Generate(context.Background(), "ahri", "Ahri")
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if out != "" {
		t.Errorf("got %q, want empty string when api key missing", out)
	}
}
