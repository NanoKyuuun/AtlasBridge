package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RoutesConfig struct {
	TaskRoutes map[string]string `yaml:"task_routes" json:"task_routes"`
}

func DefaultRoutesConfig() *RoutesConfig {
	return &RoutesConfig{
		TaskRoutes: map[string]string{
			"general_chat":          "route.default",
			"design_task":           "route.design",
			"backend_engineering":   "route.backend",
			"frontend_engineering":  "route.frontend",
			"fullstack_engineering": "route.fullstack",
			"debugging":             "route.debugging",
			"refactoring":           "route.refactoring",
			"test_generation":       "route.testing",
			"documentation":         "route.documentation",
			"architecture_design":   "route.architect",
			"security_review":       "route.security",
			"long_context_analysis": "route.long_context",
			"lightweight_task":      "route.low_cost",
			"unknown":               "route.default",
		},
	}
}

func LoadRoutes() (*RoutesConfig, error) {
	cfg := DefaultRoutesConfig()

	path := RoutesPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
				return nil, fmt.Errorf("create config dir: %w", err)
			}
			if err := SaveRoutes(cfg); err != nil {
				return nil, fmt.Errorf("save default routes: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("read routes: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse routes: %w", err)
	}

	return cfg, nil
}

func SaveRoutes(cfg *RoutesConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal routes: %w", err)
	}
	return os.WriteFile(RoutesPath(), data, 0o644)
}
