package forwarder

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
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
	StreamIdleTimeout    = 5 * time.Minute
	StreamMaxLifetime    = 30 * time.Minute
	StreamMaxBytesBudget = 256 << 20 // 256 MB default streaming byte budget
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

// StreamTransport returns the forwarder's stream transport for use by
// health checks and other internal callers that need SSRF-protected HTTP.
func (f *Forwarder) StreamTransport() http.RoundTripper {
	return f.streamClient.Transport
}

type netDialer struct {
	Timeout   time.Duration
	KeepAlive time.Duration
}

func (d *netDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed for %s: %w", host, err)
	}
	for _, ip := range ips {
		if !netutil.IsAllowedIP(ip) {
			return nil, fmt.Errorf("blocked connection to %s (%s) — IP is in a restricted range", host, ip)
		}
	}
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

// deadlineReader wraps an io.Reader and enforces a deadline. When the deadline
// fires, it calls cancel() and any blocking Read returns immediately.
type deadlineReader struct {
	r      io.Reader
	dead   <-chan time.Time
	cancel context.CancelFunc
	closed chan struct{}
	closeOnce sync.Once
}

func newDeadlineReader(r io.Reader, dead <-chan time.Time, cancel context.CancelFunc) *deadlineReader {
	dr := &deadlineReader{r: r, dead: dead, cancel: cancel, closed: make(chan struct{})}
	go dr.watch()
	return dr
}

func (dr *deadlineReader) watch() {
	select {
	case <-dr.dead:
		dr.cancel()
	case <-dr.closed:
	}
}

func (dr *deadlineReader) Read(p []byte) (int, error) {
	return dr.r.Read(p)
}

func (dr *deadlineReader) Close() error {
	dr.closeOnce.Do(func() { close(dr.closed) })
	if c, ok := dr.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
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

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.Header().Set("Content-Type", "text/event-stream")
	} else if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	} else {
		w.Header().Set("Content-Type", "text/event-stream")
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Request-ID", reqID)
	w.WriteHeader(resp.StatusCode)

	flusher, canFlush := w.(Flusher)

	idleTimer := time.NewTimer(StreamIdleTimeout)
	defer idleTimer.Stop()

	chunkCount := 0
	totalBytes := int64(0)
	deadlineCtx, deadlineCancel := context.WithCancel(ctx)
	defer deadlineCancel()
	dr := newDeadlineReader(resp.Body, idleTimer.C, deadlineCancel)
	defer dr.Close()
	reader := bufio.NewReaderSize(dr, 64*1024)

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			totalBytes += int64(len(line))
			if totalBytes > StreamMaxBytesBudget {
				log.Printf("[%s] POST %s -> stream byte budget exceeded after %d chunks (%d bytes, %v)", reqID, f.base+"/chat/completions", chunkCount, totalBytes, time.Since(start).Round(time.Microsecond))
				return fmt.Errorf("stream byte budget exceeded")
			}
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
			if deadlineCtx.Err() == context.Canceled || ctx.Err() == context.Canceled {
				log.Printf("[%s] POST %s -> stream cancelled after %d chunks (%v)", reqID, f.base+"/chat/completions", chunkCount, time.Since(start).Round(time.Microsecond))
				return fmt.Errorf("stream cancelled")
			}
			log.Printf("[%s] POST %s -> stream error after %d chunks: %v", reqID, f.base+"/chat/completions", chunkCount, err)
			return fmt.Errorf("stream read error: %w", err)
		}
	}

	log.Printf("[%s] POST %s -> %d (streamed %d chunks, %v)", reqID, f.base+"/chat/completions", resp.StatusCode, chunkCount, time.Since(start).Round(time.Microsecond))
	return nil
}
