package routing

import (
	"fmt"
	"strings"

	"github.com/smart-ai-proxy/smart-ai-proxy/internal/analyzer"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/classifier"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/config"
)

type SmartAlias string

const (
	AliasSmartAuto        SmartAlias = "smart-auto"
	AliasSmartDebug       SmartAlias = "smart-debug"
	AliasSmartCheap       SmartAlias = "smart-cheap"
	AliasSmartDocs        SmartAlias = "smart-docs"
	AliasSmartArchitect   SmartAlias = "smart-architect"
	AliasSmartCode        SmartAlias = "smart-code"
	AliasSmartFast        SmartAlias = "smart-fast"
	AliasSmartLongContext SmartAlias = "smart-long-context"
)

var smartAliasToRoute = map[SmartAlias]string{
	AliasSmartAuto:        "", // empty = auto-route by classification
	AliasSmartDebug:       "route.debugging",
	AliasSmartCheap:       "route.low_cost",
	AliasSmartDocs:        "route.documentation",
	AliasSmartArchitect:   "route.architect",
	AliasSmartCode:        "route.backend", // hybrid default
	AliasSmartFast:        "",              // resolved from config SmartFastRoute
	AliasSmartLongContext: "route.long_context",
}

var smartAliasDescriptions = map[SmartAlias]string{
	AliasSmartAuto:        "Auto-route based on request analysis",
	AliasSmartDebug:       "Force debugging route",
	AliasSmartCheap:       "Force low-cost route",
	AliasSmartDocs:        "Force documentation route",
	AliasSmartArchitect:   "Force architecture route",
	AliasSmartCode:        "Force code/engineering route",
	AliasSmartFast:        "Force fast route (configurable)",
	AliasSmartLongContext: "Force long-context analysis route",
}

// PolicyHook is a callback that can intercept and modify routing decisions.
// Return nil to use the default decision, or return a modified decision.
type PolicyHook func(model string, decision *RoutingDecision) *RoutingDecision

// PolicyHooks holds optional project-specific policy hooks.
var PolicyHooks []PolicyHook

// RegisterPolicyHook adds a project-specific policy hook.
func RegisterPolicyHook(hook PolicyHook) {
	PolicyHooks = append(PolicyHooks, hook)
}

type RoutingDecision struct {
	TaskType             string  `json:"task_type"`
	Confidence           float64 `json:"confidence"`
	RouteKey             string  `json:"route_key"`
	DownstreamAlias      string  `json:"downstream_alias"`
	RoutingReason        string  `json:"routing_reason"`
	OverrideSource       string  `json:"override_source"` // "smart_alias", "manual_model", "auto_classification", "default"
	Complexity           string  `json:"complexity"`
	ClassificationStatus string  `json:"classification_status"` // "success", "failed", "fallback"
}

func Resolve(
	model string,
	analysis *analyzer.Analysis,
	classification *classifier.Classification,
	routes *config.RoutesConfig,
	profiles *config.ProfilesConfig,
	routingCfg *config.RoutingConfig,
) *RoutingDecision {
	if !routingCfg.AutoRouting {
		d := defaultDecision(routingCfg, profiles)
		d.ClassificationStatus = "fallback"
		return d
	}

	isFallback := classification == nil || classification.ClassificationStatus == "failed"

	var d *RoutingDecision

	if alias, ok := detectSmartAlias(model); ok {
		d = resolveSmartAlias(alias, classification, profiles, routingCfg)
	} else if isManualModelOverride(model, classification) {
		d = resolveManualOverride(model, classification, routes, profiles, routingCfg)
	} else {
		d = resolveAutoClassification(classification, routes, profiles, routingCfg)
	}

	if isFallback {
		d.ClassificationStatus = "fallback"
	} else {
		d.ClassificationStatus = classification.ClassificationStatus
	}

	for _, hook := range PolicyHooks {
		if modified := hook(model, d); modified != nil {
			d = modified
		}
	}

	return d
}

func detectSmartAlias(model string) (SmartAlias, bool) {
	lower := strings.ToLower(strings.TrimSpace(model))
	for alias := range smartAliasToRoute {
		if string(alias) == lower {
			return alias, true
		}
	}
	return "", false
}

func isManualModelOverride(model string, classification *classifier.Classification) bool {
	lower := strings.ToLower(strings.TrimSpace(model))
	if lower == "" {
		return false
	}
	if classification == nil {
		return true
	}
	knownPrefixes := []string{"gpt-", "claude-", "gemini-", "llama-", "mistral-", "deepseek-", "qwen-", "phi-", "codestral-"}
	for _, prefix := range knownPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return false
		}
	}
	return true
}

