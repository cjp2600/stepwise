# Array Filters and Advanced JSONPath

Stepwise supports extended JSONPath syntax for working with arrays, allowing you to find elements by conditions rather than fixed indices.

## The Problem

Traditional approach using indices:

```yaml
validate:
  - json: "$[0].id"
    equals: 123
```

**Problem**: If the order of elements in the array changes, the test will break. The element with `id=123` might be at position 1, 2, or any other position.

## Solution: Array Filters

### Basic Filter Syntax

```yaml
validate:
  # Find the first element where id equals 123
  - json: "$[?(@.id == 123)].name"
    type: "string"
```

Here:
- `$` - JSON root
- `[?(...)]` - array filter
- `@` - current array element
- `@.id == 123` - filter condition
- `.name` - field to extract from the found element

## Supported Comparison Operators

### Equality and Inequality

```yaml
# Equality (strings)
- json: "$[?(@.status == \"active\")].id"
  type: "number"

# Equality (numbers)
- json: "$[?(@.id == 42)].name"
  type: "string"

# Inequality
- json: "$[?(@.status != \"deleted\")].id"
  type: "number"
```

### Numeric Comparisons

```yaml
# Greater than
- json: "$[?(@.price > 100)].name"
  type: "string"

# Less than
- json: "$[?(@.age < 30)].name"
  type: "string"

# Greater than or equal
- json: "$[?(@.quantity >= 10)].id"
  type: "number"

# Less than or equal
- json: "$[?(@.rating <= 5)].title"
  type: "string"
```

### Boolean Fields

```yaml
# Check for true (short form)
- json: "$[?(@.active)].id"
  type: "number"

# Check for true (full form)
- json: "$[?(@.active == true)].id"
  type: "number"

# Check for false
- json: "$[?(@.deleted == false)].name"
  type: "string"
```

## Accessing Nested Fields

### Filtering by Nested Fields

```yaml
# Find user by nested field
- json: "$[?(@.address.city == \"New York\")].name"
  type: "string"

# Get nested field from result
- json: "$[?(@.id == 1)].address.geo.lat"
  type: "string"
```

## Special Selectors

### Last Element

```yaml
# Get last element
- json: "$[last].id"
  type: "number"

# Alternative syntax
- json: "$[-1].id"
  type: "number"
```

### Wildcard (All Elements)

```yaml
# Get entire array
- json: "$[*]"
  type: "array"

# Check array length
- json: "$.length"
  greater: 0
```

### Array Slices

```yaml
# Get first 3 elements
- json: "$[0:3]"
  type: "array"

# Get elements from 5 to 10
- json: "$[5:10]"
  type: "array"

# Get all elements starting from 3
- json: "$[3:]"
  type: "array"
```

## Usage Examples

### Example 1: Find User by Name

```yaml
name: "Find User by Name"
steps:
  - name: "Get all users"
    request:
      method: "GET"
      url: "https://api.example.com/users"
    validate:
      - status: 200
      # Find user Alice regardless of position in array
      - json: "$[?(@.name == \"Alice\")].email"
        pattern: "^[^@]+@[^@]+\\.[^@]+$"
    capture:
      alice_id: "$[?(@.name == \"Alice\")].id"
      alice_email: "$[?(@.name == \"Alice\")].email"
```

### Example 2: Filtering by Multiple Conditions

```yaml
name: "Filter Products"
steps:
  - name: "Get expensive products"
    request:
      method: "GET"
      url: "https://api.example.com/products"
    validate:
      - status: 200
      # Find first expensive product
      - json: "$[?(@.price > 1000)].name"
        type: "string"
      # Find products in stock
      - json: "$[?(@.inStock == true)].id"
        type: "number"
```

### Example 3: Working with Nested Structures

```yaml
name: "Complex Nested Filter"
steps:
  - name: "Get VIP orders"
    request:
      method: "GET"
      url: "https://api.example.com/orders"
    validate:
      - status: 200
      # Find order from VIP customer
      - json: "$[?(@.customer.vip == true)].id"
        type: "number"
      # Get order total
      - json: "$[?(@.customer.vip == true)].total"
        greater: 0
```

### Example 4: Using with Capture

```yaml
name: "Capture and Reuse"
steps:
  - name: "Find active user"
    request:
      method: "GET"
      url: "https://api.example.com/users"
    validate:
      - status: 200
    capture:
      # Capture ID of active user
      active_user_id: "$[?(@.active == true)].id"
      active_user_name: "$[?(@.active == true)].name"
  
  - name: "Get user details"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{active_user_id}}"
    validate:
      - status: 200
      - json: "$.id"
        equals: "{{active_user_id}}"
```

## Comparison: Before and After

### Before (with indices)

```yaml
# Fragile code - depends on element order
validate:
  - json: "$[0].id"
    equals: 123
  - json: "$[2].name"
    equals: "Alice"
```

**Problems:**
- If elements are reordered, tests will break
- Cannot find element by condition
- No flexibility when data changes

### After (with filters)

```yaml
# Reliable code - independent of order
validate:
  - json: "$[?(@.id == 123)].id"
    equals: 123
  - json: "$[?(@.name == \"Alice\")].name"
    equals: "Alice"
```

**Advantages:**
- ✅ Independent of element order
- ✅ Finds element by any field
- ✅ Works when data changes
- ✅ More readable and understandable code

## Recommendations

1. **Use filters instead of indices** when element order may change
2. **Use unique fields** (id, email, etc.) for filtering
3. **Combine filters with capture** for subsequent value usage
4. **Use nested fields** for more precise filtering
5. **Check data types** after filtering for reliability

## Complete Example

See the complete working example in [`examples/array-filters-demo.yml`](../examples/array-filters-demo.yml).

Run:
```bash
go run main.go run examples/array-filters-demo.yml
```

## Limitations

1. Filters return the **first** found element, not all matching elements
2. If element is not found, an error is returned
3. Complex logical operations (AND, OR) are not yet supported
4. Regular expressions in filters are not yet supported

## Additional Information

- [JSONPath syntax](https://goessner.net/articles/JsonPath/)
- [Main documentation](../README.md)
- [Workflow examples](../examples/)
