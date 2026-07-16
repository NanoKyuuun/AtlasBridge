package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/atlasbridge/atlasbridge/internal/security"
	"gopkg.in/yaml.v3"
)

const (
	DefaultPort           = 20127
	DefaultDownstreamPort = 20128
	DefaultHost           = "127.0.0.1"
	DefaultDownstreamURL  = "http://127.0.0.1:20128/v1"
)

type Config struct {
	App        AppConfig        `yaml:"app" json:"app"`
	Server     ServerConfig     `yaml:"server" json:"server"`
	Downstream DownstreamConfig `yaml:"downstream" json:"downstream"`
	Security   SecurityConfig   `yaml:"security" json:"security"`
	Startup    StartupConfig    `yaml:"startup" json:"startup"`
	Routing    RoutingConfig    `yaml:"routing" json:"routing"`
	Logging    LoggingConfig    `yaml:"logging" json:"logging"`
}

type AppConfig struct {
	Name              string `yaml:"name" json:"name"`
	Mode              string `yaml:"mode" json:"mode"`
	FirstRunCompleted bool   `yaml:"first_run_completed" json:"first_run_completed"`
}

type ServerConfig struct {
	Host      string `yaml:"host" json:"host"`
	Port      int    `yaml:"port" json:"port"`
	AdminPath string `yaml:"admin_path" json:"admin_path"`
}

type DownstreamConfig struct {
	BaseURL        string `yaml:"base_url" json:"base_url"`
	TimeoutSeconds int    `yaml:"timeout_seconds" json:"timeout_seconds"`
}

type SecurityConfig struct {
	AdminAuthEnabled   bool   `yaml:"admin_auth_enabled" json:"admin_auth_enabled"`
	AdminTokenHash     string `yaml:"admin_token_hash,omitempty" json:"admin_token_hash,omitempty"`
	AdminPasswordHash  string `yaml:"admin_password_hash,omitempty" json:"admin_password_hash,omitempty"`
	BindLocalhostOnly  bool   `yaml:"bind_localhost_only" json:"bind_localhost_only"`
	AllowLANAccess     bool   `yaml:"allow_lan_access" json:"allow_lan_access"`
	SessionExpiresAt   int64  `yaml:"session_expires_at,omitempty" json:"session_expires_at,omitempty"` // Unix timestamp; 0 = no expiry
}

type StartupConfig struct {
	RunAtLogin            bool `yaml:"run_at_login" json:"run_at_login"`
	StartProxyOnAppLaunch bool `yaml:"start_proxy_on_app_launch" json:"start_proxy_on_app_launch"`
	RestartAfterCrash     bool `yaml:"restart_after_crash" json:"restart_after_crash"`
}

type RoutingConfig struct {
	AutoRouting         bool    `yaml:"auto_routing" json:"auto_routing"`
	DefaultRoute        string  `yaml:"default_route" json:"default_route"`
	LowConfidenceRoute  string  `yaml:"low_confidence_route" json:"low_confidence_route"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold" json:"confidence_threshold"`
	SmartFastRoute      string  `yaml:"smart_fast_route" json:"smart_fast_route"`
	MetadataTransport   string  `yaml:"metadata_transport" json:"metadata_transport"`
}

type LoggingConfig struct {
	Level                  string `yaml:"level" json:"level"`
	PrivacyMode            string `yaml:"privacy_mode" json:"privacy_mode"`
	PromptLoggingEnabled   bool   `yaml:"prompt_logging_enabled" json:"prompt_logging_enabled"`
	MetadataLoggingEnabled bool   `yaml:"metadata_logging_enabled" json:"metadata_logging_enabled"`
	RetentionDays          int    `yaml:"retention_days" json:"retention_days"`
}

