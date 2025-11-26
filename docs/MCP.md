# MCP Protocol Support

Stepwise поддерживает протокол MCP (Model Context Protocol) для взаимодействия с MCP серверами через JSON-RPC 2.0.

## Обзор

MCP (Model Context Protocol) - это открытый протокол для стандартизированного взаимодействия с AI моделями и внешними инструментами. Stepwise поддерживает все основные методы MCP:

- `initialize` - инициализация соединения
- `tools/list` - список доступных инструментов
- `tools/call` - вызов инструмента
- `resources/list` - список доступных ресурсов
- `resources/read` - чтение ресурса
- `prompts/list` - список доступных промптов
- `prompts/get` - получение промпта

## Транспорты

Stepwise поддерживает следующие типы транспорта для MCP:

### 1. HTTP/HTTPS

Используется для подключения к MCP серверу через HTTP:

```yaml
request:
  protocol: "mcp"
  mcp_transport: "http"  # или "https"
  mcp_url: "http://localhost:8080/mcp"
  mcp_method: "tools/list"
```

### 2. Stdio

Используется для запуска MCP сервера как локальной команды через стандартный ввод/вывод:

```yaml
request:
  protocol: "mcp"
  mcp_transport: "stdio"
  mcp_command: "npx"
  mcp_args:
    - "@modelcontextprotocol/server-filesystem"
    - "/path/to/directory"
  mcp_method: "tools/list"
```

**Особенности stdio транспорта:**
- Сервер запускается как локальный процесс
- Коммуникация происходит через stdin/stdout
- Не требует сетевого подключения
- Подходит для локальных MCP серверов
- Автоматически управляет жизненным циклом процесса

**Примеры использования:**

1. **NPM пакет MCP сервер:**
```yaml
request:
  protocol: "mcp"
  mcp_transport: "stdio"
  mcp_command: "npx"
  mcp_args:
    - "@modelcontextprotocol/server-filesystem"
    - "/tmp"
  mcp_method: "tools/list"
```

2. **Python MCP сервер:**
```yaml
request:
  protocol: "mcp"
  mcp_transport: "stdio"
  mcp_command: "python"
  mcp_args:
    - "-m"
    - "mcp_server"
    - "--config"
    - "config.json"
  mcp_method: "initialize"
```

3. **Go бинарный MCP сервер:**
```yaml
request:
  protocol: "mcp"
  mcp_transport: "stdio"
  mcp_command: "./mcp-server"
  mcp_args:
    - "--port"
    - "8080"
  mcp_method: "tools/list"
```

4. **Любой исполняемый файл:**
```yaml
request:
  protocol: "mcp"
  mcp_transport: "stdio"
  mcp_command: "/usr/local/bin/my-mcp-server"
  mcp_args:
    - "--verbose"
  mcp_method: "initialize"
```

## Конфигурация

### Базовый пример

```yaml
steps:
  - name: "List MCP Tools"
    request:
      protocol: "mcp"
      mcp_transport: "http"
      mcp_url: "http://localhost:8080/mcp"
      mcp_method: "tools/list"
      timeout: "10s"
    validate:
      - json_path: "$.tools"
        type: "array"
    capture:
      tool_count: "$.tools.length()"
```

### Вызов инструмента

```yaml
steps:
  - name: "Call MCP Tool"
    request:
      protocol: "mcp"
      mcp_transport: "http"
      mcp_url: "http://localhost:8080/mcp"
      mcp_method: "tools/call"
      mcp_params:
        name: "weather"
        arguments:
          location: "New York"
          units: "celsius"
      timeout: "10s"
    validate:
      - json_path: "$.content"
        type: "array"
    capture:
      tool_result: "$.content[0].text"
```

### Чтение ресурса

