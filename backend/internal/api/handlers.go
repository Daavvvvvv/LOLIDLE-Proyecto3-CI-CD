package api

import (
	"encoding/json"
	"net/http"

	"lolidle/backend/internal/champions"
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
