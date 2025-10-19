#!/bin/bash

echo "Testing syntax highlighting comparison..."

# Create test directory
rm -rf test-syntax-comparison
mkdir -p test-syntax-comparison
cd test-syntax-comparison

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

echo "Testing different highlighting styles:"
echo ""

echo "1. Default (github) style:"
echo "Создай простой workflow" | ../stepwise codex . 2>&1 | grep -A 20 "```yaml" | head -15

echo ""
echo "2. Monokai style:"
echo "Создай простой workflow" | ../stepwise codex --style monokai . 2>&1 | grep -A 20 "```yaml" | head -15

echo ""
echo "3. Dracula style:"
echo "Создай простой workflow" | ../stepwise codex --style dracula . 2>&1 | grep -A 20 "```yaml" | head -15

cd ..
