# MCP Test Server

Простой stub MCP сервер для тестирования поддержки MCP протокола в Stepwise.

## Запуск

```bash
cd examples/mcp-test-server
go run main.go
```

Или с указанием порта:

```bash
PORT=8080 go run main.go
```

Сервер будет доступен по адресу: `http://localhost:8080/mcp`

## Endpoints

- `POST /mcp` - MCP JSON-RPC 2.0 endpoint
- `GET /health` - Health check endpoint

## Поддерживаемые методы

### initialize
Инициализация соединения с MCP сервером.

### tools/list
Возвращает список доступных инструментов:
- `weather` - получение информации о погоде
- `calculator` - выполнение математических вычислений

### tools/call
Вызов инструмента с параметрами.

Пример для `weather`:
```json
{
  "name": "weather",
  "arguments": {
    "location": "New York",
    "units": "celsius"
  }
}
```

### resources/list
Возвращает список доступных ресурсов:
- `file:///tmp/data.json`
- `file:///tmp/config.yaml`

### resources/read
Чтение ресурса по URI.

### prompts/list
Возвращает список доступных промптов:
- `code_review` - ревью кода
- `explain_code` - объяснение кода

### prompts/get
Получение промпта с аргументами.

## Использование с Stepwise

Создайте workflow файл:

```yaml
name: "Test MCP Server"
version: "1.0"

variables:
  mcp_server_url: "http://localhost:8080/mcp"

steps:
  - name: "List Tools"
    request:
      protocol: "mcp"
      mcp_transport: "http"
      mcp_url: "{{mcp_server_url}}"
      mcp_method: "tools/list"
      timeout: "10s"
    validate:
      - json_path: "$.tools"
        type: "array"
    show_response: true

  - name: "Call Weather Tool"
    request:
      protocol: "mcp"
      mcp_transport: "http"
      mcp_url: "{{mcp_server_url}}"
      mcp_method: "tools/call"
      mcp_params:
        name: "weather"
        arguments:
          location: "Moscow"
          units: "celsius"
      timeout: "10s"
    validate:
      - json_path: "$.content"
        type: "array"
    show_response: true
```

Запустите workflow:

```bash
stepwise run workflow.yml
```

## Тестирование вручную

Вы можете протестировать сервер вручную с помощью curl:

```bash
# Initialize
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }'

# List tools
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }'

# Call tool
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "weather",
      "arguments": {
        "location": "Paris",
        "units": "celsius"
      }
    }
  }'
```



