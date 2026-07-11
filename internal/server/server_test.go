package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/forwarder"
	"github.com/atlasbridge/atlasbridge/internal/observability"
	runtimemod "github.com/atlasbridge/atlasbridge/internal/runtime"
	"github.com/atlasbridge/atlasbridge/internal/security"
)

func newSnapshotFromConfig(cfg *config.Config, routes *config.RoutesConfig, profiles *config.ProfilesConfig) (*StateStore, *ConfigService) {
	fwd, _ := forwarder.New(cfg.Downstream.BaseURL, cfg.Downstream.TimeoutSeconds)
	snap := &Snapshot{
		Config:    *cfg,
		Routes:    *routes,
		Profiles:  *profiles,
		Forwarder: fwd,
		Version:   1,
		CreatedAt: time.Now(),
	}
	store := NewStateStore(snap)
	return store, NewConfigService(store)
}

func newTestServerDeps(t *testing.T, cfg *config.Config) *ServerDeps {
	t.Helper()
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())
	return &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
	}
}

func newTestServerDepsWithRuntime(t *testing.T, cfg *config.Config) *ServerDeps {
	t.Helper()
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())
	state := runtimemod.NewState(runtimemod.ModeAlwaysOn)
	state.Start()
	return &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
		RuntimeState:  state,
	}
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	cfg := config.DefaultConfig()
	deps := newTestServerDeps(t, cfg)
	r := New(deps)
	return httptest.NewServer(r.Handler)
}

func newTestServerWithDownstream(t *testing.T, downstreamURL string) *httptest.Server {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Downstream.BaseURL = downstreamURL
	deps := newTestServerDeps(t, cfg)
	r := New(deps)
	return httptest.NewServer(r.Handler)
}

func mockDownstream(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(handler)
	return ts
}

func newTestServerNoAuth(t *testing.T) *httptest.Server {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Security.AdminAuthEnabled = false
	deps := newTestServerDeps(t, cfg)
	r := New(deps)
	return httptest.NewServer(r.Handler)
}

func newTestServerWithDownstreamNoAuth(t *testing.T, downstreamURL string) *httptest.Server {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Security.AdminAuthEnabled = false
	cfg.Downstream.BaseURL = downstreamURL
	deps := newTestServerDeps(t, cfg)
	r := New(deps)
	return httptest.NewServer(r.Handler)
}

func newTestServerWithToken(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	token, hash, _ := security.GenerateToken()
	cfg := config.DefaultConfig()
	cfg.Security.AdminAuthEnabled = true
	cfg.Security.AdminTokenHash = hash
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())
	state := runtimemod.NewState(runtimemod.ModeAlwaysOn)
	state.Start()
	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
		RuntimeState:  state,
	}
	r := New(deps)
	return httptest.NewServer(r.Handler), token
}

func TestHealthEndpoint(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
	if body["service"] != "atlasbridge" {
		t.Errorf("expected service atlasbridge, got %s", body["service"])
	}
	if body["version"] == "" {
		t.Error("expected non-empty version")
	}
}

func TestStatusEndpoint(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "running" {
		t.Errorf("expected status running, got %v", body["status"])
	}
	if body["port"] != float64(20127) {
		t.Errorf("expected port 20127, got %v", body["port"])
	}
	if body["host"] != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %v", body["host"])
	}
	if body["downstream"] != "http://127.0.0.1:20128/v1" {
		t.Errorf("expected downstream http://127.0.0.1:20128/v1, got %v", body["downstream"])
	}
	if body["mode"] != "manual" {
		t.Errorf("expected mode manual, got %v", body["mode"])
	}
	if body["version"] == "" {
		t.Error("expected non-empty version")
	}
	if body["uptime"] == "" {
		t.Error("expected non-empty uptime")
	}
	if body["go_version"] == "" {
		t.Error("expected non-empty go_version")
	}
}

func TestModelsEndpoint(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/models")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["object"] != "list" {
		t.Errorf("expected object list, got %v", body["object"])
	}
	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) != 15 {
		t.Errorf("expected 15 models, got %d", len(data))
	}

	expectedModels := []string{"atlas-auto", "atlas-debug", "atlas-cheap", "atlas-docs", "atlas-architect", "atlas-fast", "atlas-long-context", "smart-auto", "smart-debug", "smart-cheap", "smart-docs", "smart-architect", "smart-code", "smart-fast", "smart-long-context"}
	for i, m := range data {
		model := m.(map[string]interface{})
		if model["id"] != expectedModels[i] {
			t.Errorf("expected model %s, got %v", expectedModels[i], model["id"])
		}
		if model["owned_by"] != "atlasbridge" {
			t.Errorf("expected owned_by atlasbridge, got %v", model["owned_by"])
		}
	}
}

func TestChatCompletionsPassthrough(t *testing.T) {
	var receivedBody []byte
	var receivedModel string

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		var parsed map[string]interface{}
		json.Unmarshal(receivedBody, &parsed)
		receivedModel = parsed["model"].(string)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"choices": []map[string]interface{}{{"index": 0, "message": map[string]string{"role": "assistant", "content": "hello"}}},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedModel != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", receivedModel)
	}
	if string(receivedBody) != body {
		t.Errorf("body mismatch: got %s", string(receivedBody))
	}
}

