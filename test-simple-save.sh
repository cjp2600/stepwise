#!/bin/bash

echo "Testing simple save functionality..."

# Create a test directory
mkdir -p test-simple-save
cd test-simple-save

# Create a simple workflow file for context
cat > context.yml << 'EOF'
name: "Context"
version: "1.0"
steps:
  - name: "Test"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
EOF

echo "Created context file"

# Test with a timeout to avoid hanging
timeout 30s bash -c 'echo "Создай простой workflow" | ../stepwise codex .' 2>&1 | head -50

echo "Checking for generated files:"
ls -la *.yml 2>/dev/null || echo "No YAML files found"

cd ..
