package forwarder

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var forwardHeaders = []string{
	"Content-Type",
	"Accept",
	"X-Request-ID",
	"X-Route-Intent",
}

type Forwarder struct {
	client *http.Client
	base   string
}

func New(baseURL string, timeoutSeconds int) (*Forwarder, error) {
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid downstream URL: %w", err)
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 120
	}
	return &Forwarder{
		client: &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
		base:   strings.TrimRight(baseURL, "/"),
	}, nil
}

type Result struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func (f *Forwarder) buildDownstreamRequest(ctx context.Context, req *http.Request, body []byte, reqID string) (*http.Request, error) {
	downstreamURL := f.base + "/chat/completions"

	downstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, downstreamURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create downstream request: %w", err)
	}

	for _, h := range forwardHeaders {
		if v := req.Header.Get(h); v != "" {
			downstreamReq.Header.Set(h, v)
		}
	}
	if reqID != "" {
		downstreamReq.Header.Set("X-Request-ID", reqID)
	}

	return downstreamReq, nil
}

func (f *Forwarder) Forward(ctx context.Context, req *http.Request, reqID string) (*Result, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}
	defer req.Body.Close()

	downstreamReq, err := f.buildDownstreamRequest(ctx, req, body, reqID)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := f.client.Do(downstreamReq)
	latency := time.Since(start)

	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("request cancelled")
		}
		return nil, fmt.Errorf("downstream unavailable after %v: %w", latency.Round(time.Microsecond), err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read downstream response: %w", err)
	}

	log.Printf("[%s] POST %s -> %d (%v)", reqID, f.base+"/chat/completions", resp.StatusCode, latency.Round(time.Microsecond))

	return &Result{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

type Flusher interface {
	Flush()
}

func (f *Forwarder) ForwardStream(ctx context.Context, req *http.Request, w http.ResponseWriter, reqID string) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	defer req.Body.Close()

	downstreamReq, err := f.buildDownstreamRequest(ctx, req, body, reqID)
	if err != nil {
		return err
	}

	start := time.Now()
	resp, err := f.client.Do(downstreamReq)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("request cancelled")
		}
		return fmt.Errorf("downstream unavailable after %v: %w", time.Since(start).Round(time.Microsecond), err)
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Request-ID", reqID)
	w.WriteHeader(resp.StatusCode)

	flusher, canFlush := w.(Flusher)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	chunkCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)

		if canFlush {
			flusher.Flush()
		}
		chunkCount++
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() == context.Canceled {
			log.Printf("[%s] POST %s -> stream cancelled after %d chunks (%v)", reqID, f.base+"/chat/completions", chunkCount, time.Since(start).Round(time.Microsecond))
			return fmt.Errorf("stream cancelled")
		}
		log.Printf("[%s] POST %s -> stream error after %d chunks: %v", reqID, f.base+"/chat/completions", chunkCount, err)
		return fmt.Errorf("stream read error: %w", err)
	}

	log.Printf("[%s] POST %s -> %d (streamed %d chunks, %v)", reqID, f.base+"/chat/completions", resp.StatusCode, chunkCount, time.Since(start).Round(time.Microsecond))
	return nil
}

func IsStreamRequest(req *http.Request) bool {
	if req.Header.Get("Accept") == "text/event-stream" {
		return true
	}

	if req.Body == nil {
		return false
	}

	limitedReader := io.LimitReader(req.Body, 65536)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return false
	}

	rest := req.Body
	if closer, ok := rest.(io.Closer); ok {
		closer.Close()
	}
	req.Body = io.NopCloser(io.MultiReader(strings.NewReader(string(body)), rest))

	if len(body) == 0 {
		return false
	}

	var probe struct {
		Stream *bool `json:"stream"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}

	return probe.Stream != nil && *probe.Stream
}
