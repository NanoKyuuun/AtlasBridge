package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/analyzer"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/classifier"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/config"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/forwarder"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/observability"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/routing"
	runtimemod "github.com/smart-ai-proxy/smart-ai-proxy/internal/runtime"
)

const Version = "0.1.0"

var startTime = time.Now()

type ServerDeps struct {
	Config        *config.Config
	Routes        *config.RoutesConfig
	Profiles      *config.ProfilesConfig
	Fwd           *forwarder.Forwarder
	ObsLogger     *observability.Logger
	RuntimeState  *runtimemod.State
}

func New(deps *ServerDeps) *http.Server {
	// Create forwarder if not provided (for backward compatibility)
	if deps.Fwd == nil {
		fwd, err := forwarder.New(deps.Config.Downstream.BaseURL, deps.Config.Downstream.TimeoutSeconds)
		if err != nil {
			log.Printf("WARNING: failed to create forwarder: %v", err)
		}
		deps.Fwd = fwd
	}

	// Create observability logger if not provided
	if deps.ObsLogger == nil {
		deps.ObsLogger = observability.NewLogger(observability.DefaultMaxEntries)
	}

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(RequestID)
	r.Use(SafeLogger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/chat/completions", chatCompletionsHandler(deps))
		r.Get("/models", modelsHandler)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Get("/status", statusHandler(deps.Config))

			adminDeps := &AdminDeps{
				Config:        deps.Config,
				Routes:        deps.Routes,
				Profiles:      deps.Profiles,
				Forwarder:     deps.Fwd,
				Observability: deps.ObsLogger,
			}

			r.Get("/config", getConfigHandler(adminDeps))
			r.Put("/config", putConfigHandler(adminDeps))

			r.Get("/routes", getRoutesHandler(adminDeps))
			r.Put("/routes", putRoutesHandler(adminDeps))

			r.Get("/profiles", getProfilesHandler(adminDeps))
			r.Put("/profiles", putProfilesHandler(adminDeps))

			r.Post("/runtime/start", runtimeStartHandler(adminDeps))
			r.Post("/runtime/stop", runtimeStopHandler(adminDeps))
			r.Post("/runtime/restart", runtimeRestartHandler(adminDeps))

			r.Get("/startup", getStartupHandler(adminDeps))
			r.Put("/startup", putStartupHandler(adminDeps))

			r.Get("/downstream/health", downstreamHealthHandler(adminDeps))

			r.Get("/logs", logsHandler(adminDeps))
			r.Post("/logs/clear", logsClearHandler(adminDeps))
			r.Post("/diagnostics/export", diagnosticsExportHandler(adminDeps))

			r.Post("/routing/dry-run", dryRunHandler(adminDeps))

			r.Post("/combo/test", comboTestHandler(adminDeps))

			r.Post("/config/import", configImportHandler(adminDeps))
			r.Get("/config/export", configExportHandler(adminDeps))
			r.Post("/config/reset", configResetHandler(adminDeps))
		})

		r.Get("/", adminUIHandler())
		r.Get("/*", adminUIHandler())
	})

	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", deps.Config.Server.Host, deps.Config.Server.Port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

// All other helper functions...

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "atlasbridge",
		"version": Version,
	})
}

func statusHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "running",
			"version":    Version,
			"uptime":     time.Since(startTime).String(),
			"port":       cfg.Server.Port,
			"host":       cfg.Server.Host,
			"downstream": cfg.Downstream.BaseURL,
			"mode":       cfg.App.Mode,
			"privacy":    cfg.Logging.PrivacyMode,
			"go_version": runtime.Version(),
			"pid":        os.Getpid(),
		})
	}
}

func modelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data": []map[string]string{
			{"id": "atlas-auto", "object": "model", "owned_by": "atlasbridge"},
			{"id": "atlas-debug", "object": "model", "owned_by": "atlasbridge"},
			{"id": "atlas-cheap", "object": "model", "owned_by": "atlasbridge"},
			{"id": "atlas-docs", "object": "model", "owned_by": "atlasbridge"},
			{"id": "atlas-architect", "object": "model", "owned_by": "atlasbridge"},
			{"id": "atlas-fast", "object": "model", "owned_by": "atlasbridge"},
			{"id": "atlas-long-context", "object": "model", "owned_by": "atlasbridge"},
			{"id": "smart-auto", "object": "model", "owned_by": "atlasbridge"},
			{"id": "smart-debug", "object": "model", "owned_by": "atlasbridge"},
			{"id": "smart-cheap", "object": "model", "owned_by": "atlasbridge"},
			{"id": "smart-docs", "object": "model", "owned_by": "atlasbridge"},
			{"id": "smart-architect", "object": "model", "owned_by": "atlasbridge"},
			{"id": "smart-code", "object": "model", "owned_by": "atlasbridge"},
			{"id": "smart-fast", "object": "model", "owned_by": "atlasbridge"},
			{"id": "smart-long-context", "object": "model", "owned_by": "atlasbridge"},
		},
	})
}

