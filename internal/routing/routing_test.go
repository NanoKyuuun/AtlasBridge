package routing

import (
	"testing"

	"github.com/atlasbridge/atlasbridge/internal/analyzer"
	"github.com/atlasbridge/atlasbridge/internal/classifier"
	"github.com/atlasbridge/atlasbridge/internal/config"
)

func defaultRoutingCfg() *config.RoutingConfig {
	return &config.RoutingConfig{
		AutoRouting:         true,
		DefaultRoute:        "route.default",
		LowConfidenceRoute:  "route.default",
		ConfidenceThreshold: 0.55,
	}
}

func TestResolveSmartAuto(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	analysis := &analyzer.Analysis{ComplexitySignal: "medium", DetectedKeywords: []string{"backend"}}
	classification := classifier.Classify(analysis)

	decision := Resolve("smart-auto", analysis, classification, routes, profiles, routingCfg)

	if decision.OverrideSource != "smart_alias" {
		t.Errorf("expected override source smart_alias, got %s", decision.OverrideSource)
	}
	if decision.DownstreamAlias == "" {
		t.Error("expected non-empty downstream alias")
	}
	if decision.RouteKey == "" {
		t.Error("expected non-empty route key")
	}
}

func TestResolveSmartDebug(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("smart-debug", nil, nil, routes, profiles, routingCfg)

	if decision.TaskType != "debugging" {
		t.Errorf("expected task type debugging, got %s", decision.TaskType)
	}
	if decision.RouteKey != "route.debugging" {
		t.Errorf("expected route route.debugging, got %s", decision.RouteKey)
	}
	if decision.OverrideSource != "smart_alias" {
		t.Errorf("expected override source smart_alias, got %s", decision.OverrideSource)
	}
}

func TestResolveSmartCheap(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("smart-cheap", nil, nil, routes, profiles, routingCfg)

	if decision.TaskType != "lightweight_task" {
		t.Errorf("expected task type lightweight_task, got %s", decision.TaskType)
	}
	if decision.RouteKey != "route.low_cost" {
		t.Errorf("expected route route.low_cost, got %s", decision.RouteKey)
	}
}

func TestResolveSmartDocs(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("smart-docs", nil, nil, routes, profiles, routingCfg)

	if decision.TaskType != "documentation" {
		t.Errorf("expected task type documentation, got %s", decision.TaskType)
	}
	if decision.RouteKey != "route.documentation" {
		t.Errorf("expected route route.documentation, got %s", decision.RouteKey)
	}
}

func TestResolveSmartArchitect(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("smart-architect", nil, nil, routes, profiles, routingCfg)

	if decision.TaskType != "architecture_design" {
		t.Errorf("expected task type architecture_design, got %s", decision.TaskType)
	}
	if decision.RouteKey != "route.architect" {
		t.Errorf("expected route route.architect, got %s", decision.RouteKey)
	}
}

func TestResolveSmartCode(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("smart-code", nil, nil, routes, profiles, routingCfg)

	if decision.TaskType != "backend_engineering" {
		t.Errorf("expected task type backend_engineering, got %s", decision.TaskType)
	}
	if decision.RouteKey != "route.backend" {
		t.Errorf("expected route route.backend, got %s", decision.RouteKey)
	}
}

func TestResolveAutoClassificationHighConfidence(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	analysis := &analyzer.Analysis{ComplexitySignal: "medium", DetectedKeywords: []string{"backend"}}
	classification := classifier.Classify(analysis)

	decision := Resolve("gpt-4", analysis, classification, routes, profiles, routingCfg)

	if decision.OverrideSource != "auto_classification" {
		t.Errorf("expected override source auto_classification, got %s", decision.OverrideSource)
	}
	if classification.Confidence >= routingCfg.ConfidenceThreshold {
		if decision.RouteKey != "route.backend" {
			t.Errorf("expected route route.backend, got %s", decision.RouteKey)
		}
	}
}

func TestResolveAutoClassificationLowConfidence(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	analysis := &analyzer.Analysis{
		ComplexitySignal:   "low",
		MessageCount:       1,
		PromptLengthBucket: "short",
	}
	classification := classifier.Classify(analysis)

	decision := Resolve("", analysis, classification, routes, profiles, routingCfg)

	if decision.OverrideSource != "auto_classification" {
		t.Errorf("expected override source auto_classification, got %s", decision.OverrideSource)
	}
	if classification.Confidence < routingCfg.ConfidenceThreshold {
		if decision.RouteKey != routingCfg.LowConfidenceRoute {
			t.Errorf("expected low confidence route %s, got %s", routingCfg.LowConfidenceRoute, decision.RouteKey)
		}
	}
}

