package classifier

import (
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/analyzer"
)

type Classification struct {
	TaskType             string  `json:"task_type"`
	Confidence           float64 `json:"confidence"`
	Complexity           string  `json:"complexity"`
	RoutingReason        string  `json:"routing_reason"`
	ClassificationStatus string  `json:"classification_status"`
}

type rule struct {
	taskType   string
	confidence float64
	check      func(a *analyzer.Analysis) bool
	reason     string
}

var rules = []rule{
	{
		taskType:   "debugging",
		confidence: 0.9,
		check: func(a *analyzer.Analysis) bool {
			return a.HasErrorPattern
		},
		reason: "error pattern detected (stack trace, exception, panic)",
	},
	{
		taskType:   "debugging",
		confidence: 0.85,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "debugging")
		},
		reason: "debug/fix/error keyword detected",
	},
	{
		taskType:   "security_review",
		confidence: 0.9,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "security")
		},
		reason: "security keyword detected",
	},
	{
		taskType:   "documentation",
		confidence: 0.9,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "documentation")
		},
		reason: "documentation keyword detected (docs, readme, docstring)",
	},
	{
		taskType:   "architecture_design",
		confidence: 0.85,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "architecture")
		},
		reason: "architecture keyword detected",
	},
	{
		taskType:   "test_generation",
		confidence: 0.9,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "testing")
		},
		reason: "testing keyword detected (test, unittest, coverage)",
	},
	{
		taskType:   "refactoring",
		confidence: 0.85,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "refactoring")
		},
		reason: "refactoring keyword detected",
	},
	{
		taskType:   "design_task",
		confidence: 0.85,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "design")
		},
		reason: "design keyword detected (UI, UX, layout, component)",
	},
	{
		taskType:   "fullstack_engineering",
		confidence: 0.8,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "backend") && hasKeyword(a, "frontend")
		},
		reason: "both backend and frontend signals detected",
	},
	{
		taskType:   "backend_engineering",
		confidence: 0.8,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "backend")
		},
		reason: "backend keyword detected (API, database, server)",
	},
	{
		taskType:   "frontend_engineering",
		confidence: 0.8,
		check: func(a *analyzer.Analysis) bool {
			return hasKeyword(a, "frontend")
		},
		reason: "frontend keyword detected (Vue, React, CSS, browser)",
	},
	{
		taskType:   "long_context_analysis",
		confidence: 0.75,
		check: func(a *analyzer.Analysis) bool {
			return a.LongContextSignal
		},
		reason: "very long context detected (>32k chars)",
	},
	{
		taskType:   "lightweight_task",
		confidence: 0.6,
		check: func(a *analyzer.Analysis) bool {
			return a.PromptLengthBucket == "short" && !a.HasCodeBlock && !a.HasToolDefinitions
		},
		reason: "short prompt without code or tools",
	},
	{
		taskType:   "general_chat",
		confidence: 0.5,
		check: func(a *analyzer.Analysis) bool {
			return a.MessageCount <= 2 && !a.HasCodeBlock
		},
		reason: "simple conversation without code",
	},
}

func hasKeyword(a *analyzer.Analysis, category string) bool {
	for _, kw := range a.DetectedKeywords {
		if kw == category {
			return true
		}
	}
	return false
}

func hasDomain(a *analyzer.Analysis, domain string) bool {
	for _, d := range a.DomainSignals {
		if d == domain {
			return true
		}
	}
	return false
}

func Classify(a *analyzer.Analysis) *Classification {
	if a == nil {
		return &Classification{
			TaskType:             "unknown",
			Confidence:           0.0,
			Complexity:           "low",
			RoutingReason:        "nil analysis input",
			ClassificationStatus: "failed",
		}
	}

	for _, r := range rules {
		if r.check(a) {
			adjustedConfidence := adjustConfidence(r.confidence, a)
			return &Classification{
				TaskType:             r.taskType,
				Confidence:           adjustedConfidence,
				Complexity:           a.ComplexitySignal,
				RoutingReason:        r.reason,
				ClassificationStatus: "success",
			}
		}
	}

	return &Classification{
		TaskType:             "unknown",
		Confidence:           0.3,
		Complexity:           a.ComplexitySignal,
		RoutingReason:        "no matching rule",
		ClassificationStatus: "success",
	}
}

func adjustConfidence(base float64, a *analyzer.Analysis) float64 {
	confidence := base

	if a.HasCodeBlock {
		confidence += 0.05
	}
	if a.HasToolDefinitions {
		confidence += 0.05
	}
	if a.ComplexitySignal == "high" {
		confidence += 0.05
	} else if a.ComplexitySignal == "low" {
		confidence -= 0.05
	}

	if len(a.DomainSignals) > 0 {
		confidence += 0.05
	}

	if confidence > 0.95 {
		confidence = 0.95
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}