func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:              "AtlasBridge AI Proxy",
			Mode:              "manual",
			FirstRunCompleted: false,
		},
		Server: ServerConfig{
			Host:      DefaultHost,
			Port:      DefaultPort,
			AdminPath: "/admin",
		},
		Downstream: DownstreamConfig{
			BaseURL:        DefaultDownstreamURL,
			TimeoutSeconds: 120,
		},
		Security: SecurityConfig{
			AdminAuthEnabled:  true,
			BindLocalhostOnly: true,
			AllowLANAccess:    false,
		},
		Startup: StartupConfig{
			RunAtLogin:            false,
			StartProxyOnAppLaunch: true,
			RestartAfterCrash:     true,
		},
		Routing: RoutingConfig{
			AutoRouting:         true,
			DefaultRoute:        "route.default",
			LowConfidenceRoute:  "route.default",
			ConfidenceThreshold: 0.55,
			SmartFastRoute:      "route.low_cost",
			MetadataTransport:   "model_alias",
		},
		Logging: LoggingConfig{
			Level:                  "info",
			PrivacyMode:            "standard",
			PromptLoggingEnabled:   false,
			MetadataLoggingEnabled: true,
			RetentionDays:          7,
		},
	}
}

func ConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "AtlasBridge")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "AtlasBridge")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "AtlasBridge")
	}
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func RoutesPath() string {
	return filepath.Join(ConfigDir(), "routes.yaml")
}

func ProfilesPath() string {
	return filepath.Join(ConfigDir(), "profiles.yaml")
}

func TokenFilePath() string {
	return filepath.Join(ConfigDir(), ".token")
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(ConfigDir(), 0o700); err != nil {
				return nil, fmt.Errorf("create config dir: %w", err)
			}
			if cfg.Security.AdminAuthEnabled && cfg.Security.AdminTokenHash == "" {
				rawToken, err := security.EnsureToken(&cfg.Security.AdminTokenHash)
				if err == nil && rawToken != "" {
					tokenPath := TokenFilePath()
					_ = os.WriteFile(tokenPath, []byte(rawToken), 0o600)
				}
			}
			if err := Save(cfg); err != nil {
				return nil, fmt.Errorf("save default config: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Security.AdminAuthEnabled && cfg.Security.AdminTokenHash == "" {
		rawToken, err := security.EnsureToken(&cfg.Security.AdminTokenHash)
		if err == nil && rawToken != "" {
			_ = Save(cfg)
			tokenPath := TokenFilePath()
			_ = os.WriteFile(tokenPath, []byte(rawToken), 0o600)
		}
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	EnforceNetworkInvariants(cfg)

	return cfg, nil
}

func Save(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return saveWithBackup(ConfigPath(), data, 0o600)
}

// SaveAtomic writes pre-serialized config data to disk atomically.
// The caller is responsible for marshaling the data before calling this.
func SaveAtomic(data []byte) error {
	return saveWithBackup(ConfigPath(), data, 0o600)
}

func Validate(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", cfg.Server.Port)
	}
	if cfg.Server.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}
	if cfg.Downstream.BaseURL == "" {
		return fmt.Errorf("downstream URL cannot be empty")
	}
	if err := validateDownstreamURL(cfg.Downstream.BaseURL); err != nil {
		return fmt.Errorf("invalid downstream URL: %w", err)
	}
	if cfg.Downstream.TimeoutSeconds < 0 {
		return fmt.Errorf("invalid timeout: %d (must be >= 0)", cfg.Downstream.TimeoutSeconds)
	}
	validModes := map[string]bool{"always_on": true, "manual": true, "disabled": true}
	if !validModes[cfg.App.Mode] {
		return fmt.Errorf("invalid app mode: %s (must be always_on, manual, or disabled)", cfg.App.Mode)
	}
	validPrivacy := map[string]bool{"standard": true, "strict": true, "debug": true}
	if !validPrivacy[cfg.Logging.PrivacyMode] {
		return fmt.Errorf("invalid privacy mode: %s (must be standard, strict, or debug)", cfg.Logging.PrivacyMode)
	}
	validMetadataTransport := map[string]bool{"model_alias": true, "header": true}
	if cfg.Routing.MetadataTransport != "" && !validMetadataTransport[cfg.Routing.MetadataTransport] {
		return fmt.Errorf("invalid metadata_transport: %s (must be model_alias or header)", cfg.Routing.MetadataTransport)
	}
	if cfg.Security.AllowLANAccess && !cfg.Security.AdminAuthEnabled {
		return fmt.Errorf("admin auth must be enabled when LAN access is allowed")
	}
	// Auth must have at least one credential: token hash or password hash
	if cfg.Security.AdminAuthEnabled && cfg.Security.AdminTokenHash == "" && cfg.Security.AdminPasswordHash == "" {
		return fmt.Errorf("admin token hash or password hash must be set when admin auth is enabled")
	}
	return nil
}

