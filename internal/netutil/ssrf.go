package netutil

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
)

var allowedSchemes = map[string]bool{"http": true, "https": true}

// ValidateDownstreamURL validates a downstream URL for SSRF protection.
// It blocks private/link-local/metadata IPs, credentials, fragments, and query strings.
func ValidateDownstreamURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot parse URL: %v", err)
	}
	if !allowedSchemes[parsed.Scheme] {
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
	// Check IP literal first
	ip := net.ParseIP(host)
	if ip != nil && !IsAllowedIP(ip) {
		return fmt.Errorf("resolved address %s is in a blocked range (link-local/private/metadata)", ip)
	}
	// DNS-aware: resolve hostname and check all resolved IPs
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil {
			return fmt.Errorf("DNS lookup failed for %s: %v", host, err)
		}
		for _, resolved := range ips {
			if !IsAllowedIP(resolved) {
				return fmt.Errorf("resolved address %s for %s is in a blocked range (link-local/private/metadata)", resolved, host)
			}
		}
	}
	return nil
}

// IsAllowedIP checks whether an IP address is safe for outbound connections.
// It blocks link-local, multicast, RFC 1918 private ranges, and RFC 4193
// IPv6 unique local addresses (fc00::/7).
func IsAllowedIP(ip net.IP) bool {
	// Loopback is allowed — the proxy legitimately connects to local services.
	// SSRF protection targets private/link-local ranges, not localhost.
	if ip.IsLoopback() {
		return true
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	if ip.IsMulticast() {
		return false
	}
	// Block RFC 4193 IPv6 unique local addresses (fc00::/7)
	if ip.To4() == nil && len(ip) == net.IPv6len {
		if ip[0] == 0xfc || ip[0] == 0xfd {
			return false
		}
	}
	// Check IPv4 private ranges only for IPv4 addresses
	if ipv4 := ip.To4(); ipv4 != nil {
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
			if bytesCompare(ipv4, r.start.To4()) >= 0 && bytesCompare(ipv4, r.end.To4()) <= 0 {
				return false
			}
		}
		metadataIPv4 := net.ParseIP("169.254.169.254")
		if ip.Equal(metadataIPv4) {
			return false
		}
	}
	return true
}

// SafeRedirectPolicy returns a redirect policy that validates each redirect
// target against SSRF rules (including DNS resolution) and caps redirects at 5.
func SafeRedirectPolicy() func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("stopped after 5 redirects")
		}
		if err := ValidateDownstreamURL(req.URL.String()); err != nil {
			return fmt.Errorf("redirect blocked: %v", err)
		}
		// DNS-aware: resolve hostname and check all IPs
		host := req.URL.Hostname()
		if host != "" {
			ips, err := net.LookupIP(host)
			if err != nil {
				return fmt.Errorf("redirect blocked: DNS lookup failed for %s: %v", host, err)
			}
			for _, ip := range ips {
				if !IsAllowedIP(ip) {
					return fmt.Errorf("redirect blocked: resolved address %s for %s is in a blocked range", ip, host)
				}
			}
		}
		return nil
	}
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
