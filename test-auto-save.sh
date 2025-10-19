#!/bin/bash

echo "Testing auto-save functionality..."

# Test with a simple workflow creation request
echo "Создай простой workflow для тестирования API пользователей" | ./stepwise codex . 2>&1 | head -50

echo -e "\n\nChecking if files were created..."
ls -la *.yml | head -5

echo -e "\nTest completed!"
