# Array Filters и Advanced JSONPath

Stepwise поддерживает расширенный синтаксис JSONPath для работы с массивами, позволяя находить элементы по условиям, а не только по фиксированному индексу.

## Проблема

Традиционный подход с использованием индексов:

```yaml
validate:
  - json: "$[0].id"
    equals: 123
```

**Проблема**: Если порядок элементов в массиве изменится, тест сломается. Элемент с `id=123` может оказаться на позиции 1, 2 или любой другой.

## Решение: Фильтры массивов

### Базовый синтаксис фильтров

```yaml
validate:
  # Найти первый элемент где id равен 123
  - json: "$[?(@.id == 123)].name"
    type: "string"
```

Здесь:
- `$` - корень JSON
- `[?(...)]` - фильтр массива
- `@` - текущий элемент массива
- `@.id == 123` - условие фильтрации
- `.name` - поле, которое нужно извлечь из найденного элемента

## Поддерживаемые операторы сравнения

### Равенство и неравенство

```yaml
# Равенство (строки)
- json: "$[?(@.status == \"active\")].id"
  type: "number"

# Равенство (числа)
- json: "$[?(@.id == 42)].name"
  type: "string"

# Неравенство
- json: "$[?(@.status != \"deleted\")].id"
  type: "number"
```

### Числовые сравнения

```yaml
# Больше
- json: "$[?(@.price > 100)].name"
  type: "string"

# Меньше
- json: "$[?(@.age < 30)].name"
  type: "string"

# Больше или равно
- json: "$[?(@.quantity >= 10)].id"
  type: "number"

# Меньше или равно
- json: "$[?(@.rating <= 5)].title"
  type: "string"
```

### Boolean поля

```yaml
# Проверка на true (короткая форма)
- json: "$[?(@.active)].id"
  type: "number"

# Проверка на true (полная форма)
- json: "$[?(@.active == true)].id"
  type: "number"

# Проверка на false
- json: "$[?(@.deleted == false)].name"
  type: "string"
```

## Доступ к вложенным полям

### Фильтрация по вложенным полям

```yaml
# Найти пользователя по вложенному полю
- json: "$[?(@.address.city == \"New York\")].name"
  type: "string"

# Получить вложенное поле из результата
- json: "$[?(@.id == 1)].address.geo.lat"
  type: "string"
```

## Специальные селекторы

### Последний элемент

```yaml
# Получить последний элемент
- json: "$[last].id"
  type: "number"

# Альтернативный синтаксис
- json: "$[-1].id"
  type: "number"
```

### Wildcard (все элементы)

```yaml
# Получить весь массив
- json: "$[*]"
  type: "array"

# Проверить длину массива
- json: "$.length"
  greater: 0
```

### Срезы массивов

```yaml
# Получить первые 3 элемента
- json: "$[0:3]"
  type: "array"

# Получить элементы с 5 по 10
- json: "$[5:10]"
  type: "array"

# Получить все элементы начиная с 3
- json: "$[3:]"
  type: "array"
```

## Примеры использования

### Пример 1: Поиск пользователя по имени

```yaml
name: "Find User by Name"
steps:
  - name: "Get all users"
    request:
      method: "GET"
      url: "https://api.example.com/users"
    validate:
      - status: 200
      # Найти пользователя Alice независимо от позиции в массиве
      - json: "$[?(@.name == \"Alice\")].email"
        pattern: "^[^@]+@[^@]+\\.[^@]+$"
    capture:
      alice_id: "$[?(@.name == \"Alice\")].id"
      alice_email: "$[?(@.name == \"Alice\")].email"
```

### Пример 2: Фильтрация по множественным условиям

```yaml
name: "Filter Products"
steps:
  - name: "Get expensive products"
    request:
      method: "GET"
      url: "https://api.example.com/products"
    validate:
      - status: 200
      # Найти первый дорогой товар
      - json: "$[?(@.price > 1000)].name"
        type: "string"
      # Найти товары в наличии
      - json: "$[?(@.inStock == true)].id"
        type: "number"
```

### Пример 3: Работа с вложенными структурами

```yaml
name: "Complex Nested Filter"
steps:
  - name: "Get VIP orders"
    request:
      method: "GET"
      url: "https://api.example.com/orders"
    validate:
      - status: 200
      # Найти заказ VIP клиента
      - json: "$[?(@.customer.vip == true)].id"
        type: "number"
      # Получить сумму заказа
      - json: "$[?(@.customer.vip == true)].total"
        greater: 0
```

### Пример 4: Использование с capture

```yaml
name: "Capture and Reuse"
steps:
  - name: "Find active user"
    request:
      method: "GET"
      url: "https://api.example.com/users"
    validate:
      - status: 200
    capture:
      # Захватить ID активного пользователя
      active_user_id: "$[?(@.active == true)].id"
      active_user_name: "$[?(@.active == true)].name"
  
  - name: "Get user details"
    request:
      method: "GET"
      url: "https://api.example.com/users/{{active_user_id}}"
    validate:
      - status: 200
      - json: "$.id"
        equals: "{{active_user_id}}"
```

## Сравнение: До и После

### До (с индексами)

```yaml
# Хрупкий код - зависит от порядка элементов
validate:
  - json: "$[0].id"
    equals: 123
  - json: "$[2].name"
    equals: "Alice"
```

**Проблемы:**
- Если элементы переупорядочатся, тесты сломаются
- Невозможно найти элемент по условию
- Нет гибкости при изменении данных

### После (с фильтрами)

```yaml
# Надежный код - не зависит от порядка
validate:
  - json: "$[?(@.id == 123)].id"
    equals: 123
  - json: "$[?(@.name == \"Alice\")].name"
    equals: "Alice"
```

**Преимущества:**
- ✅ Не зависит от порядка элементов
- ✅ Находит элемент по любому полю
- ✅ Работает при изменении данных
- ✅ Более читаемый и понятный код

## Рекомендации

1. **Используйте фильтры вместо индексов**, когда порядок элементов может меняться
2. **Используйте уникальные поля** (id, email и т.д.) для фильтрации
3. **Комбинируйте фильтры с capture** для последующего использования значений
4. **Используйте вложенные поля** для более точной фильтрации
5. **Проверяйте типы данных** после фильтрации для надежности

## Полный пример

Смотрите полный рабочий пример в файле [`examples/array-filters-demo.yml`](../examples/array-filters-demo.yml).

Запустите:
```bash
go run main.go run examples/array-filters-demo.yml
```

## Ограничения

1. Фильтры возвращают **первый** найденный элемент, а не все подходящие
2. Если элемент не найден, возвращается ошибка
3. Сложные логические операции (AND, OR) пока не поддерживаются
4. Регулярные выражения в фильтрах пока не поддерживаются

## Дополнительная информация

- [JSONPath синтаксис](https://goessner.net/articles/JsonPath/)
- [Основная документация](../README.md)
- [Примеры workflows](../examples/)

