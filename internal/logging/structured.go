package logging

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type StructuredEntry struct {
	Timestamp     string `json:"timestamp"`
	Level         Level  `json:"level"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Message       string `json:"message"`
	Status        int    `json:"status,omitempty"`
	LatencyMs     int64  `json:"latency_ms,omitempty"`
	BytesIn       int64  `json:"bytes_in,omitempty"`
	BytesOut      int64  `json:"bytes_out,omitempty"`
	Route         string `json:"route,omitempty"`
	Outcome       string `json:"outcome,omitempty"`
	Error         string `json:"error,omitempty"`
	Prompt        string `json:"prompt,omitempty"`
}

type StructuredLogger struct {
	mu                   sync.Mutex
	out                  io.Writer
	maxEntries           int
	entries              []StructuredEntry
	retentionDays        int
	promptLoggingEnabled bool
}

func NewStructuredLogger(out io.Writer, maxEntries, retentionDays int) *StructuredLogger {
	return NewStructuredLoggerWithPromptLogging(out, maxEntries, retentionDays, false)
}

func NewStructuredLoggerWithPromptLogging(out io.Writer, maxEntries, retentionDays int, promptLogging bool) *StructuredLogger {
	if out == nil {
		out = os.Stdout
	}
	if maxEntries <= 0 {
		maxEntries = 10000
	}
	if retentionDays <= 0 {
		retentionDays = 7
	}
	return &StructuredLogger{
		out:                  out,
		maxEntries:           maxEntries,
		retentionDays:        retentionDays,
		promptLoggingEnabled: promptLogging,
		entries:              make([]StructuredEntry, 0, maxEntries),
	}
}

func (s *StructuredLogger) Log(entry StructuredEntry) {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) >= s.maxEntries {
		s.evictOld()
	}

	s.entries = append(s.entries, entry)

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return
	}
	s.out.Write(data)
	s.out.Write([]byte("\n"))
}

func (s *StructuredLogger) LogRequest(correlationID, method, path string, status int, latencyMs, bytesIn, bytesOut int64, route, outcome string) {
	s.Log(StructuredEntry{
		Level:         LevelInfo,
		CorrelationID: correlationID,
		Message:       method + " " + path,
		Status:        status,
		LatencyMs:     latencyMs,
		BytesIn:       bytesIn,
		BytesOut:      bytesOut,
		Route:         route,
		Outcome:       outcome,
	})
}

func (s *StructuredLogger) LogError(correlationID string, err error, msg string) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	s.Log(StructuredEntry{
		Level:         LevelError,
		CorrelationID: correlationID,
		Message:       msg,
		Error:         errStr,
	})
}

func (s *StructuredLogger) evictOld() {
	cutoff := time.Now().AddDate(0, 0, -s.retentionDays)
	i := 0
	for i < len(s.entries) {
		t, err := time.Parse(time.RFC3339Nano, s.entries[i].Timestamp)
		if err != nil || t.Before(cutoff) {
			i++
		} else {
			break
		}
	}
	if i > 0 {
		s.entries = append(s.entries[:0], s.entries[i:]...)
		return
	}
	if len(s.entries) >= s.maxEntries {
		s.entries = append(s.entries[:0], s.entries[1:]...)
	}
}

func (s *StructuredLogger) Entries() []StructuredEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]StructuredEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

func (s *StructuredLogger) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

func (s *StructuredLogger) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = s.entries[:0]
}

func (s *StructuredLogger) SetPromptLogging(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.promptLoggingEnabled = enabled
}

func (s *StructuredLogger) IsPromptLoggingEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.promptLoggingEnabled
}
