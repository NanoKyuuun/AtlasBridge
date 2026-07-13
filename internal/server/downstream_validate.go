package server

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/atlasbridge/atlasbridge/internal/netutil"
)

// ValidateDownstreamURL validates a downstream URL for SSRF protection.
// Deprecated: Use netutil.ValidateDownstreamURL directly.
func ValidateDownstreamURL(rawURL string) error {
	return netutil.ValidateDownstreamURL(rawURL)
}

func ValidateDownstreamURLRemote(rawURL string) error {
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

func isAllowedIP(ip net.IP) bool {
	return netutil.IsAllowedIP(ip)
}

func SafeRedirectPolicy(originalHost string) func(req *http.Request, via []*http.Request) error {
	return netutil.SafeRedirectPolicy()
}

func ResolveAndValidateIP(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot parse URL: %v", err)
	}
	host := parsed.Hostname()
	if host == "" {
		return nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("DNS lookup failed for %s: %w", host, err)
	}
	for _, ip := range ips {
		if !netutil.IsAllowedIP(ip) {
			return fmt.Errorf("resolved address %s for %s is in a blocked range", ip, host)
		}
	}
	return nil
}

func isLoopbackURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "[::1]" {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}

func formatHostPort(host string, port int) string {
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}
