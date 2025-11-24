# Polling Guide

## Overview

Polling allows you to repeatedly execute requests until certain conditions are met. This is useful for waiting for transactions to appear in the API, checking operation statuses, and other asynchronous scenarios.

## Configuration

Polling is configured through the `poll` field in a step:

```yaml
steps:
  - name: "Wait for Transaction"
    request:
      method: "GET"
      url: "{{base_url}}/transactions/{{transaction_id}}"
    poll:
      max_attempts: 10      # Maximum number of attempts
      interval: "2s"         # Delay between attempts
      until:                 # Conditions that must be met
        - status: 200
        - json: "$.status"
          equals: "completed"
```

## Parameters

### `max_attempts` (required)
Maximum number of polling attempts. If the condition is not met after all attempts, the step will fail with an error.

**Default:** 10 (if not specified)

### `interval` (optional)
Delay between polling attempts. Can be specified in Go duration format (e.g., "1s", "500ms", "2m").

**Default:** 1s (if not specified)

### `until` (required)
Array of validation rules that must be met for polling to complete successfully. Uses the same syntax as `validate`, but these rules are checked on each polling iteration.

## How It Works

1. Step is executed with `poll` configuration
2. Request is executed
3. Conditions from `poll.until` are checked
4. If all conditions are met - step completes successfully
5. If conditions are not met:
   - Wait for `interval`
   - Retry (up to `max_attempts` times)
6. If condition is not met after all attempts - step fails with error

## Examples

### Example 1: Wait for Transaction Status

```yaml
steps:
  - name: "Create Transaction"
    request:
      method: "POST"
      url: "{{base_url}}/transactions"
      body:
        amount: 100
    capture:
      transaction_id: "$.id"

  - name: "Wait for Transaction Completion"
    request:
      method: "GET"
      url: "{{base_url}}/transactions/{{transaction_id}}"
    poll:
      max_attempts: 20
      interval: "1s"
      until:
        - status: 200
        - json: "$.status"
          equals: "completed"
    validate:
      - status: 200
```

### Example 2: Poll Until Value Appears

```yaml
steps:
  - name: "Check for Data"
    request:
      method: "GET"
      url: "{{base_url}}/data/{{data_id}}"
    poll:
      max_attempts: 15
      interval: "500ms"
      until:
        - status: 200
        - json: "$.data"
          empty: false
        - json: "$.data.length"
          greater: 0
```

### Example 3: Poll with Multiple Conditions

```yaml
steps:
  - name: "Wait for Processing"
    request:
      method: "GET"
      url: "{{base_url}}/jobs/{{job_id}}"
    poll:
      max_attempts: 30
      interval: "2s"
      until:
        - status: 200
        - json: "$.status"
          equals: "done"
        - json: "$.result"
          nil: false
        - json: "$.error"
          nil: true
```

## Error Handling

If polling does not complete successfully after all attempts, the step will fail with an error containing information about failed validations:

```
polling condition not met after 10 attempts: status validation failed: expected 200, got 404; json validation failed: expected completed, got pending
```

## Best Practices

1. **Set a reasonable `max_attempts` value**: Too large a value can lead to long waits, too small - to premature errors.

2. **Choose an appropriate `interval`**: 
   - For fast APIs: 500ms - 1s
   - For slow operations: 2s - 5s
   - Don't make requests too frequent to avoid overloading the API

3. **Use specific conditions in `until`**: The more precise the conditions, the faster polling will complete when the desired state is reached.

4. **Combine with `validate`**: You can use `validate` for additional checks after successful polling.

## Differences from Retry

- **Retry** (`retry`): Repeats request on error (e.g., network errors, timeouts)
- **Polling** (`poll`): Repeats request until certain conditions in the response are met

## Notes

- Polling works with both HTTP and gRPC requests
- Conditions in `poll.until` are checked on each iteration
- If request fails with an error (network error, timeout), polling continues until `max_attempts`
- Step execution time includes all polling attempts