func TestResolveManualOverride(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	analysis := &analyzer.Analysis{ComplexitySignal: "medium"}
	classification := classifier.Classify(analysis)

	decision := Resolve("custom-model-x", analysis, classification, routes, profiles, routingCfg)

	if decision.OverrideSource != "manual_model" {
		t.Errorf("expected override source manual_model, got %s", decision.OverrideSource)
	}
}

func TestResolveDisabledAutoRouting(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()
	routingCfg.AutoRouting = false

	decision := Resolve("smart-debug", nil, nil, routes, profiles, routingCfg)

	if decision.OverrideSource != "default" {
		t.Errorf("expected override source default, got %s", decision.OverrideSource)
	}
	if decision.RouteKey != "route.default" {
		t.Errorf("expected route route.default, got %s", decision.RouteKey)
	}
}

func TestResolveNilClassification(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("gpt-4", nil, nil, routes, profiles, routingCfg)

	if decision.OverrideSource != "manual_model" {
		t.Errorf("expected override source manual_model, got %s", decision.OverrideSource)
	}
}

func TestResolveDownstreamAliasAlwaysPopulated(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	cases := []struct {
		name  string
		model string
	}{
		{"smart-auto", "smart-auto"},
		{"smart-debug", "smart-debug"},
		{"gpt-4", "gpt-4"},
		{"empty model", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decision := Resolve(tc.model, nil, nil, routes, profiles, routingCfg)
			if decision.DownstreamAlias == "" {
				t.Errorf("expected non-empty downstream alias for model %q", tc.model)
			}
		})
	}
}

func TestResolveRoutingReasonAlwaysPopulated(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("gpt-4", nil, nil, routes, profiles, routingCfg)

	if decision.RoutingReason == "" {
		t.Error("expected non-empty routing reason")
	}
}

func TestResolveComplexityPassedThrough(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	analysis := &analyzer.Analysis{ComplexitySignal: "high", DetectedKeywords: []string{"backend"}}
	classification := classifier.Classify(analysis)

	decision := Resolve("gpt-4", analysis, classification, routes, profiles, routingCfg)

	if decision.Complexity != "high" {
		t.Errorf("expected complexity high, got %s", decision.Complexity)
	}
}

func TestResolveSmartAliasWithClassification(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	analysis := &analyzer.Analysis{ComplexitySignal: "medium", DetectedKeywords: []string{"frontend"}}
	classification := classifier.Classify(analysis)

	decision := Resolve("smart-auto", analysis, classification, routes, profiles, routingCfg)

	if decision.TaskType != "frontend_engineering" {
		t.Errorf("expected task type frontend_engineering from classification, got %s", decision.TaskType)
	}
}

func TestResolveConfidenceFromClassification(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	analysis := &analyzer.Analysis{ComplexitySignal: "medium", DetectedKeywords: []string{"backend"}}
	classification := classifier.Classify(analysis)

	decision := Resolve("smart-auto", analysis, classification, routes, profiles, routingCfg)

	if decision.Confidence != classification.Confidence {
		t.Errorf("expected confidence %f from classification, got %f", classification.Confidence, decision.Confidence)
	}
}

func TestDetectSmartAlias(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
		alias    SmartAlias
	}{
		{"smart-auto", true, AliasSmartAuto},
		{"smart-debug", true, AliasSmartDebug},
		{"smart-cheap", true, AliasSmartCheap},
		{"smart-docs", true, AliasSmartDocs},
		{"smart-architect", true, AliasSmartArchitect},
		{"smart-code", true, AliasSmartCode},
		{"SMART-DEBUG", true, AliasSmartDebug},
		{"smart-debug ", true, AliasSmartDebug},
		{"gpt-4", false, ""},
		{"", false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			alias, ok := detectSmartAlias(tc.input)
			if ok != tc.expected {
				t.Errorf("expected ok=%v for input %q, got %v", tc.expected, tc.input, ok)
			}
			if ok && alias != tc.alias {
				t.Errorf("expected alias %v, got %v", tc.alias, alias)
			}
		})
	}
}

