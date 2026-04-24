package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

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

func TestCreateGame_returnsGameID(t *testing.T) {
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	rr := httptest.NewRecorder()
	h.CreateGame(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["gameId"] == "" {
		t.Error("expected non-empty gameId")
	}
}

func TestSubmitGuess_returnsFeedbackForKnownGame(t *testing.T) {
	h := newHandler(t)

	// Create game directly via Sessions to control the target
	g := h.Sessions.Create("ahri")

	body := bytes.NewBufferString(`{"championId":"yasuo"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+g.ID+"/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", g.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Correct      bool `json:"correct"`
		AttemptCount int  `json:"attemptCount"`
		Guess        struct {
			ID string `json:"id"`
		} `json:"guess"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Correct {
		t.Error("expected correct=false (yasuo != ahri)")
	}
	if resp.AttemptCount != 1 {
		t.Errorf("AttemptCount = %d, want 1", resp.AttemptCount)
	}
	if resp.Guess.ID != "yasuo" {
		t.Errorf("Guess.ID = %s, want yasuo", resp.Guess.ID)
	}
}

func TestSubmitGuess_correctChampionWinsGame(t *testing.T) {
	h := newHandler(t)
	g := h.Sessions.Create("ahri")

	body := bytes.NewBufferString(`{"championId":"ahri"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+g.ID+"/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", g.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	var resp struct{ Correct bool `json:"correct"` }
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if !resp.Correct {
		t.Error("expected correct=true for matching guess")
	}
}

func TestSubmitGuess_returns404ForUnknownGame(t *testing.T) {
	h := newHandler(t)
	body := bytes.NewBufferString(`{"championId":"ahri"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/nonexistent/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestSubmitGuess_returns400ForUnknownChampion(t *testing.T) {
	h := newHandler(t)
	g := h.Sessions.Create("ahri")
	body := bytes.NewBufferString(`{"championId":"nonexistent"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+g.ID+"/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", g.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestSubmitGuess_returns409WhenAlreadyWon(t *testing.T) {
	h := newHandler(t)
	g := h.Sessions.Create("ahri")
	g.Won = true

	body := bytes.NewBufferString(`{"championId":"ahri"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+g.ID+"/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", g.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rr.Code)
	}
}

func TestListChampions_includesImageKey(t *testing.T) {
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/champions", nil)
	rr := httptest.NewRecorder()
	h.ListChampions(rr, req)

	var body []map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body[0]["imageKey"] == "" {
		t.Errorf("expected non-empty imageKey on first entry, got %+v", body[0])
	}
}
