package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/forwarder"
	"github.com/atlasbridge/atlasbridge/internal/security"
)

// ConfigUpdateResult is returned by config mutation methods to communicate
// side effects back to the HTTP handler.
type ConfigUpdateResult struct {
	RestartRequired bool   `json:"restart_required"`
	AdminToken      string `json:"admin_token,omitempty"`
	Message         string `json:"message"`
}

// ConfigService is the single component allowed to merge, validate, persist,
// and swap runtime state. Handlers must never mutate config/routes/profiles
// directly; they call ConfigService methods instead.
type ConfigService struct {
	store *StateStore
}

// NewConfigService creates a ConfigService backed by the given store.
func NewConfigService(store *StateStore) *ConfigService {
	return &ConfigService{store: store}
}

// ApplyConfigPatch merges a partial config patch into the current config,
// validates, persists, and atomically swaps the snapshot.
func (cs *ConfigService) ApplyConfigPatch(patch map[string]json.RawMessage) (*ConfigUpdateResult, error) {
	cs.store.persistMu.Lock()
	defer cs.store.persistMu.Unlock()

	current := cs.store.Load()
	merged := current.Config

	if raw, ok := patch["app"]; ok {
		if err := json.Unmarshal(raw, &merged.App); err != nil {
			return nil, fmt.Errorf("invalid app config: %w", err)
		}
	}
	if raw, ok := patch["server"]; ok {
		if err := json.Unmarshal(raw, &merged.Server); err != nil {
			return nil, fmt.Errorf("invalid server config: %w", err)
		}
	}
	if raw, ok := patch["downstream"]; ok {
		if err := json.Unmarshal(raw, &merged.Downstream); err != nil {
			return nil, fmt.Errorf("invalid downstream config: %w", err)
		}
	}
	if raw, ok := patch["security"]; ok {
		var secUpdate SecurityUpdate
		if err := json.Unmarshal(raw, &secUpdate); err != nil {
			return nil, fmt.Errorf("invalid security config: %w", err)
		}
		if secUpdate.AdminAuthEnabled != nil {
			merged.Security.AdminAuthEnabled = *secUpdate.AdminAuthEnabled
		}
		if secUpdate.BindLocalhostOnly != nil {
			merged.Security.BindLocalhostOnly = *secUpdate.BindLocalhostOnly
		}
		if secUpdate.AllowLANAccess != nil {
			merged.Security.AllowLANAccess = *secUpdate.AllowLANAccess
		}
	}
	if raw, ok := patch["startup"]; ok {
		if err := json.Unmarshal(raw, &merged.Startup); err != nil {
			return nil, fmt.Errorf("invalid startup config: %w", err)
		}
	}
	if raw, ok := patch["routing"]; ok {
		if err := json.Unmarshal(raw, &merged.Routing); err != nil {
			return nil, fmt.Errorf("invalid routing config: %w", err)
		}
	}
	if raw, ok := patch["logging"]; ok {
		if err := json.Unmarshal(raw, &merged.Logging); err != nil {
			return nil, fmt.Errorf("invalid logging config: %w", err)
		}
	}

	var adminToken string
	if merged.Security.AdminAuthEnabled {
		var err error
		adminToken, err = security.EnsureToken(&merged.Security.AdminTokenHash)
		if err != nil {
			return nil, fmt.Errorf("generate token: %w", err)
		}
	}

	if err := config.Validate(&merged); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	restartRequired := merged.Server.Host != current.Config.Server.Host ||
		merged.Server.Port != current.Config.Server.Port

	fwd := current.Forwarder
	if merged.Downstream.BaseURL != current.Config.Downstream.BaseURL ||
		merged.Downstream.TimeoutSeconds != current.Config.Downstream.TimeoutSeconds {
		var err error
		fwd, err = forwarder.New(merged.Downstream.BaseURL, merged.Downstream.TimeoutSeconds)
		if err != nil {
			return nil, fmt.Errorf("create forwarder: %w", err)
		}
	}

	if err := config.Save(&merged); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	next := current.Clone()
	next.Config = merged
	next.Forwarder = fwd
	next.Version = current.Version + 1
	next.CreatedAt = time.Now()
	cs.store.Swap(next)

	return &ConfigUpdateResult{
		RestartRequired: restartRequired,
		AdminToken:      adminToken,
		Message:         "config updated",
	}, nil
}

// ApplyRoutes validates and replaces the routing table atomically.
func (cs *ConfigService) ApplyRoutes(body []byte) error {
	cs.store.persistMu.Lock()
	defer cs.store.persistMu.Unlock()

	current := cs.store.Load()

	var updated config.RoutesConfig
	if err := json.Unmarshal(body, &updated); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	cfgCopy := current.Config
	if err := config.ValidateFull(&cfgCopy, &updated, &current.Profiles); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := config.SaveRoutes(&updated); err != nil {
		return fmt.Errorf("save routes: %w", err)
	}

	next := current.Clone()
	next.Routes = copyRoutes(updated)
	next.Version = current.Version + 1
	next.CreatedAt = time.Now()
	cs.store.Swap(next)

	return nil
}

// ApplyProfiles validates and replaces the profiles table atomically.
func (cs *ConfigService) ApplyProfiles(body []byte) error {
	cs.store.persistMu.Lock()
	defer cs.store.persistMu.Unlock()

	current := cs.store.Load()

	var updated config.ProfilesConfig
	if err := json.Unmarshal(body, &updated); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	cfgCopy := current.Config
	if err := config.ValidateFull(&cfgCopy, &current.Routes, &updated); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := config.SaveProfiles(&updated); err != nil {
		return fmt.Errorf("save profiles: %w", err)
	}

	next := current.Clone()
	next.Profiles = copyProfiles(updated)
	next.Version = current.Version + 1
	next.CreatedAt = time.Now()
	cs.store.Swap(next)

	return nil
}

