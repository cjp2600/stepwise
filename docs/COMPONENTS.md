# Stepwise Components System

## Overview

Stepwise provides a powerful component system that allows you to create reusable workflow components. Components can be imported into workflows and provide a way to share common functionality across multiple test scenarios.

## Component Types

### 1. Step Components

Step components contain a single step that can be reused across different workflows.

```yaml
name: "HTTP GET Step"
version: "1.0"
description: "Reusable HTTP GET request step"
type: "step"
variables:
  base_url: "https://httpbin.org"
captures:
  response_status: "$.status"
  response_time: "$.duration"

steps:
  - name: "HTTP GET Request"
    description: "Performs a GET request and captures response data"
    request:
      method: "GET"
      url: "{{base_url}}/get"
      headers:
        User-Agent: "Stepwise/1.0"
    validate:
      - status: 200
      - time: "< 5000ms"
    capture:
      status: "$.status"
      url: "$.url"
      headers: "$.headers"
```

### 2. Group Components

Group components contain multiple steps that work together as a logical unit.

```yaml
name: "Authentication Group"
version: "1.0"
description: "Reusable authentication workflow group"
type: "group"
variables:
  api_base_url: "https://api.example.com"
  username: "testuser"
  password: "testpass"
captures:
  auth_token: "$.token"
  user_id: "$.user_id"

steps:
  - name: "Login User"
    description: "Authenticate user and get token"
    request:
      method: "POST"
      url: "{{api_base_url}}/auth/login"
      headers:
        Content-Type: "application/json"
      body:
        username: "{{username}}"
        password: "{{password}}"
    validate:
      - status: 200
      - time: "< 3000ms"
    capture:
      token: "$.token"
      user_id: "$.user_id"

  - name: "Validate Token"
    description: "Validate the received token"
    request:
      method: "GET"
      url: "{{api_base_url}}/auth/validate"
      headers:
        Authorization: "Bearer {{token}}"
    validate:
      - status: 200
    capture:
      is_valid: "$.valid"
```

### 3. Workflow Components

Workflow components contain complete workflows that can be imported and extended.

```yaml
name: "API Test Workflow"
version: "1.0"
description: "Complete API testing workflow"
type: "workflow"
variables:
  api_base_url: "https://api.example.com"
captures:
  user_id: "$.user_id"
  post_id: "$.post_id"

imports:
  - path: "auth-group"
    alias: "Authentication"
    variables:
      api_base_url: "{{api_base_url}}"

steps:
  - name: "Create User"
    description: "Create a new test user"
    request:
      method: "POST"
      url: "{{api_base_url}}/users"
      headers:
        Content-Type: "application/json"
      body:
        name: "Test User"
        email: "test@example.com"
    validate:
      - status: 201
    capture:
      user_id: "$.id"
```

## Using Components in Workflows

### Basic Import

```yaml
name: "My Test Workflow"
version: "1.0"
variables:
  base_url: "https://httpbin.org"

imports:
  - path: "./components/http-get-step"
    alias: "Get Request"
    variables:
      base_url: "{{base_url}}"

steps:
  - name: "Test Request"
    use: "Get Request"
    validate:
      - status: 200
    capture:
      response_data: "$"
```

### Import with Overrides

```yaml
imports:
  - path: "./components/http-post-step"
    alias: "Custom POST"
    variables:
      base_url: "{{base_url}}"
    overrides:
      name: "Custom POST Request"
      request:
        body:
          message: "Hello from Stepwise"
          timestamp: "{{timestamp}}"
```

### Import with Variables

```yaml
imports:
  - path: "./components/auth-group"
    alias: "User Authentication"
    variables:
      api_base_url: "https://api.example.com"
      username: "demo_user"
      password: "demo_pass"
```

## Component Features

### Variable Overrides

Components can accept variable overrides when imported:

```yaml
imports:
  - path: "my-component"
    variables:
      base_url: "https://custom-api.com"
      timeout: "10s"
```

### Request Overrides

You can override specific parts of requests:

```yaml
imports:
  - path: "my-component"
    overrides:
      name: "Custom Name"
      request:
        url: "https://custom-url.com/api"
        headers:
          Authorization: "Bearer {{token}}"
```

### Captures

Components can define global captures that are available throughout the workflow:

```yaml
captures:
  auth_token: "$.token"
  user_id: "$.user_id"
  response_time: "$.duration"
```

## Component Search Paths

Stepwise searches for components in the following order:

1. Current directory
2. `./components`
3. `./templates`
4. `./workflows`
5. `./steps`
6. Custom search paths specified in the workflow

## Best Practices

### 1. Component Organization

- Keep components in a dedicated `components/` directory
- Use descriptive names for components
- Group related components together

### 2. Variable Management

- Use variables for configurable values
- Provide sensible defaults
- Document required variables

### 3. Reusability

- Make components generic and reusable
- Avoid hardcoding specific values
- Use templates for dynamic content

### 4. Versioning

- Include version information in components
- Use semantic versioning
- Document breaking changes

## Example Component Structure

```
components/
├── http-get-step.yml
├── http-post-step.yml
├── auth-group.yml
├── api-test-workflow.yml
└── templates/
    ├── github-api.yml
    ├── httpbin-api.yml
    └── jsonplaceholder-api.yml
```

## Advanced Features

### Circular Import Detection

Stepwise automatically detects and prevents circular imports between components.

### Component Caching

Components are cached after loading to improve performance.

### Variable Inheritance

Variables from imported components are merged with the parent workflow's variables.

### Capture Propagation

Captures from components are available in the parent workflow and can be used by subsequent steps. 
