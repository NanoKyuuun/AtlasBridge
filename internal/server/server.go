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
	"github.com/atlasbridge/atlasbridge/internal/analyzer"
	"github.com/atlasbridge/atlasbridge/internal/classifier"
	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/observability"
	"github.com/atlasbridge/atlasbridge/internal/routing"
	runtimemod "github.com/atlasbridge/atlasbridge/internal/runtime"
	"github.com/atlasbridge/atlasbridge/internal/security"
)

var Version = "0.1.0"

var startTime = time.Now()

type ServerDeps struct {
	Store         *StateStore
	ConfigService *ConfigService
	ObsLogger     *observability.Logger
	RuntimeState  *runtimemod.State
	Bulkhead      *WeightedBulkhead
}

func New(deps *ServerDeps) *http.Server {
	if deps.ObsLogger == nil {
		deps.ObsLogger = observability.NewLogger(observability.DefaultMaxEntries)
	}
	if deps.Store == nil || deps.ConfigService == nil {
		log.Fatal("ServerDeps.Store and ServerDeps.ConfigService must be provided")
	}
	if deps.Bulkhead == nil {
		deps.Bulkhead = NewWeightedBulkhead(MaxInFlight, 10*time.Second)
	}

	snap := deps.Store.Load()

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(RequestID)
	r.Use(SafeLogger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler)

	r.Route("/v1", func(r chi.Router) {
		r.Use(bodyLimitMiddleware(MaxChatBody))
		r.Post("/chat/completions", chatCompletionsHandler(deps))
		r.Get("/models", modelsHandler)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Use(SecurityHeaders)
			store := deps.Store
			r.Use(security.AdminAuth(func() (bool, string) {
				s := store.Load()
				return s.Config.Security.AdminAuthEnabled, s.Config.Security.AdminTokenHash
			}))
			r.Use(HostGuard(snap.Config.Server.Host, snap.Config.Server.Port))
			r.Use(RequireJSON)
			r.Use(SameOriginAdmin("http://"+fmt.Sprintf("%s:%d", snap.Config.Server.Host, snap.Config.Server.Port)))
			r.Get("/status", statusHandler(deps.Store, deps.RuntimeState, deps.Bulkhead))

			adminDeps := &AdminDeps{
				Store:         deps.Store,
				ConfigService: deps.ConfigService,
				Observability: deps.ObsLogger,
				RuntimeState:  deps.RuntimeState,
			}

			r.Get("/config", getConfigHandler(adminDeps))
			r.Put("/config", limitBody(putConfigHandler(adminDeps), MaxAdminBody))

			r.Get("/routes", getRoutesHandler(adminDeps))
			r.Put("/routes", limitBody(putRoutesHandler(adminDeps), MaxAdminBody))

			r.Get("/profiles", getProfilesHandler(adminDeps))
			r.Put("/profiles", limitBody(putProfilesHandler(adminDeps), MaxAdminBody))

			r.Get("/security", getSecurityHandler(adminDeps))
			r.Post("/security/token/rotate", rotateTokenHandler(adminDeps))

			r.Post("/runtime/start", runtimeStartHandler(adminDeps))
			r.Post("/runtime/stop", runtimeStopHandler(adminDeps))
			r.Post("/runtime/restart", runtimeRestartHandler(adminDeps))

			r.Get("/startup", getStartupHandler(adminDeps))
			r.Put("/startup", limitBody(putStartupHandler(adminDeps), MaxAdminBody))

			r.Get("/downstream/health", downstreamHealthHandler(adminDeps))

			r.Get("/logs", logsHandler(adminDeps))
			r.Post("/logs/clear", logsClearHandler(adminDeps))
			r.Post("/diagnostics/export", diagnosticsExportHandler(adminDeps))

			r.Post("/routing/dry-run", limitBody(dryRunHandler(adminDeps), MaxAdminBody))

			r.Post("/combo/test", limitBody(comboTestHandler(adminDeps), MaxAdminBody))

			r.Post("/config/import", limitBody(configImportHandler(adminDeps), MaxImportBody))
			r.Get("/config/export", configExportHandler(adminDeps))
			r.Post("/config/reset", configResetHandler(adminDeps))
		})

		r.Get("/", adminUIHandler())
		r.Get("/*", adminUIHandler())
	})

	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", snap.Config.Server.Host, snap.Config.Server.Port),
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

