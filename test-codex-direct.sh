#!/bin/bash

echo "Testing codex directly..."

# Test codex command directly
codex exec --json --skip-git-repo-check "Создай простой workflow для тестирования API" 2>&1 | head -50

