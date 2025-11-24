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
go build -o stepwise main.go

# Or install globally
go install .
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

# Run with parallel execution
stepwise run test-workflows/ --parallel 4

# Run with fail-fast mode (stop on first failure)
stepwise run workflow.yml --fail-fast

# Generate HTML report
stepwise run workflow.yml --html-report

# Generate HTML report with custom path
stepwise run workflow.yml --html-report --html-report-path custom-report.html
```

### `stepwise validate`

Validate workflow syntax and configuration.

```bash
# Validate a single workflow
stepwise validate workflow.yml

# Validate all workflows in directory
stepwise validate test-workflows/
```

### `stepwise info`

Display information about a workflow.

```bash
# Show workflow information
stepwise info workflow.yml
```

### `stepwise init`

Initialize a new Stepwise project.

```bash
# Create a new workflow file
stepwise init
```

This creates a `workflow.yml` file with a basic example.

### `stepwise codex`

AI assistant for creating workflows (requires codex CLI).

```bash
# Generate workflows from directory
stepwise codex ./examples

# Use specific model
stepwise codex --model gpt-4o .
```

### `stepwise generate`

Generate test data (not yet implemented).

```bash
stepwise generate
```

### `stepwise help`

Show help message.

```bash
stepwise help
# or
stepwise --help
# or
stepwise -h
```

### `stepwise version`

Show version information.

```bash
stepwise version
# or
stepwise --version
# or
stepwise -v
```

## Command Options

### Run Command Options

```bash
stepwise run [OPTIONS] <FILE|DIR>

Options:
  --parallel, -p N        Number of parallel workflow executions (default: 1)
  --recursive, -r         Search recursively in subdirectories
  --verbose, -v           Enable verbose logging
  --fail-fast, -f         Stop execution on first test failure
  --html-report           Generate HTML report (default: test-report_TIMESTAMP.html)
  --html-report-path PATH Path for HTML report file (used with --html-report)
```

## Examples

### Basic Usage

```bash
# Run a simple workflow
stepwise run examples/simple-test.yml

# Run with verbose output
stepwise run examples/simple-test.yml --verbose

# Run with fail-fast mode
stepwise run examples/simple-test.yml --fail-fast
```

### Advanced Usage

```bash
# Run multiple workflows in parallel (non-recursive)
stepwise run test-workflows/ --parallel 4

# Run multiple workflows in parallel (recursive)
stepwise run test-workflows/ --parallel 4 -r

# Generate HTML report
stepwise run workflow.yml --html-report

# Generate HTML report with custom path
stepwise run workflow.yml --html-report --html-report-path my-report.html
```

### CI/CD Integration

```bash
#!/bin/bash
# GitHub Actions example

# Run tests with fail-fast
stepwise run test-workflows/ \
  --parallel 4 \
  --fail-fast

# Check exit code
if [ $? -eq 0 ]; then
  echo "All tests passed"
else
  echo "Some tests failed"
  exit 1
fi
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0    | Success - all tests passed |
| 1    | Failure - some tests failed or error occurred |

## Output

### Text Output (Default)

The default output shows progress with spinners and colored output:

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

### Verbose Output

When using `--verbose`, detailed debug information is shown:

```bash
stepwise run workflow.yml --verbose
```

This includes:
- Detailed request/response information
- Variable substitution details
- Validation results
- Error details

### HTML Report

Generate an HTML report with test results:

```bash
stepwise run workflow.yml --html-report
```

This creates a file named `test-report_TIMESTAMP.html` with detailed test results.

To specify a custom path:

```bash
stepwise run workflow.yml --html-report --html-report-path my-report.html
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
# Use verbose flag for detailed output
stepwise run workflow.yml --verbose
```

## Best Practices

### 1. Organization

- Keep workflows in dedicated directories
- Use descriptive file names
- Group related workflows together

### 2. CI/CD Integration

- Use appropriate exit codes
- Use `--fail-fast` for faster feedback
- Disable colors in CI environments (set `NO_COLOR=1`)

### 3. Performance

- Use parallel execution for multiple workflows
- Set appropriate timeouts in workflow files
- Monitor resource usage

### 4. Security

- Use environment variables for secrets (e.g., `${API_KEY}`)
- Validate all inputs
- Use HTTPS for all API calls
- Implement proper authentication

### 5. Monitoring

- Use HTML reports for detailed analysis
- Monitor execution times
- Track success/failure rates
