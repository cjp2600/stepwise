# Stepwise CLI Reference

## Overview

The Stepwise CLI provides a powerful command-line interface for running, validating, and managing API test workflows.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/cjp2600/stepwise.git
cd stepwise

# Build the application
go build -o stepwise cmd/stepwise/main.go

# Or install globally
go install ./cmd/stepwise
```

### Binary Download

```bash
# Download the latest release
curl -L https://github.com/cjp2600/stepwise/releases/latest/download/stepwise-darwin-amd64 -o stepwise
chmod +x stepwise
```

## Basic Commands

### `stepwise run`

Execute a workflow file or directory.

```bash
# Run a single workflow file
stepwise run workflow.yml

# Run all workflows in a directory (non-recursive)
stepwise run test-workflows/

# Run all workflows in a directory recursively
stepwise run test-workflows/ -r

# Run with verbose output
stepwise run workflow.yml --verbose

# Run with custom variables
stepwise run workflow.yml --var base_url=https://api.example.com

# Run with environment file
stepwise run workflow.yml --env production
```

### `stepwise validate`

Validate workflow syntax and configuration.

```bash
# Validate a single workflow
stepwise validate workflow.yml

# Validate all workflows in directory
stepwise validate test-workflows/

# Validate with verbose output
stepwise validate workflow.yml --verbose
```

### `stepwise info`

Display information about a workflow.

```bash
# Show workflow information
stepwise info workflow.yml

# Show detailed information
stepwise info workflow.yml --verbose
```

## Command Options

### Global Options

```bash
--verbose, -v          Enable verbose logging
--no-color            Disable colored output
--log-level LEVEL     Set log level (debug, info, warn, error)
--timeout DURATION    Set global timeout (e.g., 30s, 5m)
```

### Run Command Options

```bash
stepwise run [OPTIONS] <FILE|DIR>

Options:
  --env ENV              Environment name (loads .env.ENV file)
  --var KEY=VALUE        Set custom variables
  --parallel N           Run workflows in parallel (default: 1)
  -r, --recursive        Search recursively in subdirectories
  --output FORMAT        Output format (text, json, html, junit)
  --watch                Watch for file changes and re-run
  --dry-run             Validate without executing
  --continue-on-error   Continue execution on validation errors
```

### Validate Command Options

```bash
stepwise validate [OPTIONS] <FILE|DIR>

Options:
  --strict              Enable strict validation
  --schema-only         Validate schema only (skip network tests)
  --output FORMAT       Output format (text, json)
```

## Environment Variables

### Stepwise Configuration

```bash
# Disable colored output
NO_COLOR=1

# Set log level
STEPWISE_LOG_LEVEL=debug

# Set default timeout
STEPWISE_TIMEOUT=30s

# Set parallel execution
STEPWISE_PARALLEL=4
```

### CI/CD Integration

```bash
# Detect CI environment
CI=1

# Set output format for CI
STEPWISE_OUTPUT=junit

# Disable colors in CI
NO_COLOR=1
```

## Examples

### Basic Usage

```bash
# Run a simple workflow
stepwise run examples/simple-test.yml

# Run with verbose output
stepwise run examples/simple-test.yml --verbose

# Run with custom variables
stepwise run examples/simple-test.yml \
  --var base_url=https://api.example.com \
  --var api_key=your-api-key
```

### Advanced Usage

```bash
# Run multiple workflows in parallel (non-recursive)
stepwise run test-workflows/ --parallel 4

# Run multiple workflows in parallel (recursive)
stepwise run test-workflows/ --parallel 4 -r

# Run with environment-specific config
stepwise run workflow.yml --env staging

# Generate JUnit XML for CI
stepwise run workflow.yml --output junit > results.xml

# Watch mode for development
stepwise run workflow.yml --watch
```

### CI/CD Integration

```bash
#!/bin/bash
# GitHub Actions example

# Run tests
stepwise run test-workflows/ \
  --output junit \
  --parallel 4 \
  --continue-on-error

# Check exit code
if [ $? -eq 0 ]; then
  echo "All tests passed"
else
  echo "Some tests failed"
  exit 1
fi
```

### Docker Integration

```bash
# Run in Docker container
docker run --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  stepwise/stepwise:latest \
  run workflow.yml
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0    | Success - all tests passed |
| 1    | Failure - some tests failed or error occurred |
| 2    | Configuration error |
| 3    | Network error |
| 4    | Validation error |

