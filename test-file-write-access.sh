#!/bin/bash

echo "Testing file write access with codex..."

# Create test directory
rm -rf test-write-access
mkdir -p test-write-access
cd test-write-access

# Create a simple workflow file
cat > test-workflow.yml << 'EOF'
name: "Test Workflow"
version: "1.0"
steps:
  - name: "Simple Test"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
    validate:
      - status: 200
EOF

echo "Testing file modification through codex:"
{
  echo "update the test-workflow.yml file to add more validation checks"
  echo "/exit"
} | ../stepwise codex . 2>&1 | head -30

echo ""
echo "Checking if file was modified:"
ls -la test-workflow.yml
echo ""
echo "File content:"
cat test-workflow.yml

cd ..
