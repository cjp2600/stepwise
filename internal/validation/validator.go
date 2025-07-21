package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cjp2600/stepwise/internal/http"
	"github.com/cjp2600/stepwise/internal/logger"
)

// Validator represents a validation engine
type Validator struct {
	logger *logger.Logger
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Status   int         `yaml:"status" json:"status"`
	JSON     string      `yaml:"json" json:"json"`
	XML      string      `yaml:"xml" json:"xml"`
	Time     string      `yaml:"time" json:"time"`
	Equals   interface{} `yaml:"equals" json:"equals"`
	Contains string      `yaml:"contains" json:"contains"`
	Type     string      `yaml:"type" json:"type"`
	Greater  interface{} `yaml:"greater" json:"greater"`
	Less     interface{} `yaml:"less" json:"less"`
	Pattern  string      `yaml:"pattern" json:"pattern"`
	Custom   string      `yaml:"custom" json:"custom"`
	Value    string      `yaml:"value" json:"value"`
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Type     string      `json:"type"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
	Passed   bool        `json:"passed"`
	Error    string      `json:"error,omitempty"`
}

// NewValidator creates a new validator
func NewValidator(log *logger.Logger) *Validator {
	return &Validator{
		logger: log,
	}
}

// Validate validates a response against validation rules
func (v *Validator) Validate(response *http.Response, rules []ValidationRule) ([]ValidationResult, error) {
	var results []ValidationResult

	for _, rule := range rules {
		result := v.validateRule(response, rule)
		results = append(results, result)
	}

	return results, nil
}

// validateRule validates a single rule
func (v *Validator) validateRule(response *http.Response, rule ValidationRule) ValidationResult {
	// Status code validation
	if rule.Status != 0 {
		return v.validateStatus(response, rule.Status)
	}

	// Time validation
	if rule.Time != "" {
		return v.validateTime(response, rule.Time)
	}

	// JSON validation
	if rule.JSON != "" {
		return v.validateJSON(response, rule)
	}

	// XML validation
	if rule.XML != "" {
		return v.validateXML(response, rule)
	}

	// Default to failed validation
	return ValidationResult{
		Type:     "unknown",
		Expected: "valid rule",
		Actual:   "no matching rule found",
		Passed:   false,
		Error:    "no validation rule matched",
	}
}

// validateStatus validates HTTP status code
func (v *Validator) validateStatus(response *http.Response, expected int) ValidationResult {
	passed := response.StatusCode == expected
	return ValidationResult{
		Type:     "status",
		Expected: expected,
		Actual:   response.StatusCode,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "status code", expected, response.StatusCode),
	}
}

// validateTime validates response time
func (v *Validator) validateTime(response *http.Response, timeRule string) ValidationResult {
	duration := response.Duration
	expectedDuration, err := v.parseTimeRule(timeRule)
	if err != nil {
		return ValidationResult{
			Type:     "time",
			Expected: timeRule,
			Actual:   duration,
			Passed:   false,
			Error:    fmt.Sprintf("invalid time rule: %v", err),
		}
	}

	passed := v.compareDuration(duration, expectedDuration, timeRule)
	return ValidationResult{
		Type:     "time",
		Expected: timeRule,
		Actual:   duration,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "response time", timeRule, duration),
	}
}

// validateJSON validates JSON response
func (v *Validator) validateJSON(response *http.Response, rule ValidationRule) ValidationResult {
	// Parse JSON response
	jsonData, err := response.GetJSONBody()
	if err != nil {
		return ValidationResult{
			Type:     "json",
			Expected: rule.JSON,
			Actual:   "invalid JSON",
			Passed:   false,
			Error:    fmt.Sprintf("failed to parse JSON: %v", err),
		}
	}

	// Extract value using JSONPath-like syntax
	value, err := v.extractJSONValue(jsonData, rule.JSON)
	if err != nil {
		return ValidationResult{
			Type:     "json",
			Expected: rule.JSON,
			Actual:   "extraction failed",
			Passed:   false,
			Error:    fmt.Sprintf("failed to extract value: %v", err),
		}
	}

	// Apply validation based on rule type
	if rule.Equals != nil {
		return v.validateEquals(value, rule.Equals)
	}

	if rule.Contains != "" {
		return v.validateContains(value, rule.Contains)
	}

	if rule.Type != "" {
		return v.validateType(value, rule.Type)
	}

	if rule.Greater != nil {
		return v.validateGreater(value, rule.Greater)
	}

	if rule.Less != nil {
		return v.validateLess(value, rule.Less)
	}

	if rule.Pattern != "" {
		return v.validatePattern(value, rule.Pattern)
	}

	// Default validation - just check if value exists
	passed := value != nil
	return ValidationResult{
		Type:     "json",
		Expected: rule.JSON,
		Actual:   value,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "JSON value", rule.JSON, value),
	}
}

// validateXML validates XML response (placeholder for future implementation)
func (v *Validator) validateXML(response *http.Response, rule ValidationRule) ValidationResult {
	return ValidationResult{
		Type:     "xml",
		Expected: rule.XML,
		Actual:   "not implemented",
		Passed:   false,
		Error:    "XML validation not implemented yet",
	}
}

// validateEquals validates equality
func (v *Validator) validateEquals(actual, expected interface{}) ValidationResult {
	passed := reflect.DeepEqual(actual, expected)
	return ValidationResult{
		Type:     "equals",
		Expected: expected,
		Actual:   actual,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "value", expected, actual),
	}
}

// validateContains validates if value contains substring
func (v *Validator) validateContains(value interface{}, substring string) ValidationResult {
	strValue := fmt.Sprintf("%v", value)
	passed := strings.Contains(strValue, substring)
	return ValidationResult{
		Type:     "contains",
		Expected: substring,
		Actual:   strValue,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "contains", substring, strValue),
	}
}

// validateType validates value type
func (v *Validator) validateType(value interface{}, expectedType string) ValidationResult {
	actualType := reflect.TypeOf(value).String()
	passed := v.matchesType(value, expectedType)
	return ValidationResult{
		Type:     "type",
		Expected: expectedType,
		Actual:   actualType,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "type", expectedType, actualType),
	}
}

// validateGreater validates if value is greater than expected
func (v *Validator) validateGreater(value, expected interface{}) ValidationResult {
	passed := v.compareValues(value, expected, ">")
	return ValidationResult{
		Type:     "greater",
		Expected: expected,
		Actual:   value,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "greater than", expected, value),
	}
}

// validateLess validates if value is less than expected
func (v *Validator) validateLess(value, expected interface{}) ValidationResult {
	passed := v.compareValues(value, expected, "<")
	return ValidationResult{
		Type:     "less",
		Expected: expected,
		Actual:   value,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "less than", expected, value),
	}
}

// validatePattern validates if value matches pattern
func (v *Validator) validatePattern(value interface{}, pattern string) ValidationResult {
	strValue := fmt.Sprintf("%v", value)
	matched, err := regexp.MatchString(pattern, strValue)
	if err != nil {
		return ValidationResult{
			Type:     "pattern",
			Expected: pattern,
			Actual:   strValue,
			Passed:   false,
			Error:    fmt.Sprintf("invalid regex pattern: %v", err),
		}
	}

	return ValidationResult{
		Type:     "pattern",
		Expected: pattern,
		Actual:   strValue,
		Passed:   matched,
		Error:    v.getErrorMessage(matched, "pattern match", pattern, strValue),
	}
}

// Helper methods

func (v *Validator) parseTimeRule(timeRule string) (time.Duration, error) {
	// Parse time rules like "< 1000ms", "> 100ms", "100-500ms"
	if strings.HasPrefix(timeRule, "<") {
		durationStr := strings.TrimSpace(strings.TrimPrefix(timeRule, "<"))
		return time.ParseDuration(durationStr)
	}
	if strings.HasPrefix(timeRule, ">") {
		durationStr := strings.TrimSpace(strings.TrimPrefix(timeRule, ">"))
		return time.ParseDuration(durationStr)
	}
	if strings.Contains(timeRule, "-") {
		parts := strings.Split(timeRule, "-")
		if len(parts) == 2 {
			maxStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "ms"))
			// For now, just parse the max duration
			return time.ParseDuration(maxStr + "ms")
		}
	}
	return time.ParseDuration(timeRule)
}

func (v *Validator) compareDuration(actual, expected time.Duration, rule string) bool {
	if strings.HasPrefix(rule, "<") {
		return actual < expected
	}
	if strings.HasPrefix(rule, ">") {
		return actual > expected
	}
	if strings.Contains(rule, "-") {
		parts := strings.Split(rule, "-")
		if len(parts) == 2 {
			minStr := strings.TrimSpace(parts[0])
			maxStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "ms"))
			min, _ := time.ParseDuration(minStr + "ms")
			max, _ := time.ParseDuration(maxStr + "ms")
			return actual >= min && actual <= max
		}
	}
	return actual == expected
}

func (v *Validator) extractJSONValue(data interface{}, path string) (interface{}, error) {
	// Simple JSONPath-like extraction
	// Handle cases like "$.key", "$.nested.key", "$[0]", "$.items[0]"
	if path == "$" {
		return data, nil
	}

	if strings.HasPrefix(path, "$.") {
		// Handle nested paths like "$.data.id" or "$.items[0]"
		pathParts := strings.Split(strings.TrimPrefix(path, "$."), ".")
		current := data

		for i, part := range pathParts {
			// Check if this part contains array access like "items[0]"
			if strings.Contains(part, "[") && strings.Contains(part, "]") {
				// Extract array name and index
				openBracket := strings.Index(part, "[")
				closeBracket := strings.Index(part, "]")
				arrayName := part[:openBracket]
				indexStr := part[openBracket+1 : closeBracket]

				// Get the array
				if mapData, ok := current.(map[string]interface{}); ok {
					if array, exists := mapData[arrayName]; exists {
						if arrayData, ok := array.([]interface{}); ok {
							index, err := strconv.Atoi(indexStr)
							if err != nil {
								return nil, fmt.Errorf("invalid array index: %s", indexStr)
							}
							if index >= 0 && index < len(arrayData) {
								current = arrayData[index]
							} else {
								return nil, fmt.Errorf("array index out of bounds: %d", index)
							}
						} else {
							return nil, fmt.Errorf("key %s is not an array", arrayName)
						}
					} else {
						return nil, fmt.Errorf("key not found: %s", arrayName)
					}
				} else {
					return nil, fmt.Errorf("cannot access array on non-object")
				}
			} else {
				// Simple key access
				if mapData, ok := current.(map[string]interface{}); ok {
					if value, exists := mapData[part]; exists {
						current = value
					} else {
						return nil, fmt.Errorf("key not found: %s", part)
					}
				} else {
					return nil, fmt.Errorf("cannot access key on non-object")
				}
			}

			// If this is the last part, return the value
			if i == len(pathParts)-1 {
				return current, nil
			}
		}
	}

	if strings.HasPrefix(path, "$[") && strings.HasSuffix(path, "]") {
		indexStr := strings.TrimPrefix(strings.TrimSuffix(path, "]"), "$[")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid array index: %s", indexStr)
		}
		if arrayData, ok := data.([]interface{}); ok {
			if index >= 0 && index < len(arrayData) {
				return arrayData[index], nil
			}
		}
		return nil, fmt.Errorf("array index out of bounds: %d", index)
	}

	return nil, fmt.Errorf("unsupported JSON path: %s", path)
}

func (v *Validator) matchesType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return true
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return reflect.TypeOf(value).String() == expectedType
	}
}

func (v *Validator) compareValues(a, b interface{}, operator string) bool {
	// Convert to float64 for numeric comparison
	aFloat, aOk := v.toFloat64(a)
	bFloat, bOk := v.toFloat64(b)

	if aOk && bOk {
		switch operator {
		case ">":
			return aFloat > bFloat
		case "<":
			return aFloat < bFloat
		case ">=":
			return aFloat >= bFloat
		case "<=":
			return aFloat <= bFloat
		}
	}

	// Fallback to string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	switch operator {
	case ">":
		return aStr > bStr
	case "<":
		return aStr < bStr
	}

	return false
}

func (v *Validator) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (v *Validator) getErrorMessage(passed bool, validationType string, expected, actual interface{}) string {
	if passed {
		return ""
	}
	return fmt.Sprintf("%s validation failed: expected %v, got %v", validationType, expected, actual)
}
