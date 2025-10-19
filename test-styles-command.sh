#!/bin/bash

echo "Testing /styles command..."

# Create test directory
rm -rf test-styles-command
mkdir -p test-styles-command
cd test-styles-command

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

# Test the /styles command
{
  echo "/styles"
  echo "/exit"
} | ../stepwise codex . 2>&1

cd ..
