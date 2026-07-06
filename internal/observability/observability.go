package observability

import (
	"sync"
	"time"
)

const DefaultMaxEntries = 1000

type LogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestID    string    `json:"request_id"`
	TaskType     string    `json:"task_type"`
	Confidence   float64   `json:"confidence"`
	RouteKey     string    `json:"route_key"`
	Alias        string    `json:"alias"`
	OverrideSrc  string    `json:"override_source"`
	Status       string    `json:"classification_status"`
	Reason       string    `json:"routing_reason"`
	LatencyMs    int64     `json:"latency_ms"`
	StatusCode   int       `json:"status_code"`
	Model        string    `json:"model"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	IsStream     bool      `json:"is_stream"`
	PrivacyMode  string    `json:"privacy_mode"`
}

type Logger struct {
	mu       sync.RWMutex
	entries  []LogEntry
	maxSize  int
}

func NewLogger(maxEntries int) *Logger {
	if maxEntries <= 0 {
		maxEntries = DefaultMaxEntries
	}
	return &Logger{
		entries: make([]LogEntry, 0, maxEntries),
		maxSize: maxEntries,
	}
}

func (l *Logger) Record(entry LogEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) >= l.maxSize {
		copy(l.entries, l.entries[1:])
		l.entries = l.entries[:len(l.entries)-1]
	}

	l.entries = append(l.entries, entry)
}

func (l *Logger) GetEntries(limit int) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	total := len(l.entries)
	if limit <= 0 || limit > total {
		limit = total
	}

	result := make([]LogEntry, limit)
	copy(result, l.entries[total-limit:])
	return result
}

func (l *Logger) GetEntriesSince(since time.Time) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, e := range l.entries {
		if e.Timestamp.After(since) || e.Timestamp.Equal(since) {
			result = append(result, e)
		}
	}
	return result
}

func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = l.entries[:0]
}

func (l *Logger) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}
