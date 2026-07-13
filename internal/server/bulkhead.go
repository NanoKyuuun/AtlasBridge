package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	WeightNonstream = 1
	WeightStream    = 2
	MaxInFlight     = 64
	MaxQueued       = 32
)

type WeightedBulkhead struct {
	mu          sync.Mutex
	cond        *sync.Cond
	inFlight    int
	capacity    int
	waitQueue   int
	maxWait     time.Duration
	TotalIn     int64
	TotalWait   int64
	TotalReject int64
}

func NewWeightedBulkhead(capacity int, maxWait time.Duration) *WeightedBulkhead {
	b := &WeightedBulkhead{
		capacity: capacity,
		maxWait:  maxWait,
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *WeightedBulkhead) Acquire(weight int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	b.mu.Lock()
	defer b.mu.Unlock()

	for b.inFlight+weight > b.capacity {
		if time.Now().After(deadline) {
			b.TotalReject++
			return false
		}
		if b.waitQueue >= MaxQueued {
			b.TotalReject++
			return false
		}
		b.waitQueue++
		b.TotalWait++
		remaining := time.Until(deadline)
		if remaining <= 0 {
			b.waitQueue--
			b.TotalReject++
			return false
		}
		timer := time.NewTimer(remaining)
		done := make(chan struct{})
		go func() {
			select {
			case <-timer.C:
			case <-done:
				timer.Stop()
			}
			b.cond.Signal()
		}()
		b.cond.Wait()
		close(done)
		timer.Stop()
		b.waitQueue--
	}

	b.inFlight += weight
	b.TotalIn++
	return true
}

func (b *WeightedBulkhead) Release(weight int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.inFlight -= weight
	if b.inFlight < 0 {
		b.inFlight = 0
	}
	b.cond.Signal()
}

func (b *WeightedBulkhead) Stats() BulkheadStats {
	b.mu.Lock()
	defer b.mu.Unlock()
	return BulkheadStats{
		InFlight:    b.inFlight,
		Capacity:    b.capacity,
		WaitQueue:   b.waitQueue,
		TotalIn:     b.TotalIn,
		TotalWait:   b.TotalWait,
		TotalReject: b.TotalReject,
	}
}

type BulkheadStats struct {
	InFlight    int   `json:"in_flight"`
	Capacity    int   `json:"capacity"`
	WaitQueue   int   `json:"wait_queue"`
	TotalIn     int64 `json:"total_in"`
	TotalWait   int64 `json:"total_wait"`
	TotalReject int64 `json:"total_reject"`
}

func writeOverloadedJSON(w http.ResponseWriter, retryAfter int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": "server is at capacity, please retry later",
			"type":    "proxy_error",
			"code":    "server_overloaded",
		},
	})
}
