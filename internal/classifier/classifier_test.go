package classifier

import (
	"testing"

	"github.com/smart-ai-proxy/smart-ai-proxy/internal/analyzer"
)

func makeAnalysis(keywords []string, domains []string, opts ...func(*analyzer.Analysis)) *analyzer.Analysis {
	a := &analyzer.Analysis{
		DetectedKeywords: keywords,
		DomainSignals:    domains,
		ComplexitySignal: "low",
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func withErrorPattern(a *analyzer.Analysis)     { a.HasErrorPattern = true }
func withCodeBlock(a *analyzer.Analysis)        { a.HasCodeBlock = true }
func withLongContext(a *analyzer.Analysis)      { a.LongContextSignal = true }
func withTools(a *analyzer.Analysis)            { a.HasToolDefinitions = true }
func withHighComplexity(a *analyzer.Analysis)   { a.ComplexitySignal = "high" }
func withMediumComplexity(a *analyzer.Analysis) { a.ComplexitySignal = "medium" }
func withShort(a *analyzer.Analysis)            { a.PromptLengthBucket = "short" }
func withFewMessages(a *analyzer.Analysis)      { a.MessageCount = 1 }

func TestClassifyDebuggingErrorPattern(t *testing.T) {
	a := makeAnalysis(nil, nil, withErrorPattern)
	c := Classify(a)
	if c.TaskType != "debugging" {
		t.Errorf("expected debugging, got %s", c.TaskType)
	}
	if c.Confidence < 0.8 {
		t.Errorf("expected confidence >= 0.8, got %f", c.Confidence)
	}
	if c.ClassificationStatus != "success" {
		t.Errorf("expected success, got %s", c.ClassificationStatus)
	}
}

func TestClassifyDebuggingKeyword(t *testing.T) {
	a := makeAnalysis([]string{"debugging"}, nil)
	c := Classify(a)
	if c.TaskType != "debugging" {
		t.Errorf("expected debugging, got %s", c.TaskType)
	}
}

func TestClassifySecurityReview(t *testing.T) {
	a := makeAnalysis([]string{"security"}, nil)
	c := Classify(a)
	if c.TaskType != "security_review" {
		t.Errorf("expected security_review, got %s", c.TaskType)
	}
	if c.Confidence < 0.8 {
		t.Errorf("expected confidence >= 0.8, got %f", c.Confidence)
	}
}

func TestClassifyDocumentation(t *testing.T) {
	a := makeAnalysis([]string{"documentation"}, nil)
	c := Classify(a)
	if c.TaskType != "documentation" {
		t.Errorf("expected documentation, got %s", c.TaskType)
	}
}

func TestClassifyArchitectureDesign(t *testing.T) {
	a := makeAnalysis([]string{"architecture"}, nil)
	c := Classify(a)
	if c.TaskType != "architecture_design" {
		t.Errorf("expected architecture_design, got %s", c.TaskType)
	}
}

func TestClassifyTestGeneration(t *testing.T) {
	a := makeAnalysis([]string{"testing"}, nil)
	c := Classify(a)
	if c.TaskType != "test_generation" {
		t.Errorf("expected test_generation, got %s", c.TaskType)
	}
}

func TestClassifyRefactoring(t *testing.T) {
	a := makeAnalysis([]string{"refactoring"}, nil)
	c := Classify(a)
	if c.TaskType != "refactoring" {
		t.Errorf("expected refactoring, got %s", c.TaskType)
	}
}

func TestClassifyDesignTask(t *testing.T) {
	a := makeAnalysis([]string{"design"}, nil)
	c := Classify(a)
	if c.TaskType != "design_task" {
		t.Errorf("expected design_task, got %s", c.TaskType)
	}
}

func TestClassifyFullstackEngineering(t *testing.T) {
	a := makeAnalysis([]string{"backend", "frontend"}, nil)
	c := Classify(a)
	if c.TaskType != "fullstack_engineering" {
		t.Errorf("expected fullstack_engineering, got %s", c.TaskType)
	}
}

func TestClassifyBackendEngineering(t *testing.T) {
	a := makeAnalysis([]string{"backend"}, nil)
	c := Classify(a)
	if c.TaskType != "backend_engineering" {
		t.Errorf("expected backend_engineering, got %s", c.TaskType)
	}
}

func TestClassifyFrontendEngineering(t *testing.T) {
	a := makeAnalysis([]string{"frontend"}, nil)
	c := Classify(a)
	if c.TaskType != "frontend_engineering" {
		t.Errorf("expected frontend_engineering, got %s", c.TaskType)
	}
}

func TestClassifyLongContextAnalysis(t *testing.T) {
	a := makeAnalysis(nil, nil, withLongContext)
	c := Classify(a)
	if c.TaskType != "long_context_analysis" {
		t.Errorf("expected long_context_analysis, got %s", c.TaskType)
	}
}

func TestClassifyLightweightTask(t *testing.T) {
	a := makeAnalysis(nil, nil, withShort, withFewMessages)
	c := Classify(a)
	if c.TaskType != "lightweight_task" {
		t.Errorf("expected lightweight_task, got %s", c.TaskType)
	}
}

func TestClassifyGeneralChat(t *testing.T) {
	a := makeAnalysis(nil, nil, withFewMessages)
	c := Classify(a)
	if c.TaskType != "general_chat" {
		t.Errorf("expected general_chat, got %s", c.TaskType)
	}
}

func TestClassifyUnknown(t *testing.T) {
	a := &analyzer.Analysis{
		ComplexitySignal: "low",
		MessageCount:     10,
		HasCodeBlock:     true,
	}
	c := Classify(a)
	if c.TaskType != "unknown" {
		t.Errorf("expected unknown, got %s", c.TaskType)
	}
	if c.ClassificationStatus != "success" {
		t.Errorf("expected success, got %s", c.ClassificationStatus)
	}
}

func TestClassifyNilAnalysis(t *testing.T) {
	c := Classify(nil)
	if c.TaskType != "unknown" {
		t.Errorf("expected unknown, got %s", c.TaskType)
	}
	if c.ClassificationStatus != "failed" {
		t.Errorf("expected failed, got %s", c.ClassificationStatus)
	}
	if c.Confidence != 0.0 {
		t.Errorf("expected confidence 0.0, got %f", c.Confidence)
	}
}

func TestClassifyRulePriorityErrorOverDebug(t *testing.T) {
	a := makeAnalysis([]string{"debugging"}, nil, withErrorPattern)
	c := Classify(a)
	if c.TaskType != "debugging" {
		t.Errorf("expected debugging (error pattern wins), got %s", c.TaskType)
	}
}

func TestClassifyConfidenceBoostWithCodeBlock(t *testing.T) {
	a := &analyzer.Analysis{
		DetectedKeywords: []string{"backend"},
		HasCodeBlock:     true,
		ComplexitySignal: "medium",
	}
	c := Classify(a)
	if c.TaskType != "backend_engineering" {
		t.Errorf("expected backend_engineering, got %s", c.TaskType)
	}
	if c.Confidence < 0.85 {
		t.Errorf("expected boosted confidence >= 0.85, got %f", c.Confidence)
	}
}

func TestClassifyConfidenceBoostWithTools(t *testing.T) {
	a := &analyzer.Analysis{
		DetectedKeywords:   []string{"backend"},
		HasToolDefinitions: true,
		ComplexitySignal:   "medium",
	}
	c := Classify(a)
	if c.Confidence < 0.85 {
		t.Errorf("expected boosted confidence >= 0.85, got %f", c.Confidence)
	}
}

func TestClassifyConfidenceCap(t *testing.T) {
	a := makeAnalysis([]string{"backend"}, []string{"vue"}, withCodeBlock, withTools, withHighComplexity)
	c := Classify(a)
	if c.Confidence > 0.95 {
		t.Errorf("expected confidence capped at 0.95, got %f", c.Confidence)
	}
}

func TestClassifyConfidenceFloor(t *testing.T) {
	a := &analyzer.Analysis{
		ComplexitySignal: "low",
	}
	c := Classify(a)
	if c.Confidence < 0.1 {
		t.Errorf("expected confidence floor 0.1, got %f", c.Confidence)
	}
}

func TestClassifyRoutingReason(t *testing.T) {
	a := makeAnalysis([]string{"security"}, nil)
	c := Classify(a)
	if c.RoutingReason == "" {
		t.Error("expected non-empty routing reason")
	}
}

func TestClassifyComplexityPassedThrough(t *testing.T) {
	a := makeAnalysis([]string{"backend"}, nil, withHighComplexity)
	c := Classify(a)
	if c.Complexity != "high" {
		t.Errorf("expected complexity high, got %s", c.Complexity)
	}
}

func TestClassifyAmbiguousPrompt(t *testing.T) {
	a := makeAnalysis([]string{"backend"}, nil, withMediumComplexity)
	c := Classify(a)
	if c.TaskType != "backend_engineering" {
		t.Errorf("expected backend_engineering, got %s", c.TaskType)
	}
	if c.Confidence < 0.5 {
		t.Errorf("expected reasonable confidence >= 0.5, got %f", c.Confidence)
	}
}

func TestClassifyNoPromptPersisted(t *testing.T) {
	a := makeAnalysis([]string{"backend"}, nil)
	c := Classify(a)
	if c.RoutingReason == "" {
		t.Error("routing reason should be non-empty")
	}
}

func TestClassifyAllTaskCategories(t *testing.T) {
	tests := []struct {
		name     string
		keywords []string
		domains  []string
		opts     []func(*analyzer.Analysis)
		expected string
	}{
		{"debugging_error", nil, nil, []func(*analyzer.Analysis){withErrorPattern}, "debugging"},
		{"debugging_keyword", []string{"debugging"}, nil, nil, "debugging"},
		{"security_review", []string{"security"}, nil, nil, "security_review"},
		{"documentation", []string{"documentation"}, nil, nil, "documentation"},
		{"architecture_design", []string{"architecture"}, nil, nil, "architecture_design"},
		{"test_generation", []string{"testing"}, nil, nil, "test_generation"},
		{"refactoring", []string{"refactoring"}, nil, nil, "refactoring"},
		{"design_task", []string{"design"}, nil, nil, "design_task"},
		{"fullstack", []string{"backend", "frontend"}, nil, nil, "fullstack_engineering"},
		{"backend", []string{"backend"}, nil, nil, "backend_engineering"},
		{"frontend", []string{"frontend"}, nil, nil, "frontend_engineering"},
		{"long_context", nil, nil, []func(*analyzer.Analysis){withLongContext}, "long_context_analysis"},
		{"lightweight", nil, nil, []func(*analyzer.Analysis){withShort, withFewMessages}, "lightweight_task"},
		{"general_chat", nil, nil, []func(*analyzer.Analysis){withFewMessages}, "general_chat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := makeAnalysis(tt.keywords, tt.domains, tt.opts...)
			c := Classify(a)
			if c.TaskType != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, c.TaskType)
			}
			if c.ClassificationStatus != "success" {
				t.Errorf("expected success, got %s", c.ClassificationStatus)
			}
		})
	}
}
