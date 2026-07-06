package analyzer

import (
	"encoding/json"
	"strings"
	"testing"
)

func makeRequest(model string, messages []Message, stream *bool) string {
	req := ChatRequest{
		Model:    model,
		Stream:   stream,
		Messages: messages,
	}
	data, _ := json.Marshal(req)
	return string(data)
}

func makeMessage(role, content string) Message {
	return Message{Role: role, Content: content}
}

func TestAnalyzeBasic(t *testing.T) {
	body := makeRequest("gpt-4", []Message{
		makeMessage("user", "Hello world"),
	}, nil)

	a, err := Analyze(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.RequestedModel != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", a.RequestedModel)
	}
	if a.Stream != false {
		t.Error("expected stream false")
	}
	if a.MessageCount != 1 {
		t.Errorf("expected 1 message, got %d", a.MessageCount)
	}
}

func TestAnalyzeStreamTrue(t *testing.T) {
	stream := true
	body := makeRequest("gpt-4", []Message{
		makeMessage("user", "Hi"),
	}, &stream)

	a, err := Analyze(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.Stream {
		t.Error("expected stream true")
	}
}

func TestAnalyzeStreamFalse(t *testing.T) {
	stream := false
	body := makeRequest("gpt-4", []Message{
		makeMessage("user", "Hi"),
	}, &stream)

	a, err := Analyze(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Stream {
		t.Error("expected stream false")
	}
}

func TestAnalyzeCodeBlockDetection(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"fenced code block", "Here is some code:\n```go\nfunc main() {}\n```", true},
		{"inline code", "Use `fmt.Println()` to print", true},
		{"go func", "You can use func main() to start", true},
		{"go package", "package main is the entry point", true},
		{"python def", "Use def my_function(): to define", true},
		{"python class", "class MyClass inherits from object", true},
		{"js function", "function hello() { return true; }", true},
		{"const declaration", "const x = 10;", true},
		{"no code", "This is a plain text message about cooking", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := makeRequest("gpt-4", []Message{
				makeMessage("user", tt.content),
			}, nil)

			a, err := Analyze(strings.NewReader(body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.HasCodeBlock != tt.expected {
				t.Errorf("expected has_code_block=%v, got %v", tt.expected, a.HasCodeBlock)
			}
		})
	}
}

func TestAnalyzeErrorPatternDetection(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"error keyword", "I'm getting an error in my code", true},
		{"exception", "This throws an exception", true},
		{"panic", "The program panic when running", true},
		{"stack trace", "Here is the stack trace of the crash", true},
		{"failed", "The build failed with status 1", true},
		{"fatal", "fatal error: goroutine running", true},
		{"nil pointer", "nil pointer dereference error", true},
		{"syntax error", "syntax error unexpected token", true},
		{"no error", "This is a normal request about design", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := makeRequest("gpt-4", []Message{
				makeMessage("user", tt.content),
			}, nil)

			a, err := Analyze(strings.NewReader(body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.HasErrorPattern != tt.expected {
				t.Errorf("expected has_error_pattern=%v, got %v", tt.expected, a.HasErrorPattern)
			}
		})
	}
}

func TestAnalyzeKeywordDetection(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{"debug", "help me debug this issue", []string{"debugging"}},
		{"fix", "fix this bug in the code", []string{"debugging"}},
		{"refactor", "please refactor this function", []string{"refactoring"}},
		{"test", "write unit tests for this", []string{"testing"}},
		{"docs", "add documentation for this function", []string{"documentation"}},
		{"readme", "update the README file", []string{"documentation"}},
		{"architecture", "plan the architecture for this system", []string{"architecture"}},
		{"design", "design a new layout", []string{"design"}},
		{"ui", "improve the ui layout", []string{"design"}},
		{"backend", "improve the server endpoints", []string{"backend"}},
		{"frontend", "improve the client view rendering", []string{"frontend"}},
		{"security", "review for security vulnerabilities", []string{"security"}},
		{"no keywords", "hello world general chat", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := makeRequest("gpt-4", []Message{
				makeMessage("user", tt.content),
			}, nil)

			a, err := Analyze(strings.NewReader(body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(a.DetectedKeywords) != len(tt.expected) {
				t.Errorf("expected keywords %v, got %v", tt.expected, a.DetectedKeywords)
				return
			}
			for i, kw := range tt.expected {
				if a.DetectedKeywords[i] != kw {
					t.Errorf("expected keyword %s at index %d, got %s", kw, i, a.DetectedKeywords[i])
				}
			}
		})
	}
}

func TestAnalyzeDomainSignals(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{"vue", "build a Vue component", []string{"vue"}},
		{"react", "create a React hook", []string{"react"}},
		{"laravel", "add a Laravel route", []string{"laravel"}},
		{"express", "create an Express middleware", []string{"express"}},
		{"gin", "set up a Gin handler", []string{"gin"}},
		{"next", "build a Next.js page", []string{"nextjs"}},
		{"django", "create a Django view", []string{"django"}},
		{"no domain", "general coding question", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := makeRequest("gpt-4", []Message{
				makeMessage("user", tt.content),
			}, nil)

			a, err := Analyze(strings.NewReader(body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(a.DomainSignals) != len(tt.expected) {
				t.Errorf("expected domain signals %v, got %v", tt.expected, a.DomainSignals)
				return
			}
			for i, d := range tt.expected {
				if a.DomainSignals[i] != d {
					t.Errorf("expected domain %s at index %d, got %s", d, i, a.DomainSignals[i])
				}
			}
		})
	}
}

