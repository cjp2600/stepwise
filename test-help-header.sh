#!/bin/bash

echo "Testing help header..."

# Create test directory
rm -rf test-help-header
mkdir -p test-help-header
cd test-help-header

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

echo "Testing help header:"
{
  echo "/help"
  echo "/exit"
} | ../stepwise codex . 2>&1 | head -10

cd ..
