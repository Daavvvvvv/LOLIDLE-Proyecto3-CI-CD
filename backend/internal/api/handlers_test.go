package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/session"
)

func newHandler(t *testing.T) *Handler {
	t.Helper()
	cs, err := champions.NewStore()
	if err != nil {
		t.Fatalf("champions store: %v", err)
	}
	return &Handler{
		Champions: cs,
		Sessions:  session.NewStore(time.Minute),
	}
}

func TestHealth_returnsOK(t *testing.T) {
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %s, want ok", body["status"])
	}
}

func TestListChampions_returnsIDAndName(t *testing.T) {
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/champions", nil)
	rr := httptest.NewRecorder()
	h.ListChampions(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var body []map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) < 20 {
		t.Errorf("expected at least 20 champions, got %d", len(body))
	}
	if body[0]["id"] == "" || body[0]["name"] == "" {
		t.Errorf("expected non-empty id and name, got %+v", body[0])
	}
}
