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
	"strings"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/analyzer"
	"github.com/atlasbridge/atlasbridge/internal/classifier"
	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/forwarder"
	"github.com/atlasbridge/atlasbridge/internal/observability"
	"github.com/atlasbridge/atlasbridge/internal/routing"
	runtimemod "github.com/atlasbridge/atlasbridge/internal/runtime"
	"github.com/atlasbridge/atlasbridge/internal/security"
	"github.com/atlasbridge/atlasbridge/internal/startup"
)

type AdminDeps struct {
	Config        *config.Config
	Routes        *config.RoutesConfig
	Profiles      *config.ProfilesConfig
	Forwarder     *forwarder.Forwarder
	Observability *observability.Logger
	RuntimeState  *runtimemod.State
	AuthConfig    *AuthConfig
}

func getConfigHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		masked := maskConfig(deps.Config)
		writeJSON(w, http.StatusOK, masked)
	}
}

func putConfigHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}

		var patch map[string]json.RawMessage
		if err := json.Unmarshal(body, &patch); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		merged := *deps.Config

		if raw, ok := patch["app"]; ok {
			if err := json.Unmarshal(raw, &merged.App); err != nil {
				writeError(w, http.StatusBadRequest, "invalid app config: "+err.Error())
				return
			}
		}
		if raw, ok := patch["server"]; ok {
			if err := json.Unmarshal(raw, &merged.Server); err != nil {
				writeError(w, http.StatusBadRequest, "invalid server config: "+err.Error())
				return
			}
		}
		if raw, ok := patch["downstream"]; ok {
			if err := json.Unmarshal(raw, &merged.Downstream); err != nil {
				writeError(w, http.StatusBadRequest, "invalid downstream config: "+err.Error())
				return
			}
		}
		if raw, ok := patch["security"]; ok {
			if err := json.Unmarshal(raw, &merged.Security); err != nil {
				writeError(w, http.StatusBadRequest, "invalid security config: "+err.Error())
				return
			}
		}
		if raw, ok := patch["startup"]; ok {
			if err := json.Unmarshal(raw, &merged.Startup); err != nil {
				writeError(w, http.StatusBadRequest, "invalid startup config: "+err.Error())
				return
			}
		}
		if raw, ok := patch["routing"]; ok {
			if err := json.Unmarshal(raw, &merged.Routing); err != nil {
				writeError(w, http.StatusBadRequest, "invalid routing config: "+err.Error())
				return
			}
		}
		if raw, ok := patch["logging"]; ok {
			if err := json.Unmarshal(raw, &merged.Logging); err != nil {
				writeError(w, http.StatusBadRequest, "invalid logging config: "+err.Error())
				return
			}
		}

		if err := config.Validate(&merged); err != nil {
			writeError(w, http.StatusBadRequest, "validation failed: "+err.Error())
			return
		}

		var rawToken string
		if merged.Security.AdminAuthEnabled {
			rawToken, err = security.EnsureToken(&merged.Security.AdminTokenHash)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to generate admin token")
				return
			}
			if rawToken != "" {
				fmt.Fprint(os.Stdout, "\n")
				fmt.Fprintf(os.Stdout, "  ADMIN TOKEN: %s\n", rawToken)
				fmt.Fprint(os.Stdout, "  This token will NOT be shown again.\n")
				fmt.Fprint(os.Stdout, "  Store it safely before closing this window.\n")
				fmt.Fprint(os.Stdout, "\n")
			}
		}

		deps.AuthConfig.Set(merged.Security.AdminAuthEnabled, merged.Security.AdminTokenHash)

		*deps.Config = merged
		if err := config.Save(deps.Config); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
			return
		}

		resp := map[string]interface{}{
			"status":  "ok",
			"message": "config updated",
		}
		if rawToken != "" {
			resp["admin_token"] = rawToken
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func getRoutesHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, deps.Routes)
	}
}

func putRoutesHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}

		var updated config.RoutesConfig
		if err := json.Unmarshal(body, &updated); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if err := config.ValidateFull(deps.Config, &updated, deps.Profiles); err != nil {
			writeError(w, http.StatusBadRequest, "validation failed: "+err.Error())
			return
		}

		*deps.Routes = updated
		if err := config.SaveRoutes(deps.Routes); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save routes: "+err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "routes updated"})
	}
}

func getProfilesHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, deps.Profiles)
	}
}

func putProfilesHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}

		var updated config.ProfilesConfig
		if err := json.Unmarshal(body, &updated); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if err := config.ValidateFull(deps.Config, deps.Routes, &updated); err != nil {
			writeError(w, http.StatusBadRequest, "validation failed: "+err.Error())
			return
		}

		*deps.Profiles = updated
		if err := config.SaveProfiles(deps.Profiles); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save profiles: "+err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "profiles updated"})
	}
}

func runtimeStartHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deps.Config.App.Mode = "always_on"
		if err := config.Save(deps.Config); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
			return
		}
		if deps.RuntimeState != nil {
			_ = deps.RuntimeState.Start()
		}
		log.Printf("ADMIN: proxy engine started")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "proxy started", "mode": deps.Config.App.Mode})
	}
}

func runtimeStopHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deps.Config.App.Mode = "disabled"
		if err := config.Save(deps.Config); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
			return
		}
		if deps.RuntimeState != nil {
			_ = deps.RuntimeState.Stop()
		}
		log.Printf("ADMIN: proxy engine stopped")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "proxy stopped", "mode": deps.Config.App.Mode})
	}
}

func runtimeRestartHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.RuntimeState != nil {
			_ = deps.RuntimeState.Stop()
			_ = deps.RuntimeState.Start()
		}
		log.Printf("ADMIN: proxy engine restarted")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "proxy restarted", "mode": deps.Config.App.Mode})
	}
}

func getStartupHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, deps.Config.Startup)
	}
}

func putStartupHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}

		var updated config.StartupConfig
		if err := json.Unmarshal(body, &updated); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if updated.RunAtLogin != deps.Config.Startup.RunAtLogin {
			if err := startup.SetRunAtLogin(updated.RunAtLogin); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to update startup registration: "+err.Error())
				return
			}
		}

		deps.Config.Startup = updated
		if err := config.Save(deps.Config); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "startup settings updated"})
	}
}

func downstreamHealthHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 5 * time.Second}
		parsed, err := url.Parse(deps.Config.Downstream.BaseURL)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"status":  "unavailable",
				"message": fmt.Sprintf("invalid downstream URL: %v", err),
				"url":     deps.Config.Downstream.BaseURL,
			})
			return
		}
		healthURL := fmt.Sprintf("%s://%s/health", parsed.Scheme, parsed.Host)
		resp, err := client.Get(healthURL)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"status":  "unavailable",
				"message": fmt.Sprintf("9Router unreachable: %v", err),
				"url":     deps.Config.Downstream.BaseURL,
			})
			return
		}
		defer resp.Body.Close()

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":      "connected",
			"status_code": resp.StatusCode,
			"url":         deps.Config.Downstream.BaseURL,
		})
	}
}

func logsHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 100
		if deps.Observability != nil {
			logs := deps.Observability.GetEntries(limit)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"logs":             logs,
				"total":            deps.Observability.Count(),
				"privacy_mode":     deps.Config.Logging.PrivacyMode,
				"metadata_enabled": deps.Config.Logging.MetadataLoggingEnabled,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"logs":             []interface{}{},
			"total":            0,
			"privacy_mode":     deps.Config.Logging.PrivacyMode,
			"metadata_enabled": deps.Config.Logging.MetadataLoggingEnabled,
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
		diagnostics := map[string]interface{}{
			"version":    Version,
			"uptime":     time.Since(startTime).String(),
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"pid":        os.Getpid(),
			"config": map[string]interface{}{
				"host":           deps.Config.Server.Host,
				"port":           deps.Config.Server.Port,
				"downstream_url": deps.Config.Downstream.BaseURL,
				"mode":           deps.Config.App.Mode,
				"privacy_mode":   deps.Config.Logging.PrivacyMode,
				"auto_routing":   deps.Config.Routing.AutoRouting,
			},
			"route_profiles_count": len(deps.Profiles.RouteProfiles),
			"task_routes_count":    len(deps.Routes.TaskRoutes),
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"diagnostics.json\"")
		json.NewEncoder(w).Encode(diagnostics)
	}
}

func dryRunHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body")
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

		mockBody := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"%s"}]}`,
			req.Model, req.Prompt)

		analysis, err := analyzer.Analyze(bytes.NewReader([]byte(mockBody)))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "analysis failed: "+err.Error())
			return
		}

		classification := classifier.Classify(analysis)

		decision := routing.Resolve(req.Model, analysis, classification, deps.Routes, deps.Profiles, &deps.Config.Routing)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"analysis":       analysis,
			"classification": classification,
			"decision":       decision,
		})
	}
}

func comboTestHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body")
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

		if deps.Forwarder == nil {
			writeError(w, http.StatusServiceUnavailable, "downstream service unavailable")
			return
		}

		testBody := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"Reply with only: OK"}],"max_tokens":5}`,
			req.Model)

		start := time.Now()

		proxyReq, err := http.NewRequestWithContext(r.Context(), "POST", "/v1/chat/completions", bytes.NewReader([]byte(testBody)))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create request: "+err.Error())
			return
		}
		proxyReq.Header.Set("Content-Type", "application/json")
		proxyReq.Header.Set("X-Request-ID", "combo-test")

		result, err := deps.Forwarder.Forward(r.Context(), proxyReq, "combo-test")
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
		export := map[string]interface{}{
			"config":   maskConfig(deps.Config),
			"routes":   deps.Routes,
			"profiles": deps.Profiles,
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"atlasbridge-config.json\"")
		json.NewEncoder(w).Encode(export)
	}
}

func configImportHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}

		var imported struct {
			Config   *config.Config         `json:"config"`
			Routes   *config.RoutesConfig   `json:"routes"`
			Profiles *config.ProfilesConfig `json:"profiles"`
		}
		if err := json.Unmarshal(body, &imported); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if imported.Config != nil {
			if err := config.Validate(imported.Config); err != nil {
				writeError(w, http.StatusBadRequest, "config validation failed: "+err.Error())
				return
			}
			*deps.Config = *imported.Config
			config.Save(deps.Config)
		}

		if imported.Routes != nil {
			if imported.Profiles != nil {
				if err := config.ValidateFull(deps.Config, imported.Routes, imported.Profiles); err != nil {
					writeError(w, http.StatusBadRequest, "routes validation failed: "+err.Error())
					return
				}
			}
			*deps.Routes = *imported.Routes
			if err := config.SaveRoutes(deps.Routes); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to save routes: "+err.Error())
				return
			}
		}

		if imported.Profiles != nil {
			if imported.Routes != nil {
				if err := config.ValidateFull(deps.Config, imported.Routes, imported.Profiles); err != nil {
					writeError(w, http.StatusBadRequest, "profiles validation failed: "+err.Error())
					return
				}
			}
			*deps.Profiles = *imported.Profiles
			if err := config.SaveProfiles(deps.Profiles); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to save profiles: "+err.Error())
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "config imported"})
	}
}

func configResetHandler(deps *AdminDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*deps.Config = *config.DefaultConfig()
		*deps.Routes = *config.DefaultRoutesConfig()
		*deps.Profiles = *config.DefaultProfilesConfig()

		if err := config.Save(deps.Config); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
			return
		}
		if err := config.SaveRoutes(deps.Routes); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save routes: "+err.Error())
			return
		}
		if err := config.SaveProfiles(deps.Profiles); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save profiles: "+err.Error())
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

func maskConfig(cfg *config.Config) *config.Config {
	masked := *cfg
	if masked.Security.AdminTokenHash != "" {
		masked.Security.AdminTokenHash = maskSecret(masked.Security.AdminTokenHash)
	}
	return &masked
}

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
}
