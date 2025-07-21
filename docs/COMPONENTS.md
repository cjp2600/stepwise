# Stepwise Component System

## Overview

The Stepwise component system allows you to create reusable templates and components that can be imported into your workflows. This eliminates code duplication and promotes maintainability.

## Quick Start

### 1. Create a Component

```yaml
# components/auth/login.yml
name: "User Login"
version: "1.0"
description: "Reusable login component"
type: "step"

variables:
  username: "${USERNAME}"
  password: "${PASSWORD}"

steps:
  - name: "Login"
    request:
      method: "POST"
      url: "{{auth_url}}/login"
      body:
        username: "{{username}}"
        password: "{{password}}"
    validate:
      - status: 200
    capture:
      token: "$.token"
```

### 2. Use the Component

```yaml
# workflow.yml
name: "My Workflow"
version: "1.0"

imports:
  - path: "components/auth/login"
    alias: "User Authentication"
    variables:
      auth_url: "https://api.example.com"

steps:
  - name: "Test Protected Endpoint"
    request:
      method: "GET"
      url: "{{api_url}}/protected"
      headers:
        Authorization: "Bearer {{token}}"
    validate:
      - status: 200
```

## Component Types

### Step Components

Single reusable steps that can be imported and customized.

**New: Step Delay (wait)**

You can add a wait (delay) before a step executes:

```yaml
steps:
  - name: "Wait 2 seconds"
    wait: "2s"
  - name: "Check API"
    request:
      method: "GET"
      url: "https://example.com"
    validate:
      - status: 200
```

- `wait`: (optional) Duration to wait before executing the step. Supports Go duration format (e.g., "2s", "500ms").

```yaml
name: "Health Check"
version: "1.0"
type: "step"

variables:
  endpoint: "/health"

steps:
  - name: "API Health Check"
    request:
      method: "GET"
      url: "{{base_url}}{{endpoint}}"
    validate:
      - status: 200
```

### Group Components

Groups of related steps that can be executed together.

```yaml
name: "User CRUD Operations"
version: "1.0"
type: "group"

variables:
  user_api: "/users"

steps:
  - name: "Create User"
    request:
      method: "POST"
      url: "{{base_url}}{{user_api}}"
      body:
        name: "{{user_name}}"
        email: "{{user_email}}"
    validate:
      - status: 201

  - name: "Get User"
    request:
      method: "GET"
      url: "{{base_url}}{{user_api}}/{{user_id}}"
    validate:
      - status: 200
```

### Workflow Components

Complete workflows that can be imported and extended.

```yaml
name: "API Test Suite"
version: "1.0"
type: "workflow"

variables:
  api_base: "https://api.example.com"

steps:
  - name: "Health Check"
    request:
      method: "GET"
      url: "{{api_base}}/health"
    validate:
      - status: 200

groups:
  - name: "Authentication Tests"
    parallel: true
    steps:
      - name: "Login Test"
        request:
          method: "POST"
          url: "{{api_base}}/login"
        validate:
          - status: 200
```

## Import Options

### Basic Import

```yaml
imports:
  - path: "components/auth/login"
```

### Import with Alias

```yaml
imports:
  - path: "components/auth/login"
    alias: "User Authentication"
```

### Import with Variable Overrides

```yaml
imports:
  - path: "components/auth/login"
    variables:
      auth_url: "https://custom-auth.com"
      username: "${CUSTOM_USERNAME}"
```

### Import with Request Overrides

```yaml
imports:
  - path: "components/auth/login"
    overrides:
      request:
        url: "{{custom_url}}/login"
        headers:
          X-Custom-Header: "{{custom_value}}"
```

### Import with Version

```yaml
imports:
  - path: "components/auth/login"
    version: "1.2.0"
```

## Component Search Paths

Stepwise searches for components in the following order:

1. **Current directory** (`.`)
2. **`./components`** directory
3. **`./templates`** directory
4. **`./examples/templates`** directory
5. **Custom search paths** specified in `LoadWithImports()`

### File Structure Example

```
project/
├── components/
│   ├── auth/
│   │   ├── basic-auth.yml
│   │   ├── bearer-token.yml
│   │   └── oauth2.yml
│   ├── common/
│   │   ├── health-check.yml
│   │   └── setup.yml
│   └── api/
│       ├── user-operations.yml
│       └── post-operations.yml
├── examples/
│   └── templates/
│       ├── httpbin-api.yml
│       ├── jsonplaceholder-api.yml
│       └── github-api.yml
└── workflows/
    ├── main-test.yml
    └── integration-test.yml
```

## Best Practices

### 1. Component Design

- **Keep components focused**: Each component should have a single responsibility
- **Use descriptive names**: Make component names clear and meaningful
- **Include proper validation**: Always validate responses in components
- **Document variables**: Clearly document all required variables

### 2. Variable Management

- **Use environment variables** for sensitive data
- **Provide sensible defaults** where possible
- **Document all required variables**
- **Use consistent naming conventions**

### 3. Versioning

- **Use semantic versioning** for components
- **Maintain backward compatibility**
- **Document breaking changes**
- **Test components independently**

### 4. Organization

- **Group related components** in directories
- **Use consistent naming conventions**
- **Keep components small and focused**
- **Create reusable patterns**

## Advanced Features

### Conditional Imports

```yaml
imports:
  - path: "components/auth/basic-auth"
    condition: "{{use_basic_auth}}"
  - path: "components/auth/oauth2"
    condition: "{{use_oauth2}}"
```

