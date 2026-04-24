package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"lolidle/backend/internal/api"
	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/session"
)

func main() {
	cs, err := champions.NewStore()
	if err != nil {
		log.Fatalf("load champions: %v", err)
	}
	ss := session.NewMemoryStore(30 * time.Minute)

	h := &api.Handler{Champions: cs, Sessions: ss}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/api/health", h.Health)
	r.Get("/api/champions", h.ListChampions)
	r.Post("/api/games", h.CreateGame)
	r.Post("/api/games/{gameId}/guesses", h.SubmitGuess)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	allowed := os.Getenv("CORS_ORIGIN")
	if allowed == "" {
		allowed = "http://localhost:5173"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
