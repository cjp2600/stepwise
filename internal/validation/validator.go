package validation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cjp2600/stepwise/internal/http"
	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/variables"
)

// Validator represents a validation engine
type Validator struct {
	logger     *logger.Logger
	varManager *variables.Manager
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Status       int         `yaml:"status" json:"status"`
	JSON         string      `yaml:"json" json:"json"`
	XML          string      `yaml:"xml" json:"xml"`
	Time         string      `yaml:"time" json:"time"`
	Equals       interface{} `yaml:"equals" json:"equals"`
	Contains     string      `yaml:"contains" json:"contains"`
	Type         string      `yaml:"type" json:"type"`
	Greater      interface{} `yaml:"greater" json:"greater"`
	Less         interface{} `yaml:"less" json:"less"`
	Pattern      string      `yaml:"pattern" json:"pattern"`
	Custom       string      `yaml:"custom" json:"custom"`
	Value        string      `yaml:"value" json:"value"`
	Empty        *bool       `yaml:"empty,omitempty" json:"empty,omitempty"`   // true: must be empty, false: must not be empty
	Nil          *bool       `yaml:"nil,omitempty" json:"nil,omitempty"`       // true: must be nil, false: must not be nil
	Len          *int        `yaml:"len,omitempty" json:"len,omitempty"`       // length must be equal to this
	Decode       string      `yaml:"decode,omitempty" json:"decode,omitempty"` // "base64json"
	JSONPath     string      `yaml:"jsonpath,omitempty" json:"jsonpath,omitempty"`
	PrintDecoded bool        `yaml:"print_decoded,omitempty" json:"print_decoded,omitempty"`
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
		logger:     log,
		varManager: variables.NewManager(log),
	}
}

