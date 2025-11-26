# Ветвление (Branching)

Ветвление позволяет выполнять различные наборы шагов в зависимости от условий. Stepwise поддерживает два основных подхода к ветвлению:

1. **If-Then-Else** - простой вариант для двух веток
2. **Branches** - множественные ветки с приоритетами

## If-Then-Else

Самый простой способ ветвления - использовать конструкцию `if/then/else`. Если условие в поле `if` истинно, выполняются шаги из `then`, иначе - из `else`.

### Синтаксис

```yaml
steps:
  - name: "Branch Step"
    if: "{{variable}} == 'value'"
    then:
      - name: "Then Step 1"
        request:
          method: "GET"
          url: "{{base_url}}/endpoint1"
      - name: "Then Step 2"
        request:
          method: "GET"
          url: "{{base_url}}/endpoint2"
    else:
      - name: "Else Step 1"
        request:
          method: "GET"
          url: "{{base_url}}/endpoint3"
      - name: "Else Step 2"
        request:
          method: "GET"
          url: "{{base_url}}/endpoint4"
```

### Пример

```yaml
name: "If-Else Example"
variables:
  base_url: "https://api.example.com"
  user_id: 1

steps:
  - name: "Get User"
    request:
      method: "GET"
      url: "{{base_url}}/users/{{user_id}}"
    capture:
      user_status: "$.status"
  
  - name: "Process Based on Status"
    if: "{{user_status}} == 'active'"
    then:
      - name: "Activate User"
        request:
          method: "POST"
          url: "{{base_url}}/users/{{user_id}}/activate"
    else:
      - name: "Deactivate User"
        request:
          method: "POST"
          url: "{{base_url}}/users/{{user_id}}/deactivate"
```

## Branches (Множественные ветки)

Для более сложных сценариев с несколькими условиями используйте `branches`. Каждая ветка имеет условие и набор шагов. Ветки проверяются в порядке приоритета (если указан), и выполняется первая подходящая ветка.

### Синтаксис

```yaml
steps:
  - name: "Multi-Branch Step"
    branches:
      - condition: "{{variable}} == 'value1'"
        steps:
          - name: "Branch 1 Step"
            request:
              method: "GET"
              url: "{{base_url}}/endpoint1"
        priority: 10  # Опционально, для сортировки
      
      - condition: "{{variable}} == 'value2'"
        steps:
          - name: "Branch 2 Step"
            request:
              method: "GET"
              url: "{{base_url}}/endpoint2"
        priority: 5
      
      - condition: "true"  # Всегда истинно (default branch)
        steps:
          - name: "Default Step"
            request:
              method: "GET"
              url: "{{base_url}}/default"
        priority: 1
```

### Пример

```yaml
name: "Branches Example"
variables:
  base_url: "https://api.example.com"
  user_id: 1

steps:
  - name: "Get User"
    request:
      method: "GET"
      url: "{{base_url}}/users/{{user_id}}"
    capture:
      user_type: "$.type"
      user_status: "$.status"
  
  - name: "Route Based on User Type"
    branches:
      - condition: "{{user_type}} == 'admin' && {{user_status}} == 'active'"
        steps:
          - name: "Admin Active Flow"
            request:
              method: "GET"
              url: "{{base_url}}/admin/dashboard"
        priority: 10
      
      - condition: "{{user_type}} == 'premium'"
        steps:
          - name: "Premium User Flow"
            request:
              method: "GET"
              url: "{{base_url}}/premium/content"
        priority: 5
      
      - condition: "true"
        steps:
          - name: "Default User Flow"
            request:
              method: "GET"
              url: "{{base_url}}/default/content"
        priority: 1
```

## Операторы сравнения

Поддерживаются следующие операторы сравнения:

- `==` - равно
- `!=` - не равно
- `>` - больше
- `<` - меньше
- `>=` - больше или равно
- `<=` - меньше или равно

### Примеры условий

```yaml
# Числовое сравнение
if: "{{count}} > 10"
if: "{{age}} >= 18"
if: "{{score}} < 100"

# Строковое сравнение
if: "{{status}} == 'active'"
if: "{{email}} != ''"

# Комбинированные
if: "{{count}} > 0 && {{count}} < 100"
```

## Логические операторы

Поддерживаются логические операторы:

- `&&` - И (AND)
- `||` - ИЛИ (OR)
- `!` - НЕ (NOT)

### Примеры

```yaml
# Логическое И
if: "{{is_admin}} && {{is_active}}"
if: "{{count}} > 10 && {{count}} < 100"

# Логическое ИЛИ
if: "{{status}} == 'active' || {{status}} == 'pending'"
if: "{{user_id}} == 1 || {{user_id}} == 2"

# Логическое НЕ
if: "!{{is_blocked}}"
if: "{{status}} != 'inactive'"
```

## Вложенное ветвление

Ветвление можно вкладывать друг в друга для создания сложной логики:

