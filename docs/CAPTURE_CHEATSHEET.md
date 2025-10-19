# Шпаргалка: Захват и валидация переменных

## 🎯 Быстрый старт

### Базовый пример

```yaml
steps:
  # 1️⃣ Захватываем данные
  - name: "Get User"
    request:
      method: "GET"
      url: "https://api.example.com/users/1"
    capture:
      user_id: "$.id"
      user_name: "$.name"

  # 2️⃣ Сравниваем с захваченными данными
  - name: "Verify User"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{user_id}}"
    validate:
      - json: "$.name"
        equals: "{{user_name}}"  # ✅ Сравнение!
```

## 📦 Захват данных (Capture)

### Простые поля
```yaml
capture:
  user_id: "$.id"
  user_name: "$.name"
  user_email: "$.email"
```

### Вложенные поля
```yaml
capture:
  city: "$.address.city"
  lat: "$.address.geo.lat"
```

### Из массива с фильтром
```yaml
capture:
  title: "$[?(@.id == 5)].title"
  body: "$[?(@.id == 5)].body"
```

### Первый/последний элемент
```yaml
capture:
  first: "$[0]"
  last: "$[-1]"
  range: "$[0:3]"
```

## ✅ Валидация с переменными

### Прямое сравнение
```yaml
validate:
  - json: "$.id"
    equals: "{{saved_id}}"
  - json: "$.name"
    equals: "{{saved_name}}"
```

### Вложенные поля
```yaml
validate:
  - json: "$.address.city"
    equals: "{{saved_city}}"
  - json: "$.address.geo.lat"
    equals: "{{saved_lat}}"
```

## 🔗 Цепочка запросов

```yaml
steps:
  # Шаг 1: Получаем ID
  - name: "Get Post"
    request:
      method: "GET"
      url: "/posts/1"
    capture:
      author_id: "$.userId"

  # Шаг 2: Используем ID
  - name: "Get Author"
    request:
      method: "GET"
      url: "/users/{{author_id}}"
    validate:
      - json: "$.id"
        equals: "{{author_id}}"  # ✅
```

## 🧩 Компоненты

### Компонент с захватом
```yaml
# components/get-user.yml
name: "Get User"
type: "step"

variables:
  user_id: "1"

steps:
  - name: "Get User"
    request:
      method: "GET"
      url: "/users/{{user_id}}"
    capture:
      user_name: "$.name"
      user_email: "$.email"
```

### Использование
```yaml
imports:
  - path: "components/get-user"
    alias: "get-user"

steps:
  - name: "Get User 5"
    use: 'get-user'
    variables:
      user_id: "5"
  
  # Переменные доступны!
  - name: "Verify"
    request:
      method: "GET"
      url: "/users/5"
    validate:
      - json: "$.name"
        equals: "{{user_name}}"  # ✅
```

## 🎨 JSONPath фильтры

| Фильтр | Описание | Пример |
|--------|----------|--------|
| `$[?(@.id == 5)]` | Равенство | `$[?(@.id == 5)].title` |
| `$[?(@.id > 95)]` | Больше | `$[?(@.price > 100)]` |
| `$[?(@.id < 10)]` | Меньше | `$[?(@.age < 18)]` |
| `$[0]` | Первый элемент | `$[0].name` |
| `$[-1]` | Последний | `$[-1].id` |
| `$[0:3]` | Диапазон | `$[0:5]` |
| `$[*]` | Все элементы | `$[*].id` |

## 💡 Лучшие практики

### ✅ Хорошо
```yaml
capture:
  saved_user_id: "$.id"
  saved_user_name: "$.name"
  original_email: "$.email"
```

### ❌ Плохо
```yaml
capture:
  id: "$.id"
  n: "$.name"
  e: "$.email"
```

## 📝 Полный пример

```yaml
name: "Complete Example"
version: "1.0"

steps:
  # 1. Получаем список постов
  - name: "Get Posts"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts"
    capture:
      post_5_title: "$[?(@.id == 5)].title"
      post_5_user_id: "$[?(@.id == 5)].userId"

  # 2. Проверяем конкретный пост
  - name: "Verify Post"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/posts/5"
    validate:
      - json: "$.title"
        equals: "{{post_5_title}}"  # ✅
      - json: "$.userId"
        equals: "{{post_5_user_id}}"

  # 3. Получаем автора
  - name: "Get Author"
    request:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/users/{{post_5_user_id}}"
    capture:
      author_name: "$.name"
      author_city: "$.address.city"

  # 4. Используем все переменные
  - name: "Final Check"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
      headers:
        X-Post-Title: "{{post_5_title}}"
        X-Author-Name: "{{author_name}}"
    validate:
      - status: 200
```

## 🚀 Запуск примеров

```bash
# Простой пример
go run main.go run examples/working-capture-compare.yml

# С компонентами
go run main.go run examples/component-capture-workflow.yml

# Демонстрация всех возможностей
go run main.go run examples/capture-and-compare-demo.yml
```

## 📚 Дополнительная информация

- Полное руководство: `docs/CAPTURE_AND_VALIDATION_GUIDE.md`
- Переменные: `docs/VARIABLE_KEYS.md`
- Компоненты: `docs/COMPONENTS.md`