### Multiple Component Types

```yaml
imports:
  - path: "components/auth/basic-auth"
    type: "step"
  - path: "components/api/user-operations"
    type: "workflow"
```

### Component Composition

```yaml
# components/api/user-suite.yml
name: "User API Suite"
type: "workflow"

imports:
  - path: "components/auth/basic-auth"
  - path: "components/api/user-operations"

steps:
  - name: "Custom User Test"
    request:
      method: "GET"
      url: "{{api_url}}/users/{{user_id}}/profile"
```

## Troubleshooting

### Common Issues

1. **Component Not Found**
   ```
   Error: component not found: components/auth/login
   ```
   - Check the component path
   - Verify the component file exists
   - Check search paths configuration

2. **Variable Resolution Errors**
   ```
   Error: variable not found: auth_url
   ```
   - Ensure all required variables are defined
   - Check variable naming consistency
   - Verify environment variables are set

3. **Validation Failures**
   ```
   Error: validation failed: expected 200, got 404
   ```
   - Check response format matches expectations
   - Verify JSON paths are correct
   - Ensure validation rules match actual responses

### Debug Tips

1. **Use verbose mode** to see detailed execution:
   ```bash
   ./stepwise run workflow.yml --verbose
   ```

2. **Check component loading**:
   ```bash
   ./stepwise validate component.yml
   ```

3. **Test individual components**:
   ```bash
   ./stepwise run component.yml
   ```

4. **Check search paths**:
   ```bash
   # Add debug logging to see search paths
   ./stepwise run workflow.yml --verbose
   ```

## Examples

### Authentication Workflow

```yaml
name: "Authentication Test"
version: "1.0"

variables:
  auth_url: "https://auth.example.com"
  api_url: "https://api.example.com"

imports:
  - path: "components/auth/basic-auth"
    alias: "Login"
    variables:
      auth_url: "{{auth_url}}"

steps:
  - name: "Test Protected Endpoint"
    request:
      method: "GET"
      url: "{{api_url}}/protected"
      headers:
        Authorization: "Bearer {{token}}"
    validate:
      - status: 200
```

### API Health Check

```yaml
name: "API Health Check"
version: "1.0"

variables:
  base_url: "https://api.example.com"

imports:
  - path: "components/common/health-check"
    alias: "Health Check"
    variables:
      health_endpoint: "/status"

steps:
  - name: "Additional Check"
    request:
      method: "GET"
      url: "{{base_url}}/version"
    validate:
      - status: 200
```

### Complex Workflow with Multiple Imports

```yaml
name: "Complete API Test Suite"
version: "1.0"

variables:
  base_url: "https://api.example.com"
  auth_url: "https://auth.example.com"

imports:
  - path: "components/auth/basic-auth"
    alias: "Authentication"
    variables:
      auth_url: "{{auth_url}}"

  - path: "components/common/health-check"
    alias: "Health Check"
    variables:
      health_endpoint: "/health"

  - path: "components/api/user-operations"
    alias: "User Management"
    variables:
      user_api_base: "/api/v1/users"
    overrides:
      request:
        headers:
          Authorization: "Bearer {{token}}"

steps:
  - name: "Custom Validation"
    request:
      method: "GET"
      url: "{{base_url}}/users/{{user_id}}/profile"
    validate:
      - status: 200
```

## Migration Guide

### From Duplicated Code to Components

**Before (Duplicated Code):**
```yaml
# workflow1.yml
steps:
  - name: "Login"
    request:
      method: "POST"
      url: "https://auth.example.com/login"
      body:
        username: "user1"
        password: "pass1"
    validate:
      - status: 200

# workflow2.yml
steps:
  - name: "Login"
    request:
      method: "POST"
      url: "https://auth.example.com/login"
      body:
        username: "user2"
        password: "pass2"
    validate:
      - status: 200
```

**After (Using Components):**
```yaml
# components/auth/login.yml
name: "Login"
type: "step"
variables:
  username: "${USERNAME}"
  password: "${PASSWORD}"
steps:
  - name: "Login"
    request:
      method: "POST"
      url: "{{auth_url}}/login"
      body:
        username: "{{username}}"
        password: "{{password}}"
    validate:
      - status: 200

# workflow1.yml
imports:
  - path: "components/auth/login"
    variables:
      username: "user1"
      password: "pass1"

# workflow2.yml
imports:
  - path: "components/auth/login"
    variables:
      username: "user2"
      password: "pass2"
```

This approach eliminates duplication and makes maintenance much easier. 

### Validation Rules

You can now use advanced validation rules for JSON values:

```yaml
validate:
  - json: "$.data.items"
    empty: false   # Проверка, что массив не пустой
  - json: "$.data.items"
    len: 3         # Проверка длины массива
  - json: "$.optional"
    nil: true      # Проверка, что значение nil (отсутствует)
```

- `empty: true|false` — значение пустое/не пустое (строка, массив, map, nil)
- `nil: true|false` — значение nil/не nil
- `len: N` — длина значения (строка, массив, map) равна N 

### Advanced Validation: base64 JSON decode

You can validate fields that are base64-encoded JSON using:

```yaml
validate:
  - json: "$.widgets[0].widget"
    decode: "base64json"
    jsonpath: "$.title"
    equals: "PetShop"
```

- `decode: base64json` — decode the field as base64 and parse as JSON
- `jsonpath` — path inside the decoded JSON
 