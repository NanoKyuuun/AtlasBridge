package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 20127 {
		t.Errorf("expected port 20127, got %d", cfg.Server.Port)
	}
	if cfg.Downstream.BaseURL != "http://127.0.0.1:20128/v1" {
		t.Errorf("expected downstream URL http://127.0.0.1:20128/v1, got %s", cfg.Downstream.BaseURL)
	}
	if cfg.Security.AdminAuthEnabled != true {
		t.Error("expected admin_auth_enabled to be true by default")
	}
	if cfg.Security.BindLocalhostOnly != true {
		t.Error("expected bind_localhost_only to be true")
	}
	if cfg.Security.AllowLANAccess != false {
		t.Error("expected allow_lan_access to be false")
	}
	if cfg.Logging.PromptLoggingEnabled != false {
		t.Error("expected prompt_logging_enabled to be false")
	}
	if cfg.Logging.MetadataLoggingEnabled != true {
		t.Error("expected metadata_logging_enabled to be true")
	}
	if cfg.Routing.AutoRouting != true {
		t.Error("expected auto_routing to be true")
	}
	if cfg.Routing.DefaultRoute != "route.default" {
		t.Errorf("expected default_route route.default, got %s", cfg.Routing.DefaultRoute)
	}
	if cfg.Routing.LowConfidenceRoute != "route.default" {
		t.Errorf("expected low_confidence_route route.default, got %s", cfg.Routing.LowConfidenceRoute)
	}
	if cfg.App.Mode != "manual" {
		t.Errorf("expected app mode manual, got %s", cfg.App.Mode)
	}
}

func TestDefaultRoutesConfig(t *testing.T) {
	routes := DefaultRoutesConfig()
	if len(routes.TaskRoutes) == 0 {
		t.Error("expected non-empty task routes")
	}
	if routes.TaskRoutes["unknown"] != "route.default" {
		t.Errorf("expected unknown to map to route.default, got %s", routes.TaskRoutes["unknown"])
	}
	if routes.TaskRoutes["general_chat"] != "route.default" {
		t.Errorf("expected general_chat to map to route.default, got %s", routes.TaskRoutes["general_chat"])
	}
}

func TestDefaultProfilesConfig(t *testing.T) {
	profiles := DefaultProfilesConfig()
	if len(profiles.RouteProfiles) == 0 {
		t.Error("expected non-empty route profiles")
	}
	defaultProfile, ok := profiles.RouteProfiles["route.default"]
	if !ok {
		t.Fatal("expected route.default profile to exist")
	}
	if defaultProfile.DownstreamAlias != "combo.default" {
		t.Errorf("expected downstream_alias combo.default, got %s", defaultProfile.DownstreamAlias)
	}
	if defaultProfile.Priority != "balanced" {
		t.Errorf("expected priority balanced, got %s", defaultProfile.Priority)
	}
	if !defaultProfile.Enabled {
		t.Error("expected route.default to be enabled")
	}
}

func TestValidateValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := Validate(cfg); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestValidateInvalidPort(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.Port = 0
	if err := Validate(cfg); err == nil {
		t.Error("expected error for port 0")
	}

	cfg.Server.Port = 99999
	if err := Validate(cfg); err == nil {
		t.Error("expected error for port 99999")
	}

	cfg.Server.Port = -1
	if err := Validate(cfg); err == nil {
		t.Error("expected error for port -1")
	}
}

func TestValidateEmptyHost(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.Host = ""
	if err := Validate(cfg); err == nil {
		t.Error("expected error for empty host")
	}
}

func TestValidateEmptyDownstreamURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Downstream.BaseURL = ""
	if err := Validate(cfg); err == nil {
		t.Error("expected error for empty downstream URL")
	}
}

func TestValidateInvalidDownstreamURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"ftp scheme", "ftp://127.0.0.1:20128/v1"},
		{"no scheme", "127.0.0.1:20128/v1"},
		{"empty host", "http:///v1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Downstream.BaseURL = tt.url
			if err := Validate(cfg); err == nil {
				t.Errorf("expected error for URL %q", tt.url)
			}
		})
	}
}

func TestValidateInvalidAppMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.App.Mode = "invalid_mode"
	if err := Validate(cfg); err == nil {
		t.Error("expected error for invalid app mode")
	}
}

func TestValidateInvalidPrivacyMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Logging.PrivacyMode = "invalid_privacy"
	if err := Validate(cfg); err == nil {
		t.Error("expected error for invalid privacy mode")
	}
}

