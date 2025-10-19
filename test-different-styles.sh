#!/bin/bash

echo "Testing different syntax highlighting styles..."

# Create test directory
rm -rf test-different-styles
mkdir -p test-different-styles
cd test-different-styles

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

echo "1. Testing with monokai style:"
{
  echo "Создай простой workflow"
  sleep 2
  echo "/exit"
} | ../stepwise codex --style monokai . 2>&1 | grep -A 10 "Preview of saved workflow"

echo ""
echo "2. Testing with dracula style:"
{
  echo "Создай простой workflow"
  sleep 2
  echo "/exit"
} | ../stepwise codex --style dracula . 2>&1 | grep -A 10 "Preview of saved workflow"

echo ""
echo "3. Testing with solarized-dark style:"
{
  echo "Создай простой workflow"
  sleep 2
  echo "/exit"
} | ../stepwise codex --style solarized-dark . 2>&1 | grep -A 10 "Preview of saved workflow"

cd ..
