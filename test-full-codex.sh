#!/bin/bash

echo "Testing full stepwise codex command..."

# Create test directory
rm -rf test-full-codex
mkdir -p test-full-codex
cd test-full-codex

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

# Run codex with input
{
  echo "Создай workflow для тестирования user API"
  sleep 3
  echo "/exit"
} | ../stepwise codex .

echo ""
echo "========================"
echo "Checking for generated files:"
ls -la *.yml

cd ..
