# Stepwise - API Testing Framework Specification

## Overview

Stepwise is an open-source API testing framework written in Go, inspired by Step CI. It provides a language-agnostic, universal API testing solution with support for multiple protocols and data-driven testing.

## Core Features

### 1. Language-Agnostic Configuration
- Support for YAML, JSON, and JavaScript configuration files
- Easy-to-read workflow definitions
- Modular and extensible architecture

### 2. Universal Protocol Support
- REST APIs
- GraphQL APIs
- gRPC APIs
- SOAP APIs
- WebSocket APIs
- Custom protocols via plugins

### 3. Multi-Step Workflows
- Chain requests together using captures and variables
- Conditional execution based on response data
- Parallel and sequential execution modes
- Data flow between steps

### 4. Data-Driven Testing
- Import test data from external sources
- Generate mock data using built-in generators
- Support for CSV, JSON, XML data sources
- Dynamic data injection

### 5. Validation & Assertions
- JSON Schema validation
- XML validation
- HTML validation
- Custom matchers and conditions
- Response time assertions
- Status code validation

### 6. Performance & Security
- Load testing capabilities
- SSL certificate validation
- Basic authentication support
- Cookie management
- Header manipulation

## Architecture

### Core Components

1. **Workflow Engine**
   - Parses configuration files
   - Executes workflow steps
   - Manages state and variables
   - Handles error recovery

2. **HTTP Client**
   - Configurable HTTP client
   - Support for various authentication methods
   - Cookie and session management
   - Request/response logging

3. **Validators**
   - JSON Schema validator
   - XML validator
   - Custom validation rules
   - Response assertion engine

4. **Data Generators**
   - Mock data generation
   - Faker integration
   - Custom data generators
   - Data transformation utilities

5. **Reporters**
   - Console output
   - JSON reports
   - HTML reports
   - JUnit XML format
   - Custom report formats

### Configuration Format

#### YAML Example
```yaml
name: "API Test Suite"
version: "1.0"
description: "Comprehensive API testing workflow"

variables:
  base_url: "https://api.example.com"
  api_key: "${API_KEY}"

steps:
  - name: "Health Check"
    request:
      method: "GET"
      url: "{{base_url}}/health"
    validate:
      - status: 200
      - json: "$.status"
        equals: "healthy"
      - time: "< 1000ms"

  - name: "Create User"
    request:
      method: "POST"
      url: "{{base_url}}/users"
      headers:
        Content-Type: "application/json"
        Authorization: "Bearer {{api_key}}"
      body:
        name: "{{faker.name}}"
        email: "{{faker.email}}"
    capture:
      user_id: "$.id"
    validate:
      - status: 201
      - json: "$.name"
        type: "string"
      - json: "$.email"
        pattern: "^[^@]+@[^@]+\\.[^@]+$"

  - name: "Get User"
    request:
      method: "GET"
      url: "{{base_url}}/users/{{user_id}}"
    validate:
      - status: 200
      - json: "$.id"
        equals: "{{user_id}}"
```

#### JSON Example
```json
{
  "name": "API Test Suite",
  "version": "1.0",
  "description": "Comprehensive API testing workflow",
  "variables": {
    "base_url": "https://api.example.com",
    "api_key": "${API_KEY}"
  },
  "steps": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "url": "{{base_url}}/health"
      },
      "validate": [
        {
          "status": 200
        },
        {
          "json": "$.status",
          "equals": "healthy"
        },
        {
          "time": "< 1000ms"
        }
      ]
    }
  ]
}
```

## CLI Interface

### Commands

```bash
# Initialize a new project
stepwise init

# Run a workflow
stepwise run workflow.yml

# Run with specific environment
stepwise run workflow.yml --env production

# Run with custom variables
stepwise run workflow.yml --var base_url=https://api.example.com

# Generate test data
stepwise generate --type user --count 10

# Validate workflow configuration
stepwise validate workflow.yml

# Show workflow info
stepwise info workflow.yml

# Run in watch mode
stepwise run workflow.yml --watch

# Run with parallel execution
stepwise run workflow.yml --parallel 4
```

### Options

- `--env`: Environment configuration
- `--var`: Set custom variables
- `--parallel`: Number of parallel executions
- `--timeout`: Request timeout
- `--retry`: Number of retry attempts
- `--output`: Output format (json, html, junit)
- `--verbose`: Verbose logging
- `--quiet`: Quiet mode
- `--watch`: Watch mode for file changes

## Validation Types

### Status Code Validation
```yaml
validate:
  - status: 200
  - status: [200, 201, 204]
```

### JSON Validation
```yaml
validate:
  - json: "$.status"
    equals: "success"
  - json: "$.data"
    type: "array"
  - json: "$.count"
    greater: 0
  - json: "$.email"
    pattern: "^[^@]+@[^@]+\\.[^@]+$"
```

