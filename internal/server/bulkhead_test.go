package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWeightedBulkhead_AcquireRelease(t *testing.T) {
	b := NewWeightedBulkhead(10, time.Second)

	if !b.Acquire(1, time.Second) {
		t.Fatal("expected acquire to succeed")
	}
	b.Release(1)

	stats := b.Stats()
	if stats.InFlight != 0 {
		t.Errorf("expected 0 in-flight after release, got %d", stats.InFlight)
	}
}

func TestWeightedBulkhead_StreamWeight(t *testing.T) {
	b := NewWeightedBulkhead(3, time.Second)

	if !b.Acquire(2, time.Second) {
		t.Fatal("expected stream acquire to succeed")
	}
	if !b.Acquire(1, time.Second) {
		t.Fatal("expected nonstream acquire to succeed")
	}

	stats := b.Stats()
	if stats.InFlight != 3 {
		t.Errorf("expected 3 in-flight (2+1), got %d", stats.InFlight)
	}

	b.Release(2)
	b.Release(1)

	stats = b.Stats()
	if stats.InFlight != 0 {
		t.Errorf("expected 0 in-flight after releases, got %d", stats.InFlight)
	}
}

func TestWeightedBulkhead_RejectsWhenFull(t *testing.T) {
	b := NewWeightedBulkhead(2, 50*time.Millisecond)

	if !b.Acquire(2, time.Second) {
		t.Fatal("expected first acquire to succeed")
	}

	if b.Acquire(1, 50*time.Millisecond) {
		t.Error("expected second acquire to be rejected when full")
		b.Release(1)
	}

	stats := b.Stats()
	if stats.TotalReject == 0 {
		t.Error("expected at least one rejection counted")
	}

	b.Release(2)
}

func TestWeightedBulkhead_WaitsForSlot(t *testing.T) {
	b := NewWeightedBulkhead(2, 2*time.Second)

	if !b.Acquire(2, time.Second) {
		t.Fatal("expected acquire to succeed")
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		b.Release(2)
	}()

	if !b.Acquire(1, 2*time.Second) {
		t.Fatal("expected acquire to succeed after wait")
	}
	b.Release(1)
}

func TestWeightedBulkhead_Concurrent(t *testing.T) {
	b := NewWeightedBulkhead(3, 50*time.Millisecond)
	started := make(chan struct{})
	done := make(chan struct{})

	go func() {
		for i := 0; i < 3; i++ {
			b.Acquire(1, time.Second)
		}
		close(started)
		<-done
		for i := 0; i < 3; i++ {
			b.Release(1)
		}
	}()

	<-started

	if b.Acquire(1, 50*time.Millisecond) {
		t.Error("expected rejection when capacity full")
		b.Release(1)
	}

	stats := b.Stats()
	if stats.TotalReject == 0 {
		t.Error("expected at least one rejection")
	}

	close(done)
}

func TestWeightedBulkhead_Stats(t *testing.T) {
	b := NewWeightedBulkhead(5, time.Second)

	b.Acquire(1, time.Second)
	b.Acquire(2, time.Second)

	stats := b.Stats()
	if stats.InFlight != 3 {
		t.Errorf("expected 3 in-flight, got %d", stats.InFlight)
	}
	if stats.Capacity != 5 {
		t.Errorf("expected capacity 5, got %d", stats.Capacity)
	}
	if stats.TotalIn != 2 {
		t.Errorf("expected 2 total acquires, got %d", stats.TotalIn)
	}

	b.Release(1)
	b.Release(2)
}

func TestWriteOverloadedJSON(t *testing.T) {
	w := httptest.NewRecorder()

	writeOverloadedJSON(w, 5)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}

	retryAfter := w.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header")
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected application/json, got %s", contentType)
	}
}
