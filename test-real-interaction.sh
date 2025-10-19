#!/bin/bash

echo "Testing real interaction with codex..."

# Create a test directory
mkdir -p test-real-interaction
cd test-real-interaction

# Create a simple workflow file for context
cat > context.yml << 'EOF'
name: "Context"
version: "1.0"
steps:
  - name: "Test"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
EOF

echo "Created context file"

# Create a simple input file
cat > input.txt << 'EOF'
Создай простой workflow для тестирования API пользователей
/exit
EOF

echo "Starting codex with input file..."
cat input.txt | ../stepwise codex . 2>&1

echo "Checking for generated files:"
ls -la *.yml 2>/dev/null || echo "No YAML files found"

cd ..
