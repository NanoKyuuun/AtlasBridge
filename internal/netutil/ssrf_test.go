package netutil

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// --- IsAllowedIP tests ---

func TestIsAllowedIP_AllowsLoopback(t *testing.T) {
	ips := []string{"127.0.0.1", "127.0.0.2", "::1"}
	for _, raw := range ips {
		ip := net.ParseIP(raw)
		if ip == nil {
			t.Fatalf("failed to parse %q", raw)
		}
		if !IsAllowedIP(ip) {
			t.Errorf("expected loopback %s to be allowed", raw)
		}
	}
}

func TestIsAllowedIP_AllowsPublic(t *testing.T) {
	ips := []string{"8.8.8.8", "1.1.1.1", "203.0.113.50", "93.184.216.34"}
	for _, raw := range ips {
		ip := net.ParseIP(raw)
		if ip == nil {
			t.Fatalf("failed to parse %q", raw)
		}
		if !IsAllowedIP(ip) {
			t.Errorf("expected public IP %s to be allowed", raw)
		}
	}
}

func TestIsAllowedIP_BlocksRFC1918(t *testing.T) {
	ips := []string{
		"10.0.0.1", "10.255.255.255", "10.1.2.3",
		"172.16.0.1", "172.31.255.255", "172.20.10.5",
		"192.168.0.1", "192.168.1.100", "192.168.255.255",
	}
	for _, raw := range ips {
		ip := net.ParseIP(raw)
		if ip == nil {
			t.Fatalf("failed to parse %q", raw)
		}
		if IsAllowedIP(ip) {
			t.Errorf("expected RFC 1918 IP %s to be blocked", raw)
		}
	}
}

func TestIsAllowedIP_BlocksLinkLocal(t *testing.T) {
	ips := []string{"169.254.0.1", "169.254.10.5", "169.254.255.255"}
	for _, raw := range ips {
		ip := net.ParseIP(raw)
		if ip == nil {
			t.Fatalf("failed to parse %q", raw)
		}
		if IsAllowedIP(ip) {
			t.Errorf("expected link-local IP %s to be blocked", raw)
		}
	}
}

func TestIsAllowedIP_BlocksMetadataEndpoint(t *testing.T) {
	ip := net.ParseIP("169.254.169.254")
	if ip == nil {
		t.Fatal("failed to parse metadata IP")
	}
	if IsAllowedIP(ip) {
		t.Error("expected cloud metadata IP 169.254.169.254 to be blocked")
	}
}

func TestIsAllowedIP_BlocksSharedRange(t *testing.T) {
	ips := []string{"100.64.0.1", "100.127.255.255", "100.100.100.100"}
	for _, raw := range ips {
		ip := net.ParseIP(raw)
		if ip == nil {
			t.Fatalf("failed to parse %q", raw)
		}
		if IsAllowedIP(ip) {
			t.Errorf("expected shared range IP %s to be blocked", raw)
		}
	}
}

func TestIsAllowedIP_BlocksMulticast(t *testing.T) {
	ip := net.ParseIP("224.0.0.1")
	if ip == nil {
		t.Fatal("failed to parse multicast IP")
	}
	if IsAllowedIP(ip) {
		t.Error("expected multicast IP 224.0.0.1 to be blocked")
	}
}

// --- ValidateDownstreamURL IP blocking tests ---

func TestValidateDownstreamURL_BlocksPrivateIP(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"RFC 1918 10.x", "http://10.0.0.1:8080/v1"},
		{"RFC 1918 172.16.x", "http://172.16.0.1:8080/v1"},
		{"RFC 1918 192.168.x", "http://192.168.1.1:8080/v1"},
		{"link-local", "http://169.254.0.1:8080/v1"},
		{"shared range", "http://100.64.0.1:8080/v1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDownstreamURL(tt.url)
			if err == nil {
				t.Errorf("expected private IP URL %q to be blocked", tt.url)
			}
		})
	}
}

func TestValidateDownstreamURL_BlocksMetadataIP(t *testing.T) {
	err := ValidateDownstreamURL("http://169.254.169.254/metadata")
	if err == nil {
		t.Error("expected cloud metadata URL to be blocked")
	}
}

