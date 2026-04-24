package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestNewLogger_emitsJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerForWriter(&buf)
	logger.Info("test message", slog.String("key", "value"))

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if parsed["msg"] != "test message" {
		t.Errorf("msg = %v, want test message", parsed["msg"])
	}
	if parsed["key"] != "value" {
		t.Errorf("key = %v, want value", parsed["key"])
	}
	if parsed["level"] != "INFO" {
		t.Errorf("level = %v, want INFO", parsed["level"])
	}
}

func TestNewLogger_includesService(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerForWriter(&buf)
	logger.Info("hi")
	var parsed map[string]any
	_ = json.Unmarshal(buf.Bytes(), &parsed)
	if parsed["service"] != "lolidle-backend" {
		t.Errorf("service = %v, want lolidle-backend", parsed["service"])
	}
}
