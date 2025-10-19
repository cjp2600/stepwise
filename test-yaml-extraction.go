package main

import (
	"fmt"
	"strings"
)

// extractYAML extracts YAML content from markdown code blocks
func extractYAML(text string) string {
	lines := strings.Split(text, "\n")
	var yamlContent []string
	inYAML := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```yaml") || strings.HasPrefix(trimmed, "```yml") {
			inYAML = true
			continue
		}
		if inYAML && strings.HasPrefix(trimmed, "```") {
			break
		}
		if inYAML {
			yamlContent = append(yamlContent, line)
		}
	}

	return strings.Join(yamlContent, "\n")
}

func main() {
	testText := `Вот workflow для тестирования API:

` + "```yaml" + `
name: "API Test"
version: "1.0"
steps:
  - name: "Test API"
    request:
      method: "GET"
      url: "https://httpbin.org/get"
    validate:
      - status: 200
` + "```" + `

Этот workflow тестирует базовый API.`

	yaml := extractYAML(testText)
	fmt.Printf("Extracted YAML:\n%s\n", yaml)

	if yaml != "" {
		fmt.Println("✓ YAML extraction works!")
	} else {
		fmt.Println("✗ YAML extraction failed!")
	}
}
