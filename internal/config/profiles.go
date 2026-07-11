package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RouteProfile struct {
	Label           string `yaml:"label" json:"label"`
	Description     string `yaml:"description" json:"description"`
	DownstreamAlias string `yaml:"downstream_alias" json:"downstream_alias"`
	Priority        string `yaml:"priority" json:"priority"`
	Enabled         bool   `yaml:"enabled" json:"enabled"`
}

type ProfilesConfig struct {
	RouteProfiles map[string]RouteProfile `yaml:"route_profiles" json:"route_profiles"`
}

func DefaultProfilesConfig() *ProfilesConfig {
	return &ProfilesConfig{
		RouteProfiles: map[string]RouteProfile{
			"route.default": {
				Label:           "Default",
				Description:     "Default fallback route",
				DownstreamAlias: "combo.default",
				Priority:        "balanced",
				Enabled:         true,
			},
			"route.design": {
				Label:           "Design",
				Description:     "UI, UX, design system, layout, visual thinking",
				DownstreamAlias: "combo.design",
				Priority:        "quality",
				Enabled:         true,
			},
			"route.backend": {
				Label:           "Backend",
				Description:     "API, database, service, architecture, server logic",
				DownstreamAlias: "combo.backend",
				Priority:        "balanced",
				Enabled:         true,
			},
			"route.frontend": {
				Label:           "Frontend",
				Description:     "Vue, React, UI component, client-side implementation",
				DownstreamAlias: "combo.frontend",
				Priority:        "balanced",
				Enabled:         true,
			},
			"route.fullstack": {
				Label:           "Fullstack",
				Description:     "Tasks spanning frontend and backend",
				DownstreamAlias: "combo.fullstack",
				Priority:        "balanced",
				Enabled:         true,
			},
			"route.debugging": {
				Label:           "Debugging",
				Description:     "Error analysis, debugging, root cause",
				DownstreamAlias: "combo.debugging",
				Priority:        "quality",
				Enabled:         true,
			},
			"route.refactoring": {
				Label:           "Refactoring",
				Description:     "Code cleanup, restructuring, maintainability",
				DownstreamAlias: "combo.refactor",
				Priority:        "quality",
				Enabled:         true,
			},
			"route.testing": {
				Label:           "Testing",
				Description:     "Unit tests, integration tests, test scenarios",
				DownstreamAlias: "combo.test_generation",
				Priority:        "balanced",
				Enabled:         true,
			},
			"route.documentation": {
				Label:           "Documentation",
				Description:     "README, docstrings, technical documentation",
				DownstreamAlias: "combo.documentation",
				Priority:        "balanced",
				Enabled:         true,
			},
			"route.architect": {
				Label:           "Architecture",
				Description:     "System design, architecture planning",
				DownstreamAlias: "combo.deep_reasoning",
				Priority:        "quality",
				Enabled:         true,
			},
			"route.reasoning": {
				Label:           "Reasoning",
				Description:     "Complex reasoning, analysis",
				DownstreamAlias: "combo.deep_reasoning",
				Priority:        "quality",
				Enabled:         true,
			},
			"route.security": {
				Label:           "Security",
				Description:     "Security review, risk analysis",
				DownstreamAlias: "combo.security_review",
				Priority:        "quality",
				Enabled:         true,
			},
			"route.low_cost": {
				Label:           "Low Cost",
				Description:     "Lightweight tasks, cost-optimized",
				DownstreamAlias: "combo.low_cost",
				Priority:        "cost",
				Enabled:         true,
			},
			"route.long_context": {
				Label:           "Long Context",
				Description:     "Long context analysis, multi-file review",
				DownstreamAlias: "combo.long_context",
				Priority:        "balanced",
				Enabled:         true,
			},
		},
	}
}

func LoadProfiles() (*ProfilesConfig, error) {
	cfg := DefaultProfilesConfig()

	path := ProfilesPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(ConfigDir(), 0o700); err != nil {
				return nil, fmt.Errorf("create config dir: %w", err)
			}
			if err := SaveProfiles(cfg); err != nil {
				return nil, fmt.Errorf("save default profiles: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("read profiles: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse profiles: %w", err)
	}

	return cfg, nil
}

func SaveProfiles(cfg *ProfilesConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}
	return saveWithBackup(ProfilesPath(), data, 0o600)
}
