# Variable Capture and Validation Guide

## Overview

Stepwise provides powerful capabilities for capturing data from API responses and using it in subsequent validation and workflow steps.

## Core Features

### ✅ 1. Variable Capture

You can capture data from JSON responses using JSONPath:

```yaml
steps:
  - name: "Get User"
    request:
      method: "GET"
      url: "https://api.example.com/users/1"
    capture:
      user_id: "$.id"              # Capture ID
      user_name: "$.name"          # Capture name
      user_email: "$.email"        # Capture email
      user_city: "$.address.city"  # Capture nested field
```

### ✅ 2. Using Captured Variables

Captured variables are available in all subsequent steps:

```yaml
  - name: "Use Captured Data"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{user_id}}"  # Use in URL
      headers:
        X-User-Name: "{{user_name}}"  # Use in headers
```

### ✅ 3. Comparison in Validation

The most important feature - comparing response data with captured variables:

```yaml
  - name: "Verify Data"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{user_id}}"
    validate:
      - status: 200
      # Compare each field with captured variables
      - json: "$.id"
        equals: "{{user_id}}"
      - json: "$.name"
        equals: "{{user_name}}"
      - json: "$.email"
        equals: "{{user_email}}"
      - json: "$.address.city"
        equals: "{{user_city}}"
```

## Practical Examples

### Example 1: Simple Capture and Comparison

```yaml
name: "Simple Capture Example"
version: "1.0"

steps:
  # Step 1: Get user and capture data
  - name: "Get User"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/1"
    validate:
      - status: 200
    capture:
      saved_name: "$.name"
      saved_email: "$.email"

  # Step 2: Get the same user again and verify
  - name: "Verify User Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/1"
    validate:
      - status: 200
      - json: "$.name"
        equals: "{{saved_name}}"    # Compare with captured
      - json: "$.email"
        equals: "{{saved_email}}"   # Compare with captured
```

### Example 2: Capture from Array with Filtering

```yaml
name: "Array Filter Capture"
version: "1.0"

steps:
  # Step 1: Get array and capture data from specific element
  - name: "Get Posts Array"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts"
    capture:
      # Use JSONPath filters
      post_5_title: "$[?(@.id == 5)].title"
      post_5_body: "$[?(@.id == 5)].body"
      post_5_user_id: "$[?(@.id == 5)].userId"

  # Step 2: Get specific post and compare
  - name: "Verify Post Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts/5"
    validate:
      - status: 200
      - json: "$.title"
        equals: "{{post_5_title}}"     # Compare with data from array
      - json: "$.body"
        equals: "{{post_5_body}}"
      - json: "$.userId"
        equals: "{{post_5_user_id}}"
```

### Example 3: Chained Requests with Capture

```yaml
name: "Chained Requests"
version: "1.0"

steps:
  # Step 1: Get post and capture author ID
  - name: "Get Post"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts/1"
    capture:
      author_id: "$.userId"

  # Step 2: Use captured ID to get author
  - name: "Get Author"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/{{author_id}}"
    validate:
      - status: 200
      - json: "$.id"
        equals: "{{author_id}}"  # Verify ID correctness
    capture:
      author_name: "$.name"
      author_email: "$.email"

  # Step 3: Use all captured data
  - name: "Use All Captured Data"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
      headers:
        X-Author-Id: "{{author_id}}"
        X-Author-Name: "{{author_name}}"
    validate:
      - status: 200
```

### Example 4: Nested Fields

```yaml
name: "Nested Fields Capture"
version: "1.0"

steps:
  - name: "Get User with Nested Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/1"
    capture:
      # Capture nested fields
      user_city: "$.address.city"
      user_street: "$.address.street"
      user_lat: "$.address.geo.lat"
      user_lng: "$.address.geo.lng"
      company_name: "$.company.name"

  - name: "Verify Nested Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/1"
    validate:
      - status: 200
      # Compare nested fields
      - json: "$.address.city"
        equals: "{{user_city}}"
      - json: "$.address.street"
        equals: "{{user_street}}"
      - json: "$.address.geo.lat"
        equals: "{{user_lat}}"
      - json: "$.address.geo.lng"
        equals: "{{user_lng}}"
      - json: "$.company.name"
        equals: "{{company_name}}"
```

## Using with Components

### Creating Component with Capture

```yaml
# components/get-user-component.yml
name: "Get User Component"
version: "1.0"
type: "step"

variables:
  user_id: "1"

steps:
  - name: "Get User"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/{{user_id}}"
    validate:
      - status: 200
    capture:
      user_name: "$.name"
      user_email: "$.email"
      user_city: "$.address.city"
```

### Using Component and Its Variables

```yaml
name: "Component Usage"
version: "1.0"

imports:
  - path: "components/get-user-component"
    alias: "get-user"

steps:
  # Use component - it will capture variables
  - name: "Get User via Component"
    use: 'get-user'
    variables:
      user_id: "5"

  # Variables from component are available in next steps
  - name: "Verify Captured Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/5"
    validate:
      - status: 200
      - json: "$.name"
        equals: "{{user_name}}"    # Use variable from component
      - json: "$.email"
        equals: "{{user_email}}"
```

