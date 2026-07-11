package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestStructuredLoggerWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf, 100, 7)

	logger.LogRequest("req-123", "POST", "/v1/chat/completions", 200, 150, 1024, 2048, "route.default", "success")

	line := buf.String()
	line = strings.TrimSpace(line)
	if line == "" {
		t.Fatal("expected log output, got empty")
	}

	var entry StructuredEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nline: %s", err, line)
	}

	if entry.CorrelationID != "req-123" {
		t.Errorf("expected correlation_id req-123, got %s", entry.CorrelationID)
	}
	if entry.Status != 200 {
		t.Errorf("expected status 200, got %d", entry.Status)
	}
	if entry.LatencyMs != 150 {
		t.Errorf("expected latency_ms 150, got %d", entry.LatencyMs)
	}
	if entry.Route != "route.default" {
		t.Errorf("expected route route.default, got %s", entry.Route)
	}
}

func TestStructuredLoggerDoesNotLogRawPrompt(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf, 100, 7)

	logger.Log(StructuredEntry{
		Level:         LevelInfo,
		CorrelationID: "req-456",
		Message:       "request processed",
	})

	line := buf.String()
	if strings.Contains(line, "password") || strings.Contains(line, "secret") {
		t.Error("structured log should not contain sensitive prompt data")
	}

	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &entry); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
}

func TestStructuredLoggerRetention(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf, 5, 7)

	for i := 0; i < 10; i++ {
		logger.LogRequest("req-"+string(rune('0'+i)), "GET", "/test", 200, 10, 0, 0, "", "")
	}

	if logger.Count() > 5 {
		t.Errorf("expected max 5 entries, got %d", logger.Count())
	}
}

func TestStructuredLoggerClear(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf, 100, 7)

	logger.LogRequest("req-1", "GET", "/test", 200, 10, 0, 0, "", "")
	if logger.Count() != 1 {
		t.Fatalf("expected 1 entry, got %d", logger.Count())
	}

	logger.Clear()
	if logger.Count() != 0 {
		t.Errorf("expected 0 entries after clear, got %d", logger.Count())
	}
}

func TestStructuredLoggerError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf, 100, 7)

	logger.LogError("req-789", nil, "something went wrong")

	line := strings.TrimSpace(buf.String())
	var entry StructuredEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
	if entry.Level != LevelError {
		t.Errorf("expected level error, got %s", entry.Level)
	}
	if entry.Error != "" {
		t.Errorf("expected empty error for nil, got %q", entry.Error)
	}
}
