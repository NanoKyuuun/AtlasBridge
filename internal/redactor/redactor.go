package redactor

import (
	"regexp"
	"strings"
)

type Redactor interface {
	RedactLog(fields map[string]any) map[string]any
	RedactError(err error) string
}

type DefaultRedactor struct {
	sensitiveKeys map[string]bool
	keyPattern    *regexp.Regexp
	valuePatterns []*regexp.Regexp
}

func NewDefault() *DefaultRedactor {
	return &DefaultRedactor{
		sensitiveKeys: map[string]bool{
			"authorization": true,
			"cookie":        true,
			"set-cookie":    true,
			"x-api-key":     true,
			"api_key":       true,
			"apikey":        true,
			"token":         true,
			"secret":        true,
			"password":      true,
			"prompt":        true,
			"raw_prompt":    true,
			"body":          true,
			"request_body":  true,
		},
		keyPattern: regexp.MustCompile(`(?i)(key|token|secret|password|auth|credential|cookie)`),
		valuePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(sk-[a-zA-Z0-9]{20,})`),
			regexp.MustCompile(`(?i)(ghp_[a-zA-Z0-9]{36})`),
			regexp.MustCompile(`(?i)(Bearer\s+[a-zA-Z0-9\-._~+/]+=*)`),
		},
	}
}

func (r *DefaultRedactor) RedactLog(fields map[string]any) map[string]any {
	result := make(map[string]any, len(fields))
	for k, v := range fields {
		lk := strings.ToLower(k)
		if r.sensitiveKeys[lk] || r.keyPattern.MatchString(k) {
			result[k] = "[REDACTED]"
			continue
		}
		if sv, ok := v.(string); ok {
			result[k] = r.redactStringValue(sv)
		} else {
			result[k] = v
		}
	}
	return result
}

func (r *DefaultRedactor) RedactError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	return r.redactStringValue(msg)
}

func (r *DefaultRedactor) redactStringValue(s string) string {
	result := s
	for _, p := range r.valuePatterns {
		result = p.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}

type NopRedactor struct{}

func (NopRedactor) RedactLog(fields map[string]any) map[string]any { return fields }
func (NopRedactor) RedactError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
