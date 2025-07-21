package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Import represents an imported component
type Import struct {
	Path      string                 `yaml:"path" json:"path"`
	Version   string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Alias     string                 `yaml:"alias,omitempty" json:"alias,omitempty"`
	Variables map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
	Overrides map[string]interface{} `yaml:"overrides,omitempty" json:"overrides,omitempty"`
}

// Component represents a reusable workflow component
type Component struct {
	Name        string                 `yaml:"name" json:"name"`
	Version     string                 `yaml:"version" json:"version"`
	Description string                 `yaml:"description" json:"description"`
	Type        string                 `yaml:"type" json:"type"` // "step", "group", "workflow"
	Variables   map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
	Steps       []Step                 `yaml:"steps,omitempty" json:"steps,omitempty"`
	Groups      []StepGroup            `yaml:"groups,omitempty" json:"groups,omitempty"`
	Exports     []string               `yaml:"exports,omitempty" json:"exports,omitempty"` // Names of exported steps/groups
}

// ImportManager handles component imports and resolution
type ImportManager struct {
	searchPaths []string
	components  map[string]*Component
}

// NewImportManager creates a new import manager
func NewImportManager(searchPaths []string) *ImportManager {
	// Add current directory and common component paths
	paths := []string{".", "./components", "./templates"}
	paths = append(paths, searchPaths...)

	return &ImportManager{
		searchPaths: paths,
		components:  make(map[string]*Component),
	}
}

// AddSearchPath adds a path to search for components
func (im *ImportManager) AddSearchPath(path string) {
	im.searchPaths = append(im.searchPaths, path)
}

// LoadComponent loads a component from file
func (im *ImportManager) LoadComponent(path string) (*Component, error) {
	// Check if component is already loaded
	if component, exists := im.components[path]; exists {
		return component, nil
	}

	// Try to find the component in search paths
	fullPath := im.findComponent(path)
	if fullPath == "" {
		return nil, fmt.Errorf("component not found: %s", path)
	}

	// Read and parse the component file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read component file: %w", err)
	}

	var component Component
	if err := yaml.Unmarshal(data, &component); err != nil {
		return nil, fmt.Errorf("failed to parse component file: %w", err)
	}

	// Validate component
	if err := im.validateComponent(&component); err != nil {
		return nil, fmt.Errorf("invalid component: %w", err)
	}

	// Cache the component
	im.components[path] = &component

	return &component, nil
}

// findComponent searches for a component in the search paths
func (im *ImportManager) findComponent(path string) string {
	// Try direct path first
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Try with .yml extension
	if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
		for _, searchPath := range im.searchPaths {
			ymlPath := filepath.Join(searchPath, path+".yml")
			if _, err := os.Stat(ymlPath); err == nil {
				return ymlPath
			}
			yamlPath := filepath.Join(searchPath, path+".yaml")
			if _, err := os.Stat(yamlPath); err == nil {
				return yamlPath
			}
		}
	}

	// Try relative to search paths
	for _, searchPath := range im.searchPaths {
		fullPath := filepath.Join(searchPath, path)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return ""
}

// validateComponent validates a component structure
func (im *ImportManager) validateComponent(component *Component) error {
	if component.Name == "" {
		return fmt.Errorf("component name is required")
	}

	if component.Type == "" {
		return fmt.Errorf("component type is required")
	}

	switch component.Type {
	case "step":
		if len(component.Steps) != 1 {
			return fmt.Errorf("step component must contain exactly one step")
		}
	case "group":
		if len(component.Steps) == 0 && len(component.Groups) == 0 {
			return fmt.Errorf("group component must contain at least one step or group")
		}
	case "workflow":
		if len(component.Steps) == 0 && len(component.Groups) == 0 {
			return fmt.Errorf("workflow component must contain at least one step or group")
		}
	default:
		return fmt.Errorf("invalid component type: %s", component.Type)
	}

	return nil
}

// ResolveImports resolves all imports in a workflow
func (im *ImportManager) ResolveImports(wf *Workflow) error {
	if wf.Imports == nil {
		return nil
	}

	for _, imp := range wf.Imports {
		if err := im.resolveImport(wf, &imp); err != nil {
			return fmt.Errorf("failed to resolve import %s: %w", imp.Path, err)
		}
	}

	return nil
}

// resolveImport resolves a single import
func (im *ImportManager) resolveImport(wf *Workflow, imp *Import) error {
	// Load the component
	component, err := im.LoadComponent(imp.Path)
	if err != nil {
		return err
	}

	// Apply variable overrides
	component = im.applyOverrides(component, imp)

	// Merge component into workflow based on type
	switch component.Type {
	case "step":
		return im.mergeStepComponent(wf, component, imp)
	case "group":
		return im.mergeGroupComponent(wf, component, imp)
	case "workflow":
		return im.mergeWorkflowComponent(wf, component, imp)
	default:
		return fmt.Errorf("unknown component type: %s", component.Type)
	}
}

// applyOverrides applies variable overrides to a component
func (im *ImportManager) applyOverrides(component *Component, imp *Import) *Component {
	if imp.Variables == nil && imp.Overrides == nil {
		return component
	}

	// Create a copy of the component
	newComponent := *component

	// Apply variable overrides
	if imp.Variables != nil {
		if newComponent.Variables == nil {
			newComponent.Variables = make(map[string]interface{})
		}
		for key, value := range imp.Variables {
			newComponent.Variables[key] = value
		}
	}

	// Apply step overrides
	if imp.Overrides != nil {
		newComponent.Steps = im.applyStepOverrides(component.Steps, imp.Overrides)
	}

	return &newComponent
}

// applyStepOverrides applies overrides to steps
func (im *ImportManager) applyStepOverrides(steps []Step, overrides map[string]interface{}) []Step {
	newSteps := make([]Step, len(steps))
	for i, step := range steps {
		newStep := step

		// Apply name override
		if overrideName, exists := overrides["name"]; exists {
			if name, ok := overrideName.(string); ok {
				newStep.Name = name
			}
		}

		// Apply request overrides
		if requestOverrides, exists := overrides["request"]; exists {
			if requestMap, ok := requestOverrides.(map[string]interface{}); ok {
				newStep.Request = im.applyRequestOverrides(step.Request, requestMap)
			}
		}

		newSteps[i] = newStep
	}
	return newSteps
}

// applyRequestOverrides applies overrides to a request
func (im *ImportManager) applyRequestOverrides(request Request, overrides map[string]interface{}) Request {
	newRequest := request

	for key, value := range overrides {
		switch key {
		case "url":
			if url, ok := value.(string); ok {
				newRequest.URL = url
			}
		case "method":
			if method, ok := value.(string); ok {
				newRequest.Method = method
			}
		case "headers":
			if headers, ok := value.(map[string]interface{}); ok {
				if newRequest.Headers == nil {
					newRequest.Headers = make(map[string]string)
				}
				for hKey, hValue := range headers {
					if hStr, ok := hValue.(string); ok {
						newRequest.Headers[hKey] = hStr
					}
				}
			}
		case "body":
			newRequest.Body = value
		}
	}

	return newRequest
}

// mergeStepComponent merges a step component into the workflow
func (im *ImportManager) mergeStepComponent(wf *Workflow, component *Component, imp *Import) error {
	if len(component.Steps) != 1 {
		return fmt.Errorf("step component must contain exactly one step")
	}

	step := component.Steps[0]

	// Apply alias if specified
	if imp.Alias != "" {
		step.Name = imp.Alias
	}

	// Add step to workflow
	wf.Steps = append(wf.Steps, step)

	// Merge variables
	if component.Variables != nil {
		if wf.Variables == nil {
			wf.Variables = make(map[string]interface{})
		}
		for key, value := range component.Variables {
			wf.Variables[key] = value
		}
	}

	return nil
}

// mergeGroupComponent merges a group component into the workflow
func (im *ImportManager) mergeGroupComponent(wf *Workflow, component *Component, imp *Import) error {
	// Create a new group
	group := StepGroup{
		Name:        component.Name,
		Description: component.Description,
		Steps:       component.Steps,
		Groups:      component.Groups,
	}

	// Apply alias if specified
	if imp.Alias != "" {
		group.Name = imp.Alias
	}

	// Add group to workflow
	wf.Groups = append(wf.Groups, group)

	// Merge variables
	if component.Variables != nil {
		if wf.Variables == nil {
			wf.Variables = make(map[string]interface{})
		}
		for key, value := range component.Variables {
			wf.Variables[key] = value
		}
	}

	return nil
}

// mergeWorkflowComponent merges a workflow component into the workflow
func (im *ImportManager) mergeWorkflowComponent(wf *Workflow, component *Component, imp *Import) error {
	// Merge steps
	wf.Steps = append(wf.Steps, component.Steps...)

	// Merge groups
	wf.Groups = append(wf.Groups, component.Groups...)

	// Merge variables
	if component.Variables != nil {
		if wf.Variables == nil {
			wf.Variables = make(map[string]interface{})
		}
		for key, value := range component.Variables {
			wf.Variables[key] = value
		}
	}

	return nil
}
