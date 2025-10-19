#!/bin/bash

echo "Testing manual save functionality..."

# Create a test directory
mkdir -p test-manual-save
cd test-manual-save

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
echo "Starting codex in background..."

# Start codex in background and send input
{
  echo "Создай простой workflow для тестирования API"
  sleep 2
  echo "/exit"
} | ../stepwise codex . &

CODEX_PID=$!

# Wait a bit for processing
sleep 10

# Kill the process if it's still running
kill $CODEX_PID 2>/dev/null

echo "Checking for generated files:"
ls -la *.yml 2>/dev/null || echo "No YAML files found"

cd ..