func resolveSmartAlias(
	alias SmartAlias,
	classification *classifier.Classification,
	profiles *config.ProfilesConfig,
	routingCfg *config.RoutingConfig,
) *RoutingDecision {
	routeKey := smartAliasToRoute[alias]

	if alias == AliasSmartFast && routeKey == "" {
		if routingCfg.SmartFastRoute != "" {
			routeKey = routingCfg.SmartFastRoute
		} else {
			routeKey = "route.low_cost"
		}
	}

	if routeKey == "" && classification != nil {
		autoRoutes := config.DefaultRoutesConfig()
		if r, ok := autoRoutes.TaskRoutes[classification.TaskType]; ok {
			routeKey = r
		}
	}
	if routeKey == "" {
		routeKey = "route.default"
	}

	profile, exists := profiles.RouteProfiles[routeKey]
	if !exists || !profile.Enabled {
		profile = profiles.RouteProfiles["route.default"]
		routeKey = "route.default"
	}

	confidence := 0.9
	reason := fmt.Sprintf("smart alias %q forced route %s", alias, routeKey)
	if alias == AliasSmartAuto && classification != nil {
		confidence = classification.Confidence
		reason = classification.RoutingReason
	}

	complexity := "medium"
	if classification != nil {
		complexity = classification.Complexity
	}

	return &RoutingDecision{
		TaskType:        taskTypeForAlias(alias, classification),
		Confidence:      confidence,
		RouteKey:        routeKey,
		DownstreamAlias: profile.DownstreamAlias,
		RoutingReason:   reason,
		OverrideSource:  "smart_alias",
		Complexity:      complexity,
	}
}

func taskTypeForAlias(alias SmartAlias, classification *classifier.Classification) string {
	switch alias {
	case AliasSmartDebug:
		return "debugging"
	case AliasSmartCheap, AliasSmartFast:
		return "lightweight_task"
	case AliasSmartDocs:
		return "documentation"
	case AliasSmartArchitect:
		return "architecture_design"
	case AliasSmartLongContext:
		return "long_context_analysis"
	case AliasSmartCode:
		if classification != nil {
			return classification.TaskType
		}
		return "backend_engineering"
	default:
		if classification != nil {
			return classification.TaskType
		}
		return "general_chat"
	}
}

func resolveManualOverride(
	model string,
	classification *classifier.Classification,
	routes *config.RoutesConfig,
	profiles *config.ProfilesConfig,
	routingCfg *config.RoutingConfig,
) *RoutingDecision {
	taskType := "general_chat"
	confidence := 0.5
	reason := "manual model override"
	complexity := "medium"

	if classification != nil {
		taskType = classification.TaskType
		confidence = classification.Confidence
		reason = fmt.Sprintf("manual model %q override; original: %s (%.2f)", model, classification.TaskType, classification.Confidence)
		complexity = classification.Complexity
	}

	routeKey := routingCfg.DefaultRoute
	if r, ok := routes.TaskRoutes[taskType]; ok {
		routeKey = r
	}

	profile, exists := profiles.RouteProfiles[routeKey]
	if !exists || !profile.Enabled {
		profile = profiles.RouteProfiles["route.default"]
		routeKey = "route.default"
	}

	return &RoutingDecision{
		TaskType:        taskType,
		Confidence:      confidence,
		RouteKey:        routeKey,
		DownstreamAlias: profile.DownstreamAlias,
		RoutingReason:   reason,
		OverrideSource:  "manual_model",
		Complexity:      complexity,
	}
}

func resolveAutoClassification(
	classification *classifier.Classification,
	routes *config.RoutesConfig,
	profiles *config.ProfilesConfig,
	routingCfg *config.RoutingConfig,
) *RoutingDecision {
	if classification == nil {
		return defaultDecision(routingCfg, profiles)
	}

	taskType := classification.TaskType
	confidence := classification.Confidence

	if confidence < routingCfg.ConfidenceThreshold {
		routeKey := routingCfg.LowConfidenceRoute
		profile, exists := profiles.RouteProfiles[routeKey]
		if !exists || !profile.Enabled {
			profile = profiles.RouteProfiles["route.default"]
			routeKey = "route.default"
		}
		return &RoutingDecision{
			TaskType:        taskType,
			Confidence:      confidence,
			RouteKey:        routeKey,
			DownstreamAlias: profile.DownstreamAlias,
			RoutingReason:   fmt.Sprintf("low confidence (%.2f < %.2f threshold); fallback to %s", confidence, routingCfg.ConfidenceThreshold, routeKey),
			OverrideSource:  "auto_classification",
			Complexity:      classification.Complexity,
		}
	}

	routeKey := routingCfg.DefaultRoute
	if r, ok := routes.TaskRoutes[taskType]; ok {
		routeKey = r
	}

	profile, exists := profiles.RouteProfiles[routeKey]
	if !exists || !profile.Enabled {
		profile = profiles.RouteProfiles["route.default"]
		routeKey = "route.default"
	}

	return &RoutingDecision{
		TaskType:        taskType,
		Confidence:      confidence,
		RouteKey:        routeKey,
		DownstreamAlias: profile.DownstreamAlias,
		RoutingReason:   classification.RoutingReason,
		OverrideSource:  "auto_classification",
		Complexity:      classification.Complexity,
	}
}

func defaultDecision(routingCfg *config.RoutingConfig, profiles *config.ProfilesConfig) *RoutingDecision {
	routeKey := routingCfg.DefaultRoute
	profile, exists := profiles.RouteProfiles[routeKey]
	if !exists || !profile.Enabled {
		profile = profiles.RouteProfiles["route.default"]
		routeKey = "route.default"
	}
	return &RoutingDecision{
		TaskType:             "general_chat",
		Confidence:           0.0,
		RouteKey:             routeKey,
		DownstreamAlias:      profile.DownstreamAlias,
		RoutingReason:        "auto-routing disabled or fallback",
		OverrideSource:       "default",
		Complexity:           "medium",
		ClassificationStatus: "fallback",
	}
}
