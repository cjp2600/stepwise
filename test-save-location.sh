#!/bin/bash

echo "Testing file save location..."

# Create a test directory
mkdir -p test-save-location
cd test-save-location

# Create a simple workflow file for context
cat > test-context.yml << 'EOF'
name: "Test Context"
version: "1.0"
steps:
  - name: "Test Step"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
    validate:
      - status: 200
EOF

echo "Created test context file"

# Test the codex command with a simple request
echo "Создай workflow для тестирования API" | ../stepwise codex . 2>&1 | head -30

echo "Checking if files were created in the current directory:"
ls -la *.yml 2>/dev/null || echo "No YAML files found"

cd ..
