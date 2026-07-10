package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/security"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	cfg := config.DefaultConfig()
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	deps := &ServerDeps{
		Config:   cfg,
		Routes:   routes,
		Profiles: profiles,
	}
	r := New(deps)
	return httptest.NewServer(r.Handler)
}

func newTestServerWithDownstream(t *testing.T, downstreamURL string) *httptest.Server {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Downstream.BaseURL = downstreamURL
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	deps := &ServerDeps{
		Config:   cfg,
		Routes:   routes,
		Profiles: profiles,
	}
	r := New(deps)
	return httptest.NewServer(r.Handler)
}

func mockDownstream(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(handler)
	return ts
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
	ts := newTestServer(t)
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

func TestAdminAPIStatusContentType(t *testing.T) {
	ts := newTestServer(t)
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
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	deps := &ServerDeps{
		Config:   cfg,
		Routes:   routes,
		Profiles: profiles,
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
	ts := newTestServer(t)
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

func TestAdminGetConfigMasking(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	cfg := config.DefaultConfig()
	cfg.Security.AdminTokenHash = "supersecret123"
	body, _ := json.Marshal(cfg)
	req, _ := http.NewRequest("PUT", ts.URL+"/admin/api/config", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	resp, err := http.Get(ts.URL + "/admin/api/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	security, ok := result["security"].(map[string]interface{})
	if !ok {
		t.Fatal("expected security object in response")
	}

	hash, ok := security["admin_token_hash"].(string)
	if !ok {
		t.Fatal("expected admin_token_hash in security")
	}

	if hash == "supersecret123" {
		t.Error("admin_token_hash should be masked, got raw value")
	}
	if hash == "" {
		t.Error("admin_token_hash should not be empty when set")
	}
	if len(hash) < 4 {
		t.Error("masked hash should be at least 4 characters")
	}
}

func TestAdminPutConfig(t *testing.T) {
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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

	hash, ok := sec["admin_token_hash"].(string)
	if !ok {
		t.Fatal("expected admin_token_hash in security")
	}
	if hash == "" {
		t.Error("admin_token_hash should be populated when auth enabled")
	}
	if hash == "****" {
		t.Error("admin_token_hash should not be the masked placeholder")
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
	ts := newTestServer(t)
	defer ts.Close()

	knownToken := "my-test-token-123"
	knownHash := security.HashToken(knownToken)

	patch := fmt.Sprintf(`{"security":{"admin_auth_enabled":true,"admin_token_hash":"%s"}}`, knownHash)
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

	getReq, _ := http.NewRequest("GET", ts.URL+"/admin/api/config", nil)
	getReq.Header.Set("Authorization", "Bearer "+knownToken)
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

	hash, ok := sec["admin_token_hash"].(string)
	if !ok {
		t.Fatal("expected admin_token_hash in security")
	}
	if hash == "" {
		t.Error("admin_token_hash should not be empty")
	}
	if hash == knownHash {
		t.Error("admin_token_hash should be masked in GET response")
	}
	if len(hash) < 8 {
		t.Error("masked hash should be at least 8 characters")
	}
}

func TestAdminGetRoutes(t *testing.T) {
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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
	if body["status"] != "connected" && body["status"] != "unavailable" {
		t.Errorf("expected status connected or unavailable, got %v", body["status"])
	}
	if body["url"] == nil {
		t.Error("expected url in response")
	}
}

func TestAdminLogs(t *testing.T) {
	ts := newTestServer(t)
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

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
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
	ts := newTestServer(t)
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
	ts := newTestServer(t)
	defer ts.Close()

	knownToken := "diag-test-token-456"
	knownHash := security.HashToken(knownToken)

	cfg := config.DefaultConfig()
	cfg.Security.AdminTokenHash = knownHash
	cfg.Security.AdminAuthEnabled = true
	body, _ := json.Marshal(cfg)
	req, _ := http.NewRequest("PUT", ts.URL+"/admin/api/config", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	diagReq, _ := http.NewRequest("POST", ts.URL+"/admin/api/diagnostics/export", nil)
	diagReq.Header.Set("Authorization", "Bearer "+knownToken)
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
	ts := newTestServer(t)
	defer ts.Close()

	cfg := config.DefaultConfig()
	cfg.Security.AdminTokenHash = "mysecrettoken"
	body, _ := json.Marshal(cfg)
	req, _ := http.NewRequest("PUT", ts.URL+"/admin/api/config", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	resp, err := http.Get(ts.URL + "/admin/api/config/export")
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

	security, ok := configObj["security"].(map[string]interface{})
	if !ok {
		t.Fatal("expected security in config")
	}

	hash, ok := security["admin_token_hash"].(string)
	if !ok {
		t.Fatal("expected admin_token_hash in security")
	}

	if hash == "mysecrettoken" {
		t.Error("admin_token_hash should be masked in export")
	}
}

func TestAdminDryRun(t *testing.T) {
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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
	ts := newTestServer(t)
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

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
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
	ts := newTestServerWithDownstream(t, "http://127.0.0.1:19999/v1")
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
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	deps := &ServerDeps{
		Config:   cfg,
		Routes:   routes,
		Profiles: profiles,
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