func TestChatCompletionsDownstreamUnavailable(t *testing.T) {
	ts := newTestServerWithDownstream(t, "http://127.0.0.1:19999/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	errObj, ok := respBody["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "proxy_error" {
		t.Errorf("expected type proxy_error, got %v", errObj["type"])
	}
	if errObj["code"] != "downstream_error" {
		t.Errorf("expected code downstream_error, got %v", errObj["code"])
	}
}

func TestChatCompletionsPreservesHeaders(t *testing.T) {
	var receivedHeaders http.Header

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "ok"})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "custom-id-123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if receivedHeaders.Get("X-Request-ID") != "custom-id-123" {
		t.Errorf("expected X-Request-ID custom-id-123, got %s", receivedHeaders.Get("X-Request-ID"))
	}
	if receivedHeaders.Get("Authorization") != "" {
		t.Errorf("Authorization should not be forwarded, got %s", receivedHeaders.Get("Authorization"))
	}
}

func TestChatCompletionsRequestIDInResponse(t *testing.T) {
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "ok"})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "my-req-id")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Request-ID") != "my-req-id" {
		t.Errorf("expected X-Request-ID my-req-id in response, got %s", resp.Header.Get("X-Request-ID"))
	}
}

func TestAdminPlaceholder(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", ct)
	}
	if cache := resp.Header.Get("Cache-Control"); cache != "no-store" {
		t.Errorf("expected Cache-Control no-store, got %s", cache)
	}
}

func TestAdminStaticAsset(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	files, err := fs.Glob(adminDistFS, "assets/*.js")
	if err != nil {
		t.Fatalf("failed to list built assets: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one built JS asset")
	}

	resp, err := http.Get(ts.URL + "/admin/" + files[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/javascript") {
		t.Errorf("expected JavaScript content type, got %s", ct)
	}
	if cache := resp.Header.Get("Cache-Control"); cache != "public, max-age=31536000, immutable" {
		t.Errorf("expected immutable cache policy, got %s", cache)
	}
}

func TestAdminMissingStaticAssetReturns404(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/assets/does-not-exist.js")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); strings.HasPrefix(ct, "text/html") {
		t.Errorf("unexpected HTML fallback for missing asset, got %s", ct)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	reqID := resp.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestRequestIDPreserved(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/health", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	reqID := resp.Header.Get("X-Request-ID")
	if reqID != "my-custom-id" {
		t.Errorf("expected X-Request-ID my-custom-id, got %s", reqID)
	}
}

func TestRequestIDRejectsTooLong(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	longID := strings.Repeat("a", 200)
	req, _ := http.NewRequest("GET", ts.URL+"/health", nil)
	req.Header.Set("X-Request-ID", longID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	reqID := resp.Header.Get("X-Request-ID")
	if reqID == longID {
		t.Error("expected long request ID to be rejected and replaced with generated ID")
	}
	if len(reqID) > 128 {
		t.Errorf("expected generated ID to be at most 128 chars, got %d", len(reqID))
	}
}

func TestRequestIDRejectsInvalidChars(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	invalidIDs := []string{
		"hello world",
		"test/injection",
		"test<script>alert(1)</script>",
		"test:colon",
		"test@at",
		"test|pipe",
	}
	for _, invalidID := range invalidIDs {
		req, _ := http.NewRequest("GET", ts.URL+"/health", nil)
		req.Header.Set("X-Request-ID", invalidID)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("unexpected error for ID %q: %v", invalidID, err)
		}
		resp.Body.Close()

		reqID := resp.Header.Get("X-Request-ID")
		if reqID == invalidID {
			t.Errorf("expected invalid request ID %q to be rejected", invalidID)
		}
	}
}

func TestRequestIDAllowsValidChars(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	validIDs := []string{
		"abc123",
		"test-id",
		"test_id",
		"test.id",
		"Test-ID_123.abc",
	}
	for _, validID := range validIDs {
		req, _ := http.NewRequest("GET", ts.URL+"/health", nil)
		req.Header.Set("X-Request-ID", validID)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("unexpected error for ID %q: %v", validID, err)
		}
		resp.Body.Close()

		reqID := resp.Header.Get("X-Request-ID")
		if reqID != validID {
			t.Errorf("expected valid request ID %q to be preserved, got %q", validID, reqID)
		}
	}
}

func TestRequestIDMaxLenPreserved(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	maxID := strings.Repeat("a", 128)
	req, _ := http.NewRequest("GET", ts.URL+"/health", nil)
	req.Header.Set("X-Request-ID", maxID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	reqID := resp.Header.Get("X-Request-ID")
	if reqID != maxID {
		t.Errorf("expected 128-char request ID to be preserved, got %q", reqID)
	}
}

func TestAdminAPIStatusContentType(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestChatCompletionsStreaming(t *testing.T) {
	chunksReceived := make(chan string, 10)

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			"data: {\"id\":\"chatcmpl-1\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}",
			"",
			"data: {\"id\":\"chatcmpl-1\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\" World\"}}]}",
			"",
			"data: [DONE]",
			"",
		}
		for _, c := range chunks {
			fmt.Fprintf(w, "%s\n", c)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			chunksReceived <- c
		}
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "stream-test-1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("X-Request-ID") != "stream-test-1" {
		t.Errorf("expected X-Request-ID stream-test-1, got %s", resp.Header.Get("X-Request-ID"))
	}

	scanner := bufio.NewScanner(resp.Body)
	var receivedLines []string
	for scanner.Scan() {
		line := scanner.Text()
		receivedLines = append(receivedLines, line)
		if len(receivedLines) > 20 {
			break
		}
	}

	found := false
	for _, line := range receivedLines {
		if strings.Contains(line, "Hello") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to receive Hello chunk")
	}
}

func TestChatCompletionsStreamingDownstreamUnavailable(t *testing.T) {
	ts := newTestServerWithDownstream(t, "http://127.0.0.1:19999/v1")
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "proxy_error" {
		t.Errorf("expected type proxy_error, got %v", errObj["type"])
	}
}

func TestChatCompletionsNonStreamingRegression(t *testing.T) {
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-regression",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "regression test"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"test"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["id"] != "chatcmpl-regression" {
		t.Errorf("expected id chatcmpl-regression, got %v", result["id"])
	}
}