// SetVariableManager sets the variable manager for the validator
func (v *Validator) SetVariableManager(varManager *variables.Manager) {
	v.varManager = varManager
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

	// Сначала извлекаем значение по rule.JSON
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
	jsonData = value

	// Apply decode if needed
	if rule.Decode == "base64json" {
		v.logger.Debug("Type of jsonData for decode", "type", reflect.TypeOf(jsonData), "value", jsonData)
		if jsonData == nil {
			return ValidationResult{
				Type:     "exists",
				Expected: true,
				Actual:   false,
				Passed:   false,
				Error:    "value is nil, cannot decode base64json",
			}
		}
		strVal, ok := jsonData.(string)
		if !ok {
			// Попробовать []byte
			if b, ok2 := jsonData.([]byte); ok2 {
				strVal = string(b)
				ok = true
			}
		}
		if !ok {
			return ValidationResult{
				Type:     "decode",
				Expected: "base64json",
				Actual:   jsonData,
				Passed:   false,
				Error:    "value is not a string or []byte for base64json decode",
			}
		}
		decoded, err := base64.StdEncoding.DecodeString(strVal)
		if err != nil {
			return ValidationResult{
				Type:     "decode",
				Expected: "base64json",
				Actual:   jsonData,
				Passed:   false,
				Error:    "base64 decode error: " + err.Error(),
			}
		}
		var decodedJSON interface{}
		if err := json.Unmarshal(decoded, &decodedJSON); err != nil {
			return ValidationResult{
				Type:     "decode",
				Expected: "base64json",
				Actual:   jsonData,
				Passed:   false,
				Error:    "json decode error: " + err.Error(),
			}
		}
		if rule.PrintDecoded {
			pretty, _ := json.MarshalIndent(decodedJSON, "", "  ")
			v.logger.Debug("Decoded base64json structure", "json", string(pretty))
		}
		jsonData = decodedJSON
	}
	// Apply jsonpath if needed
	if rule.JSONPath != "" {
		val, err := v.extractJSONValue(jsonData, rule.JSONPath)
		if err != nil {
			return ValidationResult{
				Type:     "jsonpath",
				Expected: rule.JSONPath,
				Actual:   jsonData,
				Passed:   false,
				Error:    "jsonpath error: " + err.Error(),
			}
		}
		jsonData = val
	}

	value = jsonData

	// Apply validation based on rule type
	if rule.Nil != nil {
		isNil := value == nil
		passed := isNil == *rule.Nil
		return ValidationResult{
			Type:     "nil",
			Expected: *rule.Nil,
			Actual:   isNil,
			Passed:   passed,
			Error:    v.getErrorMessage(passed, "nil", *rule.Nil, isNil),
		}
	}
	if rule.Empty != nil {
		isEmpty := false
		switch val := value.(type) {
		case nil:
			isEmpty = true
		case string:
			isEmpty = val == ""
		case []interface{}:
			isEmpty = len(val) == 0
		case map[string]interface{}:
			isEmpty = len(val) == 0
		default:
			isEmpty = reflect.ValueOf(val).Len() == 0
		}
		passed := isEmpty == *rule.Empty
		return ValidationResult{
			Type:     "empty",
			Expected: *rule.Empty,
			Actual:   isEmpty,
			Passed:   passed,
			Error:    v.getErrorMessage(passed, "empty", *rule.Empty, isEmpty),
		}
	}
	if rule.Len != nil {
		var l int
		switch val := value.(type) {
		case string:
			l = len(val)
		case []interface{}:
			l = len(val)
		case map[string]interface{}:
			l = len(val)
		default:
			l = reflect.ValueOf(val).Len()
		}
		passed := l == *rule.Len
		return ValidationResult{
			Type:     "len",
			Expected: *rule.Len,
			Actual:   l,
			Passed:   passed,
			Error:    v.getErrorMessage(passed, "len", *rule.Len, l),
		}
	}
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
	// Substitute variables in expected value if it's a string
	var substitutedExpected interface{} = expected
	if expectedStr, ok := expected.(string); ok {
		if substitutedStr, err := v.varManager.Substitute(expectedStr); err == nil {
			substitutedExpected = substitutedStr
		}
	}

	// Convert both values to float64 for numeric comparison
	actualFloat, actualOk := v.toFloat64(actual)
	expectedFloat, expectedOk := v.toFloat64(substitutedExpected)

	var passed bool
	if actualOk && expectedOk {
		// Both are numeric, compare as floats
		passed = actualFloat == expectedFloat
	} else {
		// Use deep equality for non-numeric values
		passed = reflect.DeepEqual(actual, substitutedExpected)
	}

	return ValidationResult{
		Type:     "equals",
		Expected: substitutedExpected,
		Actual:   actual,
		Passed:   passed,
		Error:    v.getErrorMessage(passed, "value", substitutedExpected, actual),
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
	// Substitute variables in the path first
	substitutedPath, err := v.varManager.Substitute(path)
	if err != nil {
		return nil, fmt.Errorf("failed to substitute variables in path '%s': %w", path, err)
	}

	// Поддержка сложных путей с массивами: $.widgets[0].widget
	if substitutedPath == "$" {
		return data, nil
	}

	if strings.HasPrefix(substitutedPath, "$.") {
		pathExpr := strings.TrimPrefix(substitutedPath, "$.")
		// Разбиваем путь на части с учётом индексов
		var parts []string
		var buf strings.Builder
		inBracket := false
		for _, r := range pathExpr {
			if r == '.' && !inBracket {
				parts = append(parts, buf.String())
				buf.Reset()
			} else {
				if r == '[' {
					inBracket = true
				}
				if r == ']' {
					inBracket = false
				}
				buf.WriteRune(r)
			}
		}
		if buf.Len() > 0 {
			parts = append(parts, buf.String())
		}
		current := data
		for _, part := range parts {
			// Special handling for "length" property on arrays
			if part == "length" {
				if arr, ok := current.([]interface{}); ok {
					return len(arr), nil
				}
				return nil, fmt.Errorf("cannot get length of non-array")
			}

			// Массив с индексом или фильтром: key[index] или key[?(...)]
			if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
				openBracket := strings.Index(part, "[")
				closeBracket := strings.LastIndex(part, "]")
				key := part[:openBracket]
				indexStr := part[openBracket+1 : closeBracket]

				// Получить массив из текущего объекта или использовать текущий, если это уже массив
				var arrayData []interface{}
				if key != "" {
					if mapData, ok := current.(map[string]interface{}); ok {
						if array, exists := mapData[key]; exists {
							if arr, ok := array.([]interface{}); ok {
								arrayData = arr
							} else {
								return nil, fmt.Errorf("key %s is not an array", key)
							}
						} else {
							return nil, fmt.Errorf("key not found: %s", key)
						}
					} else {
						return nil, fmt.Errorf("cannot access key on non-object")
					}
				} else {
					// key пустой, значит применяем фильтр к текущему массиву
					if arr, ok := current.([]interface{}); ok {
						arrayData = arr
					} else {
						return nil, fmt.Errorf("cannot apply array filter to non-array")
					}
				}

				// Обработать фильтр или индекс
				result, err := v.processArrayAccessor(arrayData, indexStr)
				if err != nil {
					return nil, err
				}
				current = result
			} else {
				// Обычный ключ
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
		}
		return current, nil
	}

	if strings.HasPrefix(substitutedPath, "$[") {
		// Handle paths like $[0] or $[filter] or $[filter].field
		closeBracket := strings.Index(substitutedPath, "]")
		if closeBracket == -1 {
			return nil, fmt.Errorf("unclosed bracket in path: %s", substitutedPath)
		}

		indexStr := substitutedPath[2:closeBracket]
		remainingPath := substitutedPath[closeBracket+1:]

		// Apply array accessor
		if arrayData, ok := data.([]interface{}); ok {
			result, err := v.processArrayAccessor(arrayData, indexStr)
			if err != nil {
				return nil, err
			}

			// If there's a remaining path, continue processing
			if remainingPath != "" {
				if strings.HasPrefix(remainingPath, ".") {
					remainingPath = "$" + remainingPath
				}
				return v.extractJSONValue(result, remainingPath)
			}

			return result, nil
		}
		return nil, fmt.Errorf("root element is not an array")
	}

	return nil, fmt.Errorf("unsupported JSON path: %s", substitutedPath)
}

// processArrayAccessor handles array access with index, filter, or special selectors
func (v *Validator) processArrayAccessor(arrayData []interface{}, accessor string) (interface{}, error) {
	if len(arrayData) == 0 {
		return nil, nil
	}

	// Filter expression: ?(@.field op value) or ?(@.field)
	if strings.HasPrefix(accessor, "?(@.") && strings.HasSuffix(accessor, ")") {
		return v.filterArray(arrayData, accessor)
	}

	// Wildcard: return all elements
	if accessor == "*" {
		return arrayData, nil
	}

	// Last element
	if accessor == "last" || accessor == "-1" {
		return arrayData[len(arrayData)-1], nil
	}

	// Slice: start:end
	if strings.Contains(accessor, ":") {
		return v.sliceArray(arrayData, accessor)
	}

	// Simple numeric index
	index, err := strconv.Atoi(accessor)
	if err != nil {
		return nil, fmt.Errorf("invalid array accessor: %s", accessor)
	}

	// Handle negative indices
	if index < 0 {
		index = len(arrayData) + index
	}

	if index >= 0 && index < len(arrayData) {
		return arrayData[index], nil
	}

	return nil, fmt.Errorf("array index out of bounds: %d (length: %d)", index, len(arrayData))
}

// filterArray filters array elements based on condition
func (v *Validator) filterArray(arrayData []interface{}, filter string) (interface{}, error) {
	// Remove ?(@. prefix and ) suffix
	filter = strings.TrimPrefix(filter, "?(@.")
	filter = strings.TrimSuffix(filter, ")")

	// Parse filter expression: field op value or just field
	var field, operator, expectedValue string

	// Try to find operator
	operators := []string{"==", "!=", ">=", "<=", ">", "<", "="}
	for _, op := range operators {
		if strings.Contains(filter, op) {
			parts := strings.SplitN(filter, op, 2)
			field = strings.TrimSpace(parts[0])
			operator = op
			if len(parts) > 1 {
				expectedValue = strings.TrimSpace(parts[1])
				// Remove quotes if present
				expectedValue = strings.Trim(expectedValue, "\"'")
			}
			break
		}
	}

	// No operator found, check for boolean field
	if operator == "" {
		field = strings.TrimSpace(filter)
	}

	// Find first matching element
	for _, item := range arrayData {
		mapItem, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract field value using dot notation if needed
		fieldValue, err := v.extractFieldValue(mapItem, field)
		if err != nil {
			continue
		}

		// Check condition
		matched := false
		if operator == "" {
			// Boolean check - field exists and is truthy
			matched = v.isTruthy(fieldValue)
		} else {
			matched = v.compareFieldValue(fieldValue, operator, expectedValue)
		}

		if matched {
			return item, nil
		}
	}

	return nil, fmt.Errorf("no matching element found in array for filter: %s", filter)
}

// extractFieldValue extracts a field value, supporting dot notation
func (v *Validator) extractFieldValue(obj map[string]interface{}, field string) (interface{}, error) {
	if !strings.Contains(field, ".") {
		if val, exists := obj[field]; exists {
			return val, nil
		}
		return nil, fmt.Errorf("field not found: %s", field)
	}

	// Handle nested fields
	parts := strings.Split(field, ".")
	current := interface{}(obj)
	for _, part := range parts {
		if mapData, ok := current.(map[string]interface{}); ok {
			if val, exists := mapData[part]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("field not found: %s", part)
			}
		} else {
			return nil, fmt.Errorf("cannot access field on non-object")
		}
	}
	return current, nil
}

