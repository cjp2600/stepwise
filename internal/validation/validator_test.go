package validation

import (
	"testing"
	"time"

	"github.com/cjp2600/stepwise/internal/http"
	"github.com/cjp2600/stepwise/internal/logger"
)

func TestNewValidator(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	if validator.logger != log {
		t.Error("Logger not set correctly")
	}
}

func TestValidateStatus(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	response := &http.Response{
		StatusCode: 200,
		Duration:   100 * time.Millisecond,
	}

	// Test successful validation
	rule := ValidationRule{Status: 200}
	result := validator.validateStatus(response, rule.Status)

	if !result.Passed {
		t.Error("Status validation should pass for matching status")
	}

	if result.Type != "status" {
		t.Errorf("Expected type 'status', got %s", result.Type)
	}

	// Test failed validation
	rule = ValidationRule{Status: 404}
	result = validator.validateStatus(response, rule.Status)

	if result.Passed {
		t.Error("Status validation should fail for non-matching status")
	}
}

func TestValidateTime(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	response := &http.Response{
		StatusCode: 200,
		Duration:   500 * time.Millisecond,
	}

	// Test less than validation
	rule := ValidationRule{Time: "< 1000ms"}
	result := validator.validateTime(response, rule.Time)

	if !result.Passed {
		t.Error("Time validation should pass for duration less than limit")
	}

	// Test greater than validation
	rule = ValidationRule{Time: "> 100ms"}
	result = validator.validateTime(response, rule.Time)

	if !result.Passed {
		t.Error("Time validation should pass for duration greater than limit")
	}

	// Test failed validation
	rule = ValidationRule{Time: "< 100ms"}
	result = validator.validateTime(response, rule.Time)

	if result.Passed {
		t.Error("Time validation should fail for duration greater than limit")
	}
}

func TestValidateJSON(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	response := &http.Response{
		StatusCode: 200,
		Body:       []byte(`{"status":"success","data":{"id":123,"name":"test"}}`),
		Duration:   100 * time.Millisecond,
	}

	// Test JSON path validation
	rule := ValidationRule{JSON: "$.status", Equals: "success"}
	result := validator.validateJSON(response, rule)

	if !result.Passed {
		t.Error("JSON validation should pass for matching value")
	}

	// Test type validation
	rule = ValidationRule{JSON: "$.data.id", Type: "number"}
	result = validator.validateJSON(response, rule)

	if !result.Passed {
		t.Error("JSON type validation should pass for number type")
	}

	// Test contains validation
	rule = ValidationRule{JSON: "$.data.name", Contains: "test"}
	result = validator.validateJSON(response, rule)

	if !result.Passed {
		t.Error("JSON contains validation should pass for matching substring")
	}
}

func TestValidateEquals(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	// Test successful validation
	result := validator.validateEquals("test", "test")
	if !result.Passed {
		t.Error("Equals validation should pass for matching values")
	}

	// Test failed validation
	result = validator.validateEquals("test", "different")
	if result.Passed {
		t.Error("Equals validation should fail for non-matching values")
	}
}

func TestValidateContains(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	// Test successful validation
	result := validator.validateContains("hello world", "world")
	if !result.Passed {
		t.Error("Contains validation should pass for matching substring")
	}

	// Test failed validation
	result = validator.validateContains("hello world", "universe")
	if result.Passed {
		t.Error("Contains validation should fail for non-matching substring")
	}
}

func TestValidateType(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	// Test string type
	result := validator.validateType("test", "string")
	if !result.Passed {
		t.Error("Type validation should pass for string type")
	}

	// Test number type
	result = validator.validateType(123, "number")
	if !result.Passed {
		t.Error("Type validation should pass for number type")
	}

	// Test boolean type
	result = validator.validateType(true, "boolean")
	if !result.Passed {
		t.Error("Type validation should pass for boolean type")
	}

	// Test failed validation
	result = validator.validateType("test", "number")
	if result.Passed {
		t.Error("Type validation should fail for mismatched types")
	}
}

func TestValidatePattern(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	// Test successful validation
	result := validator.validatePattern("test@example.com", `^[^@]+@[^@]+\.[^@]+$`)
	if !result.Passed {
		t.Error("Pattern validation should pass for matching pattern")
	}

	// Test failed validation
	result = validator.validatePattern("invalid-email", `^[^@]+@[^@]+\.[^@]+$`)
	if result.Passed {
		t.Error("Pattern validation should fail for non-matching pattern")
	}
}

func TestExtractJSONValue(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	data := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":   123,
			"name": "test",
		},
		"items": []interface{}{"item1", "item2", "item3"},
	}

	// Test root path
	value, err := validator.extractJSONValue(data, "$")
	if err != nil {
		t.Errorf("Failed to extract root value: %v", err)
	}
	if value == nil {
		t.Error("Root value should not be nil")
	}

	// Test object path
	value, err = validator.extractJSONValue(data, "$.status")
	if err != nil {
		t.Errorf("Failed to extract object value: %v", err)
	}
	if value != "success" {
		t.Errorf("Expected 'success', got %v", value)
	}

	// Test nested object path
	value, err = validator.extractJSONValue(data, "$.data.id")
	if err != nil {
		t.Errorf("Failed to extract nested value: %v", err)
	}
	if value != 123 {
		t.Errorf("Expected 123, got %v", value)
	}

	// Test array path
	value, err = validator.extractJSONValue(data, "$.items[0]")
	if err != nil {
		t.Errorf("Failed to extract array value: %v", err)
	}
	if value != "item1" {
		t.Errorf("Expected 'item1', got %v", value)
	}

	// Test non-existent path
	_, err = validator.extractJSONValue(data, "$.nonexistent")
	if err == nil {
		t.Error("Should return error for non-existent path")
	}
}

func TestValidateEmptyNilLen(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	response := &http.Response{
		StatusCode: 200,
		Body:       []byte(`{"empty_str":"","nonempty_str":"abc","empty_arr":[],"arr":[1,2],"empty_map":{},"map":{"a":1},"missing":null}`),
		Duration:   10 * time.Millisecond,
	}

	// empty: true
	rule := ValidationRule{JSON: "$.empty_str", Empty: boolPtr(true)}
	result := validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected empty string to be empty")
	}
	rule = ValidationRule{JSON: "$.empty_arr", Empty: boolPtr(true)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected empty array to be empty")
	}
	rule = ValidationRule{JSON: "$.empty_map", Empty: boolPtr(true)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected empty map to be empty")
	}

	// empty: false
	rule = ValidationRule{JSON: "$.nonempty_str", Empty: boolPtr(false)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected non-empty string to not be empty")
	}
	rule = ValidationRule{JSON: "$.arr", Empty: boolPtr(false)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected non-empty array to not be empty")
	}
	rule = ValidationRule{JSON: "$.map", Empty: boolPtr(false)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected non-empty map to not be empty")
	}

	// nil: true
	rule = ValidationRule{JSON: "$.missing", Nil: boolPtr(true)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected missing to be nil")
	}

	// nil: false
	rule = ValidationRule{JSON: "$.arr", Nil: boolPtr(false)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected arr to not be nil")
	}

	// len
	rule = ValidationRule{JSON: "$.arr", Len: intPtr(2)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected arr to have len 2")
	}
	rule = ValidationRule{JSON: "$.nonempty_str", Len: intPtr(3)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected nonempty_str to have len 3")
	}
	rule = ValidationRule{JSON: "$.map", Len: intPtr(1)}
	result = validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected map to have len 1")
	}
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }
