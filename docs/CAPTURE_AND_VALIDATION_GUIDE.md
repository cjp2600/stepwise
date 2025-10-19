# Руководство по захвату переменных и валидации

## Обзор

Stepwise предоставляет мощные возможности для захвата данных из ответов API и их последующего использования в валидации и других шагах workflow.

## Основные возможности

### ✅ 1. Захват переменных (Capture)

Вы можете захватывать данные из JSON ответов используя JSONPath:

```yaml
steps:
  - name: "Получить пользователя"
    request:
      method: "GET"
      url: "https://api.example.com/users/1"
    capture:
      user_id: "$.id"              # Захватываем ID
      user_name: "$.name"          # Захватываем имя
      user_email: "$.email"        # Захватываем email
      user_city: "$.address.city"  # Захватываем вложенное поле
```

### ✅ 2. Использование захваченных переменных

Захваченные переменные доступны во всех последующих шагах:

```yaml
  - name: "Использовать захваченные данные"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{user_id}}"  # Используем в URL
      headers:
        X-User-Name: "{{user_name}}"  # Используем в заголовках
```

### ✅ 3. Сравнение в валидации

Самая важная возможность - сравнение данных из ответа с захваченными переменными:

```yaml
  - name: "Проверить данные"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{user_id}}"
    validate:
      - status: 200
      # Сравниваем каждое поле с захваченными переменными
      - json: "$.id"
        equals: "{{user_id}}"
      - json: "$.name"
        equals: "{{user_name}}"
      - json: "$.email"
        equals: "{{user_email}}"
      - json: "$.address.city"
        equals: "{{user_city}}"
```

## Практические примеры

### Пример 1: Простой захват и сравнение

```yaml
name: "Simple Capture Example"
version: "1.0"

steps:
  # Шаг 1: Получаем пользователя и захватываем данные
  - name: "Get User"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/1"
    validate:
      - status: 200
    capture:
      saved_name: "$.name"
      saved_email: "$.email"

  # Шаг 2: Снова получаем того же пользователя и проверяем
  - name: "Verify User Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/1"
    validate:
      - status: 200
      - json: "$.name"
        equals: "{{saved_name}}"    # Сравниваем с захваченным
      - json: "$.email"
        equals: "{{saved_email}}"   # Сравниваем с захваченным
```

### Пример 2: Захват из массива с фильтрацией

```yaml
name: "Array Filter Capture"
version: "1.0"

steps:
  # Шаг 1: Получаем массив и захватываем данные конкретного элемента
  - name: "Get Posts Array"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts"
    capture:
      # Используем JSONPath фильтры
      post_5_title: "$[?(@.id == 5)].title"
      post_5_body: "$[?(@.id == 5)].body"
      post_5_user_id: "$[?(@.id == 5)].userId"

  # Шаг 2: Получаем конкретный пост и сравниваем
  - name: "Verify Post Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts/5"
    validate:
      - status: 200
      - json: "$.title"
        equals: "{{post_5_title}}"     # Сравниваем с данными из массива
      - json: "$.body"
        equals: "{{post_5_body}}"
      - json: "$.userId"
        equals: "{{post_5_user_id}}"
```

### Пример 3: Цепочка запросов с захватом

```yaml
name: "Chained Requests"
version: "1.0"

steps:
  # Шаг 1: Получаем пост и захватываем ID автора
  - name: "Get Post"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts/1"
    capture:
      author_id: "$.userId"

  # Шаг 2: Используем захваченный ID для получения автора
  - name: "Get Author"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/{{author_id}}"
    validate:
      - status: 200
      - json: "$.id"
        equals: "{{author_id}}"  # Проверяем правильность ID
    capture:
      author_name: "$.name"
      author_email: "$.email"

  # Шаг 3: Используем все захваченные данные
  - name: "Use All Captured Data"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
      headers:
        X-Author-Id: "{{author_id}}"
        X-Author-Name: "{{author_name}}"
    validate:
      - status: 200
```

### Пример 4: Вложенные поля

```yaml
name: "Nested Fields Capture"
version: "1.0"

steps:
  - name: "Get User with Nested Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/1"
    capture:
      # Захват вложенных полей
      user_city: "$.address.city"
      user_street: "$.address.street"
      user_lat: "$.address.geo.lat"
      user_lng: "$.address.geo.lng"
      company_name: "$.company.name"

  - name: "Verify Nested Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/1"
    validate:
      - status: 200
      # Сравниваем вложенные поля
      - json: "$.address.city"
        equals: "{{user_city}}"
      - json: "$.address.street"
        equals: "{{user_street}}"
      - json: "$.address.geo.lat"
        equals: "{{user_lat}}"
      - json: "$.address.geo.lng"
        equals: "{{user_lng}}"
      - json: "$.company.name"
        equals: "{{company_name}}"
```

## Использование с компонентами

### Создание компонента с захватом

```yaml
# components/get-user-component.yml
name: "Get User Component"
version: "1.0"
type: "step"

variables:
  user_id: "1"

steps:
  - name: "Get User"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/{{user_id}}"
    validate:
      - status: 200
    capture:
      user_name: "$.name"
      user_email: "$.email"
      user_city: "$.address.city"
```

### Использование компонента и его переменных