func TestIsManualModelOverride(t *testing.T) {
	classification := &classifier.Classification{TaskType: "backend_engineering", Confidence: 0.8}

	cases := []struct {
		input    string
		expected bool
	}{
		{"custom-model", true},
		{"gpt-4-turbo", false},
		{"gpt-4", false},
		{"gpt-3.5-turbo", false},
		{"claude-3-opus", false},
		{"claude-3-sonnet", false},
		{"claude-3.5-sonnet", false},
		{"gemini-pro", false},
		{"llama-3.1-70b", false},
		{"mistral-large", false},
		{"deepseek-coder", false},
		{"qwen-72b", false},
		{"random-model", true},
		{"", false},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			result := isManualModelOverride(tc.input, classification)
			if result != tc.expected {
				t.Errorf("expected %v for input %q, got %v", tc.expected, tc.input, result)
			}
		})
	}
}

func TestSmartAliasDescriptions(t *testing.T) {
	aliases := []SmartAlias{
		AliasSmartAuto,
		AliasSmartDebug,
		AliasSmartCheap,
		AliasSmartDocs,
		AliasSmartArchitect,
		AliasSmartCode,
		AliasSmartFast,
		AliasSmartLongContext,
	}

	for _, alias := range aliases {
		if _, ok := smartAliasDescriptions[alias]; !ok {
			t.Errorf("missing description for alias %s", alias)
		}
	}
}

func TestResolveSmartFast(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("smart-fast", nil, nil, routes, profiles, routingCfg)

	if decision.TaskType != "lightweight_task" {
		t.Errorf("expected task type lightweight_task, got %s", decision.TaskType)
	}
	if decision.RouteKey != "route.low_cost" {
		t.Errorf("expected route route.low_cost, got %s", decision.RouteKey)
	}
	if decision.DownstreamAlias != "combo.low_cost" {
		t.Errorf("expected downstream alias combo.low_cost, got %s", decision.DownstreamAlias)
	}
	if decision.OverrideSource != "smart_alias" {
		t.Errorf("expected override source smart_alias, got %s", decision.OverrideSource)
	}
}

func TestResolveSmartFastCustomRoute(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()
	routingCfg.SmartFastRoute = "route.default"

	decision := Resolve("smart-fast", nil, nil, routes, profiles, routingCfg)

	if decision.RouteKey != "route.default" {
		t.Errorf("expected route route.default from custom config, got %s", decision.RouteKey)
	}
	if decision.DownstreamAlias != "combo.default" {
		t.Errorf("expected downstream alias combo.default, got %s", decision.DownstreamAlias)
	}
}

func TestResolveSmartLongContext(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("smart-long-context", nil, nil, routes, profiles, routingCfg)

	if decision.TaskType != "long_context_analysis" {
		t.Errorf("expected task type long_context_analysis, got %s", decision.TaskType)
	}
	if decision.RouteKey != "route.long_context" {
		t.Errorf("expected route route.long_context, got %s", decision.RouteKey)
	}
	if decision.DownstreamAlias != "combo.long_context" {
		t.Errorf("expected downstream alias combo.long_context, got %s", decision.DownstreamAlias)
	}
	if decision.OverrideSource != "smart_alias" {
		t.Errorf("expected override source smart_alias, got %s", decision.OverrideSource)
	}
}

func TestPolicyHook(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	PolicyHooks = []PolicyHook{
		func(model string, d *RoutingDecision) *RoutingDecision {
			if model == "intercept-model" {
				return &RoutingDecision{
					TaskType:             "hooked_task",
					Confidence:           1.0,
					RouteKey:             "route.architect",
					DownstreamAlias:      "combo.deep_reasoning",
					RoutingReason:        "policy hook override",
					OverrideSource:       "policy_hook",
					Complexity:           "high",
					ClassificationStatus: "success",
				}
			}
			return nil
		},
	}
	defer func() { PolicyHooks = nil }()

	decision := Resolve("intercept-model", nil, nil, routes, profiles, routingCfg)

	if decision.OverrideSource != "policy_hook" {
		t.Errorf("expected override source policy_hook, got %s", decision.OverrideSource)
	}
	if decision.TaskType != "hooked_task" {
		t.Errorf("expected task type hooked_task, got %s", decision.TaskType)
	}
	if decision.RouteKey != "route.architect" {
		t.Errorf("expected route route.architect, got %s", decision.RouteKey)
	}
}

func TestPolicyHookNotAppliedWhenNil(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	PolicyHooks = []PolicyHook{
		func(model string, d *RoutingDecision) *RoutingDecision {
			return nil
		},
	}
	defer func() { PolicyHooks = nil }()

	decision := Resolve("smart-debug", nil, nil, routes, profiles, routingCfg)

	if decision.OverrideSource != "smart_alias" {
		t.Errorf("expected override source smart_alias, got %s", decision.OverrideSource)
	}
	if decision.RouteKey != "route.debugging" {
		t.Errorf("expected route route.debugging, got %s", decision.RouteKey)
	}
}

