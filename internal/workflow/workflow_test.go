package workflow

import (
	"os"
	"testing"
	"time"

	"github.com/cjp2600/stepwise/internal/config"
	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/validation"
)

func TestLoadWorkflow(t *testing.T) {
	// Create a temporary workflow file
	content := `name: "Test Workflow"
version: "1.0"
description: "A test workflow"

variables:
  base_url: "https://api.example.com"

steps:
  - name: "Test Step"
    request:
      method: "GET"
      url: "{{base_url}}/test"
    validate:
      - status: 200
`

	tmpfile, err := os.CreateTemp("", "workflow-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Test loading the workflow
	wf, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	// Verify workflow properties
	if wf.Name != "Test Workflow" {
		t.Errorf("Expected name 'Test Workflow', got '%s'", wf.Name)
	}

	if wf.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", wf.Version)
	}

	if len(wf.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(wf.Steps))
	}

	if wf.Steps[0].Name != "Test Step" {
		t.Errorf("Expected step name 'Test Step', got '%s'", wf.Steps[0].Name)
	}
}

func TestNewExecutor(t *testing.T) {
	cfg := &config.Config{
		Timeout: 30 * time.Second,
	}
	log := logger.New()

	executor := NewExecutor(cfg, log)

	if executor.config != cfg {
		t.Error("Executor config not set correctly")
	}

	if executor.logger != log {
		t.Error("Executor logger not set correctly")
	}

	// HTTP client timeout is set internally, not exposed as a field
	// The timeout is configured when creating the client
}

func TestExecuteWorkflow(t *testing.T) {
	cfg := &config.Config{
		Timeout: 30 * time.Second,
	}
	log := logger.New()

	executor := NewExecutor(cfg, log)

	wf := &Workflow{
		Name:        "Test Workflow",
		Version:     "1.0",
		Description: "A test workflow",
		Variables:   make(map[string]interface{}),
		Steps: []Step{
			{
				Name: "Test Step 1",
				Request: Request{
					Method: "GET",
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
				},
				Validate: []validation.ValidationRule{
					{Status: 200},
				},
			},
			{
				Name: "Test Step 2",
				Request: Request{
					Method: "GET",
					URL:    "https://jsonplaceholder.typicode.com/users/1",
				},
				Validate: []validation.ValidationRule{
					{Status: 200},
				},
			},
		},
	}

	results, err := executor.Execute(wf)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Status != "passed" {
			t.Errorf("Step %d expected status 'passed', got '%s'", i+1, result.Status)
		}

		if result.Duration == 0 {
			t.Errorf("Step %d duration should not be zero", i+1)
		}
	}
}

func TestStepWait(t *testing.T) {
	cfg := &config.Config{
		Timeout: 5 * time.Second,
	}
	log := logger.New()
	executor := NewExecutor(cfg, log)

	wf := &Workflow{
		Name:        "Test Wait Step",
		Version:     "1.0",
		Description: "Workflow with wait step",
		Variables:   make(map[string]interface{}),
		Steps: []Step{
			{
				Name: "Wait Step",
				Wait: "1s",
				Request: Request{
					Method: "GET",
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
				},
				Validate: []validation.ValidationRule{
					{Status: 200},
				},
			},
		},
	}

	start := time.Now()
	results, err := executor.Execute(wf)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Status != "passed" {
		t.Errorf("Expected status 'passed', got '%s'", results[0].Status)
	}
	if elapsed < time.Second {
		t.Errorf("Expected at least 1s elapsed, got %v", elapsed)
	}
}

func TestLoadWorkflowFileNotFound(t *testing.T) {
	_, err := Load("nonexistent.yml")
	if err == nil {
		t.Error("Expected error when loading nonexistent file")
	}
}

func TestLoadWorkflowInvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	content := `name: "Test Workflow"
version: "1.0"
steps:
  - name: "Test Step"
    request:
      method: "GET"
      url: "{{base_url}}/test"
      invalid: [yaml: syntax

`

	tmpfile, err := os.CreateTemp("", "workflow-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = Load(tmpfile.Name())
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}