### XML Validation
```yaml
validate:
  - xml: "/response/status"
    equals: "success"
  - xml: "/response/data"
    type: "array"
```

### Time Validation
```yaml
validate:
  - time: "< 1000ms"
  - time: "> 100ms"
  - time: "100-500ms"
```

### Custom Matchers
```yaml
validate:
  - custom: "isValidEmail"
    value: "$.email"
  - custom: "isValidUUID"
    value: "$.id"
```

## Data Generation

### Built-in Generators
```yaml
variables:
  user_name: "{{faker.name}}"
  user_email: "{{faker.email}}"
  user_phone: "{{faker.phone}}"
  user_address: "{{faker.address}}"
  random_id: "{{faker.uuid}}"
  random_number: "{{faker.number(1, 100)}}"
```

### Custom Data Sources
```yaml
data:
  users:
    source: "csv"
    file: "users.csv"
  products:
    source: "json"
    file: "products.json"
  random_data:
    source: "generator"
    type: "user"
    count: 10
```

## Plugins & Extensions

### Custom Validators
```go
type CustomValidator interface {
    Validate(value interface{}) (bool, error)
    Name() string
}
```

### Custom Data Generators
```go
type DataGenerator interface {
    Generate() interface{}
    Name() string
}
```

### Custom Reporters
```go
type Reporter interface {
    Report(results []TestResult) error
    Name() string
}
```

## Integration Support

### CI/CD Integration
- GitHub Actions
- GitLab CI
- Jenkins
- CircleCI
- Travis CI

### IDE Integration
- VS Code extension
- IntelliJ plugin
- Vim/Neovim support

### Monitoring Integration
- Prometheus metrics
- Grafana dashboards
- Alerting systems

## Performance Features

### Load Testing
```yaml
load:
  users: 100
  duration: "5m"
  ramp_up: "30s"
  target_rps: 50
```

### Parallel Execution
```yaml
execution:
  parallel: true
  max_workers: 10
  timeout: "30s"
```

### Caching
```yaml
cache:
  enabled: true
  ttl: "1h"
  strategy: "lru"
```

## Security Features

### Authentication
- Basic Auth
- Bearer Token
- OAuth 2.0
- API Key
- Custom headers

### SSL/TLS
- Certificate validation
- Custom CA certificates
- Client certificates
- SSL verification options

### Secrets Management
- Environment variables
- Vault integration
- AWS Secrets Manager
- Azure Key Vault

## Reporting

### Console Output
```
✓ Health Check (200ms)
✓ Create User (150ms)
✓ Get User (120ms)
✗ Delete User (timeout)

Summary:
- Total: 4 tests
- Passed: 3
- Failed: 1
- Duration: 470ms
```

### JSON Report
```json
{
  "summary": {
    "total": 4,
    "passed": 3,
    "failed": 1,
    "duration": "470ms"
  },
  "results": [
    {
      "name": "Health Check",
      "status": "passed",
      "duration": "200ms",
      "validations": [
        {
          "type": "status",
          "expected": 200,
          "actual": 200,
          "passed": true
        }
      ]
    }
  ]
}
```

### HTML Report
- Interactive dashboard
- Test results visualization
- Performance charts
- Error details
- Export capabilities

## Future Enhancements

### Planned Features
1. **GraphQL Support**
   - Schema introspection
   - Query validation
   - Response parsing

2. **gRPC Support**
   - Protocol buffer support
   - Streaming capabilities
   - Service discovery

3. **WebSocket Support**
   - Real-time testing
   - Message validation
   - Connection management

4. **Advanced Analytics**
   - Performance trends
   - Anomaly detection
   - Predictive analysis

5. **Distributed Testing**
   - Multi-node execution
   - Load distribution
   - Result aggregation

### Plugin Ecosystem
- Custom protocol support
- Third-party integrations
- Community extensions
- Marketplace for plugins

## Implementation Plan

### Phase 1: Core Framework
1. Basic HTTP client
2. YAML/JSON configuration parsing
3. Simple validation engine
4. CLI interface
5. Basic reporting

### Phase 2: Advanced Features
1. Multi-step workflows
2. Data generation
3. Advanced validators
4. Plugin system
5. Performance testing

### Phase 3: Enterprise Features
1. Distributed execution
2. Advanced security
3. Enterprise integrations
4. Advanced analytics
5. UI dashboard

## Contributing

### Development Setup
```bash
git clone https://github.com/stepwise/stepwise.git
cd stepwise
go mod download
go test ./...
```

### Code Standards
- Go 1.23+ required
- Follow Go coding standards
- Comprehensive test coverage
- Documentation for all public APIs
- Performance benchmarks

### Testing Strategy
- Unit tests for all components
- Integration tests for workflows
- Performance benchmarks
- Security testing
- Cross-platform testing

This specification provides a comprehensive foundation for building Stepwise, a powerful and flexible API testing framework that rivals Step CI while leveraging Go's strengths in performance, concurrency, and ecosystem integration. 