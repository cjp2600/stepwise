#!/bin/bash

echo "Testing codex file creation approach..."

# Create test directory
rm -rf test-codex-approach
mkdir -p test-codex-approach
cd test-codex-approach

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

echo "Testing new approach - AI should ask codex to create files:"
{
  echo "create a simple workflow for testing user creation"
  echo "/exit"
} | ../stepwise codex . 2>&1 | head -20

cd ..