// isTruthy checks if a value is truthy
func (v *Validator) isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	switch val := value.(type) {
	case bool:
		return val
	case string:
		return val != ""
	case int, int8, int16, int32, int64:
		return val != 0
	case float32, float64:
		return val != 0.0
	default:
		return true
	}
}

// compareFieldValue compares a field value against expected value using operator
func (v *Validator) compareFieldValue(fieldValue interface{}, operator, expectedValue string) bool {
	// Try numeric comparison first
	fieldFloat, fieldOk := v.toFloat64(fieldValue)
	expectedFloat, expectedOk := v.toFloat64(expectedValue)

	if fieldOk && expectedOk {
		switch operator {
		case "==", "=":
			return fieldFloat == expectedFloat
		case "!=":
			return fieldFloat != expectedFloat
		case ">":
			return fieldFloat > expectedFloat
		case "<":
			return fieldFloat < expectedFloat
		case ">=":
			return fieldFloat >= expectedFloat
		case "<=":
			return fieldFloat <= expectedFloat
		}
	}

	// Try boolean comparison
	if fieldBool, ok := fieldValue.(bool); ok {
		expectedBool := expectedValue == "true"
		switch operator {
		case "==", "=":
			return fieldBool == expectedBool
		case "!=":
			return fieldBool != expectedBool
		}
	}

	// String comparison
	fieldStr := fmt.Sprintf("%v", fieldValue)
	switch operator {
	case "==", "=":
		return fieldStr == expectedValue
	case "!=":
		return fieldStr != expectedValue
	case ">":
		return fieldStr > expectedValue
	case "<":
		return fieldStr < expectedValue
	case ">=":
		return fieldStr >= expectedValue
	case "<=":
		return fieldStr <= expectedValue
	}

	return false
}

// sliceArray returns a slice of array
func (v *Validator) sliceArray(arrayData []interface{}, sliceExpr string) (interface{}, error) {
	parts := strings.Split(sliceExpr, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid slice expression: %s", sliceExpr)
	}

	start := 0
	end := len(arrayData)

	if parts[0] != "" {
		var err error
		start, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid slice start: %s", parts[0])
		}
		if start < 0 {
			start = len(arrayData) + start
		}
	}

	if parts[1] != "" {
		var err error
		end, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid slice end: %s", parts[1])
		}
		if end < 0 {
			end = len(arrayData) + end
		}
	}

	if start < 0 || end > len(arrayData) || start > end {
		return nil, fmt.Errorf("slice out of bounds: %d:%d (length: %d)", start, end, len(arrayData))
	}

	return arrayData[start:end], nil
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
