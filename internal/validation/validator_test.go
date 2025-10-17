package validation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func TestValidateBase64JSONDecode(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	// Пример JSON с base64-encoded JSON в поле "widget"
	widgetObj := map[string]interface{}{"title": "PetShop", "value": 42}
	widgetBytes, _ := json.Marshal(widgetObj)
	widgetBase64 := base64.StdEncoding.EncodeToString(widgetBytes)
	response := &http.Response{
		StatusCode: 200,
		Body:       []byte(fmt.Sprintf(`{"widgets":[{"widget":"%s"}]}`, widgetBase64)),
		Duration:   10 * time.Millisecond,
	}

	rule := ValidationRule{
		JSON:     "$.widgets[0].widget",
		Decode:   "base64json",
		JSONPath: "$.title",
		Equals:   "PetShop",
	}
	result := validator.validateJSON(response, rule)
	if !result.Passed {
		t.Errorf("Expected base64json decode and jsonpath to pass, got error: %v", result.Error)
	}
}

func TestExtractJSONValueWithFilters(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	// Test data with array of users
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"id": 1, "name": "Alice", "age": 25, "active": true},
			map[string]interface{}{"id": 2, "name": "Bob", "age": 30, "active": false},
			map[string]interface{}{"id": 3, "name": "Charlie", "age": 35, "active": true},
			map[string]interface{}{"id": 4, "name": "Diana", "age": 28, "active": true},
		},
		"products": []interface{}{
			map[string]interface{}{"id": 101, "price": 50, "name": "Widget"},
			map[string]interface{}{"id": 102, "price": 150, "name": "Gadget"},
			map[string]interface{}{"id": 103, "price": 75, "name": "Tool"},
		},
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
		wantErr  bool
	}{
		// Filter by string equality
		{
			name:     "Filter by name",
			path:     `$.users[?(@.name == "Bob")]`,
			expected: map[string]interface{}{"id": float64(2), "name": "Bob", "age": float64(30), "active": false},
			wantErr:  false,
		},
		// Filter by numeric equality
		{
			name:     "Filter by id",
			path:     `$.users[?(@.id == 3)]`,
			expected: map[string]interface{}{"id": float64(3), "name": "Charlie", "age": float64(35), "active": true},
			wantErr:  false,
		},
		// Filter by boolean field
		{
			name:     "Filter by active boolean",
			path:     `$.users[?(@.active)]`,
			expected: map[string]interface{}{"id": float64(1), "name": "Alice", "age": float64(25), "active": true},
			wantErr:  false,
		},
		// Filter by boolean equality
		{
			name:     "Filter by active == true",
			path:     `$.users[?(@.active == true)]`,
			expected: map[string]interface{}{"id": float64(1), "name": "Alice", "age": float64(25), "active": true},
			wantErr:  false,
		},
		// Filter by greater than
		{
			name:     "Filter by age > 30",
			path:     `$.users[?(@.age > 30)]`,
			expected: map[string]interface{}{"id": float64(3), "name": "Charlie", "age": float64(35), "active": true},
			wantErr:  false,
		},
		// Filter by less than
		{
			name:     "Filter by price < 100",
			path:     `$.products[?(@.price < 100)]`,
			expected: map[string]interface{}{"id": float64(101), "price": float64(50), "name": "Widget"},
			wantErr:  false,
		},
		// Filter by greater or equal
		{
			name:     "Filter by price >= 75",
			path:     `$.products[?(@.price >= 75)]`,
			expected: map[string]interface{}{"id": float64(102), "price": float64(150), "name": "Gadget"},
			wantErr:  false,
		},
		// Filter by not equal
		{
			name:     "Filter by name != Alice",
			path:     `$.users[?(@.name != "Alice")]`,
			expected: map[string]interface{}{"id": float64(2), "name": "Bob", "age": float64(30), "active": false},
			wantErr:  false,
		},
		// Last element
		{
			name:     "Get last user",
			path:     `$.users[last]`,
			expected: map[string]interface{}{"id": float64(4), "name": "Diana", "age": float64(28), "active": true},
			wantErr:  false,
		},
		// Negative index
		{
			name:     "Get last user with -1",
			path:     `$.users[-1]`,
			expected: map[string]interface{}{"id": float64(4), "name": "Diana", "age": float64(28), "active": true},
			wantErr:  false,
		},
		// Array slice
		{
			name: "Get first 2 users",
			path: `$.users[0:2]`,
			expected: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "Alice", "age": float64(25), "active": true},
				map[string]interface{}{"id": float64(2), "name": "Bob", "age": float64(30), "active": false},
			},
			wantErr: false,
		},
		// Wildcard - get all elements
		{
			name: "Get all users with wildcard",
			path: `$.users[*]`,
			expected: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "Alice", "age": float64(25), "active": true},
				map[string]interface{}{"id": float64(2), "name": "Bob", "age": float64(30), "active": false},
				map[string]interface{}{"id": float64(3), "name": "Charlie", "age": float64(35), "active": true},
				map[string]interface{}{"id": float64(4), "name": "Diana", "age": float64(28), "active": true},
			},
			wantErr: false,
		},
		// Filter not found
		{
			name:    "Filter with no match",
			path:    `$.users[?(@.name == "Unknown")]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert data to JSON and back to simulate JSON unmarshaling
			jsonBytes, _ := json.Marshal(data)
			var jsonData interface{}
			json.Unmarshal(jsonBytes, &jsonData)

			value, err := validator.extractJSONValue(jsonData, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Compare results
			expectedJSON, _ := json.Marshal(tt.expected)
			actualJSON, _ := json.Marshal(value)

			if string(expectedJSON) != string(actualJSON) {
				t.Errorf("Expected %s, got %s", string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func TestExtractJSONValueWithNestedFilters(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	// Test data with nested structures
	data := map[string]interface{}{
		"orders": []interface{}{
			map[string]interface{}{
				"id": 1,
				"customer": map[string]interface{}{
					"name": "Alice",
					"vip":  true,
				},
				"total": 150,
			},
			map[string]interface{}{
				"id": 2,
				"customer": map[string]interface{}{
					"name": "Bob",
					"vip":  false,
				},
				"total": 75,
			},
			map[string]interface{}{
				"id": 3,
				"customer": map[string]interface{}{
					"name": "Charlie",
					"vip":  true,
				},
				"total": 200,
			},
		},
	}

	// Convert to JSON and back
	jsonBytes, _ := json.Marshal(data)
	var jsonData interface{}
	json.Unmarshal(jsonBytes, &jsonData)

	// Test filter with nested field
	value, err := validator.extractJSONValue(jsonData, `$.orders[?(@.customer.vip == true)]`)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Should return first VIP customer order
	orderMap, ok := value.(map[string]interface{})
	if !ok {
		t.Errorf("Expected map, got %T", value)
		return
	}

	if orderMap["id"].(float64) != 1 {
		t.Errorf("Expected order id 1, got %v", orderMap["id"])
	}

	// Test filter with nested field and greater than
	value, err = validator.extractJSONValue(jsonData, `$.orders[?(@.total > 100)]`)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	orderMap, ok = value.(map[string]interface{})
	if !ok {
		t.Errorf("Expected map, got %T", value)
		return
	}

	if orderMap["id"].(float64) != 1 {
		t.Errorf("Expected order id 1, got %v", orderMap["id"])
	}
}

func TestExtractJSONValueFilterAndAccess(t *testing.T) {
	log := logger.New()
	validator := NewValidator(log)

	// Test data
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"id":   1,
				"name": "Alice",
				"address": map[string]interface{}{
					"city":    "New York",
					"zipcode": "10001",
				},
			},
			map[string]interface{}{
				"id":   2,
				"name": "Bob",
				"address": map[string]interface{}{
					"city":    "Los Angeles",
					"zipcode": "90001",
				},
			},
		},
	}

	// Convert to JSON and back
	jsonBytes, _ := json.Marshal(data)
	var jsonData interface{}
	json.Unmarshal(jsonBytes, &jsonData)

	// Test filter and then access nested field
	value, err := validator.extractJSONValue(jsonData, `$.users[?(@.name == "Bob")].address.city`)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if value != "Los Angeles" {
		t.Errorf("Expected 'Los Angeles', got %v", value)
	}
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }
