package variables

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/utils"
)

// Manager handles variable substitution and management
type Manager struct {
	variables map[string]interface{}
	logger    *logger.Logger
}

// NewManager creates a new variable manager
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		variables: make(map[string]interface{}),
		logger:    log,
	}
}

// Set sets a variable
func (m *Manager) Set(key string, value interface{}) {
	m.variables[key] = value
	m.logger.Debug("Set variable", "key", key, "value", value)
}

// Get gets a variable value
func (m *Manager) Get(key string) (interface{}, bool) {
	value, exists := m.variables[key]
	return value, exists
}

// GetAll gets all variables
func (m *Manager) GetAll() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range m.variables {
		result[key] = value
	}
	return result
}

// Delete removes a variable by key
func (m *Manager) Delete(key string) {
	delete(m.variables, key)
	m.logger.Debug("Delete variable", "key", key)
}

// Substitute substitutes variables in a string
func (m *Manager) Substitute(input string) (string, error) {
	if input == "" {
		return input, nil
	}

	previous := ""
	current := input
	for i := 0; i < 10; i++ { // ограничение на глубину рекурсии
		previous = current

		// Handle utils functions
		current = m.substituteUtilsFunctions(current)
		// Faker
		current = m.substituteFakerFunctions(current)
		// Variables
		current = m.substituteVariables(current)
		// Env
		current = m.substituteEnvironmentVariables(current)

		if current == previous || !strings.Contains(current, "{{") {
			break
		}
	}
	return current, nil
}

// substituteVariables substitutes {{variable}} patterns
func (m *Manager) substituteVariables(input string) string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name
		varName := strings.TrimSpace(strings.Trim(match, "{}"))

		// Skip if it's a faker function or environment variable
		if strings.HasPrefix(varName, "faker.") || strings.HasPrefix(varName, "env.") {
			return match
		}

		// Get variable value
		if value, exists := m.variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}

		// Return original if not found
		m.logger.Warn("Variable not found", "variable", varName)
		return match
	})
}