```yaml
steps:
  - name: "Read MCP Resource"
    request:
      protocol: "mcp"
      mcp_transport: "http"
      mcp_url: "http://localhost:8080/mcp"
      mcp_method: "resources/read"
      mcp_params:
        uri: "file:///tmp/data.json"
      timeout: "10s"
    validate:
      - json_path: "$.contents"
        type: "array"
    capture:
      resource_content: "$.contents[0].text"
```

### Получение промпта

```yaml
steps:
  - name: "Get MCP Prompt"
    request:
      protocol: "mcp"
      mcp_transport: "http"
      mcp_url: "http://localhost:8080/mcp"
      mcp_method: "prompts/get"
      mcp_params:
        name: "code_review"
        arguments:
          code: "function test() { return true; }"
          language: "javascript"
      timeout: "10s"
    validate:
      - json_path: "$.messages"
        type: "array"
```

## Поля конфигурации

### Обязательные поля

- `protocol`: Должно быть `"mcp"`
- `mcp_transport`: Тип транспорта (`"http"`, `"https"`, `"stdio"`)
- `mcp_method`: MCP метод для вызова

### Поля для HTTP транспорта

- `mcp_url`: URL MCP сервера

### Поля для stdio транспорта

- `mcp_command`: Команда для запуска MCP сервера
- `mcp_args`: Аргументы команды (опционально)

### Общие поля

- `mcp_params`: Параметры для MCP метода (опционально)
- `mcp_client_info`: Информация о клиенте (опционально)
  - `name`: Имя клиента
  - `version`: Версия клиента
  - `capabilities`: Возможности клиента
- `timeout`: Таймаут запроса

## Переменные

Все поля поддерживают подстановку переменных:

```yaml
variables:
  mcp_server_url: "http://localhost:8080/mcp"
  tool_name: "weather"

steps:
  - name: "Call Tool"
    request:
      protocol: "mcp"
      mcp_transport: "http"
      mcp_url: "{{mcp_server_url}}"
      mcp_method: "tools/call"
      mcp_params:
        name: "{{tool_name}}"
        arguments:
          location: "{{city}}"
```

## Валидация

MCP ответы валидируются так же, как HTTP ответы:

```yaml
validate:
  - json_path: "$.tools"
    type: "array"
  - json_path: "$.tools[0].name"
    equals: "weather"
```

## Capture

Значения из MCP ответов можно захватывать:

```yaml
capture:
  tool_result: "$.content[0].text"
  resource_data: "$.contents[0].text"
```

## Обработка ошибок

MCP ошибки обрабатываются автоматически. Если MCP сервер возвращает ошибку, она будет залогирована и шаг будет помечен как failed.

## Тестирование

Для тестирования MCP протокола в Stepwise включен простой тестовый MCP сервер.

### Запуск тестового сервера

```bash
cd examples/mcp-test-server
go run main.go
```

Сервер будет доступен по адресу: `http://localhost:8080/mcp`

### Тестовый workflow

Используйте готовый тестовый workflow:

```bash
# В первом терминале запустите тестовый сервер
cd examples/mcp-test-server
go run main.go

# Во втором терминале запустите workflow
cd ../..
stepwise run examples/mcp-test-workflow.yml
```

Тестовый сервер поддерживает все основные MCP методы:
- `initialize` - инициализация
- `tools/list` - список инструментов (weather, calculator)
- `tools/call` - вызов инструментов
- `resources/list` - список ресурсов
- `resources/read` - чтение ресурсов
- `prompts/list` - список промптов
- `prompts/get` - получение промптов

Подробнее о тестовом сервере см. `examples/mcp-test-server/README.md`.

## Примеры

Полные примеры использования MCP можно найти в:
- `examples/mcp-demo.yml` - демонстрация всех возможностей
- `examples/mcp-test-workflow.yml` - тестовый workflow для локального сервера

## Примечания

- MCP клиент автоматически инициализируется при первом использовании
- Клиент переиспользуется для одного и того же сервера
- При изменении сервера старый клиент закрывается и создается новый
- Поддержка WebSocket транспорта планируется в будущих версиях

