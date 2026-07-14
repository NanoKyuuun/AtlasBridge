package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildMockChatBody(t *testing.T) {
	body, err := buildMockChatBody("gpt-4", "Hello world")
	if err != nil {
		t.Fatal(err)
	}

	var parsed mockChatRequest
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", parsed.Model)
	}
	if len(parsed.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(parsed.Messages))
	}
	if parsed.Messages[0].Role != "user" {
		t.Errorf("expected role user, got %s", parsed.Messages[0].Role)
	}
	if parsed.Messages[0].Content != "Hello world" {
		t.Errorf("expected content 'Hello world', got %s", parsed.Messages[0].Content)
	}
}

func TestBuildMockChatBody_EscapesSpecialChars(t *testing.T) {
	prompt := `She said "hello" and then\nnewlined`
	body, err := buildMockChatBody("gpt-4", prompt)
	if err != nil {
		t.Fatal(err)
	}

	var parsed mockChatRequest
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid JSON after escaping: %v", err)
	}

	if parsed.Messages[0].Content != prompt {
		t.Errorf("content mismatch: expected %q, got %q", prompt, parsed.Messages[0].Content)
	}
}

func TestBuildComboTestBody(t *testing.T) {
	body, err := buildComboTestBody("gpt-4")
	if err != nil {
		t.Fatal(err)
	}

	var parsed comboTestRequest
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", parsed.Model)
	}
	if parsed.MaxTokens != 5 {
		t.Errorf("expected max_tokens 5, got %d", parsed.MaxTokens)
	}
	if len(parsed.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(parsed.Messages))
	}
}

func TestBuildComboTestBody_EscapesQuotes(t *testing.T) {
	model := `model"with"quotes`
	body, err := buildComboTestBody(model)
	if err != nil {
		t.Fatal(err)
	}

	var parsed comboTestRequest
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed.Model != model {
		t.Errorf("expected model %q, got %q", model, parsed.Model)
	}
}

func TestSanitizModelName_Truncates(t *testing.T) {
	long := strings.Repeat("m", 300)
	result := sanitizeModelName(long)
	if len(result) != MaxModelName {
		t.Errorf("expected length %d, got %d", MaxModelName, len(result))
	}
}

func TestDecodeChatRequest_Valid(t *testing.T) {
	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`
	env, err := decodeChatRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if env.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", env.Model)
	}
	if ve := env.Validate(); ve != nil {
		t.Errorf("expected valid, got: %s", ve.Message)
	}
}

func TestDecodeChatRequest_EmptyMessages(t *testing.T) {
	body := `{"model":"gpt-4","messages":[]}`
	env, err := decodeChatRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if ve := env.Validate(); ve == nil {
		t.Error("expected validation error for empty messages")
	}
}

func TestDecodeChatRequest_MissingMessages(t *testing.T) {
	body := `{"model":"gpt-4"}`
	env, err := decodeChatRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if ve := env.Validate(); ve == nil {
		t.Error("expected validation error for missing messages")
	}
}

func TestDecodeChatRequest_TooManyMessages(t *testing.T) {
	msgs := make([]string, MaxMessages+1)
	for i := range msgs {
		msgs[i] = `{"role":"user","content":"hi"}`
	}
	body := `{"model":"gpt-4","messages":[` + strings.Join(msgs, ",") + `]}`
	env, err := decodeChatRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if ve := env.Validate(); ve == nil {
		t.Error("expected validation error for too many messages")
	}
}

func TestDecodeChatRequest_MessageTooLong(t *testing.T) {
	longContent := strings.Repeat("x", MaxMessageLen+1)
	body := `{"model":"gpt-4","messages":[{"role":"user","content":"` + longContent + `"}]}`
	env, err := decodeChatRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if ve := env.Validate(); ve == nil {
		t.Error("expected validation error for message content too long")
	}
}

func TestDecodeChatRequest_ModelTooLong(t *testing.T) {
	longModel := strings.Repeat("g", MaxModelName+1)
	body := `{"model":"` + longModel + `","messages":[{"role":"user","content":"hi"}]}`
	env, err := decodeChatRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if ve := env.Validate(); ve == nil {
		t.Error("expected validation error for model name too long")
	}
}

func TestDecodeChatRequest_InvalidJSON(t *testing.T) {
	_, err := decodeChatRequest([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDecodeChatRequest_EmptyBody(t *testing.T) {
	_, err := decodeChatRequest([]byte{})
	if err == nil {
		t.Error("expected error for empty body")
	}
}

func TestDecodeChatRequest_Stream(t *testing.T) {
	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"stream":true}`
	env, err := decodeChatRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if !env.IsStream() {
		t.Error("expected stream to be true")
	}
}
