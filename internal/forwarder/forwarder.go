package forwarder

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/netutil"
)

var forwardHeaders = []string{
	"Content-Type",
	"Accept",
	"X-Request-ID",
	"X-Route-Intent",
}

const MaxResponseBody = 64 << 20 // 64 MB

const (
	StreamIdleTimeout = 5 * time.Minute
	StreamMaxLifetime = 30 * time.Minute
)

type Forwarder struct {
	streamClient    *http.Client
	nonstreamClient *http.Client
	base            string
}

func New(baseURL string, timeoutSeconds int) (*Forwarder, error) {
	if err := netutil.ValidateDownstreamURL(baseURL); err != nil {
		return nil, fmt.Errorf("invalid downstream URL: %w", err)
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 120
	}

	nonstreamTransport := &http.Transport{
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   5,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: time.Duration(timeoutSeconds) * time.Second,
	}

	streamTransport := &http.Transport{
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   5,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		DialContext: (&netDialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &Forwarder{
		nonstreamClient: &http.Client{
			Timeout:       time.Duration(timeoutSeconds) * time.Second,
			Transport:     nonstreamTransport,
			CheckRedirect: netutil.SafeRedirectPolicy(),
		},
		streamClient: &http.Client{
			Timeout:       0,
			Transport:     streamTransport,
			CheckRedirect: netutil.SafeRedirectPolicy(),
		},
		base: strings.TrimRight(baseURL, "/"),
	}, nil
}

type netDialer struct {
	Timeout   time.Duration
	KeepAlive time.Duration
}

func (d *netDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return (&net.Dialer{Timeout: d.Timeout, KeepAlive: d.KeepAlive}).DialContext(ctx, network, addr)
}

type Result struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func (f *Forwarder) buildDownstreamRequest(ctx context.Context, req *http.Request, body []byte, reqID string) (*http.Request, error) {
	downstreamURL := f.base + "/chat/completions"

	downstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, downstreamURL, bytes.NewReader(body))
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
	resp, err := f.nonstreamClient.Do(downstreamReq)
	latency := time.Since(start)

	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("request cancelled")
		}
		return nil, fmt.Errorf("downstream unavailable after %v: %w", latency.Round(time.Microsecond), err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBody+1))
	if err != nil {
		return nil, fmt.Errorf("read downstream response: %w", err)
	}
	if int64(len(respBody)) > MaxResponseBody {
		return nil, fmt.Errorf("downstream response exceeds %d bytes limit", MaxResponseBody)
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

	ctx, cancel := context.WithTimeout(ctx, StreamMaxLifetime)
	defer cancel()

	start := time.Now()
	resp, err := f.streamClient.Do(downstreamReq.WithContext(ctx))
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

	reader := bufio.NewReaderSize(resp.Body, 64*1024)

	idleTimer := time.NewTimer(StreamIdleTimeout)
	defer idleTimer.Stop()

	chunkCount := 0
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			w.Write(line)
			if canFlush {
				flusher.Flush()
			}
			chunkCount++
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(StreamIdleTimeout)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			if ctx.Err() == context.Canceled {
				log.Printf("[%s] POST %s -> stream cancelled after %d chunks (%v)", reqID, f.base+"/chat/completions", chunkCount, time.Since(start).Round(time.Microsecond))
				return fmt.Errorf("stream cancelled")
			}
			log.Printf("[%s] POST %s -> stream error after %d chunks: %v", reqID, f.base+"/chat/completions", chunkCount, err)
			return fmt.Errorf("stream read error: %w", err)
		}
		select {
		case <-idleTimer.C:
			log.Printf("[%s] POST %s -> stream idle timeout after %d chunks (%v)", reqID, f.base+"/chat/completions", chunkCount, time.Since(start).Round(time.Microsecond))
			return fmt.Errorf("stream idle timeout")
		default:
		}
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
	req.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body), rest))

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
