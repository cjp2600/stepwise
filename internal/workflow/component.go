package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

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
	Imports     []Import               `yaml:"imports,omitempty" json:"imports,omitempty"`
	Captures    map[string]string      `yaml:"captures,omitempty" json:"captures,omitempty"` // Global captures for the component
}

// ComponentManager handles component loading, caching, and dependency resolution
type ComponentManager struct {
	searchPaths  []string
	components   map[string]*Component
	loading      map[string]bool // For cycle detection
	depth        map[string]int  // Track recursion depth for each component
	loadAttempts map[string]int  // Track load attempts to prevent infinite loops
	mu           sync.RWMutex
}

// NewComponentManager creates a new component manager
func NewComponentManager(searchPaths []string) *ComponentManager {
	// Add default search paths
	defaultPaths := []string{
		".",
		"./components",
		"./templates",
		"./workflows",
		"./steps",
	}

	allPaths := append(defaultPaths, searchPaths...)

	return &ComponentManager{
		searchPaths:  allPaths,
		components:   make(map[string]*Component),
		loading:      make(map[string]bool),
		depth:        make(map[string]int),
		loadAttempts: make(map[string]int),
	}
}

// LoadComponent loads a component with cycle detection
func (cm *ComponentManager) LoadComponent(path string) (*Component, error) {
	return cm.LoadComponentWithTimeout(path, 10*time.Second)
}

// LoadComponentWithTimeout loads a component with cycle detection and timeout
func (cm *ComponentManager) LoadComponentWithTimeout(path string, timeout time.Duration) (*Component, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a channel for the result
	resultChan := make(chan *Component, 1)
	errChan := make(chan error, 1)

	// Run the loading in a goroutine
	go func() {
		component, err := cm.loadComponentInternal(path)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- component
	}()

	// Wait for result or timeout
	select {
	case component := <-resultChan:
		return component, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout loading component '%s' after %v", path, timeout)
	}
}

// loadComponentInternal is the internal implementation of component loading
func (cm *ComponentManager) loadComponentInternal(path string) (*Component, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Normalize the path for cycle detection
	normalizedPath := path
	if !filepath.IsAbs(path) {
		// Try to find the absolute path
		for _, searchPath := range cm.searchPaths {
			testPath := filepath.Join(searchPath, path)
			if _, err := os.Stat(testPath); err == nil {
				normalizedPath = testPath
				break
			}
			// Try with extensions
			for _, ext := range []string{".yml", ".yaml"} {
				testPath := filepath.Join(searchPath, path+ext)
				if _, err := os.Stat(testPath); err == nil {
					normalizedPath = testPath
					break
				}
			}
		}
	}

	// Check if already loaded
	if component, exists := cm.components[normalizedPath]; exists {
		return component, nil
	}

	// Check for cycles using normalized path
	if cm.loading[normalizedPath] {
		// Log the current loading stack for debugging
		loadingStack := make([]string, 0)
		for loadingPath := range cm.loading {
			loadingStack = append(loadingStack, loadingPath)
		}
		return nil, fmt.Errorf("circular import detected: %s (loading stack: %v)", path, loadingStack)
	}

	// Check for cycles by original path as well
	if cm.loading[path] {
		loadingStack := make([]string, 0)
		for loadingPath := range cm.loading {
			loadingStack = append(loadingStack, loadingPath)
		}
		return nil, fmt.Errorf("circular import detected: %s (loading stack: %v)", path, loadingStack)
	}

	// Check for cycles by path name (for relative paths)
	if cm.loading[path] {
		loadingStack := make([]string, 0)
		for loadingPath := range cm.loading {
			loadingStack = append(loadingStack, loadingPath)
		}
		return nil, fmt.Errorf("circular import detected: %s (loading stack: %v)", path, loadingStack)
	}

	// Additional check: compare with normalized paths
	for loadingPath := range cm.loading {
		if filepath.Base(loadingPath) == filepath.Base(path) {
			loadingStack := make([]string, 0)
			for lp := range cm.loading {
				loadingStack = append(loadingStack, lp)
			}
			return nil, fmt.Errorf("circular import detected: %s matches %s (loading stack: %v)", path, loadingPath, loadingStack)
		}
	}

	// Additional check: compare with normalized paths more aggressively
	for loadingPath := range cm.loading {
		// Check if the paths refer to the same file by comparing their base names
		loadingBase := filepath.Base(loadingPath)
		pathBase := filepath.Base(path)
		if loadingBase == pathBase {
			loadingStack := make([]string, 0)
			for lp := range cm.loading {
				loadingStack = append(loadingStack, lp)
			}
			return nil, fmt.Errorf("circular import detected: %s matches %s (loading stack: %v)", path, loadingPath, loadingStack)
		}
	}

	// Check recursion depth to prevent infinite loops
	const maxDepth = 100
	currentDepth := cm.depth[normalizedPath]
	if currentDepth > maxDepth {
		return nil, fmt.Errorf("maximum recursion depth exceeded for component '%s' (depth: %d)", path, currentDepth)
	}

	// Check load attempts to prevent infinite loops
	const maxAttempts = 10
	attempts := cm.loadAttempts[normalizedPath]
	if attempts > maxAttempts {
		return nil, fmt.Errorf("maximum load attempts exceeded for component '%s' (attempts: %d)", path, attempts)
	}
	cm.loadAttempts[normalizedPath] = attempts + 1

	cm.loading[normalizedPath] = true
	cm.loading[path] = true
	cm.depth[normalizedPath] = currentDepth + 1
	defer func() {
		delete(cm.loading, normalizedPath)
		delete(cm.loading, path)
		cm.depth[normalizedPath] = currentDepth // Restore previous depth
	}()

	// Find and load the component
	fullPath := cm.findComponent(path)
	if fullPath == "" {
		return nil, fmt.Errorf("component not found: %s", path)
	}

	component, err := cm.loadComponentFromFile(fullPath)
	if err != nil {
		return nil, err
	}

	// Resolve imports recursively
	if err := cm.resolveComponentImports(component); err != nil {
		return nil, fmt.Errorf("failed to resolve imports for %s: %w", path, err)
	}

	// Cache the component
	cm.components[path] = component

	return component, nil
}