// --- Safe Passthrough Tests ---

func TestSafePassthroughNilClassification(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("gpt-4", nil, nil, routes, profiles, routingCfg)

	if decision.ClassificationStatus != "fallback" {
		t.Errorf("expected classification_status fallback, got %s", decision.ClassificationStatus)
	}
	if decision.RouteKey != "route.default" {
		t.Errorf("expected route route.default, got %s", decision.RouteKey)
	}
	if decision.DownstreamAlias == "" {
		t.Error("expected non-empty downstream alias")
	}
}

func TestSafePassthroughNilAnalysis(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	classification := classifier.Classify(nil)
	decision := Resolve("smart-auto", nil, classification, routes, profiles, routingCfg)

	if decision.ClassificationStatus != "fallback" {
		t.Errorf("expected classification_status fallback, got %s", decision.ClassificationStatus)
	}
	if decision.TaskType != "unknown" {
		t.Errorf("expected task type unknown, got %s", decision.TaskType)
	}
}

func TestSafePassthroughInvalidRoute(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	// Inject an invalid route mapping
	routes.TaskRoutes["debugging"] = "route.nonexistent"

	analysis := &analyzer.Analysis{ComplexitySignal: "medium", DetectedKeywords: []string{"debugging"}}
	classification := classifier.Classify(analysis)

	decision := Resolve("gpt-4", analysis, classification, routes, profiles, routingCfg)

	if decision.RouteKey != "route.default" {
		t.Errorf("expected fallback to route.default for invalid route, got %s", decision.RouteKey)
	}
	if decision.DownstreamAlias == "" {
		t.Error("expected non-empty downstream alias after fallback")
	}
}

func TestSafePassthroughDisabledRoute(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	// Disable the debugging route
	p := profiles.RouteProfiles["route.debugging"]
	p.Enabled = false
	profiles.RouteProfiles["route.debugging"] = p

	analysis := &analyzer.Analysis{ComplexitySignal: "medium", DetectedKeywords: []string{"debugging"}}
	classification := classifier.Classify(analysis)

	decision := Resolve("gpt-4", analysis, classification, routes, profiles, routingCfg)

	if decision.RouteKey != "route.default" {
		t.Errorf("expected fallback to route.default for disabled route, got %s", decision.RouteKey)
	}
}

func TestSafePassthroughAutoRoutingDisabled(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()
	routingCfg.AutoRouting = false

	decision := Resolve("gpt-4", nil, nil, routes, profiles, routingCfg)

	if decision.ClassificationStatus != "fallback" {
		t.Errorf("expected classification_status fallback, got %s", decision.ClassificationStatus)
	}
	if decision.RouteKey != "route.default" {
		t.Errorf("expected route route.default, got %s", decision.RouteKey)
	}
}

func TestSafePassthroughSmartAliasNilClassification(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("smart-debug", nil, nil, routes, profiles, routingCfg)

	if decision.ClassificationStatus != "fallback" {
		t.Errorf("expected classification_status fallback, got %s", decision.ClassificationStatus)
	}
	if decision.RouteKey != "route.debugging" {
		t.Errorf("expected route route.debugging, got %s", decision.RouteKey)
	}
}

func TestSafePassthroughManualOverrideNilClassification(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	decision := Resolve("custom-model-x", nil, nil, routes, profiles, routingCfg)

	if decision.ClassificationStatus != "fallback" {
		t.Errorf("expected classification_status fallback, got %s", decision.ClassificationStatus)
	}
	if decision.RouteKey != "route.default" {
		t.Errorf("expected route route.default, got %s", decision.RouteKey)
	}
}

func TestSafePassthroughRequestStillForwarded(t *testing.T) {
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()
	routingCfg := defaultRoutingCfg()

	// Even with nil analysis/classification, decision has all required fields
	decision := Resolve("gpt-4", nil, nil, routes, profiles, routingCfg)

	if decision.RouteKey == "" {
		t.Error("route_key must be non-empty for safe passthrough")
	}
	if decision.DownstreamAlias == "" {
		t.Error("downstream_alias must be non-empty for safe passthrough")
	}
	if decision.RoutingReason == "" {
		t.Error("routing_reason must be non-empty for safe passthrough")
	}
	if decision.OverrideSource == "" {
		t.Error("override_source must be non-empty for safe passthrough")
	}
}
