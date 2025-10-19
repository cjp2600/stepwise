#!/bin/bash

echo "Testing simple codex interaction..."

# Test with a simple question
echo "Привет! Как дела?" | timeout 30 ./stepwise codex --model gpt-4o . 2>&1 | head -30

echo -e "\nTest completed!"
