package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/game"
	"lolidle/backend/internal/session"
)

type Handler struct {
	Champions *champions.Store
	Sessions  *session.Store
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type championListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) ListChampions(w http.ResponseWriter, r *http.Request) {
	all := h.Champions.All()
	out := make([]championListItem, 0, len(all))
	for _, c := range all {
		out = append(out, championListItem{ID: c.ID, Name: c.Name})
	}
	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

type createGameResponse struct {
	GameID string `json:"gameId"`
}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	target := h.Champions.Random()
	g := h.Sessions.Create(target.ID)
	writeJSON(w, http.StatusCreated, createGameResponse{GameID: g.ID})
}

type guessRequest struct {
	ChampionID string `json:"championId"`
}

type guessResponse struct {
	Guess        champions.Champion `json:"guess"`
	Feedback     game.Feedback      `json:"feedback"`
	Correct      bool               `json:"correct"`
	AttemptCount int                `json:"attemptCount"`
}

func (h *Handler) SubmitGuess(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameId")

	g, ok := h.Sessions.Get(gameID)
	if !ok {
		writeError(w, http.StatusNotFound, "game not found or expired")
		return
	}
	if g.Won {
		writeError(w, http.StatusConflict, "game already won")
		return
	}

	var req guessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	guess, ok := h.Champions.ByID(req.ChampionID)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown champion")
		return
	}

	target, _ := h.Champions.ByID(g.TargetID)
	fb, correct := game.Compare(guess, target)
	g.Attempts++
	if correct {
		g.Won = true
	}

	writeJSON(w, http.StatusOK, guessResponse{
		Guess:        guess,
		Feedback:     fb,
		Correct:      correct,
		AttemptCount: g.Attempts,
	})
}
