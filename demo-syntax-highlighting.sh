#!/bin/bash

echo "🎨 Stepwise Codex with Syntax Highlighting Demo"
echo "=============================================="
echo ""

# Create demo directory
rm -rf demo-syntax-highlighting
mkdir -p demo-syntax-highlighting
cd demo-syntax-highlighting

# Create context file
cat > context.yml << 'EOF'
name: "Demo Context"
version: "1.0"
steps:
  - name: "Demo Step"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
EOF

echo "📁 Created demo context file"
echo ""

echo "🚀 Testing different syntax highlighting styles:"
echo ""

echo "1️⃣  Default (github) style:"
echo "   Command: stepwise codex ."
echo "   Creating workflow with default highlighting..."
echo ""

{
  echo "Создай простой workflow для тестирования API с подсветкой синтаксиса"
  sleep 2
  echo "/exit"
} | ../stepwise codex . 2>&1 | head -30

echo ""
echo "2️⃣  Monokai style:"
echo "   Command: stepwise codex --style monokai ."
echo "   Creating workflow with monokai highlighting..."
echo ""

{
  echo "Создай простой workflow для тестирования API с подсветкой синтаксиса"
  sleep 2
  echo "/exit"
} | ../stepwise codex --style monokai . 2>&1 | head -30

echo ""
echo "3️⃣  Dracula style:"
echo "   Command: stepwise codex --style dracula ."
echo "   Creating workflow with dracula highlighting..."
echo ""

{
  echo "Создай простой workflow для тестирования API с подсветкой синтаксиса"
  sleep 2
  echo "/exit"
} | ../stepwise codex --style dracula . 2>&1 | head -30

echo ""
echo "✨ Demo completed! Check the generated files:"
ls -la *.yml 2>/dev/null || echo "No YAML files found"

echo ""
echo "🎯 Features demonstrated:"
echo "   ✅ Syntax highlighting for YAML code blocks"
echo "   ✅ Multiple highlighting styles (github, monokai, dracula)"
echo "   ✅ Automatic file saving with unique names"
echo "   ✅ Interactive chat interface"
echo "   ✅ Context-aware AI responses"
echo ""

cd ..
