#!/bin/bash

echo "Testing fixed headers..."

# Create test directory
rm -rf test-headers
mkdir -p test-headers
cd test-headers

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

echo "Testing main header:"
{
  echo "/help"
  echo "/exit"
} | ../stepwise codex . 2>&1 | head -20

cd ..
