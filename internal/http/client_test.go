package http

import (
	"testing"
	"time"

	"github.com/cjp2600/stepwise/internal/logger"
)

func TestNewClient(t *testing.T) {
	log := logger.New()
	timeout := 30 * time.Second

	client := NewClient(timeout, log)

	if client.httpClient.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, client.httpClient.Timeout)
	}

	if client.logger != log {
		t.Error("Logger not set correctly")
	}
}

func TestSerializeBody(t *testing.T) {
	log := logger.New()
	client := NewClient(30*time.Second, log)

	// Test string body
	body := "test string"
	bytes, err := client.serializeBody(body)
	if err != nil {
		t.Errorf("Failed to serialize string: %v", err)
	}
	if string(bytes) != body {
		t.Errorf("Expected %s, got %s", body, string(bytes))
	}

	// Test map body
	mapBody := map[string]interface{}{
		"key":    "value",
		"number": 123,
	}
	bytes, err = client.serializeBody(mapBody)
	if err != nil {
		t.Errorf("Failed to serialize map: %v", err)
	}
	if len(bytes) == 0 {
		t.Error("Serialized map should not be empty")
	}
}

func TestResponseMethods(t *testing.T) {
	response := &Response{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"X-Custom":     {"test-value"},
		},
		Body:     []byte(`{"status":"success"}`),
		Duration: 100 * time.Millisecond,
	}

	// Test IsSuccess
	if !response.IsSuccess() {
		t.Error("Status 200 should be considered success")
	}

	// Test GetHeader
	if response.GetHeader("Content-Type") != "application/json" {
		t.Error("Failed to get header value")
	}

	if response.GetHeader("Non-Existent") != "" {
		t.Error("Non-existent header should return empty string")
	}

	// Test GetTextBody
	textBody := response.GetTextBody()
	if textBody != `{"status":"success"}` {
		t.Errorf("Expected %s, got %s", `{"status":"success"}`, textBody)
	}

	// Test GetJSONBody
	jsonBody, err := response.GetJSONBody()
	if err != nil {
		t.Errorf("Failed to parse JSON body: %v", err)
	}

	if jsonBody == nil {
		t.Error("JSON body should not be nil")
	}
}