func TestAnalyzeFileExtensions(t *testing.T) {
	content := "Check main.go and index.ts and App.vue files"
	body := makeRequest("gpt-4", []Message{
		makeMessage("user", content),
	}, nil)

	a, err := Analyze(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]bool{"main.go": true, "index.ts": true, "app.vue": true}
	if len(a.FileExtensions) != len(expected) {
		t.Errorf("expected %d extensions, got %d: %v", len(expected), len(a.FileExtensions), a.FileExtensions)
	}
	for _, ext := range a.FileExtensions {
		if !expected[ext] {
			t.Errorf("unexpected extension %s", ext)
		}
	}
}

func TestAnalyzeLongContextDetection(t *testing.T) {
	shortContent := "Hello"
	longContent := strings.Repeat("This is a sentence for testing long context detection. ", 800)

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"short content", shortContent, false},
		{"long content", longContent, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := makeRequest("gpt-4", []Message{
				makeMessage("user", tt.content),
			}, nil)

			a, err := Analyze(strings.NewReader(body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.LongContextSignal != tt.expected {
				t.Errorf("expected long_context=%v, got %v", tt.expected, a.LongContextSignal)
			}
		})
	}
}

func TestAnalyzeToolDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		tools    json.RawMessage
		expected bool
	}{
		{"no tools", nil, false},
		{"empty tools", json.RawMessage(`[]`), false},
		{"has tools", json.RawMessage(`[{"type":"function","function":{"name":"get_weather"}}]`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := ChatRequest{
				Model:    "gpt-4",
				Messages: []Message{makeMessage("user", "hi")},
				Tools:    tt.tools,
			}
			data, _ := json.Marshal(req)

			a, err := Analyze(strings.NewReader(string(data)))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.HasToolDefinitions != tt.expected {
				t.Errorf("expected has_tool_definitions=%v, got %v", tt.expected, a.HasToolDefinitions)
			}
		})
	}
}

func TestAnalyzePromptLengthBucket(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedBucket string
	}{
		{"short", "hi", "short"},
		{"medium", strings.Repeat("word ", 200), "medium"},
		{"long", strings.Repeat("word ", 2000), "long"},
		{"very_long", strings.Repeat("word ", 8000), "very_long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := makeRequest("gpt-4", []Message{
				makeMessage("user", tt.content),
			}, nil)

			a, err := Analyze(strings.NewReader(body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.PromptLengthBucket != tt.expectedBucket {
				t.Errorf("expected bucket %s, got %s", tt.expectedBucket, a.PromptLengthBucket)
			}
		})
	}
}

func TestAnalyzeComplexity(t *testing.T) {
	tests := []struct {
		name     string
		req      ChatRequest
		expected string
	}{
		{
			name: "low complexity",
			req: ChatRequest{
				Model:    "gpt-4",
				Messages: []Message{makeMessage("user", "hi")},
			},
			expected: "low",
		},
		{
			name: "medium complexity",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					makeMessage("user", strings.Repeat("This is a test sentence for complexity analysis. ", 500)),
					makeMessage("assistant", "I understand the request."),
				},
			},
			expected: "medium",
		},
		{
			name: "high complexity",
			req: ChatRequest{
				Model:    "gpt-4",
				Messages: make([]Message, 15),
				Tools:    json.RawMessage(`[{"type":"function"}]`),
			},
			expected: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "high complexity" {
				for i := range tt.req.Messages {
					tt.req.Messages[i] = makeMessage("user", "test message with some content for analysis and complex patterns")
				}
			}
			data, _ := json.Marshal(tt.req)

			a, err := Analyze(strings.NewReader(string(data)))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.ComplexitySignal != tt.expected {
				t.Errorf("expected complexity %s, got %s (score indicators: msgs=%d, len=%d, tools=%v, code=%v, exts=%d)",
					tt.expected, a.ComplexitySignal, a.MessageCount, a.ApproxPromptLength, a.HasToolDefinitions, a.HasCodeBlock, len(a.FileExtensions))
			}
		})
	}
}

func TestAnalyzeEmptyBody(t *testing.T) {
	a, err := Analyze(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.MessageCount != 0 {
		t.Errorf("expected 0 messages, got %d", a.MessageCount)
	}
	if a.RequestedModel != "" {
		t.Errorf("expected empty model, got %s", a.RequestedModel)
	}
}

func TestAnalyzeMalformedJSON(t *testing.T) {
	a, err := Analyze(strings.NewReader("{invalid json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil analysis for malformed JSON")
	}
}

func TestAnalyzeMultipleMessages(t *testing.T) {
	body := makeRequest("gpt-4", []Message{
		makeMessage("system", "You are a helpful assistant"),
		makeMessage("user", "Hello"),
		makeMessage("assistant", "Hi there!"),
		makeMessage("user", "How are you?"),
	}, nil)

	a, err := Analyze(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.MessageCount != 4 {
		t.Errorf("expected 4 messages, got %d", a.MessageCount)
	}
}

func TestAnalyzeContentArray(t *testing.T) {
	content := []interface{}{
		map[string]interface{}{"type": "text", "text": "Check this code"},
		map[string]interface{}{"type": "text", "text": "func main() {}"},
	}
	data, _ := json.Marshal(ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: content},
		},
	})

	a, err := Analyze(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.HasCodeBlock {
		t.Error("expected has_code_block true for array content with code")
	}
}

func TestAnalyzeNoPromptInOutput(t *testing.T) {
	secretContent := "my-secret-api-key-12345 and password=abc"
	body := makeRequest("gpt-4", []Message{
		makeMessage("user", secretContent),
	}, nil)

	a, err := Analyze(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, _ := json.Marshal(a)
	if strings.Contains(string(result), "secret-api-key") {
		t.Error("analyzer output must not contain prompt content")
	}
	if strings.Contains(string(result), "password=abc") {
		t.Error("analyzer output must not contain prompt content")
	}
}