```yaml
name: "Component Usage"
version: "1.0"

imports:
  - path: "components/get-user-component"
    alias: "get-user"

steps:
  # Используем компонент - он захватит переменные
  - name: "Get User via Component"
    use: 'get-user'
    variables:
      user_id: "5"

  # Переменные из компонента доступны в следующих шагах
  - name: "Verify Captured Data"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/5"
    validate:
      - status: 200
      - json: "$.name"
        equals: "{{user_name}}"    # Используем переменную из компонента
      - json: "$.email"
        equals: "{{user_email}}"
```

## Продвинутые JSONPath фильтры

### Фильтрация по условию

```yaml
capture:
  # Равенство
  item_with_id_5: "$[?(@.id == 5)]"
  
  # Больше чем
  high_id_items: "$[?(@.id > 95)]"
  
  # Меньше чем
  low_price_items: "$[?(@.price < 100)]"
  
  # Несколько условий
  specific_item: "$[?(@.id == 5 && @.active == true)]"
```

### Специальные селекторы

```yaml
capture:
  # Первый элемент
  first_item: "$[0]"
  
  # Последний элемент
  last_item: "$[-1]"
  
  # Диапазон элементов
  first_three: "$[0:3]"
  
  # Все элементы
  all_items: "$[*]"
```

### Вложенные фильтры

```yaml
capture:
  # Фильтр с вложенным полем
  user_in_paris: "$[?(@.address.city == 'Paris')].name"
  
  # Глубоко вложенное поле после фильтра
  latitude: "$[?(@.id == 1)].address.geo.lat"
```

## Типы валидации с переменными

### 1. Прямое сравнение (equals)

```yaml
validate:
  - json: "$.id"
    equals: "{{user_id}}"
```

### 2. Проверка типа

```yaml
validate:
  - json: "$.name"
    type: "string"
  - json: "$.id"
    type: "number"
```

### 3. Регулярное выражение

```yaml
validate:
  - json: "$.email"
    pattern: "^[^@]+@[^@]+\\.[^@]+$"
```

### 4. Комбинация проверок

```yaml
validate:
  - status: 200
  - json: "$.id"
    type: "number"
  - json: "$.id"
    equals: "{{saved_id}}"
  - json: "$.email"
    equals: "{{saved_email}}"
  - json: "$.email"
    pattern: "^[^@]+@[^@]+\\.[^@]+$"
```

## Best Practices

### 1. Именование переменных

Используйте осмысленные имена:

```yaml
# ✅ Хорошо
capture:
  user_id: "$.id"
  user_email: "$.email"
  created_at: "$.createdAt"

# ❌ Плохо
capture:
  id: "$.id"
  e: "$.email"
  t: "$.createdAt"
```

### 2. Префиксы для ясности

```yaml
capture:
  saved_user_id: "$.id"        # Сохраненный ID
  current_user_name: "$.name"  # Текущее имя
  original_email: "$.email"    # Оригинальный email
```

### 3. Группировка связанных данных

```yaml
capture:
  # Данные пользователя
  user_id: "$.id"
  user_name: "$.name"
  user_email: "$.email"
  
  # Данные адреса
  address_city: "$.address.city"
  address_street: "$.address.street"
  
  # Геоданные
  geo_lat: "$.address.geo.lat"
  geo_lng: "$.address.geo.lng"
```

### 4. Валидация перед использованием

```yaml
steps:
  - name: "Capture Data"
    request:
      method: "GET"
      url: "https://api.example.com/users/1"
    validate:
      - status: 200
      # Проверяем, что поля существуют
      - json: "$.id"
        type: "number"
      - json: "$.name"
        type: "string"
    capture:
      user_id: "$.id"
      user_name: "$.name"
```

## Распространенные паттерны

### Паттерн 1: Создать → Проверить

```yaml
steps:
  - name: "Create Resource"
    request:
      method: "POST"
      url: "https://api.example.com/resources"
      body:
        name: "Test Resource"
    capture:
      created_id: "$.id"
      created_name: "$.name"

  - name: "Verify Created Resource"
    request:
      method: "GET"
      url: "https://api.example.com/resources/{{created_id}}"
    validate:
      - json: "$.id"
        equals: "{{created_id}}"
      - json: "$.name"
        equals: "{{created_name}}"
```

### Паттерн 2: Получить список → Получить детали

```yaml
steps:
  - name: "Get List"
    request:
      method: "GET"
      url: "https://api.example.com/items"
    capture:
      first_item_id: "$[0].id"

  - name: "Get Item Details"
    request:
      method: "GET"
      url: "https://api.example.com/items/{{first_item_id}}"
    validate:
      - json: "$.id"
        equals: "{{first_item_id}}"
```

### Паттерн 3: Связанные ресурсы

```yaml
steps:
  - name: "Get Post"
    request:
      method: "GET"
      url: "https://api.example.com/posts/1"
    capture:
      post_author_id: "$.userId"

  - name: "Get Author"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{post_author_id}}"
    capture:
      author_name: "$.name"

  - name: "Verify Relationship"
    request:
      method: "GET"
      url: "https://api.example.com/posts/1"
    validate:
      - json: "$.userId"
        equals: "{{post_author_id}}"
```

## Заключение

Stepwise предоставляет полный набор инструментов для:

✅ Захвата данных из JSON ответов  
✅ Использования захваченных переменных в последующих запросах  
✅ Валидации данных путем сравнения с захваченными переменными  
✅ Работы с вложенными структурами данных  
✅ Фильтрации массивов с помощью JSONPath  
✅ Переиспользования компонентов с захватом переменных  

Все эти возможности делают Stepwise мощным инструментом для тестирования API!

