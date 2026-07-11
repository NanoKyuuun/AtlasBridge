package server

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var defaultAllowedSchemes = map[string]bool{"http": true, "https": true}

func ValidateDownstreamURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot parse URL: %v", err)
	}
	if !defaultAllowedSchemes[parsed.Scheme] {
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
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	ip := net.ParseIP(host)
	if ip != nil && !isAllowedIP(ip) {
		return fmt.Errorf("resolved address %s is in a blocked range (link-local/private/metadata)", ip)
	}
	return nil
}

func ValidateDownstreamURLRemote(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot parse URL: %v", err)
	}
	if !defaultAllowedSchemes[parsed.Scheme] {
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
	if ip.IsLoopback() {
		return true
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	if ip.IsMulticast() {
		return false
	}
	privateRanges := []struct {
		start net.IP
		end   net.IP
	}{
		{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
		{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
		{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
		{net.ParseIP("169.254.0.0"), net.ParseIP("169.254.255.255")},
		{net.ParseIP("100.64.0.0"), net.ParseIP("100.127.255.255")},
	}
	for _, r := range privateRanges {
		if bytesCompare(ip.To4(), r.start.To4()) >= 0 && bytesCompare(ip.To4(), r.end.To4()) <= 0 {
			return false
		}
	}
	metadataIPv4 := net.ParseIP("169.254.169.254")
	if ip.Equal(metadataIPv4) {
		return false
	}
	return true
}

func bytesCompare(a, b net.IP) int {
	if a == nil || b == nil {
		return 0
	}
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return len(a) - len(b)
}

func SafeRedirectPolicy(originalHost string) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("stopped after 5 redirects")
		}
		redirectURL := req.URL.String()
		if err := ValidateDownstreamURL(redirectURL); err != nil {
			return fmt.Errorf("redirect blocked: %v", err)
		}
		return nil
	}
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
		if !isAllowedIP(ip) {
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