func TestValidateDownstreamURL_AllowsLoopback(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"ipv4 loopback", "http://127.0.0.1:20128/v1"},
		{"ipv6 loopback", "http://[::1]:20128/v1"},
		{"localhost hostname", "http://localhost:20128/v1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDownstreamURL(tt.url)
			if err != nil {
				t.Errorf("expected loopback URL %q to be allowed, got error: %v", tt.url, err)
			}
		})
	}
}

func TestValidateDownstreamURL_AllowsPublicURL(t *testing.T) {
	err := ValidateDownstreamURL("https://api.example.com/v1")
	if err != nil {
		t.Errorf("expected public URL to be allowed, got error: %v", err)
	}
}

// --- SafeRedirectPolicy tests ---

func TestSafeRedirectPolicy_CapsRedirects(t *testing.T) {
	policy := SafeRedirectPolicy()

	via := make([]*http.Request, 5)
	for i := range via {
		via[i] = &http.Request{URL: &url.URL{Host: "example.com"}}
	}

	req := &http.Request{URL: &url.URL{Scheme: "http", Host: "example.com", Path: "/next"}}
	err := policy(req, via)
	if err == nil {
		t.Error("expected error for more than 5 redirects")
	}
}

func TestSafeRedirectPolicy_AllowsValidRedirect(t *testing.T) {
	policy := SafeRedirectPolicy()
	req := &http.Request{URL: &url.URL{Scheme: "http", Host: "example.com:8080", Path: "/next"}}
	err := policy(req, []*http.Request{})
	if err != nil {
		t.Errorf("expected valid redirect, got error: %v", err)
	}
}

func TestSafeRedirectPolicy_BlocksRedirectToPrivateIP(t *testing.T) {
	policy := SafeRedirectPolicy()
	req := &http.Request{URL: &url.URL{Scheme: "http", Host: "10.0.0.1:8080", Path: "/secret"}}
	err := policy(req, []*http.Request{})
	if err == nil {
		t.Error("expected redirect to private IP to be blocked")
	}
}

func TestSafeRedirectPolicy_BlocksRedirectToMetadata(t *testing.T) {
	policy := SafeRedirectPolicy()
	req := &http.Request{URL: &url.URL{Scheme: "http", Host: "169.254.169.254", Path: "/latest/meta-data"}}
	err := policy(req, []*http.Request{})
	if err == nil {
		t.Error("expected redirect to metadata endpoint to be blocked")
	}
}

func TestSafeRedirectPolicy_BlocksRedirectWithCredentials(t *testing.T) {
	policy := SafeRedirectPolicy()
	req := &http.Request{URL: &url.URL{Scheme: "http", Host: "example.com", Path: "/next", User: url.User("admin")}}
	err := policy(req, []*http.Request{})
	if err == nil {
		t.Error("expected redirect with credentials to be blocked")
	}
}

// --- Integration: forwarder rejects private IP at construction ---

func TestForwarderRejectsPrivateIP(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"private 10.x", "http://10.0.0.1:8080/v1", true},
		{"private 192.168.x", "http://192.168.1.1:8080/v1", true},
		{"metadata", "http://169.254.169.254/metadata", true},
		{"loopback ok", "http://127.0.0.1:20128/v1", false},
		{"public ok", "https://api.example.com/v1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDownstreamURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDownstreamURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

// --- Integration: live redirect server test ---

func TestSafeRedirectPolicy_IntegrationWithLiveServer(t *testing.T) {
	// A server that redirects to a private IP should be blocked by the redirect policy
	privateRedirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://10.0.0.1:9999/secret", http.StatusFound)
	}))
	defer privateRedirectServer.Close()

	client := &http.Client{
		CheckRedirect: SafeRedirectPolicy(),
	}

	resp, err := client.Get(privateRedirectServer.URL + "/start")
	if resp != nil {
		resp.Body.Close()
	}

	// SafeRedirectPolicy returns an error which client.Get propagates
	if err == nil {
		t.Error("expected redirect to private IP to be blocked, but got no error")
	}
}
