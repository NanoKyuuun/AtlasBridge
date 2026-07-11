package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestValidateDownstreamURL_Valid(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"http localhost", "http://127.0.0.1:20128/v1"},
		{"https localhost", "https://127.0.0.1:20128/v1"},
		{"http hostname", "http://localhost:20128/v1"},
		{"http remote", "http://example.com:20128/v1"},
		{"https remote", "https://api.example.com/v1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateDownstreamURL(tt.url); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
	}
}

func TestValidateDownstreamURL_RejectsCredentials(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"basic auth", "http://user:pass@127.0.0.1:20128/v1"},
		{"user only", "http://user@127.0.0.1:20128/v1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDownstreamURL(tt.url)
			if err == nil {
				t.Error("expected error for credentials in URL")
			}
		})
	}
}

func TestValidateDownstreamURL_RejectsFragment(t *testing.T) {
	err := ValidateDownstreamURL("http://127.0.0.1:20128/v1#fragment")
	if err == nil {
		t.Error("expected error for URL fragment")
	}
}

func TestValidateDownstreamURL_RejectsQuery(t *testing.T) {
	err := ValidateDownstreamURL("http://127.0.0.1:20128/v1?key=value")
	if err == nil {
		t.Error("expected error for URL query")
	}
}

func TestValidateDownstreamURL_RejectsBadScheme(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"ftp", "ftp://127.0.0.1:20128/v1"},
		{"file", "file:///etc/passwd"},
		{"javascript", "javascript:alert(1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDownstreamURL(tt.url)
			if err == nil {
				t.Error("expected error for bad scheme")
			}
		})
	}
}

func TestValidateDownstreamURL_RejectsEmptyHost(t *testing.T) {
	err := ValidateDownstreamURL("http:///v1")
	if err == nil {
		t.Error("expected error for empty host")
	}
}

func TestValidateDownstreamURL_RemoteMode(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid remote", "https://api.openai.com/v1", false},
		{"credentials", "http://user:pass@api.com/v1", true},
		{"fragment", "http://api.com/v1#frag", true},
		{"query", "http://api.com/v1?q=1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDownstreamURLRemote(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDownstreamURLRemote(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestSafeRedirectPolicy_BlocksTooMany(t *testing.T) {
	allowedHost := "127.0.0.1"
	policy := SafeRedirectPolicy(allowedHost)

	via := make([]*http.Request, 6)
	for i := range via {
		via[i] = &http.Request{URL: &url.URL{Host: "127.0.0.1"}}
	}

	req := &http.Request{URL: &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/next"}}
	err := policy(req, via)
	if err == nil {
		t.Error("expected error for too many redirects")
	}
}

func TestSafeRedirectPolicy_AllowsValid(t *testing.T) {
	policy := SafeRedirectPolicy("127.0.0.1")
	req := &http.Request{URL: &url.URL{Scheme: "http", Host: "127.0.0.1:20128", Path: "/next"}}
	err := policy(req, []*http.Request{})
	if err != nil {
		t.Errorf("expected valid redirect, got: %v", err)
	}
}

func TestSafeRedirectPolicy_BlocksCredentials(t *testing.T) {
	policy := SafeRedirectPolicy("127.0.0.1")
	req := &http.Request{URL: &url.URL{Scheme: "http", Host: "127.0.0.1:20128", Path: "/next", User: url.User("admin")}}
	via := []*http.Request{{URL: &url.URL{}}}
	err := policy(req, via)
	if err == nil {
		t.Error("expected error for redirect with credentials")
	}
}

func TestIsLoopbackURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"http://127.0.0.1:20128/v1", true},
		{"http://localhost:20128/v1", true},
		{"http://[::1]:20128/v1", true},
		{"http://example.com:20128/v1", false},
		{"https://10.0.0.1:20128/v1", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := isLoopbackURL(tt.url); got != tt.want {
				t.Errorf("isLoopbackURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestDownstreamValidator_Integration(t *testing.T) {
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer downstream.Close()

	ts := newTestServerWithDownstream(t, downstream.URL+"/v1")
	defer ts.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json",
		strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRejectsURLWithUserInfo(t *testing.T) {
	err := ValidateDownstreamURL("http://admin:secret@127.0.0.1:20128/v1")
	if err == nil {
		t.Error("expected rejection of URL with userinfo")
	}
}

func TestRejectsURLWithFragment(t *testing.T) {
	err := ValidateDownstreamURL("http://127.0.0.1:20128/v1#section")
	if err == nil {
		t.Error("expected rejection of URL with fragment")
	}
}

func TestRejectsURLWithQuery(t *testing.T) {
	err := ValidateDownstreamURL("http://127.0.0.1:20128/v1?debug=true")
	if err == nil {
		t.Error("expected rejection of URL with query params")
	}
}
