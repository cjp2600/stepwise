# Flow Components Guide

## Overview

Flow components are a powerful feature in Stepwise that allows you to organize and reuse complex test scenarios. Instead of having many individual components, you can create logical groups of steps called "Flows" that can be easily imported and used in your tests.

## Problem

When you have many individual components, your main test becomes very verbose and hard to read:

```yaml
imports:
  - path: "./components/create-customer"
    alias: "create-customer"
  - path: "./components/get-token"
    alias: "get-token"
  - path: "./components/check-customer-info"
    alias: "check-customer-info"
  # ... 10+ more imports

steps:
  - use: 'create-customer'
  - use: 'get-token'
  - use: 'check-customer-info'
  # ... 10+ more steps
```

## Solutions

### 1. Flow Components (Recommended)

Create Flow components that combine logically related steps:

#### Flow Component Structure:
```
components/
├── flows/
│   ├── customer-flow.yml          # 3 steps: create-customer, get-token, check-customer-info
│   ├── payment-method-flow.yml    # 3 steps: create-payment-method, get-methods, update-wallet
│   ├── purchase-flow.yml          # 5 steps: create-purchase, create-repayment, get-invoice, check-status, pay
│   └── complete-purchase-flow.yml # Master component that combines all Flows
```

#### Usage Example:
```yaml
imports:
  - path: "./components/flows/complete-purchase-flow"
    alias: "complete-purchase-flow"

steps:
  - name: "Complete Purchase Flow"
    use: "complete-purchase-flow"
```

### 2. Hierarchical Organization

Create a hierarchy of components:

```
components/
├── customer/
│   ├── create-customer.yml
│   ├── get-token.yml
│   └── check-customer-info.yml
├── payment/
│   ├── create-payment-method.yml
│   ├── get-payment-methods.yml
│   └── update-wallet.yml
├── purchase/
│   ├── create-purchase.yml
│   ├── create-repayment.yml
│   ├── get-invoice.yml
│   ├── check-payment-status.yml
│   └── pay-invoice.yml
└── flows/
    ├── customer-flow.yml
    ├── payment-flow.yml
    └── purchase-flow.yml
```

### 3. Functional Grouping

Create components by functional areas:

```yaml
# components/auth-flow.yml
steps:
  - use: "create-customer"
  - use: "get-token"
  - use: "check-customer-info"

# components/payment-flow.yml
steps:
  - use: "create-payment-method"
  - use: "get-payment-methods"
  - use: "update-wallet"

# components/purchase-flow.yml
steps:
  - use: "create-purchase"
  - use: "create-repayment"
  - use: "get-invoice"
  - use: "check-payment-status"
  - use: "pay-invoice"
```

## Flow Component Benefits

1. **Readability**: Instead of 15+ imports, you have 3-4 Flow components
2. **Reusability**: Flow components can be used in different tests
3. **Modularity**: Easy to add/remove entire blocks of functionality
4. **Maintainability**: Simpler to maintain and modify logic

## Implementation Details

### Flow Component Structure

```yaml
name: "Customer Flow"
version: "1.0"
description: "Complete customer onboarding flow including creation, authentication and verification"
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

### Master Flow Component

```yaml
name: "Complete Purchase Flow"
version: "1.0"
description: "Complete end-to-end purchase flow including customer onboarding, payment setup, purchase and testing"
type: "workflow"

steps:
  # Customer onboarding
  - name: "Customer Onboarding"
    use: "customer-flow"

  # Payment method setup
  - name: "Payment Method Setup"
    use: "payment-method-flow"

  # Purchase and payment
  - name: "Purchase and Payment"
    use: "purchase-flow"

  # Wait for processing
  - name: "Wait for payment processing"
    wait: "10s"

  # Test operations
  - name: "Test gRPC Operations"
    use: "get-grpc-operations"

  # Summary
  - name: "Flow Summary"
    print: |
      === PURCHASE FLOW COMPLETED ===
      Customer ID: {{customer_id}}
      Access Token: {{access_token}}
      Purchase ID: {{purchase_id}}
      Invoice ID: {{invoice_id}}
      Payment Method ID: {{default_payment_method_id}}
      Wallet Balance: {{wallet_balance}} {{wallet_currency}}
      ===============================
