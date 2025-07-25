# Variable Keys Support

## Overview

Stepwise now supports using variables in JSON object keys, allowing for dynamic key generation in request bodies and validation paths.

## Features

### 1. Variables in JSON Keys

You can now use variables directly in JSON object keys:

```yaml
variables:
  purchase_id: "purchase_12345"
  user_id: "user_67890"

steps:
  - name: "Create repayment"
    request:
      method: "POST"
      url: "https://api.example.com/repayment"
      body:
        purchases:
          "{{purchase_id}}":  # Variable in key
            installments_count: 1
            user_id: "{{user_id}}"
```

### 2. Utils Functions in Keys

Support for utils functions in keys:

```yaml
variables:
  purchase_id: "purchase_12345"

steps:
  - name: "Create repayment with encoded key"
    request:
      method: "POST"
      url: "https://api.example.com/repayment"
      body:
        purchases:
          "{{utils.base64(purchase_id)}}":  # Base64 encoded key
            installments_count: 1
```

### 3. Faker Functions in Keys

Support for faker functions in keys:

```yaml
steps:
  - name: "Create user with dynamic key"
    request:
      method: "POST"
      url: "https://api.example.com/users"
      body:
        "user_{{faker.uuid}}":  # Dynamic UUID key
          name: "{{faker.name}}"
          email: "{{faker.email}}"
```

### 4. Nested Variable Keys

Support for complex nested structures with variable keys:

```yaml
variables:
  user_id: "user_123"
  order_id: "order_456"
  product_id: "product_789"

steps:
  - name: "Complex nested structure"
    request:
      method: "POST"
      url: "https://api.example.com/orders"
      body:
        "{{user_id}}_data":  # Variable in top-level key
          "{{order_id}}_details":  # Variable in nested key
            "{{product_id}}_info":  # Variable in deeply nested key
              status: "active"
              price: 150.00
```

### 5. Variables in JSONPath Validation

Support for variables in JSONPath expressions for validation:

```yaml
variables:
  purchase_id: "purchase_12345"
  user_id: "user_67890"

steps:
  - name: "Validate with variable paths"
    request:
      method: "POST"
      url: "https://api.example.com/repayment"
      body:
        purchases:
          "{{purchase_id}}":
            installments_count: 1
            user_id: "{{user_id}}"
    validate:
      - json: "$.purchases.{{purchase_id}}.installments_count"
        equals: 1
      - json: "$.purchases.{{purchase_id}}.user_id"
        equals: "{{user_id}}"
```

### 6. Variables in Capture Paths

Support for variables in capture paths:

```yaml
variables:
  purchase_id: "purchase_12345"

steps:
  - name: "Capture with variable path"
    request:
      method: "POST"
      url: "https://api.example.com/repayment"
      body:
        purchases:
          "{{purchase_id}}":
            installments_count: 1
    capture:
      invoice_id: "$.purchases.{{purchase_id}}.installments_count"
```

## Examples

### Basic Example

```yaml
name: "Basic Variable Keys Demo"
version: "1.0"

variables:
  purchase_id: "purchase_12345"
  user_id: "user_67890"

steps:
  - name: "Create repayment"
    request:
      method: "POST"
      url: "https://httpbin.org/post"
      headers:
        Content-Type: "application/json"
      body:
        purchases:
          "{{purchase_id}}":
            installments_count: 1
            user_id: "{{user_id}}"
    validate:
      - status: 200
      - json: "$.json.purchases.{{purchase_id}}.installments_count"
        equals: 1
```

### Advanced Example

```yaml
name: "Advanced Variable Keys Demo"
version: "1.0"

variables:
  user_id: "user_123"
  order_id: "order_456"
  product_id: "product_789"

steps:
  - name: "Complex nested structure"
    request:
      method: "POST"
      url: "https://httpbin.org/post"
      headers:
        Content-Type: "application/json"
      body:
        "{{user_id}}":
          profile:
            "{{faker.uuid}}":
              name: "{{faker.name}}"
              email: "{{faker.email}}"
          orders:
            "{{order_id}}":
              items:
                "{{product_id}}":
                  quantity: 1
                  price: 99.99
              metadata:
                "{{faker.uuid}}_note": "{{faker.sentence}}"
    validate:
      - status: 200
      - json: "$.json.{{user_id}}.orders.{{order_id}}.items.{{product_id}}.quantity"
        equals: 1
```

## Implementation Details

### How It Works

1. **Key Substitution**: The `SubstituteMap` function now processes variables in both keys and values
2. **Nested Processing**: Variables are processed recursively in nested objects
3. **JSONPath Support**: Variables in JSONPath expressions are substituted before extraction
4. **Validation Support**: Variables in expected values are substituted during validation

### Technical Changes

1. **Enhanced `SubstituteMap`**: Now supports variable substitution in keys
2. **Updated `extractJSONValue`**: Supports variables in JSONPath expressions
3. **Enhanced Validation**: Supports variables in expected values
4. **Improved Error Handling**: Better error messages for variable substitution failures

## Best Practices

1. **Use Descriptive Variable Names**: Make variable names clear and meaningful
2. **Test Variable Substitution**: Always test your workflows with different variable values
3. **Handle Edge Cases**: Consider what happens when variables are undefined
4. **Use Consistent Patterns**: Stick to consistent naming conventions for variable keys

## Limitations

1. **Key Uniqueness**: Ensure that variable substitution doesn't create duplicate keys
2. **Complex Nesting**: Very deeply nested structures may be harder to debug
3. **Performance**: Complex variable substitution can impact performance with large objects

## Migration Guide

### From Static Keys

Before:
```yaml
body:
  purchases:
    "static_key":
      installments_count: 1
```

After:
```yaml
variables:
  purchase_id: "dynamic_key"

body:
  purchases:
    "{{purchase_id}}":
      installments_count: 1
```

### From Hardcoded Values

Before:
```yaml
validate:
  - json: "$.purchases.static_key.installments_count"
    equals: 1
```

After:
```yaml
variables:
  purchase_id: "dynamic_key"

validate:
  - json: "$.purchases.{{purchase_id}}.installments_count"
    equals: 1
``` 