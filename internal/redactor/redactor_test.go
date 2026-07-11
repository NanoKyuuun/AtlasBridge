package redactor

import (
	"errors"
	"testing"
)

func TestRedactLogStripsSensitiveKeys(t *testing.T) {
	r := NewDefault()
	fields := map[string]any{
		"method":       "POST",
		"path":         "/v1/chat/completions",
		"authorization": "Bearer sk-test1234567890abcdef",
		"prompt":       "What is the meaning of life?",
		"api_key":      "secret-key-value",
		"status":       200,
	}

	result := r.RedactLog(fields)

	if result["method"] != "POST" {
		t.Errorf("expected method to be preserved, got %v", result["method"])
	}
	if result["path"] != "/v1/chat/completions" {
		t.Errorf("expected path to be preserved, got %v", result["path"])
	}
	if result["authorization"] != "[REDACTED]" {
		t.Errorf("expected authorization to be redacted, got %v", result["authorization"])
	}
	if result["prompt"] != "[REDACTED]" {
		t.Errorf("expected prompt to be redacted, got %v", result["prompt"])
	}
	if result["api_key"] != "[REDACTED]" {
		t.Errorf("expected api_key to be redacted, got %v", result["api_key"])
	}
	if result["status"] != 200 {
		t.Errorf("expected status to be preserved, got %v", result["status"])
	}
}

func TestRedactErrorStripsSecrets(t *testing.T) {
	r := NewDefault()

	err := errors.New("connection to sk-liveabcdef1234567890abcdef failed")
	result := r.RedactError(err)

	if result == err.Error() {
		t.Error("expected error message to be redacted, but it was not")
	}
	if len(result) == 0 {
		t.Error("expected non-empty redacted error")
	}
}

func TestRedactErrorNilReturnsEmpty(t *testing.T) {
	r := NewDefault()
	result := r.RedactError(nil)
	if result != "" {
		t.Errorf("expected empty string for nil error, got %q", result)
	}
}

func TestNopRedactorPassesThrough(t *testing.T) {
	r := NopRedactor{}
	fields := map[string]any{
		"authorization": "Bearer secret",
		"status":        200,
	}
	result := r.RedactLog(fields)
	if result["authorization"] != "Bearer secret" {
		t.Errorf("NopRedactor should pass through, got %v", result["authorization"])
	}
}

func TestRedactLogPreservesNonStringValue(t *testing.T) {
	r := NewDefault()
	fields := map[string]any{
		"count":   42,
		"enabled": true,
		"data":    []string{"a", "b"},
	}
	result := r.RedactLog(fields)
	if result["count"] != 42 {
		t.Errorf("expected count 42, got %v", result["count"])
	}
	if result["enabled"] != true {
		t.Errorf("expected enabled true, got %v", result["enabled"])
	}
}