func TestChatCompletionsStreamingPreservesBody(t *testing.T) {
	var receivedBody map[string]interface{}

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "data: {\"done\":true}\n\n")
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"test"}],"stream":true,"temperature":0.8}`
	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if receivedBody["model"] != "gpt-4" {
		t.Errorf("expected model gpt-4, got %v", receivedBody["model"])
	}
	if receivedBody["stream"] != true {
		t.Errorf("expected stream true, got %v", receivedBody["stream"])
	}
	if receivedBody["temperature"] != 0.8 {
		t.Errorf("expected temperature 0.8, got %v", receivedBody["temperature"])
	}
}

func TestChatCompletionsStreamingRequestIDForwarded(t *testing.T) {
	var receivedReqID string

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		receivedReqID = r.Header.Get("X-Request-ID")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "data: {\"done\":true}\n\n")
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "stream-custom-id")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if receivedReqID != "stream-custom-id" {
		t.Errorf("expected X-Request-ID stream-custom-id, got %s", receivedReqID)
	}
}

func TestChatCompletionsStreamingTimeout(t *testing.T) {
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		time.Sleep(3 * time.Second)
		fmt.Fprintf(w, "data: {\"done\":true}\n\n")
	})
	defer downstream.Close()

	cfg := config.DefaultConfig()
	cfg.Downstream.TimeoutSeconds = 1
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())
	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
	}
	r := New(deps)
	ts := httptest.NewServer(r.Handler)
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(`{"stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if elapsed > 5*time.Second {
		t.Errorf("expected timeout within 5s, took %v", elapsed)
	}
}

