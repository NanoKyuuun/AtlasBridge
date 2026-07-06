package analyzer

import (
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
	Tools   interface{} `json:"tools"`
}

type ChatRequest struct {
	Model    string          `json:"model"`
	Stream   *bool           `json:"stream"`
	Messages []Message       `json:"messages"`
	Tools    json.RawMessage `json:"tools"`
}

type Analysis struct {
	RequestedModel     string   `json:"requested_model"`
	Stream             bool     `json:"stream"`
	MessageCount       int      `json:"message_count"`
	ApproxPromptLength int      `json:"approx_prompt_length"`
	PromptLengthBucket string   `json:"prompt_length_bucket"`
	HasCodeBlock       bool     `json:"has_code_block"`
	HasErrorPattern    bool     `json:"has_error_pattern"`
	HasToolDefinitions bool     `json:"has_tool_definitions"`
	FileExtensions     []string `json:"file_extensions"`
	DetectedKeywords   []string `json:"detected_keywords"`
	DomainSignals      []string `json:"domain_signals"`
	ComplexitySignal   string   `json:"complexity_signal"`
	LongContextSignal  bool     `json:"long_context_signal"`
}

var codeBlockPatterns = regexp.MustCompile("(?s)(```[\\s\\S]*?```|`[^`]+`|\\bfunc\\s+\\w+\\b|\\bpackage\\s+\\w+\\b|\\bimport\\s+[\"(]|\\bdef\\s+\\w+\\b|\\bclass\\s+\\w+\\b|\\bfunction\\s+\\w+\\b|\\bconst\\s+\\w+\\b|\\blet\\s+\\w+\\b|\\bvar\\s+\\w+\\b)")

var errorPatterns = regexp.MustCompile("(?i)(error|exception|panic|traceback|stack\\s*trace|failed|fatal|errno|segfault|nil\\s*pointer|index\\s*out\\s*of\\s*range|type\\s*mismatch|syntax\\s*error|undefined\\s*(variable|function)|cannot\\s*(read|find|open)|no\\s*such\\s*file|permission\\s*denied|connection\\s*refused|timeout|deadlock)")

var fileExtensions = regexp.MustCompile(`\b[\w.-]+\.(go|ts|tsx|js|jsx|vue|py|java|rb|rs|cpp|c|cs|php|swift|kt|sql|yaml|yml|json|toml|md|sh|bash|ps1|dockerfile|tf|hcl)\b`)

var keywordMap = map[string]string{
	"debug":         "debugging",
	"fix":           "debugging",
	"error":         "debugging",
	"bug":           "debugging",
	"crash":         "debugging",
	"issue":         "debugging",
	"refactor":      "refactoring",
	"restructure":   "refactoring",
	"clean":         "refactoring",
	"rewrite":       "refactoring",
	"optimize":      "refactoring",
	"test":          "testing",
	"tests":         "testing",
	"unittest":      "testing",
	"spec":          "testing",
	"coverage":      "testing",
	"mock":          "testing",
	"docs":          "documentation",
	"readme":        "documentation",
	"document":      "documentation",
	"comment":       "documentation",
	"docstring":     "documentation",
	"architecture":  "architecture",
	"design":        "design",
	"ui":            "design",
	"ux":            "design",
	"layout":        "design",
	"component":     "design",
	"style":         "design",
	"backend":       "backend",
	"api":           "backend",
	"database":      "backend",
	"server":        "backend",
	"endpoint":      "backend",
	"sql":           "backend",
	"query":         "backend",
	"frontend":      "frontend",
	"client":        "frontend",
	"browser":       "frontend",
	"dom":           "frontend",
	"render":        "frontend",
	"security":      "security",
	"vulnerability": "security",
	"auth":          "security",
	"permission":    "security",
	"encrypt":       "security",
}

var domainKeywords = map[string]string{
	"vue":      "vue",
	"react":    "react",
	"angular":  "angular",
	"next":     "nextjs",
	"nuxt":     "nuxt",
	"laravel":  "laravel",
	"express":  "express",
	"gin":      "gin",
	"echo":     "echo",
	"fiber":    "fiber",
	"nestjs":   "nestjs",
	"django":   "django",
	"flask":    "flask",
	"rails":    "rails",
	"spring":   "spring",
	"fastapi":  "fastapi",
	"svelte":   "svelte",
	"tailwind": "tailwind",
	"daisyui":  "daisyui",
}

func Analyze(body io.Reader) (*Analysis, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var req ChatRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return &Analysis{}, nil
	}

	a := &Analysis{
		RequestedModel: req.Model,
		Stream:         req.Stream != nil && *req.Stream,
		MessageCount:   len(req.Messages),
	}

	allContent := extractAllContent(req.Messages)

	a.ApproxPromptLength = len(allContent)
	a.PromptLengthBucket = bucketLength(a.ApproxPromptLength)
	a.HasCodeBlock = detectCodeBlock(allContent)
	a.HasErrorPattern = detectErrorPattern(allContent)
	a.HasToolDefinitions = len(req.Tools) > 2 && string(req.Tools) != "null"
	a.FileExtensions = detectFileExtensions(allContent)
	a.DetectedKeywords = detectKeywords(allContent)
	a.DomainSignals = detectDomainSignals(allContent)
	a.ComplexitySignal = classifyComplexity(a)
	a.LongContextSignal = a.ApproxPromptLength > 32000

	return a, nil
}

func extractAllContent(messages []Message) string {
	var parts []string
	for _, m := range messages {
		switch c := m.Content.(type) {
		case string:
			parts = append(parts, c)
		case []interface{}:
			for _, item := range c {
				if obj, ok := item.(map[string]interface{}); ok {
					if text, ok := obj["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
		}
	}
	return strings.Join(parts, " ")
}

func bucketLength(length int) string {
	switch {
	case length < 500:
		return "short"
	case length < 4000:
		return "medium"
	case length < 16000:
		return "long"
	default:
		return "very_long"
	}
}

func detectCodeBlock(content string) bool {
	return codeBlockPatterns.MatchString(content)
}

func detectErrorPattern(content string) bool {
	return errorPatterns.MatchString(content)
}

func detectFileExtensions(content string) []string {
	matches := fileExtensions.FindAllString(content, -1)
	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		ext := strings.ToLower(m)
		if !seen[ext] {
			seen[ext] = true
			result = append(result, ext)
		}
	}
	return result
}

func detectKeywords(content string) []string {
	lower := strings.ToLower(content)
	seen := make(map[string]bool)
	var result []string
	for keyword, category := range keywordMap {
		if strings.Contains(lower, keyword) {
			if !seen[category] {
				seen[category] = true
				result = append(result, category)
			}
		}
	}
	return result
}

func detectDomainSignals(content string) []string {
	lower := strings.ToLower(content)
	seen := make(map[string]bool)
	var result []string
	for keyword, domain := range domainKeywords {
		if strings.Contains(lower, keyword) {
			if !seen[domain] {
				seen[domain] = true
				result = append(result, domain)
			}
		}
	}
	return result
}

func classifyComplexity(a *Analysis) string {
	score := 0

	if a.MessageCount > 10 {
		score += 2
	} else if a.MessageCount > 5 {
		score += 1
	}

	if a.ApproxPromptLength > 16000 {
		score += 2
	} else if a.ApproxPromptLength > 4000 {
		score += 1
	}

	if a.HasToolDefinitions {
		score += 1
	}

	if a.HasCodeBlock {
		score += 1
	}

	if len(a.FileExtensions) > 2 {
		score += 1
	}

	switch {
	case score >= 3:
		return "high"
	case score >= 2:
		return "medium"
	default:
		return "low"
	}
}
