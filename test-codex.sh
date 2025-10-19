#!/bin/bash

echo "Testing Stepwise Codex integration..."

# Test 1: Simple question
echo "Test 1: Simple question"
echo "Привет! Как дела?" | ./stepwise codex --model gpt-4o . 2>&1 | head -20

echo -e "\n\nTest completed!"
