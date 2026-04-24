package observability

import (
	"io"
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	return NewLoggerForWriter(os.Stdout)
}

func NewLoggerForWriter(w io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler).With(
		slog.String("service", "lolidle-backend"),
	)
}
