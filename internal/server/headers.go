package server

import "net/http"

var hopByHopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true,
	"Trailer":             true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
	"Set-Cookie":          true,
}

var allowedDownstreamHeaders = map[string]bool{
	"Content-Type":   true,
	"X-Request-ID":   true,
	"X-RateLimit-*":  true,
	"Retry-After":    true,
	"Cache-Control":  true,
	"ETag":           true,
	"Date":           true,
}

func isAllowedHeader(key string) bool {
	if hopByHopHeaders[key] {
		return false
	}
	if allowedDownstreamHeaders[key] {
		return true
	}
	return false
}

func copyAllowedHeaders(dst http.ResponseWriter, src http.Header) {
	for k, vv := range src {
		if isAllowedHeader(k) {
			for _, v := range vv {
				dst.Header().Add(k, v)
			}
		}
	}
}
