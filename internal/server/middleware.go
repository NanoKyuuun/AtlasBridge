package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"mime"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const maxRequestIDLen = 128

var validRequestIDChars = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" || len(id) > maxRequestIDLen || !validRequestIDChars.MatchString(id) {
			b := make([]byte, 8)
			if _, err := rand.Read(b); err != nil {
				id = "req-error"
			} else {
				id = hex.EncodeToString(b)
			}
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SafeLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)

		reqID, _ := r.Context().Value(RequestIDKey).(string)

		log.Printf("[%s] %s %s %d %v",
			reqID,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration.Round(time.Microsecond),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

var _ http.ResponseWriter = (*responseWriter)(nil)
var _ http.Flusher = (*responseWriter)(nil)

// RequireJSON rejects state-changing requests (POST, PUT, PATCH) that do not
// have Content-Type: application/json. This prevents CSRF-style attacks that
// rely on text/plain or multipart form submissions.
func RequireJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost ||
			r.Method == http.MethodPut ||
			r.Method == http.MethodPatch {
			mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || mediaType != "application/json" {
				http.Error(w, `{"error":{"message":"application/json content type required","type":"invalid_request_error"}}`, http.StatusUnsupportedMediaType)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// SameOriginAdmin rejects requests that carry an Origin header that does not
// match the allowed origin. Requests without an Origin header (CLI, curl) are
// allowed through. This blocks browser-based cross-origin state changes.
func SameOriginAdmin(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && origin != allowedOrigin {
				http.Error(w, `{"error":{"message":"forbidden origin","type":"auth_error"}}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds standard security headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// HostGuard validates the Host header against the expected listener address
// to mitigate DNS rebinding attacks. Requests with no Host header (HTTP/1.0)
// or from allowed hosts are passed through. Loopback addresses skip port
// verification since local connections are not vulnerable to DNS rebinding.
func HostGuard(expectedHost string, expectedPort int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hostHeader := r.Host
			if hostHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			hostname, portStr, err := net.SplitHostPort(hostHeader)
			if err != nil {
				hostname = hostHeader
			}

			allowedHosts := map[string]bool{
				expectedHost:     true,
				"localhost":      true,
				"127.0.0.1":      true,
				"[::1]":          true,
				"0.0.0.0":        true,
			}

			if !allowedHosts[hostname] {
				http.Error(w, `{"error":{"message":"forbidden host","type":"auth_error"}}`, http.StatusForbidden)
				return
			}

			// Skip port check for loopback addresses — not vulnerable to DNS rebinding.
			isLoopback := hostname == "localhost" || hostname == "127.0.0.1" || hostname == "[::1]" || hostname == "0.0.0.0"
			if !isLoopback && portStr != "" {
				port, err := strconv.Atoi(portStr)
				if err != nil || port != expectedPort {
					http.Error(w, `{"error":{"message":"forbidden host","type":"auth_error"}}`, http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