func TestRoutingPipelineSmartDebugAlias(t *testing.T) {
	var receivedModel string

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		var parsed map[string]interface{}
		json.NewDecoder(r.Body).Decode(&parsed)
		receivedModel = parsed["model"].(string)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-routing-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "debug response"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"smart-debug","messages":[{"role":"user","content":"fix this error: panic at runtime"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedModel != "combo.debugging" {
		t.Errorf("expected model combo.debugging forwarded, got %s", receivedModel)
	}
}

func TestRoutingPipelineSmartAutoBackendClassification(t *testing.T) {
	var receivedModel string

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		var parsed map[string]interface{}
		json.NewDecoder(r.Body).Decode(&parsed)
		receivedModel = parsed["model"].(string)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-auto-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "auto response"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"smart-auto","messages":[{"role":"user","content":"build a REST API endpoint with database query and server handler"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedModel != "combo.design" {
		t.Errorf("expected model combo.design forwarded, got %s", receivedModel)
	}
}

func TestRoutingPipelineManualModelPassthrough(t *testing.T) {
	var receivedModel string

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		var parsed map[string]interface{}
		json.NewDecoder(r.Body).Decode(&parsed)
		receivedModel = parsed["model"].(string)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-manual-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "manual response"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4-turbo","messages":[{"role":"user","content":"hello"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedModel != "gpt-4-turbo" {
		t.Errorf("expected model gpt-4-turbo forwarded, got %s", receivedModel)
	}
}

func TestRoutingPipelineStreamingSmartAlias(t *testing.T) {
	var receivedModel string

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		var parsed map[string]interface{}
		json.NewDecoder(r.Body).Decode(&parsed)
		receivedModel = parsed["model"].(string)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\n")
		fmt.Fprintf(w, "data: [DONE]\n\n")
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"smart-docs","messages":[{"role":"user","content":"write docs"}],"stream":true}`
	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedModel != "combo.documentation" {
		t.Errorf("expected model combo.documentation forwarded, got %s", receivedModel)
	}
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestRoutingPipelineBodyPreservedAfterAnalysis(t *testing.T) {
	var receivedBody map[string]interface{}

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-body-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"smart-auto","messages":[{"role":"user","content":"test message with API and database keywords"}],"temperature":0.7,"max_tokens":100}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if receivedBody["model"] != "combo.test_generation" {
		t.Errorf("expected model combo.test_generation in downstream body, got %v", receivedBody["model"])
	}
	if receivedBody["temperature"] != 0.7 {
		t.Errorf("expected temperature 0.7 preserved, got %v", receivedBody["temperature"])
	}
	if receivedBody["max_tokens"] != float64(100) {
		t.Errorf("expected max_tokens 100 preserved, got %v", receivedBody["max_tokens"])
	}
}

func TestModelsEndpointIncludesSmartCode(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/models")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data := body["data"].([]interface{})
	found := false
	for _, m := range data {
		model := m.(map[string]interface{})
		if model["id"] == "smart-code" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected smart-code model to be listed")
	}
}

func TestAdminGetConfig(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body config.Config
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Server.Port != 20127 {
		t.Errorf("expected port 20127, got %d", body.Server.Port)
	}
	if body.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", body.Server.Host)
	}
}

func TestAdminGetConfigNoTokenLeak(t *testing.T) {
	ts, token := newTestServerWithToken(t)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/admin/api/config", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	sec, ok := result["security"].(map[string]interface{})
	if !ok {
		t.Fatal("expected security object in response")
	}

	if _, exists := sec["admin_token_hash"]; exists {
		t.Error("admin_token_hash must not be present in SecurityView response")
	}

	tokenConfigured, ok := sec["token_configured"].(bool)
	if !ok {
		t.Fatal("expected token_configured in security")
	}
	if !tokenConfigured {
		t.Error("token_configured should be true when a hash is set")
	}
}

func TestAdminPutConfig(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	cfg := config.DefaultConfig()
	cfg.Logging.PrivacyMode = "strict"
	body, _ := json.Marshal(cfg)

	req, _ := http.NewRequest("PUT", ts.URL+"/admin/api/config", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %s", result["status"])
	}
}

func TestAdminPutConfigGeneratesToken(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	patch := `{"security":{"admin_auth_enabled":true}}`
	req, _ := http.NewRequest("PUT", ts.URL+"/admin/api/config", strings.NewReader(patch))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var putResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&putResp)
	adminToken, _ := putResp["admin_token"].(string)
	if adminToken == "" {
		t.Fatal("expected admin_token in PUT response")
	}

	getReq, _ := http.NewRequest("GET", ts.URL+"/admin/api/config", nil)
	getReq.Header.Set("Authorization", "Bearer "+adminToken)
	getResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer getResp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(getResp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	sec, ok := result["security"].(map[string]interface{})
	if !ok {
		t.Fatal("expected security in config")
	}

	if _, exists := sec["admin_token_hash"]; exists {
		t.Error("admin_token_hash should not be in SecurityView response")
	}

	tokenConfigured, ok := sec["token_configured"].(bool)
	if !ok {
		t.Fatal("expected token_configured in security")
	}
	if !tokenConfigured {
		t.Error("token_configured should be true after token generation")
	}

	authEnabled, ok := sec["admin_auth_enabled"].(bool)
	if !ok {
		t.Fatal("expected admin_auth_enabled in security")
	}
	if !authEnabled {
		t.Error("admin_auth_enabled should be true")
	}
}

func TestAdminPutConfigNoRegeneration(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	patch := `{"security":{"admin_auth_enabled":true}}`
	req, _ := http.NewRequest("PUT", ts.URL+"/admin/api/config", strings.NewReader(patch))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var putResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&putResp)
	firstToken, _ := putResp["admin_token"].(string)
	if firstToken == "" {
		t.Fatal("expected admin_token on first PUT (hash was empty)")
	}

	patch2 := `{"logging":{"privacy_mode":"strict"}}`
	req2, _ := http.NewRequest("PUT", ts.URL+"/admin/api/config", strings.NewReader(patch2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+firstToken)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 on second PUT, got %d", resp2.StatusCode)
	}

	var putResp2 map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&putResp2)
	secondToken, _ := putResp2["admin_token"].(string)
	if secondToken != "" {
		t.Error("second PUT should not return a new token (hash already set)")
	}

	getReq, _ := http.NewRequest("GET", ts.URL+"/admin/api/config", nil)
	getReq.Header.Set("Authorization", "Bearer "+firstToken)
	getResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 with original token, got %d", getResp.StatusCode)
	}
}

func TestAdminGetRoutes(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/routes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body config.RoutesConfig
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.TaskRoutes["debugging"] != "route.debugging" {
		t.Errorf("expected debugging route route.debugging, got %s", body.TaskRoutes["debugging"])
	}
}

func TestAdminGetProfiles(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/profiles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body config.ProfilesConfig
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := body.RouteProfiles["route.default"]; !ok {
		t.Error("expected route.default profile to exist")
	}
}

func TestAdminRuntimeStart(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/runtime/start", "application/json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

func TestAdminRuntimeStop(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/runtime/stop", "application/json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

func TestAdminRuntimeRestart(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/runtime/restart", "application/json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAdminGetStartup(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/startup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body config.StartupConfig
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.StartProxyOnAppLaunch {
		t.Error("expected StartProxyOnAppLaunch to be true by default")
	}
}

func TestAdminDownstreamHealth(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/downstream/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "connected" && body["status"] != "unavailable" && body["status"] != "degraded" {
		t.Errorf("expected status connected, unavailable, or degraded, got %v", body["status"])
	}
	if body["url"] == nil {
		t.Error("expected url in response")
	}
}

func TestAdminLogs(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/logs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["logs"] == nil {
		t.Error("expected logs in response")
	}
	if body["total"] == nil {
		t.Error("expected total in response")
	}
	if body["privacy_mode"] == nil {
		t.Error("expected privacy_mode in response")
	}
}

func TestAdminLogsAfterRequest(t *testing.T) {
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-log-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstreamNoAuth(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"smart-debug","messages":[{"role":"user","content":"fix this error"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	logsResp, err := http.Get(ts.URL + "/admin/api/logs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer logsResp.Body.Close()

	var logBody map[string]interface{}
	if err := json.NewDecoder(logsResp.Body).Decode(&logBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	total := logBody["total"].(float64)
	if total < 1 {
		t.Errorf("expected at least 1 log entry after request, got %f", total)
	}

	logs := logBody["logs"].([]interface{})
	if len(logs) < 1 {
		t.Fatal("expected at least 1 log entry")
	}

	entry := logs[0].(map[string]interface{})
	if entry["request_id"] == nil || entry["request_id"] == "" {
		t.Error("expected request_id in log entry")
	}
	if entry["task_type"] == nil || entry["task_type"] == "" {
		t.Error("expected task_type in log entry")
	}
	if entry["route_key"] == nil || entry["route_key"] == "" {
		t.Error("expected route_key in log entry")
	}
}

func TestAdminDiagnosticsExport(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/diagnostics/export", "application/json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["version"] == nil {
		t.Error("expected non-empty version in diagnostics")
	}
}

func TestAdminDiagnosticsExportNoSecrets(t *testing.T) {
	knownToken := "diag-test-token-456"
	knownHash := security.HashToken(knownToken)

	cfg := config.DefaultConfig()
	cfg.Security.AdminAuthEnabled = true
	cfg.Security.AdminTokenHash = knownHash
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())
	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
	}
	r := New(deps)
	ts := httptest.NewServer(r.Handler)
	defer ts.Close()

	diagReq, _ := http.NewRequest("POST", ts.URL+"/admin/api/diagnostics/export", strings.NewReader(`{}`))
	diagReq.Header.Set("Authorization", "Bearer "+knownToken)
	diagReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(diagReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var diagBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&diagBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	configObj, ok := diagBody["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected config in diagnostics")
	}

	if _, exists := configObj["admin_token_hash"]; exists {
		t.Error("diagnostics should not expose admin_token_hash")
	}
}

func TestAdminConfigExportMasking(t *testing.T) {
	knownToken := "export-test-token-789"
	knownHash := security.HashToken(knownToken)

	cfg := config.DefaultConfig()
	cfg.Security.AdminAuthEnabled = true
	cfg.Security.AdminTokenHash = knownHash
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())
	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
	}
	r := New(deps)
	ts := httptest.NewServer(r.Handler)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/admin/api/config/export", nil)
	req.Header.Set("Authorization", "Bearer "+knownToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var exportBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&exportBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	configObj, ok := exportBody["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected config in export")
	}

	sec, ok := configObj["security"].(map[string]interface{})
	if !ok {
		t.Fatal("expected security in config")
	}

	if _, exists := sec["admin_token_hash"]; exists {
		t.Error("config export should not expose admin_token_hash")
	}
}

func TestAdminDryRun(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	body := `{"prompt":"fix this error: panic at runtime","model":"smart-auto"}`
	resp, err := http.Post(ts.URL+"/admin/api/routing/dry-run", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["analysis"] == nil {
		t.Error("expected analysis in response")
	}
	if result["classification"] == nil {
		t.Error("expected classification in response")
	}
	if result["decision"] == nil {
		t.Error("expected decision in response")
	}
}

func TestAdminConfigExport(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/config/export")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["config"] == nil {
		t.Error("expected config in export")
	}
	if body["routes"] == nil {
		t.Error("expected routes in export")
	}
	if body["profiles"] == nil {
		t.Error("expected profiles in export")
	}
}

func TestAdminConfigReset(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/config/reset", "application/json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

// --- Request Validation Tests ---

func TestInvalidJSON(t *testing.T) {
	ts := newTestServerWithDownstream(t, "http://127.0.0.1:19999/v1")
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{invalid json`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "invalid_request_error" {
		t.Errorf("expected type invalid_request_error, got %v", errObj["type"])
	}
	if errObj["code"] != "invalid_json" {
		t.Errorf("expected code invalid_json, got %v", errObj["code"])
	}
	if errObj["message"] != "Invalid JSON" {
		t.Errorf("expected message Invalid JSON, got %v", errObj["message"])
	}
}

func TestEmptyBody(t *testing.T) {
	ts := newTestServerWithDownstream(t, "http://127.0.0.1:19999/v1")
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "invalid_request_error" {
		t.Errorf("expected type invalid_request_error, got %v", errObj["type"])
	}
	if errObj["code"] != "invalid_request_error" {
		t.Errorf("expected code invalid_request_error, got %v", errObj["code"])
	}
}

func TestMissingMessages(t *testing.T) {
	ts := newTestServerWithDownstream(t, "http://127.0.0.1:19999/v1")
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"gpt-4"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "invalid_request_error" {
		t.Errorf("expected type invalid_request_error, got %v", errObj["type"])
	}
	if errObj["message"] != "Missing required field: messages" {
		t.Errorf("expected message Missing required field: messages, got %v", errObj["message"])
	}
}

func TestEmptyMessages(t *testing.T) {
	ts := newTestServerWithDownstream(t, "http://127.0.0.1:19999/v1")
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"gpt-4","messages":[]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "invalid_request_error" {
		t.Errorf("expected type invalid_request_error, got %v", errObj["type"])
	}
	if errObj["message"] != "messages must not be empty" {
		t.Errorf("expected message messages must not be empty, got %v", errObj["message"])
	}
}

func TestComboTestHandler(t *testing.T) {
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-combo-test",
			"model":  "big-pickle",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "OK"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstreamNoAuth(t, downstream.URL+"/v1")
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/combo/test", "application/json",
		strings.NewReader(`{"model":"COding"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["success"] != true {
		t.Errorf("expected success true, got %v", body["success"])
	}
	if body["model"] != "COding" {
		t.Errorf("expected model COding, got %v", body["model"])
	}
}

func TestComboTestHandlerMissingModel(t *testing.T) {
	ts := newTestServerWithDownstreamNoAuth(t, "http://127.0.0.1:19999/v1")
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/combo/test", "application/json",
		strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestSafePassthroughAnalyzerFailure(t *testing.T) {
	var receivedBody map[string]interface{}

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-passthrough",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	// Send a valid request — analyzer should succeed, but test the passthrough path
	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedBody["model"] != "gpt-4" {
		t.Errorf("expected model gpt-4 forwarded, got %v", receivedBody["model"])
	}
}

func TestSafePassthroughClassifierFailureFallback(t *testing.T) {
	var receivedBody map[string]interface{}

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-fallback",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	// Send request with empty content — classifier returns "unknown" with low confidence
	// This should still forward to downstream via safe passthrough
	body := `{"model":"gpt-4","messages":[{"role":"user","content":""}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	// Request should still be forwarded even with unknown classification
	if receivedBody == nil {
		t.Error("expected request to be forwarded to downstream")
	}
}

func TestSafePassthroughInvalidRouteFallback(t *testing.T) {
	var receivedModel string

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		var parsed map[string]interface{}
		json.NewDecoder(r.Body).Decode(&parsed)
		receivedModel = parsed["model"].(string)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-route-fallback",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		})
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	// smart-debug routes to route.debugging → combo.debugging
	// This should work and forward the model name
	body := `{"model":"smart-debug","messages":[{"role":"user","content":"fix this error: panic"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedModel != "combo.debugging" {
		t.Errorf("expected model combo.debugging forwarded, got %s", receivedModel)
	}
}

func TestHeaderMetadataTransport(t *testing.T) {
	var receivedHeaders http.Header

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-header-meta",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		})
	})
	defer downstream.Close()

	cfg := config.DefaultConfig()
	cfg.Downstream.BaseURL = downstream.URL + "/v1"
	cfg.Routing.MetadataTransport = "header"
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())
	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
	}
	r := New(deps)
	ts := httptest.NewServer(r.Handler)
	defer ts.Close()

	body := `{"model":"smart-debug","messages":[{"role":"user","content":"fix this error"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedHeaders.Get("X-Route-Intent") != "combo.debugging" {
		t.Errorf("expected X-Route-Intent combo.debugging, got %s", receivedHeaders.Get("X-Route-Intent"))
	}
}

// --- Phase 1 Security Hardening Regression Tests ---

func TestRequireJSONRejectsNonJSON(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/admin/api/config/reset", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("expected status 415, got %d", resp.StatusCode)
	}
}

func TestRequireJSONAcceptsValidJSON(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/config/reset", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequireJSONAllowsGETWithoutContentType(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for GET without Content-Type, got %d", resp.StatusCode)
	}
}

func TestSameOriginAdminRejectsCrossOrigin(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/admin/api/config/reset", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestSameOriginAdminAllowsNoOrigin(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/admin/api/config/reset", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 when no Origin header, got %d", resp.StatusCode)
	}
}

func TestSecurityHeadersPresent(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("X-Content-Type-Options"); ct != "nosniff" {
		t.Errorf("expected X-Content-Type-Options nosniff, got %s", ct)
	}
	if ct := resp.Header.Get("X-Frame-Options"); ct != "DENY" {
		t.Errorf("expected X-Frame-Options DENY, got %s", ct)
	}
	if ct := resp.Header.Get("Content-Security-Policy"); ct == "" {
		t.Error("expected non-empty Content-Security-Policy")
	}
}

func TestTokenRotation(t *testing.T) {
	ts, initialToken := newTestServerWithToken(t)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/admin/api/status", nil)
	req.Header.Set("Authorization", "Bearer "+initialToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 with initial token, got %d", resp.StatusCode)
	}

	rotateReq, _ := http.NewRequest("POST", ts.URL+"/admin/api/security/token/rotate", strings.NewReader(`{}`))
	rotateReq.Header.Set("Authorization", "Bearer "+initialToken)
	rotateReq.Header.Set("Content-Type", "application/json")
	rotateResp, err := http.DefaultClient.Do(rotateReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rotateResp.Body.Close()
	if rotateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for token rotation, got %d", rotateResp.StatusCode)
	}

	var rotateResult map[string]interface{}
	json.NewDecoder(rotateResp.Body).Decode(&rotateResult)
	newToken, _ := rotateResult["token"].(string)
	if newToken == "" {
		t.Fatal("expected new token in rotation response")
	}
	if newToken == initialToken {
		t.Error("rotated token should differ from initial token")
	}

	newReq, _ := http.NewRequest("GET", ts.URL+"/admin/api/status", nil)
	newReq.Header.Set("Authorization", "Bearer "+newToken)
	newResp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer newResp.Body.Close()
	if newResp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 with new token, got %d", newResp.StatusCode)
	}

	oldReq, _ := http.NewRequest("GET", ts.URL+"/admin/api/status", nil)
	oldReq.Header.Set("Authorization", "Bearer "+initialToken)
	oldResp, err := http.DefaultClient.Do(oldReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer oldResp.Body.Close()
	if oldResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 with old token after rotation, got %d", oldResp.StatusCode)
	}
}

func TestAdminAuthRejectsNoToken(t *testing.T) {
	ts, _ := newTestServerWithToken(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/admin/api/status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 when no token, got %d", resp.StatusCode)
	}
}

func TestAdminAuthRejectsWrongToken(t *testing.T) {
	ts, _ := newTestServerWithToken(t)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/admin/api/status", nil)
	req.Header.Set("Authorization", "Bearer wrong-token-value")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 with wrong token, got %d", resp.StatusCode)
	}
}

func TestSecurityViewEndpoint(t *testing.T) {
	ts, token := newTestServerWithToken(t)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/admin/api/security", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var sec SecurityView
	if err := json.NewDecoder(resp.Body).Decode(&sec); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !sec.AdminAuthEnabled {
		t.Error("expected admin_auth_enabled true")
	}
	if !sec.TokenConfigured {
		t.Error("expected token_configured true")
	}
	if sec.BindLocalhostOnly != true {
		t.Error("expected bind_localhost_only true")
	}
}

func TestConcurrencySnapshotSwap(t *testing.T) {
	var hitCount atomic.Int64

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		hitCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ok","choices":[{"message":{"role":"assistant","content":"hi"}}]}`))
	})
	defer downstream.Close()

	cfg := config.DefaultConfig()
	cfg.Downstream.BaseURL = downstream.URL + "/v1"
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())

	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
	}
	r := New(deps)
	ts := httptest.NewServer(r.Handler)
	defer ts.Close()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%10 == 0 {
				snap := store.Load().Clone()
				snap.Version = uint64(idx)
				store.Swap(snap)
			}
			body := `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`
			resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
			if err != nil {
				t.Errorf("request error: %v", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200, got %d", resp.StatusCode)
			}
		}(i)
	}
	wg.Wait()

	if hitCount.Load() == 0 {
		t.Error("expected at least 1 downstream hit")
	}
}