// findComponent searches for a component in the search paths
func (cm *ComponentManager) findComponent(path string) string {
	// Try direct path first
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err == nil {
			return path
		}
		// Try with extensions
		for _, ext := range []string{"", ".yml", ".yaml"} {
			testPath := path + ext
			if _, err := os.Stat(testPath); err == nil {
				return testPath
			}
		}
	}

	// Try relative paths
	for _, searchPath := range cm.searchPaths {
		// Try without extension
		fullPath := filepath.Join(searchPath, path)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}

		// Try with extensions
		for _, ext := range []string{".yml", ".yaml"} {
			testPath := fullPath + ext
			if _, err := os.Stat(testPath); err == nil {
				return testPath
			}
		}
	}

	return ""
}

// loadComponentFromFile loads a component from a file
func (cm *ComponentManager) loadComponentFromFile(path string) (*Component, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read component file: %w", err)
	}

	var component Component
	if err := yaml.Unmarshal(data, &component); err != nil {
		return nil, fmt.Errorf("failed to parse component file: %w", err)
	}

	// Validate component
	if err := cm.validateComponent(&component); err != nil {
		return nil, fmt.Errorf("invalid component: %w", err)
	}

	return &component, nil
}

// validateComponent validates a component structure
func (cm *ComponentManager) validateComponent(component *Component) error {
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

// resolveComponentImports resolves all imports in a component
func (cm *ComponentManager) resolveComponentImports(component *Component) error {
	if component.Imports == nil {
		return nil
	}

	for _, imp := range component.Imports {
		if err := cm.resolveComponentImport(component, &imp); err != nil {
			return fmt.Errorf("failed to resolve import %s: %w", imp.Path, err)
		}
	}

	return nil
}

// resolveComponentImport resolves a single import in a component
func (cm *ComponentManager) resolveComponentImport(component *Component, imp *Import) error {
	// Check if this import would create a circular dependency
	// by checking if the imported component would try to import the current component
	if importedComponent, exists := cm.components[imp.Path]; exists {
		// Component already loaded, check if it imports the current component
		for _, importItem := range importedComponent.Imports {
			if importItem.Path == component.Name || importItem.Path == "./"+component.Name {
				return fmt.Errorf("circular dependency detected: component %s imports %s which imports %s",
					component.Name, imp.Path, component.Name)
			}
		}
	}

	// Additional check: if we're currently loading the target component, it's a cycle
	if cm.loading[imp.Path] {
		return fmt.Errorf("circular dependency detected: component %s is trying to import %s which is currently being loaded",
			component.Name, imp.Path)
	}

	// Check for circular dependencies by comparing normalized paths
	// This handles cases where relative paths might not match exactly
	for loadingPath := range cm.loading {
		// Check if the loading path and import path refer to the same file
		if loadingPath == imp.Path ||
			filepath.Base(loadingPath) == filepath.Base(imp.Path) ||
			strings.HasSuffix(loadingPath, imp.Path) ||
			strings.HasSuffix(imp.Path, loadingPath) {
			return fmt.Errorf("circular dependency detected: component %s is trying to import %s while %s is being loaded",
				component.Name, imp.Path, loadingPath)
		}
	}

	// Load the imported component with timeout
	importedComponent, err := cm.LoadComponentWithTimeout(imp.Path, 10*time.Second)
	if err != nil {
		return err
	}

	// Apply variable overrides
	importedComponent = cm.applyImportOverrides(importedComponent, imp)

	// Merge component based on type
	switch importedComponent.Type {
	case "step":
		return cm.mergeStepIntoComponent(component, importedComponent, imp)
	case "group":
		return cm.mergeGroupIntoComponent(component, importedComponent, imp)
	case "workflow":
		return cm.mergeWorkflowIntoComponent(component, importedComponent, imp)
	default:
		return fmt.Errorf("unknown component type: %s", importedComponent.Type)
	}
}

// applyImportOverrides applies variable overrides to an imported component
func (cm *ComponentManager) applyImportOverrides(component *Component, imp *Import) *Component {
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
		newComponent.Steps = cm.applyStepOverrides(component.Steps, imp.Overrides)
	}

	return &newComponent
}

