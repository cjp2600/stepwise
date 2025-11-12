# Polling Guide

## Overview

Polling позволяет выполнять запросы повторно до тех пор, пока не будут выполнены определенные условия. Это полезно для ожидания появления транзакций в API, проверки статусов операций и других асинхронных сценариев.

## Configuration

Polling настраивается через поле `poll` в шаге:

```yaml
steps:
  - name: "Wait for Transaction"
    request:
      method: "GET"
      url: "{{base_url}}/transactions/{{transaction_id}}"
    poll:
      max_attempts: 10      # Максимальное количество попыток
      interval: "2s"         # Задержка между попытками
      until:                 # Условия, которые должны быть выполнены
        - status: 200
        - json: "$.status"
          equals: "completed"
```

## Parameters

### `max_attempts` (required)
Максимальное количество попыток поллинга. Если условие не выполнено после всех попыток, шаг завершится с ошибкой.

**Default:** 10 (если не указано)

### `interval` (optional)
Задержка между попытками поллинга. Может быть указана в формате Go duration (например, "1s", "500ms", "2m").

**Default:** 1s (если не указано)

### `until` (required)
Массив правил валидации, которые должны быть выполнены для успешного завершения поллинга. Используется тот же синтаксис, что и в `validate`, но эти правила проверяются на каждой итерации поллинга.

## How It Works

1. Шаг выполняется с конфигурацией `poll`
2. Выполняется запрос
3. Проверяются условия из `poll.until`
4. Если все условия выполнены - шаг завершается успешно
5. Если условия не выполнены:
   - Ожидается `interval`
   - Повторяется попытка (до `max_attempts` раз)
6. Если после всех попыток условие не выполнено - шаг завершается с ошибкой

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

Если поллинг не завершился успешно после всех попыток, шаг завершится с ошибкой, содержащей информацию о неудачных валидациях:

```
polling condition not met after 10 attempts: status validation failed: expected 200, got 404; json validation failed: expected completed, got pending
```

## Best Practices

1. **Установите разумное значение `max_attempts`**: Слишком большое значение может привести к долгому ожиданию, слишком маленькое - к преждевременным ошибкам.

2. **Выберите подходящий `interval`**: 
   - Для быстрых API: 500ms - 1s
   - Для медленных операций: 2s - 5s
   - Не делайте слишком частые запросы, чтобы не перегружать API

3. **Используйте конкретные условия в `until`**: Чем точнее условия, тем быстрее поллинг завершится при достижении нужного состояния.

4. **Комбинируйте с `validate`**: Вы можете использовать `validate` для дополнительных проверок после успешного поллинга.

## Differences from Retry

- **Retry** (`retry`): Повторяет запрос при ошибке (например, сетевые ошибки, таймауты)
- **Polling** (`poll`): Повторяет запрос до выполнения определенных условий в ответе

## Notes

- Поллинг работает как с HTTP, так и с gRPC запросами
- Условия в `poll.until` проверяются на каждой итерации
- Если запрос завершается с ошибкой (сетевая ошибка, таймаут), поллинг продолжается до `max_attempts`
- Время выполнения шага включает все попытки поллинга

