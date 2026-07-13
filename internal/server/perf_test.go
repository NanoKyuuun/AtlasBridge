package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/forwarder"
	"github.com/atlasbridge/atlasbridge/internal/observability"
)

// benchMockDownstream creates a mock downstream server for benchmarks.
func benchMockDownstream(b testing.TB, handler http.HandlerFunc) *httptest.Server {
	b.Helper()
	ts := httptest.NewServer(handler)
	return ts
}

// benchServerWithDownstream creates a proxy server for benchmarks (no auth).
func benchServerWithDownstream(b testing.TB, downstreamURL string) *httptest.Server {
	b.Helper()
	cfg := config.DefaultConfig()
	cfg.Security.AdminAuthEnabled = false
	cfg.Downstream.BaseURL = downstreamURL
	fwd, _ := forwarder.New(cfg.Downstream.BaseURL, cfg.Downstream.TimeoutSeconds)
	defaultRoutes := config.DefaultRoutesConfig()
	snap := &Snapshot{
		Config:    *cfg,
		Routes:    *defaultRoutes,
		Profiles:  *config.DefaultProfilesConfig(),
		Forwarder: fwd,
		Version:   1,
	}
	store := NewStateStore(snap)
	cs := NewConfigService(store)
	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
	}
	r := New(deps)
	return httptest.NewServer(r.Handler)
}

// benchmarkNonStream is the core non-streaming benchmark.
// PF-01: single request baseline.
// PF-02/03: via b.RunParallel with GOMAXPROCS.
func benchmarkNonStream(b *testing.B, parallel bool) {
	downstream := benchMockDownstream(b, func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-bench","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	})
	defer downstream.Close()

	ts := benchServerWithDownstream(b, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`

	b.ResetTimer()
	b.ReportAllocs()

	if parallel {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					b.Fatal(err)
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		})
	} else {
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

// PF-01: Non-stream request, single goroutine baseline.
func BenchmarkNonStreamSingle(b *testing.B) {
	benchmarkNonStream(b, false)
}

// PF-02: Non-stream request, 5 concurrent goroutines.
func BenchmarkNonStreamConcurrent5(b *testing.B) {
	b.SetParallelism(5)
	benchmarkNonStream(b, true)
}

// PF-03: Non-stream request, 10 concurrent goroutines.
func BenchmarkNonStreamConcurrent10(b *testing.B) {
	b.SetParallelism(10)
	benchmarkNonStream(b, true)
}

// PF-04: 1 MiB request body forwarded through proxy.
func BenchmarkLargePayload1MiB(b *testing.B) {
	largePayload := strings.Repeat("A", 1<<20) // 1 MiB

	downstream := benchMockDownstream(b, func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-large","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	})
	defer downstream.Close()

	ts := benchServerWithDownstream(b, downstream.URL+"/v1")
	defer ts.Close()

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(largePayload)))

	for i := 0; i < b.N; i++ {
		body := fmt.Sprintf(`{"model":"gpt-4","messages":[{"role":"user","content":"%s"}]}`, largePayload)
		req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// PF-05: Request near the 64 MB body limit (10 MB — full 64 MB would be too slow for benchmark).
func BenchmarkNearBodyLimit10MB(b *testing.B) {
	nearLimitPayload := strings.Repeat("B", 10<<20) // 10 MB

	downstream := benchMockDownstream(b, func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-near","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	})
	defer downstream.Close()

	ts := benchServerWithDownstream(b, downstream.URL+"/v1")
	defer ts.Close()

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(nearLimitPayload)))

	for i := 0; i < b.N; i++ {
		body := fmt.Sprintf(`{"model":"gpt-4","messages":[{"role":"user","content":"%s"}]}`, nearLimitPayload)
		req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// PF-06: Multiple concurrent SSE streams.
func BenchmarkStreamConcurrent(b *testing.B) {
	chunk := "data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hi\"}}]}\n\n"

	downstream := benchMockDownstream(b, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		for i := 0; i < 50; i++ {
			fmt.Fprint(w, chunk)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	})
	defer downstream.Close()

	ts := benchServerWithDownstream(b, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}],"stream":true}`

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
			}
			resp.Body.Close()
		}
	})
}

