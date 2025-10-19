#!/bin/bash

echo "Testing updated prompt with production patterns..."

# Create test directory
rm -rf test-updated-prompt
mkdir -p test-updated-prompt
cd test-updated-prompt

# Create context file
cat > context.yml << 'EOF'
name: "Context"
version: "1.0"
steps:
  - name: "Test"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
EOF

echo "Testing with production patterns:"
echo ""

{
  echo "Создай workflow для тестирования API с компонентами и array filters"
  sleep 3
  echo "/exit"
} | ../stepwise codex . 2>&1

echo ""
echo "========================"
echo "Checking for generated files:"
ls -la *.yml 2>/dev/null || echo "No YAML files found"

cd ..
