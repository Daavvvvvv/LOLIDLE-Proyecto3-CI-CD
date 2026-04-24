package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"lolidle/backend/internal/api"
	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/lore"
	"lolidle/backend/internal/observability"
	"lolidle/backend/internal/session"
)

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

func main() {
	logger := observability.NewLogger()

	cs, err := champions.NewStore()
	if err != nil {
		logger.Error("load champions failed", "err", err)
		os.Exit(1)
	}

	var ss session.Store
	var lo *lore.Service

	switch os.Getenv("STORE_BACKEND") {
	case "dynamodb":
		dynClient, err := newDynamoClient()
		if err != nil {
			logger.Error("dynamo client failed", "err", err)
			os.Exit(1)
		}
		sessTable := os.Getenv("SESSIONS_TABLE")
		ss = session.NewDynamoDBStore(dynClient, sessTable, 30*time.Minute)
		loreTable := os.Getenv("LORE_CACHE_TABLE")
		if loreTable != "" {
			cache := lore.NewDynamoDBCache(dynClient, loreTable)
			lo = lore.New(geminiEndpoint, os.Getenv("GEMINI_API_KEY"), cache)
		}
	default:
		if backend := os.Getenv("STORE_BACKEND"); backend != "" {
			logger.Warn("unknown STORE_BACKEND, falling back to memory store", "backend", backend)
		}
		ss = session.NewMemoryStore(30 * time.Minute)
	}

	h := &api.Handler{Champions: cs, Sessions: ss, Lore: lo}

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
	logger.Info("listening", "addr", addr, "store", os.Getenv("STORE_BACKEND"))
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func newDynamoClient() (*dynamodb.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(getenvDefault("AWS_REGION", "us-east-1")),
	)
	if err != nil {
		return nil, err
	}
	endpoint := os.Getenv("DYNAMO_ENDPOINT")
	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	}), nil
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
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