// applyStepOverrides applies overrides to steps
func (cm *ComponentManager) applyStepOverrides(steps []Step, overrides map[string]interface{}) []Step {
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
				newStep.Request = cm.applyRequestOverrides(step.Request, requestMap)
			}
		}

		newSteps[i] = newStep
	}
	return newSteps
}

// applyRequestOverrides applies overrides to a request
func (cm *ComponentManager) applyRequestOverrides(request Request, overrides map[string]interface{}) Request {
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

// mergeStepIntoComponent merges a step component into another component
func (cm *ComponentManager) mergeStepIntoComponent(target *Component, source *Component, imp *Import) error {
	if len(source.Steps) != 1 {
		return fmt.Errorf("step component must contain exactly one step")
	}

	step := source.Steps[0]

	// Apply alias if specified
	if imp.Alias != "" {
		step.Name = imp.Alias
	}

	// Add step to target component
	target.Steps = append(target.Steps, step)

	// Merge variables
	if source.Variables != nil {
		if target.Variables == nil {
			target.Variables = make(map[string]interface{})
		}
		for key, value := range source.Variables {
			target.Variables[key] = value
		}
	}

	// Merge captures
	if source.Captures != nil {
		if target.Captures == nil {
			target.Captures = make(map[string]string)
		}
		for key, value := range source.Captures {
			target.Captures[key] = value
		}
	}

	return nil
}

// mergeGroupIntoComponent merges a group component into another component
func (cm *ComponentManager) mergeGroupIntoComponent(target *Component, source *Component, imp *Import) error {
	// Create a new group
	group := StepGroup{
		Name:        source.Name,
		Description: source.Description,
		Steps:       source.Steps,
		Groups:      source.Groups,
	}

	// Apply alias if specified
	if imp.Alias != "" {
		group.Name = imp.Alias
	}

	// Add group to target component
	target.Groups = append(target.Groups, group)

	// Merge variables
	if source.Variables != nil {
		if target.Variables == nil {
			target.Variables = make(map[string]interface{})
		}
		for key, value := range source.Variables {
			target.Variables[key] = value
		}
	}

	// Merge captures
	if source.Captures != nil {
		if target.Captures == nil {
			target.Captures = make(map[string]string)
		}
		for key, value := range source.Captures {
			target.Captures[key] = value
		}
	}

	return nil
}

// mergeWorkflowIntoComponent merges a workflow component into another component
func (cm *ComponentManager) mergeWorkflowIntoComponent(target *Component, source *Component, imp *Import) error {
	// Merge steps
	target.Steps = append(target.Steps, source.Steps...)

	// Merge groups
	target.Groups = append(target.Groups, source.Groups...)

	// Merge variables
	if source.Variables != nil {
		if target.Variables == nil {
			target.Variables = make(map[string]interface{})
		}
		for key, value := range source.Variables {
			target.Variables[key] = value
		}
	}

	// Merge captures
	if source.Captures != nil {
		if target.Captures == nil {
			target.Captures = make(map[string]string)
		}
		for key, value := range source.Captures {
			target.Captures[key] = value
		}
	}

	return nil
}

// GetComponent retrieves a cached component
func (cm *ComponentManager) GetComponent(path string) (*Component, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	component, exists := cm.components[path]
	return component, exists
}

// ClearCache clears the component cache
func (cm *ComponentManager) ClearCache() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.components = make(map[string]*Component)
	cm.loading = make(map[string]bool)
}

