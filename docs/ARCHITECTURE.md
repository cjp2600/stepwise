# Stepwise Architecture

## Overview

Stepwise is designed as a modular, extensible API testing framework with a clear separation of concerns. The architecture follows Go best practices and is built for performance, maintainability, and extensibility.

## High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │  Workflow       │    │   HTTP Client   │
│                 │    │  Engine         │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • Command       │    │ • Parser        │    │ • Request       │
│   Handling      │    │ • Executor      │    │   Builder       │
│ • Help/Version  │    │ • Validator     │    │ • Response      │
│ • Output        │    │ • State Mgmt    │    │   Handler       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Core Services │
                    │                 │
                    ├─────────────────┤
                    │ • Config        │
                    │ • Logger        │
                    │ • Data Gen      │
                    │ • Reporters     │
                    └─────────────────┘
```

## Core Components

### 1. CLI Layer (`cmd/stepwise/`)

The CLI layer provides the user interface and command handling.

**Responsibilities:**
- Parse command-line arguments
- Handle different commands (init, run, validate, etc.)
- Format and display output
- Manage user interaction

**Key Files:**
- `cmd/stepwise/main.go` - Application entry point
- `internal/cli/app.go` - CLI application logic

### 2. Workflow Engine (`internal/workflow/`)

The workflow engine is the core component that orchestrates test execution.

**Responsibilities:**
- Parse YAML/JSON configuration files
- Execute workflow steps sequentially or in parallel
- Manage workflow state and variables
- Handle step dependencies and conditions
- Coordinate with other components

**Key Types:**
```go
type Workflow struct {
    Name        string
    Version     string
    Description string
    Variables   map[string]interface{}
    Steps       []Step
}

type Step struct {
    Name        string
    Description string
    Request     Request
    Validate    []Validation
    Capture     map[string]string
    Condition   string
}
```

### 3. HTTP Client (`internal/http/`)

The HTTP client handles all HTTP communication.

**Responsibilities:**
- Execute HTTP requests
- Handle different HTTP methods
- Manage headers and authentication
- Process responses
- Handle timeouts and retries

**Features:**
- Support for all HTTP methods (GET, POST, PUT, DELETE, etc.)
- Custom header management
- Authentication (Basic, Bearer, OAuth)
- Request/response logging
- Timeout handling
- Retry logic

### 4. Validation Engine (`internal/validation/`)

The validation engine handles all response validation.

**Responsibilities:**
- Validate HTTP status codes
- Validate JSON responses using JSONPath
- Validate XML responses using XPath
- Validate response times
- Execute custom validation functions
- Generate validation reports

**Validation Types:**
- Status code validation
- JSON path validation
- XML path validation
- Time validation
- Custom matchers
- Pattern matching
- Type checking

### 5. Data Generation (`internal/generator/`)

The data generation component provides mock data and test data management.

**Responsibilities:**
- Generate fake data (names, emails, UUIDs, etc.)
- Load data from external sources (CSV, JSON, XML)
- Transform data according to templates
- Manage data sets for data-driven testing

**Features:**
- Built-in faker functions
- Custom data generators
- External data source integration
- Template-based data transformation

### 6. Configuration Management (`internal/config/`)

The configuration component manages application settings.

**Responsibilities:**
- Load configuration from environment variables
- Provide default values
- Validate configuration
- Support different environments

**Configuration Options:**
- Log level
- Timeout settings
- Parallel execution settings
- Output format
- Environment-specific settings

### 7. Logging (`internal/logger/`)

The logging component provides structured logging throughout the application.

**Responsibilities:**
- Provide consistent logging interface
- Support different log levels
- Format log messages
- Support structured logging

**Features:**
- Multiple log levels (DEBUG, INFO, WARN, ERROR)
- Structured logging with fields
- Configurable output formats
- Performance logging

## Data Flow

### 1. Workflow Execution Flow

```
1. CLI parses command and arguments
2. Load workflow configuration file
3. Parse YAML/JSON into Workflow struct
4. Initialize Executor with configuration
5. For each step:
   a. Substitute variables
   b. Execute HTTP request
   c. Capture response data
   d. Run validations
   e. Store results
6. Generate and display report
```

### 2. Request Execution Flow

```
1. Parse request configuration
2. Substitute variables in URL, headers, body
3. Build HTTP request
4. Execute request with timeout
5. Capture response
6. Parse response body (JSON/XML)
7. Run validations
8. Capture specified values
9. Return result
```

### 3. Validation Flow

```
1. For each validation rule:
   a. Determine validation type
   b. Extract value from response
   c. Apply validation logic
   d. Compare expected vs actual
   e. Record result
