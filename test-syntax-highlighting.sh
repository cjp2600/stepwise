#!/bin/bash

echo "Testing syntax highlighting..."

# Create test directory
rm -rf test-syntax-highlighting
mkdir -p test-syntax-highlighting
cd test-syntax-highlighting

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

echo "Testing with default style (github):"
{
  echo "Создай простой workflow для тестирования API"
  sleep 3
  echo "/exit"
} | ../stepwise codex . 2>&1 | head -50

echo ""
echo "========================"
echo "Testing with monokai style:"
{
  echo "Создай простой workflow для тестирования API"
  sleep 3
  echo "/exit"
} | ../stepwise codex --style monokai . 2>&1 | head -50

cd ..
