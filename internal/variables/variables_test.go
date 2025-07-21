package variables

import (
	"strings"
	"testing"

	"github.com/cjp2600/stepwise/internal/logger"
)

func TestNewManager(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	if manager.logger != log {
		t.Error("Logger not set correctly")
	}

	if len(manager.variables) != 0 {
		t.Error("Variables map should be empty initially")
	}
}

func TestSetAndGet(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Test setting and getting variables
	manager.Set("test_key", "test_value")
	manager.Set("number_key", 123)
	manager.Set("bool_key", true)

	// Test string variable
	if value, exists := manager.Get("test_key"); !exists {
		t.Error("Variable should exist")
	} else if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}

	// Test number variable
	if value, exists := manager.Get("number_key"); !exists {
		t.Error("Variable should exist")
	} else if value != 123 {
		t.Errorf("Expected 123, got %v", value)
	}

	// Test boolean variable
	if value, exists := manager.Get("bool_key"); !exists {
		t.Error("Variable should exist")
	} else if value != true {
		t.Errorf("Expected true, got %v", value)
	}

	// Test non-existent variable
	if _, exists := manager.Get("non_existent"); exists {
		t.Error("Non-existent variable should not exist")
	}
}

func TestSubstituteVariables(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Set up variables
	manager.Set("base_url", "https://api.example.com")
	manager.Set("user_id", 123)
	manager.Set("api_key", "secret-key")

	// Test simple variable substitution
	input := "{{base_url}}/users/{{user_id}}"
	expected := "https://api.example.com/users/123"
	result, err := manager.Substitute(input)
	if err != nil {
		t.Errorf("Failed to substitute variables: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test with non-existent variable
	input = "{{base_url}}/users/{{non_existent}}"
	result, err = manager.Substitute(input)
	if err != nil {
		t.Errorf("Failed to substitute variables: %v", err)
	}
	// Should substitute known variables but keep unknown ones
	expected = "https://api.example.com/users/{{non_existent}}"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestSubstituteFakerFunctions(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Test faker functions
	tests := []struct {
		input    string
		expected string
	}{
		{"{{faker.name}}", "John Doe"},              // Will be one of the names
		{"{{faker.email}}", "john.doe@example.com"}, // Will be generated email
		{"{{faker.uuid}}", "test-uuid"},             // Will be generated UUID
		{"{{faker.number(1, 10)}}", "5"},            // Will be random number
	}

	for _, test := range tests {
		result, err := manager.Substitute(test.input)
		if err != nil {
			t.Errorf("Failed to substitute faker function %s: %v", test.input, err)
		}
		if result == test.input {
			t.Errorf("Faker function %s was not substituted", test.input)
		}
	}
}

func TestSubstituteEnvironmentVariables(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Set up environment-like variables
	manager.Set("API_KEY", "secret-key")
	manager.Set("BASE_URL", "https://api.example.com")

	// Test environment variable substitution
	input := "Authorization: Bearer {{env.API_KEY}}"
	expected := "Authorization: Bearer secret-key"
	result, err := manager.Substitute(input)
	if err != nil {
		t.Errorf("Failed to substitute environment variable: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestSubstituteMap(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Set up variables
	manager.Set("user_id", 123)
	manager.Set("api_key", "secret-key")

	// Test map substitution
	input := map[string]interface{}{
		"url": "{{base_url}}/users/{{user_id}}",
		"headers": map[string]interface{}{
			"Authorization": "Bearer {{api_key}}",
		},
		"body": map[string]interface{}{
			"user_id": "{{user_id}}",
			"name":    "{{faker.name}}",
		},
	}

	// Set base_url for substitution
	manager.Set("base_url", "https://api.example.com")

	result, err := manager.SubstituteMap(input)
	if err != nil {
		t.Errorf("Failed to substitute map: %v", err)
	}

	// Check that substitution happened
	if url, ok := result["url"].(string); ok {
		if url == "{{base_url}}/users/{{user_id}}" {
			t.Error("URL was not substituted")
		}
	}
}

func TestGenerateFakerData(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Test name generation
	name := manager.generateName()
	if name == "" {
		t.Error("Generated name should not be empty")
	}

	// Test email generation
	email := manager.generateEmail()
	if email == "" {
		t.Error("Generated email should not be empty")
	}
	if !strings.Contains(email, "@") {
		t.Error("Generated email should contain @")
	}

	// Test UUID generation
	uuid := manager.generateUUID()
	if uuid == "" {
		t.Error("Generated UUID should not be empty")
	}
	if !strings.Contains(uuid, "-") {
		t.Error("Generated UUID should contain hyphens")
	}

	// Test number generation
	number := manager.generateNumber([]string{"1", "10"})
	if number == "" {
		t.Error("Generated number should not be empty")
	}

	// Test date generation
	date := manager.generateDate()
	if date == "" {
		t.Error("Generated date should not be empty")
	}
}