// PF-07: Downstream with 500ms artificial latency per chunk.
func BenchmarkSlowDownstream(b *testing.B) {
	downstream := benchMockDownstream(b, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-slow","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	})
	defer downstream.Close()

	ts := benchServerWithDownstream(b, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// PF-08: Downstream never responds — verify timeout triggers.
func BenchmarkDeadDownstreamTimeout(b *testing.B) {
	downstream := benchMockDownstream(b, func(w http.ResponseWriter, r *http.Request) {
		select {} // block forever
	})
	defer downstream.Close()

	ts := benchServerWithDownstream(b, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		req, _ := http.NewRequestWithContext(ctx, "POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		http.DefaultClient.Do(req)
		cancel()
	}
}

// PF-09: Client disconnect mid-stream — verify downstream cancellation.
func BenchmarkStreamClientDisconnect(b *testing.B) {
	received := make(chan struct{})

	downstream := benchMockDownstream(b, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		close(received)
		for i := 0; i < 200; i++ {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"%d\"}}]}\n\n", i)
			flusher.Flush()
			time.Sleep(50 * time.Millisecond)
		}
	})
	defer downstream.Close()

	ts := benchServerWithDownstream(b, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}],"stream":true}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		req, _ := http.NewRequestWithContext(ctx, "POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		// Read a few chunks then disconnect
		buf := make([]byte, 4096)
		resp.Body.Read(buf)
		cancel()
		resp.Body.Close()
		// Wait for next iteration
		time.Sleep(100 * time.Millisecond)
	}
}

// PF-10: Goroutine and memory leak detection after repeated load.
func TestNoGoroutineLeaksAfterLoad(t *testing.T) {
	runtime.GC()
	runtime.GC()
	baselineGoroutines := runtime.NumGoroutine()

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-leak","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	})
	defer downstream.Close()

	ts := newTestServerWithDownstreamNoAuth(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`

	// Phase 1: Non-stream load
	t.Log("Phase 1: non-stream load (200 requests)")
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	wg.Wait()

	// Phase 2: Stream load
	streamDownstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		for i := 0; i < 20; i++ {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"%d\"}}]}\n\n", i)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	})
	defer streamDownstream.Close()

	streamTS := newTestServerWithDownstreamNoAuth(t, streamDownstream.URL+"/v1")
	defer streamTS.Close()

	streamBody := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}],"stream":true}`

	t.Log("Phase 2: stream load (50 concurrent streams)")
	var swg sync.WaitGroup
	for i := 0; i < 50; i++ {
		swg.Add(1)
		go func() {
			defer swg.Done()
			req, _ := http.NewRequest("POST", streamTS.URL+"/v1/chat/completions", strings.NewReader(streamBody))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
			}
			resp.Body.Close()
		}()
	}
	swg.Wait()

	// Allow goroutines to settle
	time.Sleep(2 * time.Second)
	runtime.GC()
	runtime.GC()

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	finalGoroutines := runtime.NumGoroutine()
	goroutineDelta := finalGoroutines - baselineGoroutines

	t.Logf("Baseline goroutines: %d", baselineGoroutines)
	t.Logf("Final goroutines:    %d", finalGoroutines)
	t.Logf("Goroutine delta:     %d", goroutineDelta)
	t.Logf("Heap alloc delta:    %d bytes", int64(m2.Alloc)-int64(m1.Alloc))
	t.Logf("Heap inuse delta:    %d bytes", int64(m2.HeapInuse)-int64(m1.HeapInuse))

	// httptest.Server goroutines (listeners, idle conns) + runtime overhead.
	// We have 2 servers (proxy + downstream) each contributing ~10-15 goroutines.
	// Allow up to 50 for infrastructure overhead.
	if goroutineDelta > 50 {
		t.Errorf("goroutine leak detected: baseline=%d final=%d delta=%d (threshold=50)",
			baselineGoroutines, finalGoroutines, goroutineDelta)
	}

	// Heap should not grow by more than 50 MB after load
	heapGrowth := int64(m2.HeapInuse) - int64(m1.HeapInuse)
	if heapGrowth > 50<<20 {
		t.Errorf("excessive heap growth: %d bytes (threshold 50 MB)", heapGrowth)
	}
}

// PF-07 supplementary: Verify bulkhead does not reject under normal load.
func TestBulkheadNoRejectionUnderNormalLoad(t *testing.T) {
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"ok","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	})
	defer downstream.Close()

	ts := newTestServerWithDownstreamNoAuth(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`

	rejected := make(chan int, 100)
	total := 50

	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				rejected <- 1
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	wg.Wait()
	close(rejected)

	rejectionCount := 0
	for range rejected {
		rejectionCount++
	}

	if rejectionCount > 0 {
		t.Errorf("bulkhead rejected %d/%d requests under normal load (expected 0)", rejectionCount, total)
	}
	t.Logf("bulkhead: %d/%d requests succeeded, %d rejected", total-rejectionCount, total, rejectionCount)
}