func chatCompletionsHandler(deps *ServerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check runtime state only if provided
		if deps.RuntimeState != nil {
			mode := deps.RuntimeState.GetMode()
			if mode == runtimemod.ModeDisabled {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"message": "proxy is disabled",
						"type":    "proxy_error",
						"code":    "proxy_disabled",
					},
				})
				return
			}

			// Only check status if mode is not AlwaysOn (always on allows requests even when stopped)
			if mode != runtimemod.ModeAlwaysOn {
				status := deps.RuntimeState.GetStatus()
				if status != runtimemod.StatusRunning {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": map[string]interface{}{
							"message": "proxy is not running",
							"type":    "proxy_error",
							"code":    "proxy_not_running",
						},
					})
					return
				}
			}
		}

		reqID, _ := r.Context().Value(RequestIDKey).(string)

		if deps.Fwd == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "downstream service unavailable",
					"type":    "proxy_error",
					"code":    "downstream_unavailable",
				},
			})
			return
		}

		body := readBodyForAnalysis(r)

		if err := validateRequestBody(body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": err.Error(),
					"type":    "invalid_request_error",
					"code":    err.Code,
				},
			})
			return
		}

		decision := analyzeAndRoute(r, body, deps.Routes, deps.Profiles, &deps.Config.Routing)
		logRoutingDecision(reqID, decision)
		recordObservation(deps.ObsLogger, reqID, r, decision, body)

		if deps.Config.Routing.MetadataTransport == "header" && decision.DownstreamAlias != "" {
			r.Header.Set("X-Route-Intent", decision.DownstreamAlias)
		}

		isStream := forwarder.IsStreamRequest(r)

		if isStream {
			err := deps.Fwd.ForwardStream(r.Context(), r, w, reqID)
			if err != nil {
				log.Printf("[%s] stream error: %v", reqID, err)
				if !headersWritten(w) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadGateway)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": map[string]interface{}{
							"message": fmt.Sprintf("downstream error: %v", err),
							"type":    "proxy_error",
							"code":    "downstream_error",
						},
					})
				}
			}
			return
		}

		result, err := deps.Fwd.Forward(r.Context(), r, reqID)
		if err != nil {
			log.Printf("[%s] forward error: %v", reqID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": fmt.Sprintf("downstream error: %v", err),
					"type":    "proxy_error",
					"code":    "downstream_error",
				},
			})
			return
		}

		for k, vv := range result.Headers {
			if k == "Content-Type" || k == "X-Request-ID" {
				continue
			}
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", reqID)
		w.WriteHeader(result.StatusCode)
		w.Write(result.Body)
	}
}

// ValidationError represents an invalid request error with an OpenAI-compatible code.
type ValidationError struct {
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// validateRequestBody checks that the body is valid JSON with a non-empty messages array.
func validateRequestBody(body []byte) *ValidationError {
	if len(body) == 0 {
		return &ValidationError{Message: "Request body must not be empty", Code: "invalid_request_error"}
	}

	var req struct {
		Messages []interface{} `json:"messages"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return &ValidationError{Message: "Invalid JSON", Code: "invalid_json"}
	}

	if req.Messages == nil {
		return &ValidationError{Message: "Missing required field: messages", Code: "invalid_request_error"}
	}
	if len(req.Messages) == 0 {
		return &ValidationError{Message: "messages must not be empty", Code: "invalid_request_error"}
	}

	return nil
}

func readBodyForAnalysis(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil
	}
	r.Body = newCloserReader(bytes.NewReader(body))
	return body
}

func analyzeAndRoute(
	r *http.Request,
	body []byte,
	routes *config.RoutesConfig,
	profiles *config.ProfilesConfig,
	routingCfg *config.RoutingConfig,
) *routing.RoutingDecision {
	model := extractModelFromBody(body)

	analysis, err := analyzer.Analyze(bytes.NewReader(body))
	if err != nil {
		log.Printf("analysis error: %v", err)
		return routing.Resolve(model, nil, nil, routes, profiles, routingCfg)
	}

	classification := classifier.Classify(analysis)

	return routing.Resolve(model, analysis, classification, routes, profiles, routingCfg)
}

func extractModelFromBody(body []byte) string {
	if body == nil {
		return ""
	}
	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}
	return req.Model
}

func logRoutingDecision(reqID string, decision *routing.RoutingDecision) {
	if decision == nil {
		return
	}
	log.Printf("[%s] ROUTE task=%s confidence=%.2f route=%s alias=%s source=%s status=%s reason=%s",
		reqID,
		decision.TaskType,
		decision.Confidence,
		decision.RouteKey,
		decision.DownstreamAlias,
		decision.OverrideSource,
		decision.ClassificationStatus,
		decision.RoutingReason,
	)
}

func recordObservation(obsLogger *observability.Logger, reqID string, r *http.Request, decision *routing.RoutingDecision, body []byte) {
	if obsLogger == nil || decision == nil {
		return
	}

	model := extractModelFromBody(body)
	isStream := strings.Contains(strings.ToLower(string(body)), `"stream":true`)

	entry := observability.LogEntry{
		RequestID:   reqID,
		TaskType:    decision.TaskType,
		Confidence:  decision.Confidence,
		RouteKey:    decision.RouteKey,
		Alias:       decision.DownstreamAlias,
		OverrideSrc: decision.OverrideSource,
		Status:      decision.ClassificationStatus,
		Reason:      decision.RoutingReason,
		Model:       model,
		Method:      r.Method,
		Path:        r.URL.Path,
		IsStream:    isStream,
	}

	obsLogger.Record(entry)
}

func headersWritten(w http.ResponseWriter) bool {
	return false
}

func adminPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Smart AI Proxy Admin</title>
<style>body{font-family:system-ui,-apple-system,sans-serif;max-width:600px;margin:40px auto;padding:0 20px;color:#333}
h1{color:#1a1a2e}p{color:#666}</style></head>
<body>
<h1>Smart AI Proxy Admin</h1>
<p>Web UI will be available in Phase 3.</p>
<p>Health: <a href="/health">/health</a></p>
<p>Status: <a href="/admin/api/status">/admin/api/status</a></p>
</body></html>`)
}

type closerReader struct {
	*bytes.Reader
}

func (c *closerReader) Close() error {
	return nil
}

func newCloserReader(r *bytes.Reader) *closerReader {
	return &closerReader{r}
}

func mimeContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".html":
		return "text/html"
	case ".json":
		return "application/json"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	default:
		return mime.TypeByExtension(ext)
	}
}