```

## File Examples

### Basic Test (Complex):
- `examples/purchase-overview-test.yml` - 15+ imports, 15+ steps

### Simplified Test:
- `examples/purchase-overview-test-simplified.yml` - 4 imports, 6 steps

### Ultimate Simplified:
- `examples/purchase-overview-test-ultimate.yml` - 1 import, 1 step

## Comparison

| Approach | Imports | Steps | Readability |
|----------|---------|-------|-------------|
| Original | 15+ | 15+ | ❌ Poor |
| Simplified | 4 | 6 | ✅ Good |
| Ultimate | 1 | 1 | ✅ Excellent |

## Usage Examples

### Simple Flow Component

```yaml
# components/flows/simple-flow.yml
name: "Simple Flow"
version: "1.0"
description: "Simple flow component without circular dependencies"
type: "workflow"

steps:
  - name: "Step 1"
    print: "Step 1 executed"
  
  - name: "Step 2"
    wait: "1s"
  
  - name: "Step 3"
    print: "Step 3 executed - Flow completed!"
```

### Test Using Flow Component

```yaml
# examples/simple-flow-test.yml
name: "Simple Flow Test"
version: "1.0"

imports:
  - path: "./components/flows/simple-flow"
    alias: "simple-flow"

steps:
  - name: "Execute Simple Flow"
    use: "simple-flow"
  
  - name: "Test Summary"
    print: "Simple flow test completed successfully!"
```

## Folder Execution

Flow components work perfectly when running tests at the folder level:

```bash
go run main.go run test-folder/
```

### Example Output:
```
✓ Found 2 workflow files
[DEBUG] Loading component: ./components/flows/simple-flow
✓ Workflow completed: flow-folder-test.yml
✓ Workflow completed: simple-flow-test.yml

OVERALL SUMMARY
===============
Workflows: 2
Tests Passed: 14
Tests Failed: 0
Total Duration: 4004ms
Success Rate: 100.0%
```

## Circular Dependency Protection

Stepwise includes built-in protection against circular dependencies:

### At Component Loading Level
- Detects circular imports during component loading
- Prevents infinite recursion
- Provides detailed error messages with loading stack

### At Execution Level
- Tracks currently executing workflows
- Prevents workflow components from calling themselves
- Maintains execution context across nested components

### Example Error:
```
[DEBUG] Circular import detected: ./flow-b (loading stack: [./flow-a])
circular import detected: ./flow-b (loading stack: [./flow-a])
```

## Best Practices

1. **Start with Flow components** for logically related operations
2. **Create master components** for complete end-to-end scenarios
3. **Use descriptive names** for Flow components
4. **Add documentation** to each Flow component
5. **Group by domains** (customer, payment, purchase, etc.)
6. **Keep flows focused** on a single responsibility
7. **Use meaningful step names** within flows
8. **Include wait steps** for realistic test scenarios

## Recommendations

1. **Begin with Flow components** for logically related operations
2. **Create master components** for complete end-to-end scenarios
3. **Use descriptive names** for Flow components
4. **Add documentation** to each Flow component
5. **Group by domains** (customer, payment, purchase, etc.)

## Technical Implementation

### Flow Component Support
- `type: "workflow"` components are fully supported
- Imports are resolved recursively
- Variables are properly scoped and merged
- Circular dependency detection at both loading and execution levels

### Execution Context
- Each workflow component maintains its own execution context
- Variables are properly isolated between components
- Results are aggregated correctly
- Error handling preserves context information

### Performance
- Components are cached after first load
- Efficient path resolution and normalization
- Minimal overhead for flow component execution
- Parallel execution support for folder-level runs 