// substituteFakerFunctions substitutes {{faker.function}} patterns
func (m *Manager) substituteFakerFunctions(input string) string {
	re := regexp.MustCompile(`\{\{faker\.([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract function name and parameters
		funcCall := strings.TrimSpace(strings.TrimPrefix(strings.Trim(match, "{}"), "faker."))

		// Parse function call
		parts := strings.Split(funcCall, "(")
		funcName := parts[0]

		var params []string
		if len(parts) > 1 {
			// Extract parameters
			paramStr := strings.TrimSuffix(parts[1], ")")
			if paramStr != "" {
				params = strings.Split(paramStr, ",")
				for i, param := range params {
					params[i] = strings.TrimSpace(param)
				}
			}
		}

		// Generate fake data based on function name
		result := m.generateFakerData(funcName, params)
		return result
	})
}

// substituteEnvironmentVariables substitutes {{env.VARIABLE}} patterns
func (m *Manager) substituteEnvironmentVariables(input string) string {
	re := regexp.MustCompile(`\{\{env\.([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract environment variable name
		envVar := strings.TrimSpace(strings.TrimPrefix(strings.Trim(match, "{}"), "env."))

		// Get from environment variables
		if value, exists := m.variables[envVar]; exists {
			return fmt.Sprintf("%v", value)
		}

		// Return original if not found
		m.logger.Warn("Environment variable not found", "variable", envVar)
		return match
	})
}

// substituteUtilsFunctions substitutes {{utils.function}} patterns
func (m *Manager) substituteUtilsFunctions(input string) string {
	re := regexp.MustCompile(`\{\{utils\.([a-zA-Z0-9_]+)\((.*?)\)\}\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract function name and argument string
		parts := re.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		funcName := parts[1]
		argStr := strings.TrimSpace(parts[2])

		// Поддержка вложенных шаблонов: если аргумент содержит {{...}}, подставить переменные
		if strings.Contains(argStr, "{{") {
			argStr, _ = m.Substitute(argStr)
		}
		argStr = strings.Trim(argStr, "\"'") // убрать кавычки

		switch funcName {
		case "base64":
			return utils.Base64Encode(argStr)
		case "base64_decode":
			decoded, err := utils.Base64Decode(argStr)
			if err != nil {
				m.logger.Warn("base64_decode error", "arg", argStr, "err", err)
				return ""
			}
			return decoded
		// Здесь можно добавить другие утилиты
		default:
			m.logger.Warn("Unknown utils function", "function", funcName)
			return match
		}
	})
}

// generateFakerData generates fake data based on function name
func (m *Manager) generateFakerData(funcName string, params []string) string {
	switch funcName {
	case "name":
		return m.generateName()
	case "email":
		return m.generateEmail()
	case "phone":
		return m.generatePhone()
	case "address":
		return m.generateAddress()
	case "uuid":
		return m.generateUUID()
	case "number":
		return m.generateNumber(params)
	case "date":
		return m.generateDate()
	case "sentence":
		return m.generateSentence()
	case "paragraph":
		return m.generateParagraph()
	case "sha":
		return m.generateSHA()
	default:
		m.logger.Warn("Unknown faker function", "function", funcName)
		return fmt.Sprintf("{{faker.%s}}", funcName)
	}
}

// Faker data generation methods
func (m *Manager) generateName() string {
	names := []string{
		"John Doe", "Jane Smith", "Bob Johnson", "Alice Brown",
		"Charlie Wilson", "Diana Davis", "Edward Miller", "Fiona Garcia",
		"George Martinez", "Helen Anderson", "Ian Taylor", "Julia Thomas",
		"Kevin Jackson", "Laura White", "Michael Harris", "Nancy Clark",
	}
	return names[time.Now().UnixNano()%int64(len(names))]
}

func (m *Manager) generateEmail() string {
	domains := []string{"example.com", "test.org", "demo.net", "sample.io"}
	name := strings.ToLower(strings.ReplaceAll(m.generateName(), " ", "."))
	domain := domains[time.Now().UnixNano()%int64(len(domains))]
	return fmt.Sprintf("%s@%s", name, domain)
}

func (m *Manager) generatePhone() string {
	// Generate a simple phone number
	return fmt.Sprintf("+1-%03d-%03d-%04d",
		time.Now().UnixNano()%900+100,
		time.Now().UnixNano()%900+100,
		time.Now().UnixNano()%9000+1000)
}

func (m *Manager) generateAddress() string {
	streets := []string{"Main St", "Oak Ave", "Pine Rd", "Elm Blvd"}
	cities := []string{"New York", "Los Angeles", "Chicago", "Houston"}
	street := streets[time.Now().UnixNano()%int64(len(streets))]
	number := time.Now().UnixNano()%9999 + 1
	city := cities[time.Now().UnixNano()%int64(len(cities))]
	return fmt.Sprintf("%d %s, %s", number, street, city)
}

func (m *Manager) generateUUID() string {
	// Simple UUID generation
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		timestamp&0xffffffff,
		(timestamp>>32)&0xffff,
		(timestamp>>48)&0xffff,
		(timestamp>>64)&0xffff,
		timestamp&0xffffffffffff)
}

func (m *Manager) generateNumber(params []string) string {
	if len(params) >= 2 {
		min, _ := strconv.Atoi(params[0])
		max, _ := strconv.Atoi(params[1])
		if max > min {
			result := min + int(time.Now().UnixNano()%int64(max-min+1))
			return strconv.Itoa(result)
		}
	}
	// Default to 1-100
	return strconv.Itoa(int(time.Now().UnixNano()%100) + 1)
}

func (m *Manager) generateDate() string {
	// Generate a random date within the last year
	daysAgo := time.Now().UnixNano() % 365
	date := time.Now().AddDate(0, 0, -int(daysAgo))
	return date.Format("2006-01-02")
}

func (m *Manager) generateSentence() string {
	sentences := []string{
		"This is a test sentence.",
		"Another example sentence for testing.",
		"A third sentence to demonstrate functionality.",
		"Testing the sentence generation feature.",
		"Sample sentence for API testing.",
	}
	return sentences[time.Now().UnixNano()%int64(len(sentences))]
}

func (m *Manager) generateParagraph() string {
	paragraphs := []string{
		"This is a test paragraph that contains multiple sentences. It demonstrates the paragraph generation functionality. The content is suitable for testing purposes.",
		"Another example paragraph with different content. This paragraph has multiple sentences to test the generation feature. It provides realistic test data.",
		"A third paragraph example for demonstration. This content shows how the paragraph generator works. It includes various sentence structures.",
	}
	return paragraphs[time.Now().UnixNano()%int64(len(paragraphs))]
}

func (m *Manager) generateSHA() string {
	// Simple SHA-like string generation
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%040x", timestamp)
}

// SubstituteMap substitutes variables in a map
func (m *Manager) SubstituteMap(input map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range input {
		result[key] = value
	}

	for i := 0; i < 10; i++ {
		changed := false
		for key, value := range result {
			switch v := value.(type) {
			case string:
				substituted, err := m.Substitute(v)
				if err != nil {
					return nil, err
				}
				if substituted != v {
					changed = true
				}
				result[key] = substituted
			case map[string]interface{}:
				substituted, err := m.SubstituteMap(v)
				if err != nil {
					return nil, err
				}
				result[key] = substituted
			case []interface{}:
				substituted, err := m.SubstituteSlice(v)
				if err != nil {
					return nil, err
				}
				result[key] = substituted
			}
		}
		if !changed {
			break
		}
	}
	return result, nil
}

// SubstituteSlice substitutes variables in a slice
func (m *Manager) SubstituteSlice(input []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(input))

	for i, value := range input {
		switch v := value.(type) {
		case string:
			substituted, err := m.Substitute(v)
			if err != nil {
				return nil, err
			}
			result[i] = substituted
		case map[string]interface{}:
			substituted, err := m.SubstituteMap(v)
			if err != nil {
				return nil, err
			}
			result[i] = substituted
		case []interface{}:
			substituted, err := m.SubstituteSlice(v)
			if err != nil {
				return nil, err
			}
			result[i] = substituted
		default:
			result[i] = value
		}
	}

	return result, nil
}
