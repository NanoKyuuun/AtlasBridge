package forwarder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewValidURL(t *testing.T) {
	f, err := New("http://127.0.0.1:20128/v1", 120)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.base != "http://127.0.0.1:20128/v1" {
		t.Errorf("expected base http://127.0.0.1:20128/v1, got %s", f.base)
	}
}

func TestNewInvalidURL(t *testing.T) {
	_, err := New("://invalid", 120)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestNewDefaultTimeout(t *testing.T) {
	f, err := New("http://127.0.0.1:20128/v1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.nonstreamClient.Timeout != 120*time.Second {
		t.Errorf("expected default timeout 120s, got %v", f.nonstreamClient.Timeout)
	}
}

func TestForwardSuccess(t *testing.T) {
	var receivedBody []byte
	var receivedAuth string
	var receivedCT string

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		receivedAuth = r.Header.Get("Authorization")
		receivedCT = r.Header.Get("Content-Type")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "hello"}},
			},
		})
	}))
	defer mock.Close()

	f, err := New(mock.URL+"/v1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret-key")

	result, err := f.Forward(context.Background(), req, "req-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
	if string(receivedBody) != body {
		t.Errorf("body mismatch: got %s", string(receivedBody))
	}
	if receivedAuth != "" {
		t.Errorf("Authorization should not be forwarded, got %s", receivedAuth)
	}
	if receivedCT != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", receivedCT)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["id"] != "chatcmpl-test" {
		t.Errorf("expected id chatcmpl-test, got %v", resp["id"])
	}
}

func TestForwardPreservesUnknownFields(t *testing.T) {
	var receivedBody map[string]interface{}

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "ok"})
	}))
	defer mock.Close()

	f, _ := New(mock.URL+"/v1", 10)

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"temperature":0.7,"max_tokens":100,"custom_field":"custom_value"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	_, err := f.Forward(context.Background(), req, "req-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["temperature"] != 0.7 {
		t.Errorf("expected temperature 0.7, got %v", receivedBody["temperature"])
	}
	if receivedBody["max_tokens"] != float64(100) {
		t.Errorf("expected max_tokens 100, got %v", receivedBody["max_tokens"])
	}
	if receivedBody["custom_field"] != "custom_value" {
		t.Errorf("expected custom_field custom_value, got %v", receivedBody["custom_field"])
	}
}

func TestForwardDownstreamUnavailable(t *testing.T) {
	f, err := New("http://127.0.0.1:19999/v1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	_, err = f.Forward(context.Background(), req, "req-err")
	if err == nil {
		t.Error("expected error for unavailable downstream")
	}
}

func TestForwardDownstreamError(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "provider error"},
		})
	}))
	defer mock.Close()

	f, _ := New(mock.URL+"/v1", 10)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	result, err := f.Forward(context.Background(), req, "req-502")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", result.StatusCode)
	}
}

func TestForwardRequestIDHeader(t *testing.T) {
	var receivedReqID string

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedReqID = r.Header.Get("X-Request-ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "ok"})
	}))
	defer mock.Close()

	f, _ := New(mock.URL+"/v1", 10)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	_, err := f.Forward(context.Background(), req, "my-req-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedReqID != "my-req-id" {
		t.Errorf("expected X-Request-ID my-req-id, got %s", receivedReqID)
	}
}

func TestForwardContextCancelled(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	f, _ := New(mock.URL+"/v1", 10)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	_, err := f.Forward(ctx, req, "req-cancel")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestForwardStreamSuccess(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		}
	}))
	defer mock.Close()

	f, _ := New(mock.URL+"/v1", 10)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	err := f.ForwardStream(context.Background(), req, rec, "stream-req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("X-Request-ID") != "stream-req-1" {
		t.Errorf("expected X-Request-ID stream-req-1, got %s", rec.Header().Get("X-Request-ID"))
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Hello") {
		t.Error("expected body to contain Hello")
	}
	if !strings.Contains(body, "World") {
		t.Error("expected body to contain World")
	}
	if !strings.Contains(body, "[DONE]") {
		t.Error("expected body to contain [DONE]")
	}
}

func TestForwardStreamDownstreamUnavailable(t *testing.T) {
	f, err := New("http://127.0.0.1:19999/v1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	err = f.ForwardStream(context.Background(), req, rec, "stream-err")
	if err == nil {
		t.Error("expected error for unavailable downstream")
	}
}

func TestForwardStreamDownstreamError(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "provider error"},
		})
	}))
	defer mock.Close()

	f, _ := New(mock.URL+"/v1", 10)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	err := f.ForwardStream(context.Background(), req, rec, "stream-502")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestForwardStreamPreservesFields(t *testing.T) {
	var receivedBody map[string]interface{}

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "data: {\"done\":true}\n\n")
	}))
	defer mock.Close()

	f, _ := New(mock.URL+"/v1", 10)

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"stream":true,"temperature":0.5}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	err := f.ForwardStream(context.Background(), req, rec, "stream-fields")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["stream"] != true {
		t.Errorf("expected stream true, got %v", receivedBody["stream"])
	}
	if receivedBody["temperature"] != 0.5 {
		t.Errorf("expected temperature 0.5, got %v", receivedBody["temperature"])
	}
}

func TestIsStreamRequestWithStreamTrue(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	if !IsStreamRequest(req) {
		t.Error("expected stream true")
	}
}

func TestIsStreamRequestWithStreamFalse(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"stream":false}`))
	req.Header.Set("Content-Type", "application/json")
	if IsStreamRequest(req) {
		t.Error("expected stream false")
	}
}

func TestIsStreamRequestWithNoStream(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gpt-4"}`))
	req.Header.Set("Content-Type", "application/json")
	if IsStreamRequest(req) {
		t.Error("expected no stream")
	}
}

func TestIsStreamRequestWithAcceptHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Accept", "text/event-stream")
	if !IsStreamRequest(req) {
		t.Error("expected stream via Accept header")
	}
}

func TestIsStreamRequestEmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(``))
	if IsStreamRequest(req) {
		t.Error("expected no stream for empty body")
	}
}

func TestForwardStreamContextCancelled(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		time.Sleep(2 * time.Second)
		fmt.Fprintf(w, "data: {\"done\":true}\n\n")
	}))
	defer mock.Close()

	f, _ := New(mock.URL+"/v1", 10)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	err := f.ForwardStream(ctx, req, rec, "stream-cancel")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
