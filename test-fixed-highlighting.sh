#!/bin/bash

echo "Testing fixed syntax highlighting..."

# Create test directory
rm -rf test-fixed-highlighting
mkdir -p test-fixed-highlighting
cd test-fixed-highlighting

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

echo "Testing with streaming response and syntax highlighting preview:"
echo ""

{
  echo "Создай простой workflow для тестирования API"
  sleep 3
  echo "/exit"
} | ../stepwise codex . 2>&1

echo ""
echo "========================"
echo "Checking for generated files:"
ls -la *.yml 2>/dev/null || echo "No YAML files found"

cd ..