func statusHandler(store *StateStore, runtimeState *runtimemod.State, bulkhead *WeightedBulkhead) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := store.Load()
		cfg := &snap.Config
		runtimeStatus := string(runtimemod.StatusRunning)
		runtimeMode := cfg.App.Mode
		runtimeUptime := time.Since(startTime).String()
		if runtimeState != nil {
			runtimeStatus = string(runtimeState.GetStatus())
			runtimeMode = string(runtimeState.GetMode())
			if u := runtimeState.GetUptime(); u > 0 {
				runtimeUptime = u.String()
			}
		}
		var bulkheadStats BulkheadStats
		if bulkhead != nil {
			bulkheadStats = bulkhead.Stats()
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     runtimeStatus,
			"version":    Version,
			"uptime":     runtimeUptime,
			"port":       cfg.Server.Port,
			"host":       cfg.Server.Host,
			"downstream": cfg.Downstream.BaseURL,
			"mode":       runtimeMode,
			"privacy":    cfg.Logging.PrivacyMode,
			"go_version": runtime.Version(),
			"pid":        os.Getpid(),
			"config_version": snap.Version,
			"bulkhead":   bulkheadStats,
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

type chatRequestEnvelope struct {
	Model    string          `json:"model"`
	Stream   *bool           `json:"stream"`
	Messages json.RawMessage `json:"messages"`
	Raw      json.RawMessage `json:"-"`
}

func decodeChatRequest(body []byte) (*chatRequestEnvelope, error) {
	var env chatRequestEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		if len(body) == 0 {
			return nil, &ValidationError{Message: "Request body must not be empty", Code: "invalid_request_error"}
		}
		return nil, &ValidationError{Message: "Invalid JSON", Code: "invalid_json"}
	}
	env.Raw = body
	return &env, nil
}

func (env *chatRequestEnvelope) IsStream() bool {
	return env.Stream != nil && *env.Stream
}

func (env *chatRequestEnvelope) Validate() *ValidationError {
	if env.Messages == nil {
		return &ValidationError{Message: "Missing required field: messages", Code: "invalid_request_error"}
	}

	var msgs []json.RawMessage
	if err := json.Unmarshal(env.Messages, &msgs); err != nil || len(msgs) == 0 {
		return &ValidationError{Message: "messages must not be empty", Code: "invalid_request_error"}
	}
	if len(msgs) > MaxMessages {
		return &ValidationError{
			Message: fmt.Sprintf("messages array exceeds maximum of %d entries", MaxMessages),
			Code:    "invalid_request_error",
		}
	}
	if len(env.Model) > MaxModelName {
		return &ValidationError{
			Message: fmt.Sprintf("model name exceeds maximum length of %d characters", MaxModelName),
			Code:    "invalid_request_error",
		}
	}
	for i, raw := range msgs {
		var msg struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(raw, &msg); err == nil {
			if len(msg.Content) > MaxMessageLen {
				return &ValidationError{
					Message: fmt.Sprintf("message %d content exceeds maximum length of %d characters", i, MaxMessageLen),
					Code:    "invalid_request_error",
				}
			}
		}
	}
	return nil
}

