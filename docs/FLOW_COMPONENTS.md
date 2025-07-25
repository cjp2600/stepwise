# Flow Components

## Overview

Flow components are a powerful feature in Stepwise that allows you to organize and reuse complex test scenarios. Instead of having many individual components, you can create logical groups of steps called "Flows" that can be easily imported and used in your tests.

## Quick Start

### Create a Flow Component

```yaml
# components/flows/simple-flow.yml
name: "Simple Flow"
version: "1.0"
description: "Simple flow component"
type: "workflow"

steps:
  - name: "Step 1"
    print: "Step 1 executed"
  
  - name: "Step 2"
    wait: "1s"
  
  - name: "Step 3"
    print: "Step 3 executed - Flow completed!"
```

### Use Flow Component in Test

```yaml
# test.yml
name: "Flow Test"
version: "1.0"

imports:
  - path: "./components/flows/simple-flow"
    alias: "simple-flow"

steps:
  - name: "Execute Simple Flow"
    use: "simple-flow"
  
  - name: "Test Summary"
    print: "Flow test completed successfully!"
```

## Benefits

- **Readability**: Instead of 15+ imports, you have 3-4 Flow components
- **Reusability**: Flow components can be used in different tests
- **Modularity**: Easy to add/remove entire blocks of functionality
- **Maintainability**: Simpler to maintain and modify logic

## Folder Execution

Flow components work perfectly when running tests at the folder level:

```bash
go run main.go run test-folder/
```

## Circular Dependency Protection

Stepwise includes built-in protection against circular dependencies at both component loading and execution levels.

## Best Practices

1. Start with Flow components for logically related operations
2. Create master components for complete end-to-end scenarios
3. Use descriptive names for Flow components
4. Add documentation to each Flow component
5. Group by domains (customer, payment, purchase, etc.)

For detailed information, see [FLOW_ORGANIZATION.md](FLOW_ORGANIZATION.md). 