## Advanced JSONPath Filters

### Filtering by Condition

```yaml
capture:
  # Equality
  item_with_id_5: "$[?(@.id == 5)]"
  
  # Greater than
  high_id_items: "$[?(@.id > 95)]"
  
  # Less than
  low_price_items: "$[?(@.price < 100)]"
  
  # Multiple conditions
  specific_item: "$[?(@.id == 5 && @.active == true)]"
```

### Special Selectors

```yaml
capture:
  # First element
  first_item: "$[0]"
  
  # Last element
  last_item: "$[-1]"
  
  # Element range
  first_three: "$[0:3]"
  
  # All elements
  all_items: "$[*]"
```

### Nested Filters

```yaml
capture:
  # Filter with nested field
  user_in_paris: "$[?(@.address.city == 'Paris')].name"
  
  # Deeply nested field after filter
  latitude: "$[?(@.id == 1)].address.geo.lat"
```

## Validation Types with Variables

### 1. Direct Comparison (equals)

```yaml
validate:
  - json: "$.id"
    equals: "{{user_id}}"
```

### 2. Type Check

```yaml
validate:
  - json: "$.name"
    type: "string"
  - json: "$.id"
    type: "number"
```

### 3. Regular Expression

```yaml
validate:
  - json: "$.email"
    pattern: "^[^@]+@[^@]+\\.[^@]+$"
```

### 4. Combination of Checks

```yaml
validate:
  - status: 200
  - json: "$.id"
    type: "number"
  - json: "$.id"
    equals: "{{saved_id}}"
  - json: "$.email"
    equals: "{{saved_email}}"
  - json: "$.email"
    pattern: "^[^@]+@[^@]+\\.[^@]+$"
```

## Best Practices

### 1. Variable Naming

Use meaningful names:

```yaml
# ✅ Good
capture:
  user_id: "$.id"
  user_email: "$.email"
  created_at: "$.createdAt"

# ❌ Bad
capture:
  id: "$.id"
  e: "$.email"
  t: "$.createdAt"
```

### 2. Prefixes for Clarity

```yaml
capture:
  saved_user_id: "$.id"        # Saved ID
  current_user_name: "$.name"  # Current name
  original_email: "$.email"    # Original email
```

### 3. Grouping Related Data

```yaml
capture:
  # User data
  user_id: "$.id"
  user_name: "$.name"
  user_email: "$.email"
  
  # Address data
  address_city: "$.address.city"
  address_street: "$.address.street"
  
  # Geo data
  geo_lat: "$.address.geo.lat"
  geo_lng: "$.address.geo.lng"
```

### 4. Validate Before Using

```yaml
steps:
  - name: "Capture Data"
    request:
      method: "GET"
      url: "https://api.example.com/users/1"
    validate:
      - status: 200
      # Check that fields exist
      - json: "$.id"
        type: "number"
      - json: "$.name"
        type: "string"
    capture:
      user_id: "$.id"
      user_name: "$.name"
```

## Common Patterns

### Pattern 1: Create → Verify

```yaml
steps:
  - name: "Create Resource"
    request:
      method: "POST"
      url: "https://api.example.com/resources"
      body:
        name: "Test Resource"
    capture:
      created_id: "$.id"
      created_name: "$.name"

  - name: "Verify Created Resource"
    request:
      method: "GET"
      url: "https://api.example.com/resources/{{created_id}}"
    validate:
      - json: "$.id"
        equals: "{{created_id}}"
      - json: "$.name"
        equals: "{{created_name}}"
```

### Pattern 2: Get List → Get Details

```yaml
steps:
  - name: "Get List"
    request:
      method: "GET"
      url: "https://api.example.com/items"
    capture:
      first_item_id: "$[0].id"

  - name: "Get Item Details"
    request:
      method: "GET"
      url: "https://api.example.com/items/{{first_item_id}}"
    validate:
      - json: "$.id"
        equals: "{{first_item_id}}"
```

### Pattern 3: Related Resources

```yaml
steps:
  - name: "Get Post"
    request:
      method: "GET"
      url: "https://api.example.com/posts/1"
    capture:
      post_author_id: "$.userId"

  - name: "Get Author"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{post_author_id}}"
    capture:
      author_name: "$.name"

  - name: "Verify Relationship"
    request:
      method: "GET"
      url: "https://api.example.com/posts/1"
    validate:
      - json: "$.userId"
        equals: "{{post_author_id}}"
```

## Conclusion

Stepwise provides a complete set of tools for:

✅ Capturing data from JSON responses  
✅ Using captured variables in subsequent requests  
✅ Validating data by comparing with captured variables  
✅ Working with nested data structures  
✅ Filtering arrays using JSONPath  
✅ Reusing components with variable capture  

All these capabilities make Stepwise a powerful tool for API testing!
