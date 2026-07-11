package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/analyzer"
	"github.com/atlasbridge/atlasbridge/internal/classifier"
	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/observability"
	"github.com/atlasbridge/atlasbridge/internal/routing"
	runtimemod "github.com/atlasbridge/atlasbridge/internal/runtime"
	"github.com/atlasbridge/atlasbridge/internal/startup"
)

// SecurityView is the DTO returned to clients. It never exposes the token hash.
type SecurityView struct {
	AdminAuthEnabled  bool `json:"admin_auth_enabled"`
	TokenConfigured   bool `json:"token_configured"`
	BindLocalhostOnly bool `json:"bind_localhost_only"`
	AllowLANAccess    bool `json:"allow_lan_access"`
}

// SecurityUpdate is the DTO accepted from clients for partial updates.
// All fields are pointers with omitempty so absent fields are not overwritten.
// admin_token_hash is intentionally excluded — token rotation has a dedicated endpoint.
type SecurityUpdate struct {
	AdminAuthEnabled  *bool `json:"admin_auth_enabled,omitempty"`
	BindLocalhostOnly *bool `json:"bind_localhost_only,omitempty"`
	AllowLANAccess    *bool `json:"allow_lan_access,omitempty"`
}

type AdminDeps struct {
	Store         *StateStore
	ConfigService *ConfigService
	Observability *observability.Logger
	RuntimeState  *runtimemod.State
}

func getConfigHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		secView := SecurityView{
			AdminAuthEnabled:  snap.Config.Security.AdminAuthEnabled,
			TokenConfigured:   snap.Config.Security.AdminTokenHash != "",
			BindLocalhostOnly: snap.Config.Security.BindLocalhostOnly,
			AllowLANAccess:    snap.Config.Security.AllowLANAccess,
		}

		cfgCopy := snap.Config
		cfgCopy.Security = config.SecurityConfig{
			AdminAuthEnabled:  secView.AdminAuthEnabled,
			BindLocalhostOnly: secView.BindLocalhostOnly,
			AllowLANAccess:    secView.AllowLANAccess,
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"app":            cfgCopy.App,
			"server":         cfgCopy.Server,
			"downstream":     cfgCopy.Downstream,
			"security":       secView,
			"startup":        cfgCopy.Startup,
			"routing":        cfgCopy.Routing,
			"logging":        cfgCopy.Logging,
			"config_version": snap.Version,
		})
	}
}

func putConfigHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if handleBodyReadError(w, err) {
			return
		}

		var patch map[string]json.RawMessage
		if err := json.Unmarshal(body, &patch); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		result, err := deps.ConfigService.ApplyConfigPatch(patch)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		if result.AdminToken != "" {
			fmt.Fprint(os.Stdout, "\n")
			fmt.Fprintf(os.Stdout, "  ADMIN TOKEN: %s\n", result.AdminToken)
			fmt.Fprint(os.Stdout, "  This token will NOT be shown again.\n")
			fmt.Fprint(os.Stdout, "  Store it safely before closing this window.\n")
			fmt.Fprint(os.Stdout, "\n")
		}

		resp := map[string]interface{}{
			"status":           "ok",
			"message":          result.Message,
			"restart_required": result.RestartRequired,
		}
		if result.AdminToken != "" {
			resp["admin_token"] = result.AdminToken
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func getRoutesHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		writeJSON(w, http.StatusOK, &snap.Routes)
	}
}

func putRoutesHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if handleBodyReadError(w, err) {
			return
		}

		if err := deps.ConfigService.ApplyRoutes(body); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "routes updated"})
	}
}

func getProfilesHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		writeJSON(w, http.StatusOK, &snap.Profiles)
	}
}

func putProfilesHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if handleBodyReadError(w, err) {
			return
		}

		if err := deps.ConfigService.ApplyProfiles(body); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "profiles updated"})
	}
}

func runtimeStartHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deps.ConfigService.UpdateMode("always_on"); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
			return
		}
		if deps.RuntimeState != nil {
			_ = deps.RuntimeState.Start()
		}
		log.Printf("ADMIN: proxy engine started")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "proxy started", "mode": "always_on"})
	}
}

func runtimeStopHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deps.ConfigService.UpdateMode("manual"); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
			return
		}
		if deps.RuntimeState != nil {
			_ = deps.RuntimeState.Stop()
		}
		log.Printf("ADMIN: proxy engine stopped")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "proxy stopped", "mode": "manual"})
	}
}

func runtimeRestartHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		if deps.RuntimeState != nil {
			_ = deps.RuntimeState.Stop()
			_ = deps.RuntimeState.Start()
		}
		log.Printf("ADMIN: proxy engine restarted")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "proxy restarted", "mode": snap.Config.App.Mode})
	}
}

func getStartupHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		writeJSON(w, http.StatusOK, snap.Config.Startup)
	}
}

func putStartupHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if handleBodyReadError(w, err) {
			return
		}

		var updated config.StartupConfig
		if err := json.Unmarshal(body, &updated); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		snap := deps.Store.Load()
		if updated.RunAtLogin != snap.Config.Startup.RunAtLogin {
			if err := startup.SetRunAtLogin(updated.RunAtLogin); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to update startup registration: "+err.Error())
				return
			}
		}

		if err := deps.ConfigService.ApplyStartup(updated); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "startup settings updated"})
	}
}

func getSecurityHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		writeJSON(w, http.StatusOK, SecurityView{
			AdminAuthEnabled:  snap.Config.Security.AdminAuthEnabled,
			TokenConfigured:   snap.Config.Security.AdminTokenHash != "",
			BindLocalhostOnly: snap.Config.Security.BindLocalhostOnly,
			AllowLANAccess:    snap.Config.Security.AllowLANAccess,
		})
	}
}

func rotateTokenHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newToken, err := deps.ConfigService.RotateToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to rotate token: "+err.Error())
			return
		}

		tokenPath := config.TokenFilePath()
		if writeErr := os.WriteFile(tokenPath, []byte(newToken), 0o600); writeErr != nil {
			log.Printf("WARNING: failed to update token file for CLI access: %v", writeErr)
		}

		log.Printf("ADMIN: admin token rotated")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"message": "token rotated — this is the only time the raw token is shown",
			"token":   newToken,
		})
	}
}

func downstreamHealthHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		client := &http.Client{Timeout: 5 * time.Second}
		parsed, err := url.Parse(snap.Config.Downstream.BaseURL)
		if err != nil {
			log.Printf("downstream health: invalid URL: %v", err)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"status":  "unavailable",
				"message": "downstream URL configuration is invalid",
			})
			return
		}
		healthURL := fmt.Sprintf("%s://%s/health", parsed.Scheme, parsed.Host)
		resp, err := client.Get(healthURL)
		if err != nil {
			log.Printf("downstream health: unreachable: %v", err)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"status":  "unavailable",
				"message": "downstream service is unreachable",
			})
			return
		}
		defer resp.Body.Close()

		status := "connected"
		msg := "downstream is healthy"
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			status = "degraded"
			msg = fmt.Sprintf("health endpoint returned HTTP %d", resp.StatusCode)
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":      status,
			"status_code": resp.StatusCode,
			"url":         snap.Config.Downstream.BaseURL,
			"message":     msg,
		})
	}
}

func logsHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		limit := 100
		if deps.Observability != nil {
			logs := deps.Observability.GetEntries(limit)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"logs":             logs,
				"total":            deps.Observability.Count(),
				"privacy_mode":     snap.Config.Logging.PrivacyMode,
				"metadata_enabled": snap.Config.Logging.MetadataLoggingEnabled,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"logs":             []interface{}{},
			"total":            0,
			"privacy_mode":     snap.Config.Logging.PrivacyMode,
			"metadata_enabled": snap.Config.Logging.MetadataLoggingEnabled,
		})
	}
}

func logsClearHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Observability != nil {
			deps.Observability.Clear()
		}
		log.Printf("ADMIN: metadata logs cleared")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "logs cleared"})
	}
}

func diagnosticsExportHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		diagnostics := map[string]interface{}{
			"version":    Version,
			"uptime":     time.Since(startTime).String(),
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"pid":        os.Getpid(),
			"config": map[string]interface{}{
				"host":           snap.Config.Server.Host,
				"port":           snap.Config.Server.Port,
				"downstream_url": snap.Config.Downstream.BaseURL,
				"mode":           snap.Config.App.Mode,
				"privacy_mode":   snap.Config.Logging.PrivacyMode,
				"auto_routing":   snap.Config.Routing.AutoRouting,
			},
			"route_profiles_count": len(snap.Profiles.RouteProfiles),
			"task_routes_count":    len(snap.Routes.TaskRoutes),
			"config_version":       snap.Version,
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"diagnostics.json\"")
		json.NewEncoder(w).Encode(diagnostics)
	}
}

func dryRunHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		body, err := io.ReadAll(r.Body)
		if handleBodyReadError(w, err) {
			return
		}

		var req struct {
			Prompt string `json:"prompt"`
			Model  string `json:"model"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if req.Prompt == "" {
			writeError(w, http.StatusBadRequest, "prompt is required")
			return
		}

		mockBody, err := buildMockChatBody(req.Model, req.Prompt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to build request body: "+err.Error())
			return
		}

		analysis, err := analyzer.Analyze(bytes.NewReader(mockBody))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "analysis failed: "+err.Error())
			return
		}

		classification := classifier.Classify(analysis)

		decision := routing.Resolve(req.Model, analysis, classification, &snap.Routes, &snap.Profiles, &snap.Config.Routing)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"analysis":       analysis,
			"classification": classification,
			"decision":       decision,
		})
	}
}

func comboTestHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		body, err := io.ReadAll(r.Body)
		if handleBodyReadError(w, err) {
			return
		}

		var req struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if req.Model == "" {
			writeError(w, http.StatusBadRequest, "model is required")
			return
		}

		if snap.Forwarder == nil {
			writeError(w, http.StatusServiceUnavailable, "downstream service unavailable")
			return
		}

		testBody, err := buildComboTestBody(req.Model)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to build request body: "+err.Error())
			return
		}

		start := time.Now()

		proxyReq, err := http.NewRequestWithContext(r.Context(), "POST", "/v1/chat/completions", bytes.NewReader(testBody))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create request: "+err.Error())
			return
		}
		proxyReq.Header.Set("Content-Type", "application/json")
		proxyReq.Header.Set("X-Request-ID", "combo-test")

		result, err := snap.Forwarder.Forward(r.Context(), proxyReq, "combo-test")
		latency := time.Since(start)

		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"model":   req.Model,
				"success": false,
				"error":   err.Error(),
				"latency": latency.Milliseconds(),
			})
			return
		}

		var parsed struct {
			Model string `json:"model"`
		}
		json.Unmarshal(result.Body, &parsed)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"model":         req.Model,
			"resolved_model": parsed.Model,
			"success":       result.StatusCode >= 200 && result.StatusCode < 300,
			"status_code":   result.StatusCode,
			"latency":       latency.Milliseconds(),
		})
	}
}

func configExportHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := deps.Store.Load()
		secView := SecurityView{
			AdminAuthEnabled:  snap.Config.Security.AdminAuthEnabled,
			TokenConfigured:   snap.Config.Security.AdminTokenHash != "",
			BindLocalhostOnly: snap.Config.Security.BindLocalhostOnly,
			AllowLANAccess:    snap.Config.Security.AllowLANAccess,
		}
		cfgCopy := snap.Config
		cfgCopy.Security = config.SecurityConfig{
			AdminAuthEnabled:  secView.AdminAuthEnabled,
			BindLocalhostOnly: secView.BindLocalhostOnly,
			AllowLANAccess:    secView.AllowLANAccess,
		}
		export := map[string]interface{}{
			"config":   cfgCopy,
			"routes":   &snap.Routes,
			"profiles": &snap.Profiles,
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"atlasbridge-config.json\"")
		json.NewEncoder(w).Encode(export)
	}
}

func configImportHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if handleBodyReadError(w, err) {
			return
		}

		if err := deps.ConfigService.Import(body); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "config imported"})
	}
}

func configResetHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deps.ConfigService.Reset(); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reset: "+err.Error())
			return
		}

		log.Printf("ADMIN: config reset to defaults")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "config reset to defaults"})
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"message": message,
			"type":    "admin_error",
		},
	})
}

func handleBodyReadError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if isMaxBytesError(err) {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]interface{}{
			"error": map[string]string{
				"message": "request body too large",
				"type":    "invalid_request_error",
				"code":    "payload_too_large",
			},
		})
		return true
	}
	writeError(w, http.StatusBadRequest, "failed to read request body")
	return true
}