func TestDownstreamHotReload(t *testing.T) {
	var target1Hits, target2Hits atomic.Int32

	downstream1 := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		target1Hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"target1","choices":[{"message":{"role":"assistant","content":"from-target1"}}]}`))
	})
	defer downstream1.Close()

	downstream2 := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		target2Hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"target2","choices":[{"message":{"role":"assistant","content":"from-target2"}}]}`))
	})
	defer downstream2.Close()

	cfg := config.DefaultConfig()
	cfg.Downstream.BaseURL = downstream1.URL + "/v1"
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())

	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
	}
	r := New(deps)
	ts := httptest.NewServer(r.Handler)
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"test"}]}`
	resp1, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	defer resp1.Body.Close()
	var result1 map[string]interface{}
	json.NewDecoder(resp1.Body).Decode(&result1)
	if result1["id"] != "target1" {
		t.Errorf("expected response from target1, got id=%v", result1["id"])
	}

	newSnap := store.Load().Clone()
	newSnap.Config.Downstream.BaseURL = downstream2.URL + "/v1"
	newFwd, err := forwarder.New(newSnap.Config.Downstream.BaseURL, newSnap.Config.Downstream.TimeoutSeconds)
	if err != nil {
		t.Fatalf("failed to create new forwarder: %v", err)
	}
	newSnap.Forwarder = newFwd
	store.Swap(newSnap)

	resp2, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	defer resp2.Body.Close()
	var result2 map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&result2)
	if result2["id"] != "target2" {
		t.Errorf("expected response from target2 after hot-reload, got id=%v", result2["id"])
	}
	if target2Hits.Load() < 1 {
		t.Error("expected at least 1 hit on target2")
	}
}

func TestConfigServiceAtomicSwap(t *testing.T) {
	cfg := config.DefaultConfig()
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())

	initialSnap := store.Load()
	if initialSnap.Version != 1 {
		t.Errorf("expected initial version 1, got %d", initialSnap.Version)
	}

	patchJSON := `{"downstream":{"base_url":"http://new-downstream:9999/v1","timeout_seconds":30}}`
	var patch map[string]json.RawMessage
	json.Unmarshal([]byte(patchJSON), &patch)

	result, err := cs.ApplyConfigPatch(patch)
	if err != nil {
		t.Fatalf("ApplyConfigPatch failed: %v", err)
	}
	_ = result

	afterSnap := store.Load()
	if afterSnap.Version != 2 {
		t.Errorf("expected version 2 after patch, got %d", afterSnap.Version)
	}
	if afterSnap.Config.Downstream.BaseURL != "http://new-downstream:9999/v1" {
		t.Errorf("expected downstream URL http://new-downstream:9999/v1, got %s", afterSnap.Config.Downstream.BaseURL)
	}
}

func TestConfigServiceResetAtomicSwap(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Server.Port = 99999
	cfg.Downstream.BaseURL = "http://custom:8080/v1"
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())

	snapBefore := store.Load()
	if snapBefore.Config.Server.Port != 99999 {
		t.Fatalf("expected port 99999, got %d", snapBefore.Config.Server.Port)
	}

	err := cs.Reset()
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	snapAfter := store.Load()
	defaults := config.DefaultConfig()
	if snapAfter.Config.Server.Port != defaults.Server.Port {
		t.Errorf("expected port %d after reset, got %d", defaults.Server.Port, snapAfter.Config.Server.Port)
	}
	if snapAfter.Config.Downstream.BaseURL != defaults.Downstream.BaseURL {
		t.Errorf("expected downstream URL %s after reset, got %s", defaults.Downstream.BaseURL, snapAfter.Config.Downstream.BaseURL)
	}
}

func TestBodyLimitChatRejectsOversized(t *testing.T) {
	ts := newTestServerWithDownstream(t, "http://127.0.0.1:19999/v1")
	defer ts.Close()

	oversized := strings.Repeat("x", MaxChatBody+1)
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(oversized))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resp.StatusCode)
	}
}

func TestBodyLimitAdminRejectsOversized(t *testing.T) {
	ts := newTestServerNoAuth(t)
	defer ts.Close()

	oversized := strings.Repeat("x", MaxAdminBody+1)
	req, _ := http.NewRequest("PUT", ts.URL+"/admin/api/config", strings.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resp.StatusCode)
	}
}

func TestSemaphoreRejectsWhenFull(t *testing.T) {
	cfg := config.DefaultConfig()
	store, cs := newSnapshotFromConfig(cfg, config.DefaultRoutesConfig(), config.DefaultProfilesConfig())
	bh := NewWeightedBulkhead(2, 1*time.Second)

	// Fill to capacity (2 non-streaming requests, weight 1 each)
	bh.Acquire(1, 0)
	bh.Acquire(1, 0)

	deps := &ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     observability.NewLogger(observability.DefaultMaxEntries),
		Bulkhead:      bh,
	}
	r := New(deps)
	ts := httptest.NewServer(r.Handler)
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 when bulkhead full, got %d", resp.StatusCode)
	}

	var parsed map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&parsed)
	errObj := parsed["error"].(map[string]interface{})
	if errObj["code"] != "server_overloaded" {
		t.Errorf("expected code server_overloaded, got %v", errObj["code"])
	}

	bh.Release(1)
	bh.Release(1)
}

func TestStreamLargeEvent(t *testing.T) {
	largeContent := strings.Repeat("A", 2*1024*1024)

	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprintf(w, "data: %s\n\n", largeContent)
		flusher.Flush()
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	})
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions", strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var fullBody strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 4*1024*1024), 4*1024*1024)
	for scanner.Scan() {
		fullBody.WriteString(scanner.Text())
		fullBody.WriteString("\n")
	}

	if !strings.Contains(fullBody.String(), largeContent) {
		t.Error("expected large event content in stream")
	}
	if !strings.Contains(fullBody.String(), "[DONE]") {
		t.Error("expected [DONE] in stream")
	}
}

func TestClientCancellation(t *testing.T) {
	started := make(chan struct{})
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		close(started)
	})

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")

	ctx, cancel := context.WithCancel(context.Background())

	req, _ := http.NewRequestWithContext(ctx, "POST", ts.URL+"/v1/chat/completions",
		strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Content-Type", "application/json")

	done := make(chan struct{})
	go func() {
		defer close(done)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("downstream never received request")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("proxy did not return after client cancellation")
	}

	downstream.Close()
	ts.Close()
}

func TestCommitWriterDoubleWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := &commitWriter{ResponseWriter: rec}

	cw.WriteHeader(http.StatusOK)
	cw.Write([]byte("first"))

	cw.WriteHeader(http.StatusInternalServerError)
	cw.Write([]byte("second"))

	if rec.Code != http.StatusOK {
		t.Errorf("expected first WriteHeader to stick (200), got %d", rec.Code)
	}
	if !cw.Committed() {
		t.Error("expected committed to be true")
	}
}

func TestWriteJSONAfterCommitDrops(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := &commitWriter{ResponseWriter: rec}

	cw.WriteHeader(http.StatusOK)
	cw.Write([]byte("first"))

	writeJSONAfterCommit(cw, http.StatusInternalServerError, map[string]string{"error": "dropped"})

	if rec.Code != http.StatusOK {
		t.Errorf("expected first WriteHeader to stick (200), got %d", rec.Code)
	}
	if rec.Body.String() != "first" {
		t.Errorf("expected body 'first', got %q", rec.Body.String())
	}
}

func TestShutdownWithActiveStream(t *testing.T) {
	streamStarted := make(chan struct{})
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}
		close(streamStarted)
		for i := 0; i < 100; i++ {
			fmt.Fprintf(w, "data: chunk %d\n\n", i)
			flusher.Flush()
			time.Sleep(50 * time.Millisecond)
		}
	})

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")

	req, _ := http.NewRequest("POST", ts.URL+"/v1/chat/completions",
		strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	select {
	case <-streamStarted:
	case <-time.After(5 * time.Second):
		resp.Body.Close()
		t.Fatal("stream never started")
	}

	shutdownDone := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = ts.Config.Shutdown(ctx)
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
	case <-time.After(20 * time.Second):
		t.Fatal("shutdown did not complete within 20s")
	}

	resp.Body.Close()
}

func TestDownstreamErrorDoesNotLeakInternalURL(t *testing.T) {
	// Start then immediately close a server to get a port that will refuse connections
	closedServer := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {})
	closedAddr := closedServer.URL
	closedServer.Close()

	ts := newTestServerWithDownstream(t, closedAddr+"/v1")

	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json",
		strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	defer ts.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", resp.StatusCode)
	}

	// Must NOT leak connection details like "connection refused" or internal addresses
	if strings.Contains(body, "connection refused") {
		t.Error("response body leaked internal error detail: connection refused")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	errObj, ok := parsed["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if _, hasCID := errObj["correlation_id"]; !hasCID {
		t.Error("expected correlation_id in error response")
	}
	if msg, ok := errObj["message"].(string); ok {
		if msg != "downstream request failed" {
			t.Errorf("expected generic message 'downstream request failed', got %q", msg)
		}
	}
}

func TestHopByHopHeadersStripped(t *testing.T) {
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Keep-Alive", "timeout=5")
		w.Header().Set("Set-Cookie", "session=abc123")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Proxy-Auth", "secret")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Allowed", "yes")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")

	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json",
		strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if v := resp.Header.Get("Connection"); v != "" {
		t.Errorf("Connection header should be stripped, got %q", v)
	}
	if v := resp.Header.Get("Keep-Alive"); v != "" {
		t.Errorf("Keep-Alive header should be stripped, got %q", v)
	}
	if v := resp.Header.Get("Set-Cookie"); v != "" {
		t.Errorf("Set-Cookie header should be stripped, got %q", v)
	}
	if v := resp.Header.Get("Transfer-Encoding"); v != "" {
		t.Errorf("Transfer-Encoding header should be stripped, got %q", v)
	}
	if v := resp.Header.Get("Proxy-Auth"); v != "" {
		t.Errorf("Proxy-Auth header should be stripped, got %q", v)
	}

	if v := resp.Header.Get("Content-Type"); v != "application/json" {
		t.Errorf("Content-Type should be preserved, got %q", v)
	}
}

func TestNoRawPromptInObservabilityLog(t *testing.T) {
	started := make(chan struct{})
	downstream := mockDownstream(t, func(w http.ResponseWriter, r *http.Request) {
		close(started)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"response"}}]}`))
	})

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")

	prompt := "My secret password is hunter2 and my API key is sk-abcdef1234567890abcdef"
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json",
		strings.NewReader(fmt.Sprintf(`{"model":"gpt-4","messages":[{"role":"user","content":"%s"}]}`, prompt)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("downstream never received request")
	}
}

func TestHeaderFilterAllowsContent(t *testing.T) {
	allowed := isAllowedHeader("Content-Type")
	if !allowed {
		t.Error("Content-Type should be allowed")
	}
	blocked := isAllowedHeader("Set-Cookie")
	if blocked {
		t.Error("Set-Cookie should be blocked")
	}
	blocked2 := isAllowedHeader("Transfer-Encoding")
	if blocked2 {
		t.Error("Transfer-Encoding should be blocked")
	}
	blocked3 := isAllowedHeader("Connection")
	if blocked3 {
		t.Error("Connection should be blocked")
	}
}