## Output Formats

### Text Output (Default)

```
✅ Stepwise v1.0.0
🔍 Running workflow: examples/simple-test.yml

📋 Step 1/2: Health Check
   🌐 GET https://httpbin.org/status/200
   ✅ Status: 200 OK
   ✅ Time: 245ms
   ✅ Validation: All checks passed

📋 Step 2/2: JSON Validation
   🌐 GET https://httpbin.org/json
   ✅ Status: 200 OK
   ✅ Time: 189ms
   ✅ Validation: All checks passed

🎉 Summary: 2/2 steps passed (0 failed)
⏱️  Total time: 434ms
```

### JSON Output

```bash
stepwise run workflow.yml --output json
```

```json
{
  "version": "1.0.0",
  "workflow": "examples/simple-test.yml",
  "start_time": "2024-01-15T10:30:00Z",
  "end_time": "2024-01-15T10:30:01Z",
  "duration": "1.2s",
  "results": {
    "total_steps": 2,
    "passed": 2,
    "failed": 0,
    "steps": [
      {
        "name": "Health Check",
        "status": "passed",
        "duration": "245ms",
        "request": {
          "method": "GET",
          "url": "https://httpbin.org/status/200"
        },
        "response": {
          "status": 200,
          "time": "245ms"
        },
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
}
```

### JUnit XML Output

```bash
stepwise run workflow.yml --output junit > results.xml
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="examples/simple-test.yml" tests="2" failures="0" time="1.2">
    <testcase name="Health Check" time="0.245">
      <system-out>GET https://httpbin.org/status/200 - 200 OK</system-out>
    </testcase>
    <testcase name="JSON Validation" time="0.189">
      <system-out>GET https://httpbin.org/json - 200 OK</system-out>
    </testcase>
  </testsuite>
</testsuites>
```

## Configuration Files

### Environment Files

Create `.env` files for different environments:

```bash
# .env.development
API_BASE_URL=https://dev-api.example.com
API_KEY=dev-key-123

# .env.staging
API_BASE_URL=https://staging-api.example.com
API_KEY=staging-key-456

# .env.production
API_BASE_URL=https://api.example.com
API_KEY=prod-key-789
```

### Stepwise Configuration

Create `stepwise.yml` for global configuration:

```yaml
# stepwise.yml
defaults:
  timeout: 30s
  parallel: 4
  output: text

environments:
  development:
    variables:
      api_base_url: https://dev-api.example.com
  staging:
    variables:
      api_base_url: https://staging-api.example.com
  production:
    variables:
      api_base_url: https://api.example.com

logging:
  level: info
  color: true
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   chmod +x stepwise
   ```

2. **Workflow Not Found**
   ```bash
   # Check file exists
   ls -la workflow.yml
   
   # Validate syntax
   stepwise validate workflow.yml
   ```

3. **Network Errors**
   ```bash
   # Check connectivity
   curl https://api.example.com
   
   # Run with verbose output
   stepwise run workflow.yml --verbose
   ```

4. **Validation Failures**
   ```bash
   # Check response format
   stepwise run workflow.yml --verbose
   
   # Validate JSON paths
   stepwise validate workflow.yml
   ```

### Debug Mode

```bash
# Enable debug logging
STEPWISE_LOG_LEVEL=debug stepwise run workflow.yml

# Or use verbose flag
stepwise run workflow.yml --verbose
```

### Performance Issues

```bash
# Run with profiling
STEPWISE_PROFILE=1 stepwise run workflow.yml

# Check memory usage
STEPWISE_MEMORY=1 stepwise run workflow.yml
```

## Best Practices

### 1. Organization

- Keep workflows in dedicated directories
- Use descriptive file names
- Group related workflows together
- Use environment-specific configurations

### 2. CI/CD Integration

- Use appropriate exit codes
- Generate JUnit XML for CI systems
- Disable colors in CI environments
- Set reasonable timeouts

### 3. Performance

- Use parallel execution for multiple workflows
- Set appropriate timeouts
- Monitor resource usage
- Use caching where possible

### 4. Security

- Use environment variables for secrets
- Validate all inputs
- Use HTTPS for all API calls
- Implement proper authentication

### 5. Monitoring

- Use structured logging
- Monitor execution times
- Track success/failure rates
- Set up alerts for failures 