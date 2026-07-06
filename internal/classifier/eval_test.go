package classifier

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/smart-ai-proxy/smart-ai-proxy/internal/analyzer"
)

type EvalExample struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Prompt           string `json:"prompt"`
	ExpectedTaskType string `json:"expected_task_type"`
	ExpectedRoute    string `json:"expected_route"`
}

type EvalCategory struct {
	TaskType      string        `json:"task_type"`
	ExpectedRoute string        `json:"expected_route"`
	Examples      []EvalExample `json:"examples"`
}

type EvalDataset struct {
	Version    string         `json:"version"`
	Categories []EvalCategory `json:"categories"`
}

func loadEvalDataset(t *testing.T) *EvalDataset {
	t.Helper()

	data, err := os.ReadFile("../../testdata/classification/eval_dataset.json")
	if err != nil {
		t.Fatalf("failed to load eval dataset: %v", err)
	}

	var dataset EvalDataset
	if err := json.Unmarshal(data, &dataset); err != nil {
		t.Fatalf("failed to parse eval dataset: %v", err)
	}

	return &dataset
}

func buildRequest(prompt string) string {
	return `{"model":"smart-auto","messages":[{"role":"user","content":"` + strings.ReplaceAll(prompt, `"`, `\"`) + `"}]}`
}

func analyzePrompt(t *testing.T, prompt string) *analyzer.Analysis {
	t.Helper()

	body := strings.NewReader(buildRequest(prompt))
	a, err := analyzer.Analyze(body)
	if err != nil {
		t.Fatalf("analysis failed for prompt: %v", err)
	}
	return a
}

func TestEvalDatasetExists(t *testing.T) {
	dataset := loadEvalDataset(t)

	if len(dataset.Categories) == 0 {
		t.Error("eval dataset has no categories")
	}

	totalExamples := 0
	for _, cat := range dataset.Categories {
		totalExamples += len(cat.Examples)
	}

	t.Logf("Eval dataset: %d categories, %d total examples", len(dataset.Categories), totalExamples)
}

func TestEvalDatasetCoversAllCategories(t *testing.T) {
	dataset := loadEvalDataset(t)

	requiredCategories := []string{
		"design_task",
		"backend_engineering",
		"frontend_engineering",
		"fullstack_engineering",
		"code_generation",
		"debugging",
		"refactoring",
		"test_generation",
		"documentation",
		"code_review",
		"architecture_design",
		"security_review",
		"long_context_analysis",
		"lightweight_task",
		"general_chat",
		"unknown",
	}

	foundCategories := make(map[string]bool)
	for _, cat := range dataset.Categories {
		foundCategories[cat.TaskType] = true
	}

	for _, required := range requiredCategories {
		if !foundCategories[required] {
			t.Errorf("missing required category: %s", required)
		}
	}
}

func TestEvalDatasetExamplesHaveRequiredFields(t *testing.T) {
	dataset := loadEvalDataset(t)

	for _, cat := range dataset.Categories {
		for _, ex := range cat.Examples {
			if ex.ID == "" {
				t.Errorf("category %s: example missing ID", cat.TaskType)
			}
			if ex.Name == "" {
				t.Errorf("category %s: example %s missing Name", cat.TaskType, ex.ID)
			}
			if ex.Prompt == "" {
				t.Errorf("category %s: example %s missing Prompt", cat.TaskType, ex.ID)
			}
			if ex.ExpectedTaskType == "" {
				t.Errorf("category %s: example %s missing ExpectedTaskType", cat.TaskType, ex.ID)
			}
			if ex.ExpectedRoute == "" {
				t.Errorf("category %s: example %s missing ExpectedRoute", cat.TaskType, ex.ID)
			}
		}
	}
}

func TestEvalDatasetContainsNoSecrets(t *testing.T) {
	dataset := loadEvalDataset(t)

	secretPatterns := []string{
		"-----BEGIN",
		"sk-",
		"ghp_",
		"gho_",
		"AKIA",
	}

	for _, cat := range dataset.Categories {
		for _, ex := range cat.Examples {
			lower := strings.ToLower(ex.Prompt)
			for _, pattern := range secretPatterns {
				if strings.Contains(lower, pattern) {
					t.Errorf("category %s: example %s may contain actual secrets (pattern: %s)", cat.TaskType, ex.ID, pattern)
				}
			}
		}
	}
}