```yaml
steps:
  - name: "Outer Branch"
    if: "{{user_type}} == 'premium'"
    then:
      - name: "Inner Branch"
        if: "{{subscription_days}} > 30"
        then:
          - name: "Long-term Premium"
            request:
              method: "GET"
              url: "{{base_url}}/premium/long-term"
        else:
          - name: "Short-term Premium"
            request:
              method: "GET"
              url: "{{base_url}}/premium/short-term"
    else:
      - name: "Regular User"
        request:
          method: "GET"
          url: "{{base_url}}/regular"
```

## Приоритеты веток

В `branches` можно указать приоритет для сортировки. Ветки с более высоким приоритетом проверяются первыми:

```yaml
branches:
  - condition: "{{user_type}} == 'admin'"
    priority: 10  # Проверяется первым
    steps:
      - name: "Admin Flow"
        request: ...
  
  - condition: "{{user_type}} == 'user'"
    priority: 5   # Проверяется вторым
    steps:
      - name: "User Flow"
        request: ...
  
  - condition: "true"
    priority: 1   # Проверяется последним (default)
    steps:
      - name: "Default Flow"
        request: ...
```

## Использование компонентов (use) внутри веток

Вы можете использовать компоненты (`use`) внутри веток ветвления:

```yaml
imports:
  - path: "components/get-user-by-id.yml"
    alias: "get-user"

steps:
  - name: "Branch with Component"
    if: "{{user_id}} > 5"
    then:
      - name: "Use Component in Then"
        use: "get-user"
        variables:
          user_id: "{{user_id}}"
        validate:
          - status: 200
    else:
      - name: "Use Component in Else"
        use: "get-user"
        variables:
          user_id: "{{user_id}}"
        validate:
          - status: 200
```

Переменные в `variables` автоматически подставляются перед выполнением компонента.

## Совместимость с другими фичами

Ветвление совместимо с другими фичами Stepwise:

### С condition (пропуск шага)

```yaml
steps:
  - name: "Branch Step"
    if: "{{user_id}} > 5"
    then:
      - name: "High ID Step"
        condition: "{{skip_this}} == false"  # Может быть пропущен
        request:
          method: "GET"
          url: "{{base_url}}/high-id"
```

### С repeat

`repeat` полностью поддерживается внутри веток:

```yaml
steps:
  - name: "Branch with Repeat"
    if: "{{use_repeat}} == true"
    then:
      - name: "Repeated Step"
        repeat:
          count: 3
          delay: "100ms"
        request:
          method: "GET"
          url: "{{base_url}}/endpoint"
    else:
      - name: "Single Step"
        request:
          method: "GET"
          url: "{{base_url}}/endpoint"
```

### С poll

`poll` полностью поддерживается внутри веток:

```yaml
steps:
  - name: "Branch with Poll"
    if: "{{use_poll}} == true"
    then:
      - name: "Polling Step"
        poll:
          max_attempts: 10
          interval: "1s"
          until:
            - json: "$.status"
              equals: "ready"
        request:
          method: "GET"
          url: "{{base_url}}/status"
    else:
      - name: "Regular Step"
        request:
          method: "GET"
          url: "{{base_url}}/status"
```

## Обработка ошибок

Если ни одна ветка не подходит (в `branches`), шаг помечается как `failed` с ошибкой "no branch condition matched".

В режиме `fail-fast` выполнение останавливается при первой ошибке в любой ветке.

## Результаты выполнения

Результаты выполнения веток сохраняются в `TestResult`:

- Если все шаги в ветке прошли успешно, общий статус шага - `passed`
- Если хотя бы один шаг в ветке упал, общий статус - `failed`
- Пропущенные шаги (из-за `condition`) помечаются как `skipped`

## Примеры использования

См. примеры в директории `examples/`:

- `branching-if-else-demo.yml` - простые примеры if/then/else
- `branching-multiple-demo.yml` - множественные ветки
- `branching-complex-demo.yml` - сложные сценарии с вложенностью
- `branching-with-use-demo.yml` - использование компонентов (`use`) внутри веток
- `branching-with-repeat-poll-demo.yml` - использование `repeat` и `poll` внутри веток
- `branching-verification-demo.yml` - проверка корректности работы ветвления

## Ограничения

1. Условия оцениваются последовательно, без оптимизации короткого замыкания (хотя логически работает как короткое замыкание)
2. Вложенное ветвление поддерживается, но рекомендуется избегать слишком глубокой вложенности (более 3 уровней) для читаемости

## Рекомендации

1. Используйте `if/then/else` для простых случаев с двумя ветками
2. Используйте `branches` для множественных условий
3. Всегда добавляйте ветку с `condition: "true"` в `branches` как fallback
4. Используйте приоритеты для явного контроля порядка проверки условий
5. Избегайте слишком глубокой вложенности (более 3 уровней)