// RotateToken generates a new admin token, persists the hash, and returns
// the raw token. The raw token is shown to the user exactly once.
func (cs *ConfigService) RotateToken() (string, error) {
	cs.store.persistMu.Lock()
	defer cs.store.persistMu.Unlock()

	newToken, newHash, err := security.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	current := cs.store.Load()
	merged := current.Config
	merged.Security.AdminTokenHash = newHash

	if err := config.Save(&merged); err != nil {
		return "", fmt.Errorf("save config: %w", err)
	}

	next := current.Clone()
	next.Config = merged
	next.Version = current.Version + 1
	next.CreatedAt = time.Now()
	cs.store.Swap(next)

	return newToken, nil
}

// UpdateMode sets the app mode (always_on, manual, disabled), persists,
// and swaps the snapshot.
func (cs *ConfigService) UpdateMode(mode string) error {
	cs.store.persistMu.Lock()
	defer cs.store.persistMu.Unlock()

	current := cs.store.Load()
	merged := current.Config
	merged.App.Mode = mode

	if err := config.Save(&merged); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	next := current.Clone()
	next.Config = merged
	next.Version = current.Version + 1
	next.CreatedAt = time.Now()
	cs.store.Swap(next)

	return nil
}

// ApplyStartup persists startup config changes atomically.
func (cs *ConfigService) ApplyStartup(startupCfg config.StartupConfig) error {
	cs.store.persistMu.Lock()
	defer cs.store.persistMu.Unlock()

	current := cs.store.Load()
	merged := current.Config
	merged.Startup = startupCfg

	if err := config.Save(&merged); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	next := current.Clone()
	next.Config = merged
	next.Version = current.Version + 1
	next.CreatedAt = time.Now()
	cs.store.Swap(next)

	return nil
}

// Reset restores all config, routes, and profiles to defaults atomically.
// All three files are persisted before the snapshot is swapped.
// The current admin token hash is preserved to prevent lockout.
func (cs *ConfigService) Reset() error {
	cs.store.persistMu.Lock()
	defer cs.store.persistMu.Unlock()

	current := cs.store.Load()

	cfg := config.DefaultConfig()
	cfg.Security.AdminTokenHash = current.Config.Security.AdminTokenHash
	routes := config.DefaultRoutesConfig()
	profiles := config.DefaultProfilesConfig()

	fwd, err := forwarder.New(cfg.Downstream.BaseURL, cfg.Downstream.TimeoutSeconds)
	if err != nil {
		return fmt.Errorf("create forwarder: %w", err)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	if err := config.SaveRoutes(routes); err != nil {
		return fmt.Errorf("save routes: %w", err)
	}
	if err := config.SaveProfiles(profiles); err != nil {
		return fmt.Errorf("save profiles: %w", err)
	}

	next := &Snapshot{
		Config:    *cfg,
		Routes:    copyRoutes(*routes),
		Profiles:  copyProfiles(*profiles),
		Forwarder: fwd,
		Version:   1,
		CreatedAt: time.Now(),
	}
	cs.store.Swap(next)

	return nil
}

// Import performs an all-or-nothing import of config, routes, and profiles.
// Each piece is only replaced if provided in the import bundle. The
// in-memory snapshot is only swapped after all pieces are validated and
// persisted successfully.
func (cs *ConfigService) Import(body []byte) error {
	cs.store.persistMu.Lock()
	defer cs.store.persistMu.Unlock()

	current := cs.store.Load()

	var imported struct {
		Config   *config.Config         `json:"config"`
		Routes   *config.RoutesConfig   `json:"routes"`
		Profiles *config.ProfilesConfig `json:"profiles"`
	}
	if err := json.Unmarshal(body, &imported); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	cfg := current.Config
	routes := current.Routes
	profiles := current.Profiles

	if imported.Config != nil {
		cfg = *imported.Config
		if cfg.Security.AdminTokenHash == "" {
			cfg.Security.AdminTokenHash = current.Config.Security.AdminTokenHash
		}
	}
	if imported.Routes != nil {
		routes = *imported.Routes
	}
	if imported.Profiles != nil {
		profiles = *imported.Profiles
	}

	if err := config.Validate(&cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	if err := config.ValidateFull(&cfg, &routes, &profiles); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fwd := current.Forwarder
	if imported.Config != nil && (cfg.Downstream.BaseURL != current.Config.Downstream.BaseURL ||
		cfg.Downstream.TimeoutSeconds != current.Config.Downstream.TimeoutSeconds) {
		newFwd, fwdErr := forwarder.New(cfg.Downstream.BaseURL, cfg.Downstream.TimeoutSeconds)
		if fwdErr != nil {
			return fmt.Errorf("create forwarder: %w", fwdErr)
		}
		fwd = newFwd
	}

	if err := config.Save(&cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	if err := config.SaveRoutes(&routes); err != nil {
		return fmt.Errorf("save routes: %w", err)
	}
	if err := config.SaveProfiles(&profiles); err != nil {
		return fmt.Errorf("save profiles: %w", err)
	}

	next := &Snapshot{
		Config:    cfg,
		Routes:    copyRoutes(routes),
		Profiles:  copyProfiles(profiles),
		Forwarder: fwd,
		Version:   current.Version + 1,
		CreatedAt: time.Now(),
	}
	cs.store.Swap(next)

	return nil
}
