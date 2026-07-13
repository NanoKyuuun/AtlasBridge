package server

import (
	"encoding/json"
)

const (
	MaxMessages     = 128
	MaxMessageLen   = 32 << 10
	MaxModelName    = 256
	MaxRequestIDLen = 128
	MaxRouteKeyLen  = 128
)

type mockChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type mockChatRequest struct {
	Model    string            `json:"model"`
	Messages []mockChatMessage `json:"messages"`
}

type comboTestRequest struct {
	Model     string            `json:"model"`
	Messages  []mockChatMessage `json:"messages"`
	MaxTokens int               `json:"max_tokens"`
}

func buildMockChatBody(model, prompt string) ([]byte, error) {
	req := mockChatRequest{
		Model: model,
		Messages: []mockChatMessage{
			{Role: "user", Content: prompt},
		},
	}
	return json.Marshal(req)
}

func buildComboTestBody(model string) ([]byte, error) {
	req := comboTestRequest{
		Model: model,
		Messages: []mockChatMessage{
			{Role: "user", Content: "Reply with only: OK"},
		},
		MaxTokens: 5,
	}
	return json.Marshal(req)
}

func validateRequestID(id string) string {
	if len(id) > MaxRequestIDLen {
		return id[:MaxRequestIDLen]
	}
	return id
}

func validateRouteKey(key string) string {
	if len(key) > MaxRouteKeyLen {
		return key[:MaxRouteKeyLen]
	}
	return key
}

func sanitizeModelName(model string) string {
	if len(model) > MaxModelName {
		return model[:MaxModelName]
	}
	return model
}
