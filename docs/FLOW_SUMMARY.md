# Flow Components Summary

## What are Flow Components?

Flow components are a powerful feature in Stepwise that allows you to organize complex test scenarios into logical, reusable groups. Instead of having many individual components, you can create meaningful "Flows" that combine related steps.

## Key Benefits

| Benefit | Description |
|---------|-------------|
| **Readability** | Instead of 15+ imports, you have 3-4 Flow components |
| **Reusability** | Flow components can be used across different tests |
| **Modularity** | Easy to add/remove entire blocks of functionality |
| **Maintainability** | Simpler to maintain and modify logic |

## Quick Example

### Before (Complex)
```yaml
imports:
  - path: "./components/create-customer"
  - path: "./components/get-token"
  - path: "./components/check-customer-info"
  - path: "./components/create-payment-method"
  - path: "./components/get-payment-methods"
  - path: "./components/update-wallet"
  # ... 10+ more imports

steps:
  - use: 'create-customer'
  - use: 'get-token'
  - use: 'check-customer-info'
  - use: 'create-payment-method'
  - use: 'get-payment-methods'
  - use: 'update-wallet'
  # ... 10+ more steps
```

### After (Simple)
```yaml
imports:
  - path: "./components/flows/customer-flow"
  - path: "./components/flows/payment-flow"

steps:
  - name: "Customer Onboarding"
    use: "customer-flow"
  - name: "Payment Setup"
    use: "payment-flow"
```

## Implementation Status

✅ **Fully Implemented**
- Flow component support (`type: "workflow"`)
- Recursive import resolution
- Variable scoping and merging
- Circular dependency detection (loading level)
- Folder-level execution support
- Component caching and performance optimization

⚠️ **Partially Implemented**
- Circular dependency detection (execution level) - requires additional refinement

## Usage

### Create Flow Component
```yaml
# components/flows/customer-flow.yml
name: "Customer Flow"
version: "1.0"
type: "workflow"

imports:
  - path: "../create-customer"
  - path: "../get-token"

steps:
  - name: "Create Customer"
    use: "create-customer"
  - name: "Get Token"
    use: "get-token"
```

### Use in Test
```yaml
# test.yml
imports:
  - path: "./components/flows/customer-flow"
    alias: "customer-flow"

steps:
  - name: "Customer Onboarding"
    use: "customer-flow"
```

### Run Tests
```bash
# Single test
go run main.go run test.yml

# Folder execution
go run main.go run test-folder/
```

## Documentation

- [FLOW_COMPONENTS.md](FLOW_COMPONENTS.md) - Quick start guide
- [FLOW_ORGANIZATION.md](FLOW_ORGANIZATION.md) - Detailed documentation
- [COMPONENTS.md](COMPONENTS.md) - General component system

## Examples

See the `examples/` directory for working Flow component examples:
- `simple-flow-test.yml` - Basic Flow component usage
- `flow-folder-test.yml` - Folder execution example
- `test-flows/` - Directory with Flow component tests 