// resolveWorkflowImports resolves all imports in a workflow
func (cm *ComponentManager) resolveWorkflowImports(wf *Workflow) error {
	if wf.Imports == nil {
		return nil
	}

	for _, imp := range wf.Imports {
		if err := cm.resolveWorkflowImport(wf, &imp); err != nil {
			return fmt.Errorf("failed to resolve import %s: %w", imp.Path, err)
		}
	}

	return nil
}

// resolveWorkflowImport resolves a single import in a workflow
func (cm *ComponentManager) resolveWorkflowImport(wf *Workflow, imp *Import) error {
	// Load the imported component with timeout
	importedComponent, err := cm.LoadComponentWithTimeout(imp.Path, 10*time.Second)
	if err != nil {
		return err
	}

	// Apply variable overrides
	importedComponent = cm.applyImportOverrides(importedComponent, imp)

	// Merge component based on type
	switch importedComponent.Type {
	case "step":
		return cm.mergeStepIntoWorkflow(wf, importedComponent, imp)
	case "group":
		return cm.mergeGroupIntoWorkflow(wf, importedComponent, imp)
	case "workflow":
		return cm.mergeWorkflowIntoWorkflow(wf, importedComponent, imp)
	default:
		return fmt.Errorf("unknown component type: %s", importedComponent.Type)
	}
}

// mergeStepIntoWorkflow merges a step component into a workflow
func (cm *ComponentManager) mergeStepIntoWorkflow(wf *Workflow, component *Component, imp *Import) error {
	// Не добавляем шаги из компонента в wf.Steps!
	// Просто регистрируем переменные и captures, если нужно

	// Merge variables
	if component.Variables != nil {
		if wf.Variables == nil {
			wf.Variables = make(map[string]interface{})
		}
		for key, value := range component.Variables {
			wf.Variables[key] = value
		}
	}

	// Merge captures
	if component.Captures != nil {
		if wf.Captures == nil {
			wf.Captures = make(map[string]string)
		}
		for key, value := range component.Captures {
			wf.Captures[key] = value
		}
	}

	return nil
}

// mergeGroupIntoWorkflow merges a group component into a workflow
func (cm *ComponentManager) mergeGroupIntoWorkflow(wf *Workflow, component *Component, imp *Import) error {
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

	// Merge captures
	if component.Captures != nil {
		if wf.Captures == nil {
			wf.Captures = make(map[string]string)
		}
		for key, value := range component.Captures {
			wf.Captures[key] = value
		}
	}

	return nil
}

// mergeWorkflowIntoWorkflow merges a workflow component into a workflow
func (cm *ComponentManager) mergeWorkflowIntoWorkflow(wf *Workflow, component *Component, imp *Import) error {
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

	// Merge captures
	if component.Captures != nil {
		if wf.Captures == nil {
			wf.Captures = make(map[string]string)
		}
		for key, value := range component.Captures {
			wf.Captures[key] = value
		}
	}

	return nil
}
