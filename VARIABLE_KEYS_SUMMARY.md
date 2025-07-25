# Variable Keys Support - Implementation Summary

## Overview

Successfully implemented support for variables in JSON object keys in Stepwise, allowing dynamic key generation in request bodies and validation paths.

## Changes Made

### 1. Enhanced Variable Substitution (`internal/variables/variables.go`)

**File**: `internal/variables/variables.go`
**Function**: `SubstituteMap`

**Changes**:
- Added support for variable substitution in JSON object keys
- Implemented two-pass processing for nested substitutions
- Enhanced error handling with detailed error messages
- Added support for utils functions and faker functions in keys

**Before**:
```go
func (m *Manager) SubstituteMap(input map[string]interface{}) (map[string]interface{}, error) {
    result := make(map[string]interface{})
    for key, value := range input {
        result[key] = value  // Only values were processed
    }
    // ... rest of implementation
}
```

**After**:
```go
func (m *Manager) SubstituteMap(input map[string]interface{}) (map[string]interface{}, error) {
    result := make(map[string]interface{})
    
    // First pass: substitute variables in keys and values
    for key, value := range input {
        // Substitute variables in the key
        substitutedKey, err := m.Substitute(key)
        if err != nil {
            return nil, fmt.Errorf("failed to substitute key '%s': %w", key, err)
        }
        
        // Substitute variables in the value
        // ... enhanced value processing
        result[substitutedKey] = substitutedValue
    }
    
    // Second pass: handle nested substitutions
    // ... additional processing for complex cases
}
```

### 2. Enhanced JSONPath Support (`internal/workflow/workflow.go`)

**File**: `internal/workflow/workflow.go`
**Function**: `extractJSONValue`

**Changes**:
- Added variable substitution in JSONPath expressions
- Support for variables in capture paths and validation paths

**Before**:
```go
func (e *Executor) extractJSONValue(data interface{}, path string) (interface{}, error) {
    if path == "$" {
        return data, nil
    }
    // ... rest of implementation without variable substitution
}
```

**After**:
```go
func (e *Executor) extractJSONValue(data interface{}, path string) (interface{}, error) {
    // Substitute variables in the path first
    substitutedPath, err := e.varManager.Substitute(path)
    if err != nil {
        return nil, fmt.Errorf("failed to substitute variables in path '%s': %w", path, err)
    }
    
    // Use substitutedPath for all operations
    // ... rest of implementation
}
```

### 3. Enhanced Validation Support (`internal/validation/validator.go`)

**File**: `internal/validation/validator.go`
**Changes**:
- Added variable manager to validator
- Enhanced `extractJSONValue` function for variable support
- Updated `validateEquals` to support variables in expected values
- Added `SetVariableManager` method

**New Features**:
```go
type Validator struct {
    logger     *logger.Logger
    varManager *variables.Manager  // Added
}

func (v *Validator) SetVariableManager(varManager *variables.Manager) {
    v.varManager = varManager
}

func (v *Validator) validateEquals(actual, expected interface{}) ValidationResult {
    // Substitute variables in expected value if it's a string
    var substitutedExpected interface{} = expected
    if expectedStr, ok := expected.(string); ok {
        if substitutedStr, err := v.varManager.Substitute(expectedStr); err == nil {
            substitutedExpected = substitutedStr
        }
    }
    // ... rest of implementation
}
```

### 4. Updated Executor (`internal/workflow/workflow.go`)

**File**: `internal/workflow/workflow.go`
**Function**: `NewExecutor`

**Changes**:
- Connected variable manager to validator
- Ensured proper initialization of variable support

```go
func NewExecutor(cfg *config.Config, log *logger.Logger) *Executor {
    executor := &Executor{
        // ... existing fields
    }
    
    // Set the variable manager in the validator
    executor.validator.SetVariableManager(executor.varManager)
    
    return executor
}
```

## New Features Supported

### 1. Variables in JSON Keys
```yaml
variables:
  purchase_id: "purchase_12345"

body:
  purchases:
    "{{purchase_id}}":  # Variable in key
      installments_count: 1
```

### 2. Utils Functions in Keys
```yaml
body:
  purchases:
    "{{utils.base64(purchase_id)}}":  # Base64 encoded key
      installments_count: 1
```

### 3. Faker Functions in Keys
```yaml
body:
  "user_{{faker.uuid}}":  # Dynamic UUID key
    name: "{{faker.name}}"
    email: "{{faker.email}}"
```

### 4. Nested Variable Keys
```yaml
body:
  "{{user_id}}_data":  # Variable in top-level key
    "{{order_id}}_details":  # Variable in nested key
      "{{product_id}}_info":  # Variable in deeply nested key
        status: "active"
```

### 5. Variables in JSONPath Validation
```yaml
validate:
  - json: "$.purchases.{{purchase_id}}.installments_count"
    equals: 1
```

### 6. Variables in Capture Paths
```yaml
capture:
  invoice_id: "$.purchases.{{purchase_id}}.installments_count"
```

## Testing

### Unit Tests Added
- `TestSubstituteMapWithVariableKeys`
- `TestSubstituteMapWithNestedVariableKeys`
- `TestSubstituteMapWithUtilsInKeys`
- `TestSubstituteMapWithFakerInKeys`

### Integration Tests
- `examples/variable-keys-demo.yml` - Basic functionality tests
- `examples/working-variable-keys-demo.yml` - Working examples with httpbin.org
- `examples/tabby-repayment-demo.yml` - Real-world scenario example

## Documentation

### New Documentation Files
- `docs/VARIABLE_KEYS.md` - Comprehensive documentation
- `VARIABLE_KEYS_SUMMARY.md` - This implementation summary

### Updated Files
- `README.md` - Added feature to feature list

## Examples Created

1. **Basic Variable Keys Demo** (`examples/variable-keys-demo.yml`)
   - Demonstrates basic variable substitution in keys
   - Shows utils and faker functions in keys
   - Tests nested structures

2. **Working Variable Keys Demo** (`examples/working-variable-keys-demo.yml`)
   - Working examples with httpbin.org
   - All tests pass successfully
   - Demonstrates real-world usage

3. **Tabby Repayment Demo** (`examples/tabby-repayment-demo.yml`)
   - Real-world scenario based on user's example
   - Shows complex nested structures
   - Demonstrates practical use cases

## Backward Compatibility

✅ **Fully Backward Compatible**
- All existing workflows continue to work without changes
- No breaking changes to existing functionality
- Enhanced functionality is additive only

## Performance Impact

✅ **Minimal Performance Impact**
- Variable substitution is efficient
- Two-pass processing only when needed
- Caching of substituted values where possible

## Error Handling

✅ **Enhanced Error Handling**
- Detailed error messages for variable substitution failures
- Graceful fallback for missing variables
- Clear indication of which key/value failed substitution

## Future Enhancements

Potential future improvements:
1. Support for more complex variable expressions in keys
2. Performance optimizations for large objects
3. Additional utils functions for key manipulation
4. Support for conditional key generation

## Conclusion

The implementation successfully adds support for variables in JSON object keys while maintaining full backward compatibility. The feature is well-tested, documented, and ready for production use. 