func TestValidateFullValid(t *testing.T) {
	cfg := DefaultConfig()
	routes := DefaultRoutesConfig()
	profiles := DefaultProfilesConfig()
	if err := ValidateFull(cfg, routes, profiles); err != nil {
		t.Errorf("expected valid full config, got error: %v", err)
	}
}

func TestValidateFullNilProfiles(t *testing.T) {
	cfg := DefaultConfig()
	routes := DefaultRoutesConfig()
	if err := ValidateFull(cfg, routes, nil); err == nil {
		t.Error("expected error for nil profiles")
	}
}

func TestValidateFullNilRoutes(t *testing.T) {
	cfg := DefaultConfig()
	profiles := DefaultProfilesConfig()
	if err := ValidateFull(cfg, nil, profiles); err == nil {
		t.Error("expected error for nil routes")
	}
}

func TestValidateFullMissingDefaultRoute(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Routing.DefaultRoute = "route.nonexistent"
	routes := DefaultRoutesConfig()
	profiles := DefaultProfilesConfig()
	if err := ValidateFull(cfg, routes, profiles); err == nil {
		t.Error("expected error for missing default route")
	}
}

func TestValidateFullMissingLowConfidenceRoute(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Routing.LowConfidenceRoute = "route.nonexistent"
	routes := DefaultRoutesConfig()
	profiles := DefaultProfilesConfig()
	if err := ValidateFull(cfg, routes, profiles); err == nil {
		t.Error("expected error for missing low confidence route")
	}
}

func TestValidateFullDisabledDefaultRoute(t *testing.T) {
	cfg := DefaultConfig()
	profiles := DefaultProfilesConfig()
	disabled := profiles.RouteProfiles["route.default"]
	disabled.Enabled = false
	profiles.RouteProfiles["route.default"] = disabled
	routes := DefaultRoutesConfig()
	if err := ValidateFull(cfg, routes, profiles); err == nil {
		t.Error("expected error for disabled default route")
	}
}

func TestValidateFullDisabledLowConfidenceRoute(t *testing.T) {
	cfg := DefaultConfig()
	profiles := DefaultProfilesConfig()
	disabled := profiles.RouteProfiles["route.default"]
	disabled.Enabled = false
	profiles.RouteProfiles["route.default"] = disabled
	routes := DefaultRoutesConfig()
	if err := ValidateFull(cfg, routes, profiles); err == nil {
		t.Error("expected error for disabled low confidence route")
	}
}

func TestValidateFullTaskRouteMissingProfile(t *testing.T) {
	cfg := DefaultConfig()
	routes := DefaultRoutesConfig()
	routes.TaskRoutes["custom_task"] = "route.nonexistent"
	profiles := DefaultProfilesConfig()
	if err := ValidateFull(cfg, routes, profiles); err == nil {
		t.Error("expected error for task route referencing missing profile")
	}
}

func TestValidateFullEmptyDownstreamAlias(t *testing.T) {
	cfg := DefaultConfig()
	routes := DefaultRoutesConfig()
	profiles := DefaultProfilesConfig()
	bad := profiles.RouteProfiles["route.default"]
	bad.DownstreamAlias = ""
	profiles.RouteProfiles["route.default"] = bad
	if err := ValidateFull(cfg, routes, profiles); err == nil {
		t.Error("expected error for empty downstream_alias")
	}
}

func TestValidateFullInvalidProfilePriority(t *testing.T) {
	cfg := DefaultConfig()
	routes := DefaultRoutesConfig()
	profiles := DefaultProfilesConfig()
	bad := profiles.RouteProfiles["route.default"]
	bad.Priority = "invalid_priority"
	profiles.RouteProfiles["route.default"] = bad
	if err := ValidateFull(cfg, routes, profiles); err == nil {
		t.Error("expected error for invalid profile priority")
	}
}

func TestLoadWithTempDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := os.Getenv("APPDATA")
	t.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", origDir)

	cfgDir := filepath.Join(tmpDir, "AtlasBridge")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected to load config, got error: %v", err)
	}
	if cfg.Server.Port != 20127 {
		t.Errorf("expected port 20127, got %d", cfg.Server.Port)
	}

	if _, err := os.Stat(ConfigPath()); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}

func TestLoadRoutesWithTempDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := os.Getenv("APPDATA")
	t.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", origDir)

	routes, err := LoadRoutes()
	if err != nil {
		t.Fatalf("expected to load routes, got error: %v", err)
	}
	if len(routes.TaskRoutes) == 0 {
		t.Error("expected non-empty task routes")
	}
}

func TestLoadProfilesWithTempDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := os.Getenv("APPDATA")
	t.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", origDir)

	profiles, err := LoadProfiles()
	if err != nil {
		t.Fatalf("expected to load profiles, got error: %v", err)
	}
	if len(profiles.RouteProfiles) == 0 {
		t.Error("expected non-empty route profiles")
	}
}

func TestLoadFullWithTempDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := os.Getenv("APPDATA")
	t.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", origDir)

	cfg, routes, profiles, err := LoadFull()
	if err != nil {
		t.Fatalf("expected to load full config, got error: %v", err)
	}
	if cfg.Server.Port != 20127 {
		t.Errorf("expected port 20127, got %d", cfg.Server.Port)
	}
	if len(routes.TaskRoutes) == 0 {
		t.Error("expected non-empty task routes")
	}
	if len(profiles.RouteProfiles) == 0 {
		t.Error("expected non-empty route profiles")
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("expected non-empty config dir")
	}
}

func TestValidateLANModeRequiresAuth(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Security.AllowLANAccess = true
	cfg.Security.AdminAuthEnabled = false
	if err := Validate(cfg); err == nil {
		t.Error("expected error when LAN access enabled without admin auth")
	}
}

func TestValidateLANModeWithAuthOK(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Security.AllowLANAccess = true
	cfg.Security.AdminAuthEnabled = true
	if err := Validate(cfg); err != nil {
		t.Errorf("expected valid config with LAN + auth, got error: %v", err)
	}
}

func TestEnforceNetworkInvariantsForcesLoopback(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.Host = "0.0.0.0"
	cfg.Security.AllowLANAccess = false

	EnforceNetworkInvariants(cfg)

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host forced to 127.0.0.1, got %s", cfg.Server.Host)
	}
}

func TestEnforceNetworkInvariantsPreservesHostWhenLANAllowed(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.Host = "0.0.0.0"
	cfg.Security.AllowLANAccess = true

	EnforceNetworkInvariants(cfg)

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host preserved as 0.0.0.0, got %s", cfg.Server.Host)
	}
}

func TestAtomicWriteCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yaml")
	data := []byte("key: value\n")

	if err := atomicWriteFile(path, data, 0o600); err != nil {
		t.Fatalf("atomicWriteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("content mismatch: got %q, want %q", got, data)
	}
}

func TestAtomicWriteStrictPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	if err := atomicWriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("atomicWriteFile failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	if runtime.GOOS != "windows" {
		perm := info.Mode().Perm()
		if perm != 0o600 {
			t.Errorf("expected file permission 0600, got %o", perm)
		}
	} else {
		if info.Mode().Perm() == 0o666 {
			t.Skip("Windows does not enforce Unix file permissions via os.Chmod")
		}
	}
}

func TestAtomicWriteDirPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "config.yaml")

	if err := atomicWriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("atomicWriteFile failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(tmpDir, "subdir"))
	if err != nil {
		t.Fatalf("stat dir failed: %v", err)
	}

	if runtime.GOOS != "windows" {
		perm := info.Mode().Perm()
		if perm != 0o700 {
			t.Errorf("expected dir permission 0700, got %o", perm)
		}
	} else {
		if info.Mode().Perm() == 0o777 {
			t.Skip("Windows does not enforce Unix file permissions via os.Chmod")
		}
	}
}

func TestSaveWithBackupCreatesBackup(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := saveWithBackup(path, []byte("updated"), 0o600); err != nil {
		t.Fatalf("saveWithBackup failed: %v", err)
	}

	backupPath := backupFilePath(path)
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
	if string(backup) != "original" {
		t.Errorf("backup content mismatch: got %q, want %q", backup, "original")
	}

	main, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("main file read failed: %v", err)
	}
	if string(main) != "updated" {
		t.Errorf("main content mismatch: got %q, want %q", main, "updated")
	}
}

func TestConfigSaveAndReloadAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := os.Getenv("APPDATA")
	t.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", origDir)

	cfg := DefaultConfig()
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load after Save failed: %v", err)
	}
	if reloaded.Server.Port != cfg.Server.Port {
		t.Errorf("port mismatch: got %d, want %d", reloaded.Server.Port, cfg.Server.Port)
	}

	backup := backupFilePath(ConfigPath())
	if _, err := os.Stat(backup); err == nil {
		t.Log("backup file exists as expected on first save (may be from default creation)")
	}
}