2. Aggregate validation results
3. Determine overall step status
```

## Configuration Format

### YAML Structure

```yaml
name: "Workflow Name"
version: "1.0"
description: "Workflow description"

variables:
  base_url: "https://api.example.com"
  api_key: "${API_KEY}"

steps:
  - name: "Step Name"
    description: "Step description"
    request:
      method: "GET"
      url: "{{base_url}}/endpoint"
      headers:
        Authorization: "Bearer {{api_key}}"
      body:
        key: "value"
    validate:
      - status: 200
      - json: "$.status"
        equals: "success"
    capture:
      user_id: "$.user.id"
```

### JSON Structure

```json
{
  "name": "Workflow Name",
  "version": "1.0",
  "description": "Workflow description",
  "variables": {
    "base_url": "https://api.example.com",
    "api_key": "${API_KEY}"
  },
  "steps": [
    {
      "name": "Step Name",
      "description": "Step description",
      "request": {
        "method": "GET",
        "url": "{{base_url}}/endpoint",
        "headers": {
          "Authorization": "Bearer {{api_key}}"
        }
      },
      "validate": [
        {
          "status": 200
        },
        {
          "json": "$.status",
          "equals": "success"
        }
      ],
      "capture": {
        "user_id": "$.user.id"
      }
    }
  ]
}
```

## Extension Points

### 1. Custom Validators

```go
type CustomValidator interface {
    Validate(value interface{}) (bool, error)
    Name() string
}

// Example implementation
type EmailValidator struct{}

func (v *EmailValidator) Name() string {
    return "isValidEmail"
}

func (v *EmailValidator) Validate(value interface{}) (bool, error) {
    email, ok := value.(string)
    if !ok {
        return false, fmt.Errorf("value is not a string")
    }
    
    // Email validation logic
    return isValidEmail(email), nil
}
```

### 2. Custom Data Generators

```go
type DataGenerator interface {
    Generate() interface{}
    Name() string
}

// Example implementation
type UUIDGenerator struct{}

func (g *UUIDGenerator) Name() string {
    return "uuid"
}

func (g *UUIDGenerator) Generate() interface{} {
    return uuid.New().String()
}
```

### 3. Custom Reporters

```go
type Reporter interface {
    Report(results []TestResult) error
    Name() string
}

// Example implementation
type HTMLReporter struct{}

func (r *HTMLReporter) Name() string {
    return "html"
}

func (r *HTMLReporter) Report(results []TestResult) error {
    // Generate HTML report
    return nil
}
```

## Performance Considerations

### 1. Parallel Execution

- Steps can be executed in parallel when independent
- Configurable number of concurrent workers
- Resource management and limits

### 2. Caching

- HTTP response caching
- Configuration caching
- Data generation caching

### 3. Memory Management

- Streaming response processing for large responses
- Efficient JSON/XML parsing
- Memory pool for HTTP clients

## Security Features

### 1. Authentication Support

- Basic authentication
- Bearer token authentication
- OAuth 2.0 support
- API key authentication

### 2. SSL/TLS

- Certificate validation
- Custom CA certificates
- Client certificate support
- SSL verification options

### 3. Secrets Management

- Environment variable support
- External secrets management integration
- Secure credential storage

## Testing Strategy

### 1. Unit Tests

- Each component has comprehensive unit tests
- Mock interfaces for external dependencies
- High test coverage (>90%)

### 2. Integration Tests

- End-to-end workflow testing
- Real API integration tests
- Performance benchmarks

### 3. Test Data

- Mock APIs for testing
- Sample workflows for validation
- Performance test suites

## Deployment

### 1. Binary Distribution

- Single binary for easy deployment
- Cross-platform compilation
- Minimal dependencies

### 2. Docker Support

- Dockerfile for containerized deployment
- Multi-stage builds for optimization
- Docker Compose for development

### 3. CI/CD Integration

- GitHub Actions workflows
- GitLab CI configuration
- Jenkins pipeline support

## Monitoring and Observability

### 1. Metrics

- Request/response metrics
- Performance metrics
- Error rates and types

### 2. Logging

- Structured logging
- Log levels and filtering
- Log aggregation support

### 3. Tracing

- Request tracing
- Step execution tracing
- Performance profiling

## Future Enhancements

### 1. Plugin System

- Dynamic plugin loading
- Plugin marketplace
- Custom protocol support

### 2. Distributed Execution

- Multi-node execution
- Load distribution
- Result aggregation

### 3. UI Dashboard

- Web-based interface
- Real-time monitoring
- Visual workflow editor

This architecture provides a solid foundation for building a powerful, extensible API testing framework that can grow with user needs while maintaining simplicity and performance. 