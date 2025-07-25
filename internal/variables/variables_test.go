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

func TestUtilsBase64(t *testing.T) {
	log := logger.New()
	vm := NewManager(log)

	// Прямая строка
	encoded, _ := vm.Substitute("{{utils.base64('hello')}}")
	if encoded != "aGVsbG8=" {
		t.Errorf("Expected aGVsbG8=, got %s", encoded)
	}

	// Декодирование
	decoded, _ := vm.Substitute("{{utils.base64_decode('aGVsbG8=')}}")
	if decoded != "hello" {
		t.Errorf("Expected hello, got %s", decoded)
	}

	// Вложенная переменная
	vm.Set("myvar", "test123")
	encodedVar, _ := vm.Substitute("{{utils.base64({{myvar}})}}")
	if encodedVar != "dGVzdDEyMw==" {
		t.Errorf("Expected dGVzdDEyMw==, got %s", encodedVar)
	}

	decodedVar, _ := vm.Substitute("{{utils.base64_decode('dGVzdDEyMw==')}}")
	if decodedVar != "test123" {
		t.Errorf("Expected test123, got %s", decodedVar)
	}
}

func TestSubstituteMapWithVariableKeys(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Set up variables
	manager.Set("purchase_id", "purchase_12345")
	manager.Set("user_id", "user_67890")
	manager.Set("order_id", "order_11111")

	// Test map with variable keys
	input := map[string]interface{}{
		"purchases": map[string]interface{}{
			"{{purchase_id}}": map[string]interface{}{
				"installments_count": 1,
			},
		},
		"orders": map[string]interface{}{
			"{{order_id}}": map[string]interface{}{
				"user": "{{user_id}}",
			},
		},
	}

	result, err := manager.SubstituteMap(input)
	if err != nil {
		t.Errorf("Failed to substitute map with variable keys: %v", err)
	}

	// Check that keys were substituted
	purchases, ok := result["purchases"].(map[string]interface{})
	if !ok {
		t.Error("Expected purchases to be a map")
	}

	if _, exists := purchases["purchase_12345"]; !exists {
		t.Error("Expected purchase_12345 key to exist after substitution")
	}

	orders, ok := result["orders"].(map[string]interface{})
	if !ok {
		t.Error("Expected orders to be a map")
	}

	if _, exists := orders["order_11111"]; !exists {
		t.Error("Expected order_11111 key to exist after substitution")
	}

	orderData, ok := orders["order_11111"].(map[string]interface{})
	if !ok {
		t.Error("Expected order data to be a map")
	}

	if user, exists := orderData["user"]; !exists || user != "user_67890" {
		t.Errorf("Expected user to be 'user_67890', got %v", user)
	}
}

func TestSubstituteMapWithNestedVariableKeys(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Set up variables
	manager.Set("user_id", "user_123")
	manager.Set("order_id", "order_456")
	manager.Set("product_id", "product_789")

	// Test nested structure with variable keys
	input := map[string]interface{}{
		"{{user_id}}_data": map[string]interface{}{
			"{{order_id}}_details": map[string]interface{}{
				"{{product_id}}_info": map[string]interface{}{
					"status": "active",
					"price":  150.00,
				},
			},
		},
	}

	result, err := manager.SubstituteMap(input)
	if err != nil {
		t.Errorf("Failed to substitute nested map with variable keys: %v", err)
	}

	// Check nested structure
	userData, ok := result["user_123_data"].(map[string]interface{})
	if !ok {
		t.Error("Expected user_123_data to be a map")
	}

	orderDetails, ok := userData["order_456_details"].(map[string]interface{})
	if !ok {
		t.Error("Expected order_456_details to be a map")
	}

	productInfo, ok := orderDetails["product_789_info"].(map[string]interface{})
	if !ok {
		t.Error("Expected product_789_info to be a map")
	}

	if status, exists := productInfo["status"]; !exists || status != "active" {
		t.Errorf("Expected status to be 'active', got %v", status)
	}

	if price, exists := productInfo["price"]; !exists || price != 150.00 {
		t.Errorf("Expected price to be 150.00, got %v", price)
	}
}

func TestSubstituteMapWithUtilsInKeys(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Test map with utils functions in keys
	input := map[string]interface{}{
		"{{utils.base64('test_key')}}": map[string]interface{}{
			"value": "test_value",
		},
		"{{utils.base64('another_key')}}": map[string]interface{}{
			"value": "another_value",
		},
	}

	result, err := manager.SubstituteMap(input)
	if err != nil {
		t.Errorf("Failed to substitute map with utils in keys: %v", err)
	}

	// Check that base64 encoded keys exist
	expectedKey1 := "dGVzdF9rZXk="     // base64 of "test_key"
	expectedKey2 := "YW5vdGhlcl9rZXk=" // base64 of "another_key"

	if _, exists := result[expectedKey1]; !exists {
		t.Errorf("Expected key %s to exist after substitution", expectedKey1)
	}

	if _, exists := result[expectedKey2]; !exists {
		t.Errorf("Expected key %s to exist after substitution", expectedKey2)
	}
}

func TestSubstituteMapWithFakerInKeys(t *testing.T) {
	log := logger.New()
	manager := NewManager(log)

	// Test map with faker functions in keys
	input := map[string]interface{}{
		"user_{{faker.uuid}}": map[string]interface{}{
			"name":  "{{faker.name}}",
			"email": "{{faker.email}}",
		},
		"order_{{faker.uuid}}": map[string]interface{}{
			"status": "pending",
		},
	}

	result, err := manager.SubstituteMap(input)
	if err != nil {
		t.Errorf("Failed to substitute map with faker in keys: %v", err)
	}

	// Check that at least one key was generated (we can't predict exact values)
	foundUserKey := false
	foundOrderKey := false

	for key := range result {
		if strings.HasPrefix(key, "user_") && len(key) > 5 {
			foundUserKey = true
		}
		if strings.HasPrefix(key, "order_") && len(key) > 6 {
			foundOrderKey = true
		}
	}

	if !foundUserKey {
		t.Error("Expected to find a user key with faker.uuid")
	}

	if !foundOrderKey {
		t.Error("Expected to find an order key with faker.uuid")
	}
}
