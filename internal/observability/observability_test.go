package observability

import (
	"testing"
	"time"
)

func TestRecordAndGetEntries(t *testing.T) {
	logger := NewLogger(10)

	logger.Record(LogEntry{
		RequestID: "req-1",
		TaskType:  "debugging",
		RouteKey:  "route.debugging",
		Alias:     "combo.debugging",
		Status:    "success",
	})

	entries := logger.GetEntries(10)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].RequestID != "req-1" {
		t.Errorf("expected request ID req-1, got %s", entries[0].RequestID)
	}
	if entries[0].TaskType != "debugging" {
		t.Errorf("expected task type debugging, got %s", entries[0].TaskType)
	}
}

func TestRecordOverflow(t *testing.T) {
	logger := NewLogger(3)

	for i := 0; i < 5; i++ {
		logger.Record(LogEntry{RequestID: "req-" + string(rune('A'+i))})
	}

	entries := logger.GetEntries(10)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	if entries[0].RequestID != "req-C" {
		t.Errorf("expected oldest entry req-C, got %s", entries[0].RequestID)
	}
	if entries[2].RequestID != "req-E" {
		t.Errorf("expected newest entry req-E, got %s", entries[2].RequestID)
	}
}

func TestGetEntriesSince(t *testing.T) {
	logger := NewLogger(10)

	now := time.Now()
	logger.Record(LogEntry{RequestID: "req-1", Timestamp: now.Add(-2 * time.Hour)})
	logger.Record(LogEntry{RequestID: "req-2", Timestamp: now.Add(-30 * time.Minute)})
	logger.Record(LogEntry{RequestID: "req-3", Timestamp: now})

	entries := logger.GetEntriesSince(now.Add(-1 * time.Hour))
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].RequestID != "req-2" {
		t.Errorf("expected req-2, got %s", entries[0].RequestID)
	}
	if entries[1].RequestID != "req-3" {
		t.Errorf("expected req-3, got %s", entries[1].RequestID)
	}
}

func TestClear(t *testing.T) {
	logger := NewLogger(10)

	logger.Record(LogEntry{RequestID: "req-1"})
	logger.Record(LogEntry{RequestID: "req-2"})

	if logger.Count() != 2 {
		t.Fatalf("expected 2 entries, got %d", logger.Count())
	}

	logger.Clear()

	if logger.Count() != 0 {
		t.Fatalf("expected 0 entries after clear, got %d", logger.Count())
	}
}

func TestCount(t *testing.T) {
	logger := NewLogger(10)

	if logger.Count() != 0 {
		t.Errorf("expected 0, got %d", logger.Count())
	}

	logger.Record(LogEntry{RequestID: "req-1"})
	if logger.Count() != 1 {
		t.Errorf("expected 1, got %d", logger.Count())
	}

	logger.Record(LogEntry{RequestID: "req-2"})
	if logger.Count() != 2 {
		t.Errorf("expected 2, got %d", logger.Count())
	}
}

func TestGetEntriesLimit(t *testing.T) {
	logger := NewLogger(100)

	for i := 0; i < 10; i++ {
		logger.Record(LogEntry{RequestID: "req-" + string(rune('A'+i))})
	}

	entries := logger.GetEntries(5)
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}

	if entries[0].RequestID != "req-F" {
		t.Errorf("expected req-F, got %s", entries[0].RequestID)
	}
}

func TestGetEntriesAll(t *testing.T) {
	logger := NewLogger(100)

	logger.Record(LogEntry{RequestID: "req-1"})
	logger.Record(LogEntry{RequestID: "req-2"})

	entries := logger.GetEntries(0)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries with limit 0, got %d", len(entries))
	}
}

func TestTimestampAutoFill(t *testing.T) {
	logger := NewLogger(10)

	before := time.Now()
	logger.Record(LogEntry{RequestID: "req-1"})
	after := time.Now()

	entries := logger.GetEntries(1)
	if entries[0].Timestamp.Before(before) || entries[0].Timestamp.After(after) {
		t.Errorf("timestamp not auto-filled correctly")
	}
}

func TestDefaultMaxEntries(t *testing.T) {
	logger := NewLogger(0)
	if logger.maxSize != DefaultMaxEntries {
		t.Errorf("expected default max size %d, got %d", DefaultMaxEntries, logger.maxSize)
	}
}