func chatCompletionsHandler(deps *ServerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cw := &commitWriter{ResponseWriter: w}

		snap := deps.Store.Load()

		if deps.RuntimeState != nil {
			mode := deps.RuntimeState.GetMode()
			if mode == runtimemod.ModeDisabled {
				cw.Header().Set("Content-Type", "application/json")
				cw.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(cw).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"message": "proxy is disabled",
						"type":    "proxy_error",
						"code":    "proxy_disabled",
					},
				})
				return
			}

			if mode != runtimemod.ModeAlwaysOn {
				status := deps.RuntimeState.GetStatus()
				if status != runtimemod.StatusRunning {
					cw.Header().Set("Content-Type", "application/json")
					cw.WriteHeader(http.StatusServiceUnavailable)
					json.NewEncoder(cw).Encode(map[string]interface{}{
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
		reqID = validateRequestID(reqID)

		if snap.Forwarder == nil {
			cw.Header().Set("Content-Type", "application/json")
			cw.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(cw).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "downstream service unavailable",
					"type":    "proxy_error",
					"code":    "downstream_unavailable",
				},
			})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			if isMaxBytesError(err) {
				cw.Header().Set("Content-Type", "application/json")
				cw.WriteHeader(http.StatusRequestEntityTooLarge)
				json.NewEncoder(cw).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"message": "request body too large",
						"type":    "invalid_request_error",
						"code":    "payload_too_large",
					},
				})
				return
			}
			cw.Header().Set("Content-Type", "application/json")
			cw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(cw).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "failed to read request body",
					"type":    "invalid_request_error",
					"code":    "body_read_error",
				},
			})
			return
		}
		r.Body = newCloserReader(bytes.NewReader(body))

		env, err := decodeChatRequest(body)
		if err != nil {
			ve := err.(*ValidationError)
			cw.Header().Set("Content-Type", "application/json")
			cw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(cw).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": ve.Message,
					"type":    "invalid_request_error",
					"code":    ve.Code,
				},
			})
			return
		}

		if ve := env.Validate(); ve != nil {
			cw.Header().Set("Content-Type", "application/json")
			cw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(cw).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": ve.Message,
					"type":    "invalid_request_error",
					"code":    ve.Code,
				},
			})
			return
		}

		weight := WeightNonstream
		if env.IsStream() {
			weight = WeightStream
		}

		if !deps.Bulkhead.Acquire(weight, 5*time.Second) {
			writeOverloadedJSON(cw, 2)
			return
		}
		defer deps.Bulkhead.Release(weight)

		decision := analyzeAndRoute(r, body, env.Model, env.IsStream(), &snap.Routes, &snap.Profiles, &snap.Config.Routing)
		if snap.Config.Logging.MetadataLoggingEnabled {
			logRoutingDecision(reqID, decision)
			recordObservation(deps.ObsLogger, reqID, r, decision, body)
		}

		if decision.DownstreamAlias != "" && decision.OverrideSource == "smart_alias" {
			if snap.Config.Routing.MetadataTransport == "header" {
				r.Header.Set("X-Route-Intent", decision.DownstreamAlias)
			} else if snap.Config.Routing.MetadataTransport == "model_alias" {
				rewritten, err := rewriteModelInBody(body, decision.DownstreamAlias)
				if err != nil {
					log.Printf("[%s] failed to rewrite model in body: %v", reqID, err)
				} else {
					body = rewritten
					r.Body = newCloserReader(bytes.NewReader(body))
					r.ContentLength = int64(len(body))
				}
			}
		}

		if env.IsStream() {
			err := snap.Forwarder.ForwardStream(r.Context(), r, cw, reqID)
			if err != nil {
				log.Printf("[%s] stream error: %v", reqID, err)
				writeJSONAfterCommit(cw, http.StatusBadGateway, map[string]interface{}{
					"error": map[string]interface{}{
						"message":        "downstream request failed",
						"type":           "proxy_error",
						"code":           "downstream_error",
						"correlation_id": reqID,
					},
				})
			}
			return
		}

		result, err := snap.Forwarder.Forward(r.Context(), r, reqID)
		if err != nil {
			log.Printf("[%s] forward error: %v", reqID, err)
			cw.Header().Set("Content-Type", "application/json")
			cw.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(cw).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message":        "downstream request failed",
					"type":           "proxy_error",
					"code":           "downstream_error",
					"correlation_id": reqID,
				},
			})
			return
		}

		copyAllowedHeaders(cw, result.Headers)
		cw.Header().Set("Content-Type", "application/json")
		cw.Header().Set("X-Request-ID", reqID)
		cw.WriteHeader(result.StatusCode)
		cw.Write(result.Body)
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

func analyzeAndRoute(
	r *http.Request,
	body []byte,
	model string,
	isStream bool,
	routes *config.RoutesConfig,
	profiles *config.ProfilesConfig,
	routingCfg *config.RoutingConfig,
) *routing.RoutingDecision {
	analysis, err := analyzer.Analyze(bytes.NewReader(body))
	if err != nil {
		log.Printf("analysis error: %v", err)
		return routing.Resolve(model, nil, nil, routes, profiles, routingCfg)
	}

	classification := classifier.Classify(analysis)

	return routing.Resolve(model, analysis, classification, routes, profiles, routingCfg)
}

func rewriteModelInBody(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	payload["model"] = model
	return json.Marshal(payload)
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

	model := extractModelFromRaw(body)

	var isStream bool
	if idx := bytes.Index(body, []byte(`"stream"`)); idx >= 0 {
		isStream = bytes.Contains(body[idx:min(idx+20, len(body))], []byte("true"))
	}

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

func extractModelFromRaw(body []byte) string {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func adminPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>AtlasBridge Admin</title>
<style>body{font-family:system-ui,-apple-system,sans-serif;max-width:600px;margin:40px auto;padding:0 20px;color:#333}
h1{color:#1a1a2e}p{color:#666}</style></head>
<body>
<h1>AtlasBridge Dashboard</h1>
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