func TestEvalClassifierTaskTypeAccuracy(t *testing.T) {
	dataset := loadEvalDataset(t)

	totalExamples := 0
	correctTaskType := 0
	classificationFailures := []string{}

	for _, cat := range dataset.Categories {
		for _, ex := range cat.Examples {
			totalExamples++
			a := analyzePrompt(t, ex.Prompt)
			c := Classify(a)

			if c.TaskType == ex.ExpectedTaskType {
				correctTaskType++
			} else {
				classificationFailures = append(classificationFailures,
					ex.ID+": expected="+ex.ExpectedTaskType+" got="+c.TaskType)
			}
		}
	}

	accuracy := float64(correctTaskType) / float64(totalExamples) * 100
	t.Logf("Task type accuracy: %d/%d (%.1f%%)", correctTaskType, totalExamples, accuracy)

	for _, failure := range classificationFailures {
		t.Logf("  FAIL: %s", failure)
	}

	if accuracy < 30.0 {
		t.Errorf("task type accuracy %.1f%% is below 30%% threshold", accuracy)
	}
}

func TestEvalClassifierRoutingAccuracy(t *testing.T) {
	dataset := loadEvalDataset(t)

	totalExamples := 0
	correctRoute := 0
	routeFailures := []string{}

	for _, cat := range dataset.Categories {
		for _, ex := range cat.Examples {
			totalExamples++

			a := analyzePrompt(t, ex.Prompt)
			c := Classify(a)

			actualRoute := routeForTaskType(c.TaskType)
			if actualRoute == ex.ExpectedRoute {
				correctRoute++
			} else {
				routeFailures = append(routeFailures,
					ex.ID+": expected_route="+ex.ExpectedRoute+" got_route="+actualRoute+" (task="+c.TaskType+")")
			}
		}
	}

	accuracy := float64(correctRoute) / float64(totalExamples) * 100
	t.Logf("Route accuracy: %d/%d (%.1f%%)", correctRoute, totalExamples, accuracy)

	for _, failure := range routeFailures {
		t.Logf("  FAIL: %s", failure)
	}

	if accuracy < 30.0 {
		t.Errorf("route accuracy %.1f%% is below 30%% threshold", accuracy)
	}
}

func TestEvalClassifierConfidence(t *testing.T) {
	dataset := loadEvalDataset(t)

	totalExamples := 0
	totalConfidence := 0.0

	for _, cat := range dataset.Categories {
		for _, ex := range cat.Examples {
			totalExamples++
			a := analyzePrompt(t, ex.Prompt)
			c := Classify(a)
			totalConfidence += c.Confidence
		}
	}

	avgConfidence := totalConfidence / float64(totalExamples)
	t.Logf("Average confidence: %.3f across %d examples", avgConfidence, totalExamples)

	if avgConfidence < 0.3 {
		t.Errorf("average confidence %.3f is below 0.3 threshold", avgConfidence)
	}
}

func TestEvalNoNilAnalysis(t *testing.T) {
	dataset := loadEvalDataset(t)

	for _, cat := range dataset.Categories {
		for _, ex := range cat.Examples {
			a := analyzePrompt(t, ex.Prompt)
			if a == nil {
				t.Errorf("category %s: example %s produced nil analysis", cat.TaskType, ex.ID)
			}
			c := Classify(a)
			if c == nil {
				t.Errorf("category %s: example %s produced nil classification", cat.TaskType, ex.ID)
			}
			if c.TaskType == "" {
				t.Errorf("category %s: example %s produced empty task type", cat.TaskType, ex.ID)
			}
		}
	}
}

func routeForTaskType(taskType string) string {
	routes := map[string]string{
		"general_chat":          "route.default",
		"design_task":           "route.design",
		"backend_engineering":   "route.backend",
		"frontend_engineering":  "route.frontend",
		"fullstack_engineering": "route.fullstack",
		"code_generation":       "route.backend",
		"debugging":             "route.debugging",
		"refactoring":           "route.refactoring",
		"test_generation":       "route.testing",
		"documentation":         "route.documentation",
		"architecture_design":   "route.architect",
		"security_review":       "route.security",
		"long_context_analysis": "route.long_context",
		"lightweight_task":      "route.low_cost",
		"unknown":               "route.default",
	}

	if route, ok := routes[taskType]; ok {
		return route
	}
	return "route.default"
}