// EnforceNetworkInvariants applies one-way security invariants to the config.
// If LAN access is not allowed, the server host is forced to loopback
// regardless of what Host says. BindLocalhostOnly is also enforced as a
// belt-and-suspenders check.
func EnforceNetworkInvariants(cfg *Config) {
	if !cfg.Security.AllowLANAccess || cfg.Security.BindLocalhostOnly {
		cfg.Server.Host = "127.0.0.1"
	}
}

func validateDownstreamURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot parse URL: %v", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if parsed.User != nil {
		return fmt.Errorf("credentials in URL are not allowed")
	}
	if parsed.Fragment != "" {
		return fmt.Errorf("URL fragments are not allowed")
	}
	if len(parsed.RawQuery) > 0 {
		return fmt.Errorf("URL query strings are not allowed for downstream")
	}
	return nil
}

func ValidateFull(cfg *Config, routes *RoutesConfig, profiles *ProfilesConfig) error {
	if err := Validate(cfg); err != nil {
		return err
	}
	if profiles == nil {
		return fmt.Errorf("profiles config cannot be nil")
	}
	if routes == nil {
		return fmt.Errorf("routes config cannot be nil")
	}
	profileNames := make(map[string]bool)
	for name, profile := range profiles.RouteProfiles {
		profileNames[name] = true
		if profile.DownstreamAlias == "" {
			return fmt.Errorf("profile %q has empty downstream_alias", name)
		}
		validPriorities := map[string]bool{"quality": true, "balanced": true, "cost": true}
		if !validPriorities[profile.Priority] {
			return fmt.Errorf("profile %q has invalid priority %q (must be quality, balanced, or cost)", name, profile.Priority)
		}
	}
	if cfg.Routing.DefaultRoute != "" {
		if !profileNames[cfg.Routing.DefaultRoute] {
			return fmt.Errorf("default_route %q references non-existent profile", cfg.Routing.DefaultRoute)
		}
		defaultProfile := profiles.RouteProfiles[cfg.Routing.DefaultRoute]
		if !defaultProfile.Enabled {
			return fmt.Errorf("default_route %q references disabled profile", cfg.Routing.DefaultRoute)
		}
	}
	if cfg.Routing.LowConfidenceRoute != "" {
		if !profileNames[cfg.Routing.LowConfidenceRoute] {
			return fmt.Errorf("low_confidence_route %q references non-existent profile", cfg.Routing.LowConfidenceRoute)
		}
		lowProfile := profiles.RouteProfiles[cfg.Routing.LowConfidenceRoute]
		if !lowProfile.Enabled {
			return fmt.Errorf("low_confidence_route %q references disabled profile", cfg.Routing.LowConfidenceRoute)
		}
	}
	for taskName, routeRef := range routes.TaskRoutes {
		if !profileNames[routeRef] {
			return fmt.Errorf("task_route %q references non-existent profile %q", taskName, routeRef)
		}
	}
	return nil
}

func LoadFull() (*Config, *RoutesConfig, *ProfilesConfig, error) {
	cfg, err := Load()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load config: %w", err)
	}
	routes, err := LoadRoutes()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load routes: %w", err)
	}
	profiles, err := LoadProfiles()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load profiles: %w", err)
	}
	if err := ValidateFull(cfg, routes, profiles); err != nil {
		return nil, nil, nil, fmt.Errorf("validate config: %w", err)
	}
	EnforceNetworkInvariants(cfg)
	return cfg, routes, profiles, nil
}
