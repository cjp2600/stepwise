#!/bin/bash

echo "🎨 Stepwise Codex - Final Demo with Syntax Highlighting"
echo "======================================================"
echo ""

# Create demo directory
rm -rf final-demo
mkdir -p final-demo
cd final-demo

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

echo "🚀 Testing the complete workflow:"
echo ""

echo "1️⃣  Creating workflow with default highlighting..."
{
  echo "Создай workflow для тестирования API пользователей с полным CRUD"
  sleep 3
  echo "/exit"
} | ../stepwise codex . 2>&1 | head -50

echo ""
echo "2️⃣  Testing /styles command..."
{
  echo "/styles"
  echo "/exit"
} | ../stepwise codex . 2>&1 | head -20

echo ""
echo "3️⃣  Testing with custom style (monokai)..."
{
  echo "Создай простой workflow для тестирования API"
  sleep 2
  echo "/exit"
} | ../stepwise codex --style monokai . 2>&1 | grep -A 15 "Preview of saved workflow"

echo ""
echo "✨ Demo completed! Check the generated files:"
ls -la *.yml 2>/dev/null || echo "No YAML files found"

echo ""
echo "🎯 Features demonstrated:"
echo "   ✅ Interactive chat interface with streaming responses"
echo "   ✅ Syntax highlighting for saved YAML workflows"
echo "   ✅ Multiple highlighting styles (github, monokai, dracula, etc.)"
echo "   ✅ Automatic file saving with unique names"
echo "   ✅ Context-aware AI responses with Stepwise knowledge"
echo "   ✅ Commands: /help, /clear, /save, /styles, /exit"
echo "   ✅ Real-time progress indicators"
echo ""

cd ..
