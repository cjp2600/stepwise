# Stepwise - API Testing Framework

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/cjp2600/stepwise)](https://goreportcard.com/report/github.com/cjp2600/stepwise)

Stepwise is an open-source API testing framework written in Go, inspired by [Step CI](https://stepci.com/). It provides a language-agnostic, universal API testing solution with support for multiple protocols and data-driven testing.

## Features

- **Language-Agnostic Configuration**: Support for YAML, JSON, and JavaScript configuration files
- **Universal Protocol Support**: REST, GraphQL, gRPC, SOAP, and WebSocket APIs
- **Multi-Step Workflows**: Chain requests together using captures and variables
- **Advanced Component System**: Reusable components with imports, overrides, cycle detection, and Flow components
- **Recursive Execution**: Control over directory search with `-r` flag
- **Beautiful Spinners**: Animated progress indicators with CI-aware behavior
- **Data-Driven Testing**: Import test data or generate mock data
- **Variable Keys Support**: Use variables in JSON object keys for dynamic key generation
- **Comprehensive Validation**: JSON Schema, XML, HTML validation with custom matchers
- **Performance Testing**: Load testing capabilities with parallel execution
- **Security Features**: SSL certificate validation, authentication support
- **CI/CD Integration**: Works with GitHub Actions, GitLab CI, and more
- **Colorful Output**: Colored terminal output with CI/CD compatibility
- **Verbose Logging**: Detailed debug information with `--verbose` flag

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/cjp2600/stepwise.git
cd stepwise

# Install dependencies
go mod download

# Build the application
go build -o stepwise cmd/stepwise/main.go

# Or install globally
go install ./cmd/stepwise
```

### Initialize a Project

```bash
# Create a new Stepwise project
stepwise init
```

This creates a `workflow.yml` file with a basic example.

### Run Your First Test

```bash
# Run the example workflow
stepwise run workflow.yml

# Run all workflows in a directory (non-recursive)
stepwise run examples/

# Run all workflows in a directory recursively
stepwise run examples/ -r

# Run with beautiful spinners
stepwise run workflow.yml --verbose
```

## Configuration

### Basic Workflow Example

```yaml
name: "API Test Suite"
version: "1.0"
description: "A simple example workflow"

variables:
  base_url: "https://jsonplaceholder.typicode.com"
  timeout: "5s"

steps:
  - name: "Health Check"
    request:
      method: "GET"
      url: "{{base_url}}/posts/1"
      headers:
        Accept: "application/json"
    validate:
      - status: 200
      - json: "$.id"
        equals: 1
      - time: "< 2000ms"
    capture:
      post_id: "$.id"
```

### Component System Example

```yaml
# components/http-get-step.yml
name: "HTTP GET Step"
version: "1.0"
type: "step"
variables:
  base_url: "https://httpbin.org"
steps:
  - name: "HTTP GET Request"
    request:
      method: "GET"
      url: "{{base_url}}/get"
    validate:
      - status: 200
    capture:
      status: "$.status"
```

```yaml
# workflow.yml
name: "Component Test"
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
```

### Flow Components Example

Flow components allow you to organize complex test scenarios into logical groups:

```yaml
# components/flows/customer-flow.yml
name: "Customer Flow"
version: "1.0"
description: "Complete customer onboarding flow"
type: "workflow"

imports:
  - path: "../create-customer"
    alias: "create-customer"
  - path: "../get-token"
    alias: "get-token"
  - path: "../check-customer-info"
    alias: "check-customer-info"

steps:
  - name: "Create new customer"
    use: "create-customer"
  
  - name: "Authenticate customer"
    use: "get-token"
  
  - name: "Verify customer information"
    use: "check-customer-info"
```

```yaml
# workflow.yml
name: "Complete Purchase Test"
version: "1.0"

imports:
  - path: "./components/flows/customer-flow"
    alias: "customer-flow"

steps:
  - name: "Customer Onboarding"
    use: "customer-flow"
  
  - name: "Test Summary"
    print: "Customer flow completed successfully!"
```

### Advanced Workflow Example

```yaml
name: "Advanced API Test Suite"
version: "1.0"
description: "Advanced workflow with authentication and complex validations"

variables:
  base_url: "https://api.github.com"
  api_token: "${GITHUB_TOKEN}"

steps:
  - name: "Authenticate User"
    request:
      method: "GET"
      url: "{{base_url}}/user"
      headers:
        Authorization: "Bearer {{api_token}}"
        Accept: "application/vnd.github.v3+json"
    validate:
      - status: 200
      - json: "$.login"
        type: "string"
    capture:
      user_login: "$.login"

  - name: "Create Repository"
    request:
      method: "POST"
      url: "{{base_url}}/user/repos"
      headers:
        Authorization: "Bearer {{api_token}}"
        Content-Type: "application/json"
      body:
        name: "test-repo-{{faker.uuid}}"
        private: true
    validate:
      - status: 201
      - json: "$.private"
        equals: true
```

## CLI Commands

### Basic Commands

```bash
# Initialize a new project
stepwise init

# Run a workflow
stepwise run workflow.yml

# Validate a workflow
stepwise validate workflow.yml

# Show workflow information
stepwise info workflow.yml

# Generate test data
stepwise generate --type user --count 10
```

### Advanced Options

```bash
# Run with environment variables
stepwise run workflow.yml --env production

# Set custom variables
stepwise run workflow.yml --var base_url=https://api.example.com

# Run with parallel execution
stepwise run workflow.yml --parallel 4

# Run with custom timeout
stepwise run workflow.yml --timeout 30s

# Generate different output formats
stepwise run workflow.yml --output json
stepwise run workflow.yml --output html

# Watch mode for development
stepwise run workflow.yml --watch

# Run with verbose logging
stepwise run workflow.yml --verbose

# Disable colors for CI
NO_COLOR=1 stepwise run workflow.yml
```

## Component System

Stepwise supports reusable components and templates to eliminate code duplication and promote maintainability.

### Creating Components

Create reusable components in any directory:

```yaml
# components/auth/basic-auth.yml
name: "Basic Authentication"
version: "1.0"
description: "Reusable basic authentication step"
type: "step"

variables:
  auth_username: "${AUTH_USERNAME}"
  auth_password: "${AUTH_PASSWORD}"

steps:
  - name: "Basic Auth Login"
    request:
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

### Using Components

Import components in your workflows:

```yaml
name: "API Test Suite"
version: "1.0"

imports:
  # Basic import
  - path: "components/auth/basic-auth"
  
  # Import with alias
  - path: "components/auth/basic-auth"
    alias: "User Login"
  
  # Import with variable overrides
  - path: "components/auth/basic-auth"
    variables:
      auth_url: "https://custom-auth.com"
  
  # Import with request overrides
  - path: "components/auth/basic-auth"
    overrides:
      request:
        url: "{{custom_url}}/login"
        headers:
          X-Custom-Header: "{{custom_value}}"

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

### Component Types

- **Step Components** (`type: "step"`): Single reusable steps
- **Group Components** (`type: "group"`): Groups of related steps
- **Workflow Components** (`type: "workflow"`): Complete workflows

### Component Search Paths

Stepwise searches for components in:
1. Current directory (`.`)
2. `./components` directory
3. `./templates` directory
4. `./examples/templates` directory
5. Custom search paths

### Example Templates

See `examples/templates/` for ready-to-use templates:
- `jsonplaceholder-api.yml` - JSONPlaceholder API testing
- `github-api.yml` - GitHub API testing
- `httpbin-api.yml` - HTTPBin API testing

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
```

## Examples

See the `examples/` directory for complete workflow examples:

### Basic Examples
- `examples/basic-workflow.yml` - Basic API testing workflow
- `examples/advanced-workflow.yml` - Advanced workflow with authentication
- `examples/simple-test.yml` - Simple HTTP testing

### Component Examples
- `examples/simple-template-demo.yml` - Using imported templates
- `examples/use-httpbin-template.yml` - HTTPBin API template usage
- `examples/use-jsonplaceholder-template.yml` - JSONPlaceholder API template usage

### Mixed Protocol Examples
- `examples/demo-mixed-workflow.yml` - HTTP and gRPC testing
- `examples/mixed-protocol-test.yml` - Mixed protocol workflow

### Templates
- `examples/templates/httpbin-api.yml` - HTTPBin API template
- `examples/templates/jsonplaceholder-api.yml` - JSONPlaceholder API template
- `examples/templates/github-api.yml` - GitHub API template

## Architecture

### Core Components

1. **Workflow Engine**: Parses configuration files and executes workflow steps
2. **HTTP Client**: Configurable HTTP client with authentication support
3. **Validators**: JSON Schema, XML, and custom validation rules
4. **Data Generators**: Mock data generation and external data sources
5. **Reporters**: Console, JSON, HTML, and JUnit XML output formats

### Project Structure

```
stepwise/
├── cmd/stepwise/          # CLI application
├── internal/              # Internal packages
│   ├── cli/              # CLI handling
│   ├── config/           # Configuration management
│   ├── logger/           # Logging functionality
│   └── workflow/         # Workflow execution engine
├── components/            # Reusable components
│   ├── auth/             # Authentication components
│   ├── common/           # Common utilities
│   └── api/              # API-specific components
├── examples/             # Example workflows
│   └── templates/        # Example templates
├── docs/                # Documentation
└── tests/               # Test files
```

## Development

### Prerequisites

- Go 1.23 or higher
- Git

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/cjp2600/stepwise.git
cd stepwise

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the application
go build -o stepwise cmd/stepwise/main.go
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code Style

This project follows Go coding standards:

- Use `gofmt` for code formatting
- Follow Go naming conventions
- Write comprehensive tests
- Document all public APIs

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Code Standards

- Follow Go coding standards
- Write comprehensive tests
- Document all public APIs
- Keep commits atomic and well-described

## Roadmap

### Phase 1: Core Framework ✅
- [x] Basic HTTP client
- [x] YAML/JSON configuration parsing
- [x] Simple validation engine
- [x] CLI interface
- [x] Basic reporting
- [x] gRPC support
- [x] Colorful output with CI compatibility
- [x] Verbose logging

### Phase 2: Advanced Features ✅
- [x] Multi-step workflows
- [x] Component system with imports
- [x] Advanced validators
- [x] Recursive directory execution
- [x] Performance testing
- [x] Proper exit codes for CI/CD

### Phase 3: Enterprise Features 📋
- [ ] Distributed execution
- [ ] Advanced security
- [ ] Enterprise integrations
- [ ] Advanced analytics
- [ ] UI dashboard

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [Step CI](https://stepci.com/)
- Built with Go and modern testing practices
- Community-driven development

## Documentation

- **[Getting Started](README.md)** - This guide
- **[CLI Reference](docs/CLI.md)** - Complete command-line interface guide
- **[Component System](docs/COMPONENTS.md)** - Reusable components and templates
- **[Import System](docs/IMPORTS.md)** - Advanced import functionality
- **[API Reference](docs/API.md)** - Complete API documentation

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/cjp2600/stepwise/issues)
- **Discussions**: [GitHub Discussions](https://github.com/cjp2600/stepwise/discussions)
- **Email**: team@stepwise.dev

---

**Stepwise** - Making API testing simple, powerful, and accessible. 

## Установка CLI

Установить CLI можно одной командой:

```sh
go install github.com/cjp2600/stepwise@latest
```

После этого бинарник появится в вашем $GOPATH/bin или $GOBIN.

## Использование

```sh
stepwise [аргументы]
``` 