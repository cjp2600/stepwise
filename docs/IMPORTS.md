# Stepwise Import System

## Overview

The Stepwise import system allows you to create reusable components and import them into your workflows. This eliminates code duplication and promotes maintainability.

## Component Types

### 1. Step Components (`type: "step"`)
Single reusable steps that can be imported and customized.

```yaml
name: "Basic Authentication"
version: "1.0"
description: "Reusable basic authentication step"
type: "step"

variables:
  auth_username: "${AUTH_USERNAME}"
  auth_password: "${AUTH_PASSWORD}"

steps:
  - name: "Basic Auth Login"
    description: "Authenticate using basic authentication"
    request:
      protocol: "http"
      method: "POST"
      url: "{{auth_url}}/login"
      headers:
        Content-Type: "application/json"
      body:
        username: "{{auth_username}}"
        password: "{{auth_password}}"
    validate:
      - status: 200
    capture:
      auth_token: "$.token"
```

### 2. Group Components (`type: "group"`)
Groups of related steps that can be executed together.

```yaml
name: "User CRUD Operations"
version: "1.0"
description: "Complete user CRUD operations"
type: "group"

variables:
  user_api_base: "/users"

steps:
  - name: "Create User"
    request:
      method: "POST"
      url: "{{base_url}}{{user_api_base}}"
      body:
        name: "{{user_name}}"
        email: "{{user_email}}"
    validate:
      - status: 201

  - name: "Get User"
    request:
      method: "GET"
      url: "{{base_url}}{{user_api_base}}/{{user_id}}"
    validate:
      - status: 200
```

### 3. Workflow Components (`type: "workflow"`)
Complete workflows that can be imported and extended.

```yaml
name: "API Test Suite"
version: "1.0"
description: "Complete API test suite"
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

## Importing Components

### Basic Import
```yaml
imports:
  - path: "components/auth/basic-auth"
```

### Import with Alias
```yaml
imports:
  - path: "components/auth/basic-auth"
    alias: "User Login"
```

### Import with Variable Overrides
```yaml
imports:
  - path: "components/auth/basic-auth"
    alias: "Admin Login"
    variables:
      auth_url: "https://admin.example.com"
      auth_username: "${ADMIN_USERNAME}"
```

### Import with Request Overrides
```yaml
imports:
  - path: "components/auth/basic-auth"
    alias: "Custom Login"
    overrides:
      name: "Custom Authentication"
      request:
        url: "{{custom_auth_url}}/login"
        headers:
          X-Custom-Header: "{{custom_value}}"
```

## Component Search Paths

Stepwise searches for components in the following order:

1. Current directory (`.`)
2. `./components` directory
3. `./templates` directory
4. Custom search paths specified in `LoadWithImports()`

## File Structure

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
├── workflows/
│   ├── main-test.yml
│   └── integration-test.yml
└── templates/
    └── base-workflow.yml
```

## Examples

### Example 1: Authentication Workflow
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
        Authorization: "Bearer {{auth_token}}"
    validate:
      - status: 200
```

### Example 2: API Health Check
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

### Example 3: Complex Workflow with Multiple Imports
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
          Authorization: "Bearer {{auth_token}}"

steps:
  - name: "Custom Validation"
    request:
      method: "GET"
      url: "{{base_url}}/users/{{user_id}}/profile"
    validate:
      - status: 200
```

## Best Practices

### 1. Component Design
- Keep components focused and single-purpose
- Use descriptive names and descriptions
- Include proper validation rules
- Document expected variables and outputs

### 2. Variable Management
- Use environment variables for sensitive data
- Provide sensible defaults where possible
- Document all required variables

### 3. Versioning
- Use semantic versioning for components
- Maintain backward compatibility
- Document breaking changes

### 4. Organization
- Group related components in directories
- Use consistent naming conventions
- Keep components small and focused

## Advanced Features

### Conditional Imports
```yaml
imports:
  - path: "components/auth/basic-auth"
    condition: "{{use_basic_auth}}"
  - path: "components/auth/oauth2"
    condition: "{{use_oauth2}}"
```

### Version-Specific Imports
```yaml
imports:
  - path: "components/auth/basic-auth"
    version: "1.2.0"
```

### Multiple Component Types
```yaml
imports:
  - path: "components/auth/basic-auth"
    type: "step"
  - path: "components/api/user-operations"
    type: "workflow"
```

## Troubleshooting

### Common Issues

1. **Component Not Found**
   - Check the component path
   - Verify the component file exists
   - Check search paths configuration

2. **Variable Resolution Errors**
   - Ensure all required variables are defined
   - Check variable naming consistency
   - Verify environment variables are set

3. **Validation Failures**
   - Check response format matches expectations
   - Verify JSON paths are correct
   - Ensure validation rules match actual responses

### Debug Tips

1. Use verbose mode to see detailed execution:
   ```bash
   ./stepwise run workflow.yml --verbose
   ```

2. Check component loading:
   ```bash
   ./stepwise validate component.yml
   ```

3. Test individual components:
   ```bash
   ./stepwise run component.yml
   ``` 