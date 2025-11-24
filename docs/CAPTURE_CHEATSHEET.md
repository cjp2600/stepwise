# Cheat Sheet: Variable Capture and Validation

## 🎯 Quick Start

### Basic Example

```yaml
steps:
  # 1️⃣ Capture data
  - name: "Get User"
    request:
      method: "GET"
      url: "https://api.example.com/users/1"
    capture:
      user_id: "$.id"
      user_name: "$.name"

  # 2️⃣ Compare with captured data
  - name: "Verify User"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{user_id}}"
    validate:
      - json: "$.name"
        equals: "{{user_name}}"  # ✅ Comparison!
```

## 📦 Data Capture

### Simple Fields
```yaml
capture:
  user_id: "$.id"
  user_name: "$.name"
  user_email: "$.email"
```

### Nested Fields
```yaml
capture:
  city: "$.address.city"
  lat: "$.address.geo.lat"
```

### From Array with Filter
```yaml
capture:
  title: "$[?(@.id == 5)].title"
  body: "$[?(@.id == 5)].body"
```

### First/Last Element
```yaml
capture:
  first: "$[0]"
  last: "$[-1]"
  range: "$[0:3]"
```

## ✅ Validation with Variables

### Direct Comparison
```yaml
validate:
  - json: "$.id"
    equals: "{{saved_id}}"
  - json: "$.name"
    equals: "{{saved_name}}"
```

### Nested Fields
```yaml
validate:
  - json: "$.address.city"
    equals: "{{saved_city}}"
  - json: "$.address.geo.lat"
    equals: "{{saved_lat}}"
```

## 🔗 Chained Requests

```yaml
steps:
  # Step 1: Get ID
  - name: "Get Post"
    request:
      method: "GET"
      url: "/posts/1"
    capture:
      author_id: "$.userId"

  # Step 2: Use ID
  - name: "Get Author"
    request:
      method: "GET"
      url: "/users/{{author_id}}"
    validate:
      - json: "$.id"
        equals: "{{author_id}}"  # ✅
```

## 🧩 Components

### Component with Capture
```yaml
# components/get-user.yml
name: "Get User"
type: "step"

variables:
  user_id: "1"

steps:
  - name: "Get User"
    request:
      method: "GET"
      url: "/users/{{user_id}}"
    capture:
      user_name: "$.name"
      user_email: "$.email"
```

### Usage
```yaml
imports:
  - path: "components/get-user"
    alias: "get-user"

steps:
  - name: "Get User 5"
    use: 'get-user'
    variables:
      user_id: "5"
  
  # Variables are available!
  - name: "Verify"
    request:
      method: "GET"
      url: "/users/5"
    validate:
      - json: "$.name"
        equals: "{{user_name}}"  # ✅
```

## 🎨 JSONPath Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `$[?(@.id == 5)]` | Equality | `$[?(@.id == 5)].title` |
| `$[?(@.id > 95)]` | Greater than | `$[?(@.price > 100)]` |
| `$[?(@.id < 10)]` | Less than | `$[?(@.age < 18)]` |
| `$[0]` | First element | `$[0].name` |
| `$[-1]` | Last element | `$[-1].id` |
| `$[0:3]` | Range | `$[0:5]` |
| `$[*]` | All elements | `$[*].id` |

## 💡 Best Practices

### ✅ Good
```yaml
capture:
  saved_user_id: "$.id"
  saved_user_name: "$.name"
  original_email: "$.email"
```

### ❌ Bad
```yaml
capture:
  id: "$.id"
  n: "$.name"
  e: "$.email"
```

## 📝 Complete Example

```yaml
name: "Complete Example"
version: "1.0"

steps:
  # 1. Get posts list
  - name: "Get Posts"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts"
    capture:
      post_5_title: "$[?(@.id == 5)].title"
      post_5_user_id: "$[?(@.id == 5)].userId"

  # 2. Verify specific post
  - name: "Verify Post"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts/5"
    validate:
      - json: "$.title"
        equals: "{{post_5_title}}"  # ✅
      - json: "$.userId"
        equals: "{{post_5_user_id}}"

  # 3. Get author
  - name: "Get Author"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/{{post_5_user_id}}"
    capture:
      author_name: "$.name"
      author_city: "$.address.city"

  # 4. Use all variables
  - name: "Final Check"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
      headers:
        X-Post-Title: "{{post_5_title}}"
        X-Author-Name: "{{author_name}}"
    validate:
      - status: 200
```

## 🚀 Running Examples

```bash
# Simple example
go run main.go run examples/working-capture-compare.yml

# With components
go run main.go run examples/component-capture-workflow.yml

# Full capabilities demo
go run main.go run examples/capture-and-compare-demo.yml
```

## 📚 Additional Information

- Full guide: `docs/CAPTURE_AND_VALIDATION_GUIDE.md`
- Variables: `docs/VARIABLE_KEYS.md`
- Components: `docs/COMPONENTS.md`
